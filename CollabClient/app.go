package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	isHost     bool      // 标记是否为房主
	forceClose bool      // 标记是否确认强制退出
	loggedIn   bool      // 标记是否已登录（未登录时允许直接关闭窗口）
	serverCmd  *exec.Cmd // 后端子进程句柄
}

type LANServer struct {
	Name        string `json:"name"`
	IP          string `json:"ip"`
	Tag         string `json:"tag"`
	Recommended bool   `json:"recommended"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// =============================================================================
// startup Wails OnStartup 生命周期
// =============================================================================
// 在此阶段自动检测并拉起后端服务
// =============================================================================
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.isHost = a.startBackendServer()
}

// =============================================================================
// shutdown Wails OnShutdown 生命周期
// =============================================================================
// 确保在客户端关闭时同步杀死后端子进程，防止后端残留
// =============================================================================
func (a *App) shutdown(ctx context.Context) {
	a.killBackendServer()
}

// IsHostUser 供前端查询是否为房主
func (a *App) IsHostUser() bool {
	return a.isHost
}

// CheckServerHealth 供前端调用的健康检查（绕过 WebView CORS 限制）
// 通过 Go HTTP 客户端直接请求后端 /ping，不受浏览器同源策略影响
func (a *App) CheckServerHealth() bool {
	return a.checkPortAlive()
}

// ConfirmExit 前端确认退出后调用
func (a *App) ConfirmExit() {
	a.forceClose = true
	wailsRuntime.Quit(a.ctx)
}

// SetLoggedIn 前端登录成功后调用，标记已登录状态
// 只有已登录状态才会触发关闭拦截（防止登录页无法关闭窗口）
func (a *App) SetLoggedIn(status bool) {
	a.loggedIn = status
}

// ScanLANServers scans the local network for running CollabServer instances.
func (a *App) ScanLANServers() []LANServer {
	const discoveryPort = 9999
	const defaultHTTPPort = "8080"

	found := make(map[string]LANServer)

	if a.checkPortAlive() {
		found["localhost:"+defaultHTTPPort] = LANServer{
			Name: "本机服务器",
			IP:   "localhost:" + defaultHTTPPort,
			Tag:  "本机",
		}
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		log.Printf("⚠️ 局域网扫描启动失败: %v", err)
		return prioritizeLANServers(found)
	}
	defer conn.Close()

	deadline := time.Now().Add(1500 * time.Millisecond)
	if err := conn.SetDeadline(deadline); err != nil {
		log.Printf("⚠️ 设置局域网扫描超时失败: %v", err)
	}

	message := []byte("WHOIS_COLLAB_HOST")
	for _, target := range discoveryTargets(discoveryPort) {
		if _, err := conn.WriteToUDP(message, target); err != nil {
			log.Printf("⚠️ 发送局域网扫描包失败 target=%s err=%v", target.String(), err)
		}
	}

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			log.Printf("⚠️ 读取局域网扫描响应失败: %v", err)
			break
		}

		payload := strings.TrimSpace(string(buf[:n]))
		parts := strings.Split(payload, "|")
		if len(parts) < 2 || parts[0] != "IAM_HOST" {
			continue
		}

		name := strings.TrimSpace(parts[1])
		if name == "" {
			name = "CollabStudio 主机"
		}

		port := defaultHTTPPort
		if len(parts) >= 3 && strings.TrimSpace(parts[2]) != "" {
			port = strings.TrimSpace(parts[2])
		}

		host := remoteAddr.IP.String()
		address := net.JoinHostPort(host, port)
		if strings.Contains(host, ".") {
			address = host + ":" + port
		}

		found[address] = LANServer{Name: name, IP: address}
	}

	return prioritizeLANServers(found)
}

// =============================================================================
// startBackendServer 自动启动后端服务
// =============================================================================
// 逻辑流程：
// 1. 先检查端口是否有已运行的服务 → 如果有，直接当访客
// 2. 查找同路径下的 CollabServer.exe → 启动并设置隐藏窗口
// 3. 等待健康检查通过 → 成为房主
// =============================================================================
func (a *App) startBackendServer() bool {
	// 1. 抢答环节：先检查端口是否已经被占用了
	if a.checkPortAlive() {
		log.Println("🔍 检测到已有服务器运行中，自动以 [访客] 身份启动。")
		return false
	}

	// --- 以下是没人占端口，自己尝试当房主的逻辑 ---

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("⚠️ 无法获取当前可执行文件路径: %v", err)
		return false
	}
	exeDir := filepath.Dir(exePath)

	serverExeName := "CollabServer"
	if runtime.GOOS == "windows" {
		serverExeName += ".exe"
	}
	serverPath := filepath.Join(exeDir, serverExeName)

	// 检查文件是否存在
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		log.Printf("⚠️ 未找到本地服务器文件: %s，放弃启动。", serverPath)
		return false
	}

	cmd := exec.Command(serverPath)
	cmd.Dir = exeDir

	// 📝 捕获后端 stdout/stderr 到日志文件，方便诊断静默运行时的崩溃
	stderrLogPath := filepath.Join(exeDir, "server_stderr.log")
	stderrFile, fileErr := os.OpenFile(stderrLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if fileErr == nil {
		cmd.Stdout = stderrFile
		cmd.Stderr = stderrFile
		log.Printf("📝 后端输出将写入: %s", stderrLogPath)
	}

	// 🔐 Windows 平台：使用 CREATE_NO_WINDOW 标志隐藏后端 CMD 窗口
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	if err := cmd.Start(); err != nil {
		log.Printf("❌ 后端启动失败: %v，转为 [访客]", err)
		return false
	}

	a.serverCmd = cmd

	// 2. 验证环节：等待后端就绪
	if a.waitForHealthCheck() {
		log.Printf("✅ 后端服务启动成功 (PID: %d)，我是 [👑 房主]", cmd.Process.Pid)
		return true
	}

	log.Println("⚠️ 后端启动超时或失败，转为 [访客]")
	return false
}

// =============================================================================
// checkPortAlive 快速检查端口是否存活（只查一次）
// =============================================================================
func (a *App) checkPortAlive() bool {
	client := http.Client{Timeout: 500 * time.Millisecond}

	// 优先检查 80 端口 (生产环境)
	resp, err := client.Get("http://localhost/ping")
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		return true
	}

	// 备选检查：如果 80 不通，再检查一下 8080 (兼容旧版或开发模式)
	resp2, err2 := client.Get("http://localhost:8080/ping")
	if err2 == nil && resp2.StatusCode == 200 {
		resp2.Body.Close()
		return true
	}

	return false
}

// =============================================================================
// waitForHealthCheck 循环检查健康状态（用于等待后端启动）
// =============================================================================
func (a *App) waitForHealthCheck() bool {
	// 尝试 20 次，共 5 秒
	for i := 0; i < 20; i++ {
		if a.checkPortAlive() {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

// =============================================================================
// killBackendServer 杀死后端子进程
// =============================================================================
// 先尝试 Process.Kill()，如果失败或留下孤儿进程，
// 使用 taskkill /F /T /PID 强制清理整个进程树（Windows）
// =============================================================================
func (a *App) killBackendServer() {
	if a.serverCmd == nil || a.serverCmd.Process == nil {
		return
	}

	pid := a.serverCmd.Process.Pid
	log.Printf("🛑 正在关闭后台服务 (PID: %d)...", pid)

	// 第一步：常规 Kill
	a.serverCmd.Process.Kill()

	// 第二步（Windows）：如果常规 Kill 可能遗留子进程，用 taskkill 清理进程树
	if runtime.GOOS == "windows" {
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
		killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := killCmd.Run(); err != nil {
			log.Printf("⚠️ taskkill fallback: %v（进程可能已退出）", err)
		} else {
			log.Printf("✅ 进程树已清理 (PID: %d)", pid)
		}
	}

	a.serverCmd = nil
}

// SaveFile 保存文件
func (a *App) SaveFile(content string) string {
	filepath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "保存协作文档",
		DefaultFilename: "collab_doc.txt",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "Markdown (*.md)", Pattern: "*.md"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})

	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	if filepath == "" {
		return "CANCELLED"
	}

	err = os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("Error saving file: %s", err)
	}
	return "SUCCESS"
}

// OpenFile 打开文件
func (a *App) OpenFile() string {
	filepath, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "导入本地文件",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text/Markdown", Pattern: "*.txt;*.md;*.js;*.json;*.go;*.py"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})

	if err != nil {
		return "Error"
	}
	if filepath == "" {
		return "CANCELLED"
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %s", err)
	}
	return string(data)
}

func discoveryTargets(port int) []*net.UDPAddr {
	targets := map[string]*net.UDPAddr{
		fmt.Sprintf("255.255.255.255:%d", port): {IP: net.IPv4bcast, Port: port},
		fmt.Sprintf("127.0.0.1:%d", port):       {IP: net.IPv4(127, 0, 0, 1), Port: port},
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return serversFromTargetMap(targets)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || len(ipNet.Mask) != net.IPv4len {
				continue
			}

			broadcast := net.IPv4(
				ip[0]|^ipNet.Mask[0],
				ip[1]|^ipNet.Mask[1],
				ip[2]|^ipNet.Mask[2],
				ip[3]|^ipNet.Mask[3],
			)
			key := fmt.Sprintf("%s:%d", broadcast.String(), port)
			targets[key] = &net.UDPAddr{IP: broadcast, Port: port}
		}
	}

	return serversFromTargetMap(targets)
}

func serversFromTargetMap(targets map[string]*net.UDPAddr) []*net.UDPAddr {
	keys := make([]string, 0, len(targets))
	for key := range targets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]*net.UDPAddr, 0, len(keys))
	for _, key := range keys {
		result = append(result, targets[key])
	}
	return result
}

func prioritizeLANServers(found map[string]LANServer) []LANServer {
	localHostname, _ := os.Hostname()
	localKey := normalizeServerName(localHostname)

	bestByHost := make(map[string]LANServer)
	bestRankByHost := make(map[string]int)

	for _, server := range found {
		server = decorateLANServer(server, localKey)
		hostKey := dedupeHostKey(server, localKey)
		rank := lanServerRank(server, localKey)

		if existingRank, exists := bestRankByHost[hostKey]; !exists || rank < existingRank || (rank == existingRank && server.IP < bestByHost[hostKey].IP) {
			bestByHost[hostKey] = server
			bestRankByHost[hostKey] = rank
		}
	}

	result := make([]LANServer, 0, len(bestByHost))
	for _, server := range bestByHost {
		result = append(result, server)
	}

	sort.Slice(result, func(i, j int) bool {
		leftRank := lanServerRank(result[i], localKey)
		rightRank := lanServerRank(result[j], localKey)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].IP < result[j].IP
	})

	return result
}

func decorateLANServer(server LANServer, localKey string) LANServer {
	host := hostPart(server.IP)
	ip := net.ParseIP(host)
	isLocal := isLocalServer(server, localKey)

	switch {
	case isUsableLANIP(ip) && !isLocal:
		server.Tag = "推荐"
		server.Recommended = true
	case isUsableLANIP(ip) && isLocal:
		server.Tag = "本机"
		server.Recommended = false
	case isLoopbackHost(host, ip):
		server.Tag = "本机"
		server.Recommended = false
	default:
		server.Tag = "备用"
		server.Recommended = false
	}

	return server
}

func dedupeHostKey(server LANServer, localKey string) string {
	nameKey := normalizeServerName(server.Name)
	if isLocalServer(server, localKey) {
		return "local"
	}
	if nameKey != "" && nameKey != "collabstudio 主机" {
		return "host:" + nameKey
	}
	return "addr:" + hostPart(server.IP)
}

func lanServerRank(server LANServer, localKey string) int {
	host := hostPart(server.IP)
	ip := net.ParseIP(host)
	isLocal := isLocalServer(server, localKey)

	switch {
	case isUsableLANIP(ip) && !isLocal:
		return 10
	case isUsableLANIP(ip) && isLocal:
		return 30
	case isLoopbackHost(host, ip):
		return 60
	case isLinkLocalIPv4(ip):
		return 80
	case isBenchmarkIPv4(ip):
		return 90
	default:
		return 70
	}
}

func isLocalServer(server LANServer, localKey string) bool {
	if localKey == "" {
		return strings.EqualFold(server.Name, "本机服务器")
	}
	nameKey := normalizeServerName(server.Name)
	return nameKey == localKey || strings.EqualFold(server.Name, "本机服务器")
}

func normalizeServerName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func hostPart(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return strings.Trim(host, "[]")
	}

	if idx := strings.LastIndex(address, ":"); idx > -1 {
		return strings.Trim(address[:idx], "[]")
	}
	return strings.Trim(address, "[]")
}

func isLoopbackHost(host string, ip net.IP) bool {
	return strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())
}

func isUsableLANIP(ip net.IP) bool {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}
	if isLinkLocalIPv4(ipv4) || isBenchmarkIPv4(ipv4) || ipv4[0] == 127 || ipv4[0] == 0 {
		return false
	}
	return ipv4[0] == 10 ||
		(ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31) ||
		(ipv4[0] == 192 && ipv4[1] == 168)
}

func isLinkLocalIPv4(ip net.IP) bool {
	ipv4 := ip.To4()
	return ipv4 != nil && ipv4[0] == 169 && ipv4[1] == 254
}

func isBenchmarkIPv4(ip net.IP) bool {
	ipv4 := ip.To4()
	return ipv4 != nil && ipv4[0] == 198 && (ipv4[1] == 18 || ipv4[1] == 19)
}
