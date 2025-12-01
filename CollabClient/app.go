package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	isHost     bool // 标记是否为房主
	forceClose bool // 标记是否确认强制退出
}

// ServerInfo 定义扫描结果结构体
type ServerInfo struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// IsHostUser 供前端查询是否为房主
func (a *App) IsHostUser() bool {
	return a.isHost
}

// ConfirmExit 前端确认退出后调用
func (a *App) ConfirmExit() {
	a.forceClose = true
	runtime.Quit(a.ctx)
}

// 🟢 核心修复：全频段雷达扫描
// 遍历所有网卡发送 UDP 广播，解决多网卡环境下广播发不出去的问题
func (a *App) ScanLanServers() []ServerInfo {
	var servers []ServerInfo
	port := 9999
	uniqueIPs := make(map[string]bool)

	// 1. 监听一个随机端口，准备接收来自四面八方的回信
	// 这里监听 0.0.0.0:0，让系统自动分配端口
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		// 极少情况会失败
		return servers
	}
	defer conn.Close()

	// 2. 获取本机所有网络接口 (Wi-Fi, 以太网, 虚拟网卡...)
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			// 过滤掉挂掉的、不支持广播的、回环接口
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}

			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				// 只处理 IPv4
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					// 🔥 计算该网段的定向广播地址
					// 例如本机 IP 192.168.1.5 -> 广播地址 192.168.1.255
					// 这样能确保信号强制走这个网卡发出
					ip := ipnet.IP.To4()
					mask := ipnet.Mask
					broadcastIP := net.IP(make([]byte, 4))
					for i := 0; i < 4; i++ {
						broadcastIP[i] = ip[i] | ^mask[i]
					}

					// 发送探测信号 "WHOIS_COLLAB_HOST"
					dstAddr := &net.UDPAddr{IP: broadcastIP, Port: port}
					conn.WriteTo([]byte("WHOIS_COLLAB_HOST"), dstAddr)
				}
			}
		}
	}

	// 同时也发一份给通用广播地址（双重保险，有些路由器只认这个）
	universalAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	conn.WriteTo([]byte("WHOIS_COLLAB_HOST"), universalAddr)

	// 3. 设置超时时间 (1.5秒)
	conn.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))

	// 4. 循环读取回信
	buffer := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			break // 超时结束
		}

		msg := string(buffer[:n])
		// 验证暗号格式: "IAM_HOST|主机名"
		if strings.HasPrefix(msg, "IAM_HOST|") {
			parts := strings.Split(msg, "|")
			hostname := "Unknown"
			if len(parts) > 1 {
				hostname = parts[1]
			}

			// 获取对方 IP
			var remoteIP string
			if udpAddr, ok := addr.(*net.UDPAddr); ok {
				remoteIP = udpAddr.IP.String()
			} else {
				continue
			}

			// 去重并添加到结果列表
			if !uniqueIPs[remoteIP] {
				uniqueIPs[remoteIP] = true
				servers = append(servers, ServerInfo{
					IP:   fmt.Sprintf("%s:8080", remoteIP), // 拼接端口
					Name: hostname,
				})
			}
		}
	}

	return servers
}

// SaveFile 保存文件
func (a *App) SaveFile(content string) string {
	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存协作文档",
		DefaultFilename: "collab_doc.txt",
		Filters: []runtime.FileFilter{
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
	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入本地文件",
		Filters: []runtime.FileFilter{
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
