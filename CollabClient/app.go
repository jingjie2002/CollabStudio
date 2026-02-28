package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
func (a *App) killBackendServer() {
	if a.serverCmd != nil && a.serverCmd.Process != nil {
		log.Printf("🛑 正在关闭后台服务 (PID: %d)...", a.serverCmd.Process.Pid)
		a.serverCmd.Process.Kill()
		a.serverCmd = nil
	}
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
