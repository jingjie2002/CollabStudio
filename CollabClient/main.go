package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var serverCmd *exec.Cmd

func main() {
	app := NewApp()

	// 🟢 逻辑修正：先尝试启动/检测后端，获取身份
	app.isHost = startBackendServer()

	err := wails.Run(&options.App{
		Title:  "CollabStudio",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,

		// 拦截关闭逻辑
		OnBeforeClose: func(ctx context.Context) bool {
			// 只有房主才会被拦截
			if app.isHost && !app.forceClose {
				wailsRuntime.EventsEmit(ctx, "show-exit-warning")
				return true
			}
			return false
		},

		OnShutdown: func(ctx context.Context) {
			killBackendServer()
		},
		Bind: []interface{}{
			app,
		},
	})

	killBackendServer()

	if err != nil {
		println("Error:", err.Error())
	}
}

func startBackendServer() bool {
	// 🟢 1. 抢答环节：先检查端口是否已经被占用了
	// 如果现在 ping 8080 能通，说明已经有房主了，我直接当访客
	if checkPortAlive() {
		log.Println("🔍 检测到已有服务器运行中，自动以 [访客] 身份启动。")
		return false
	}

	// --- 以下是没人占端口，我自己尝试当房主的逻辑 ---

	exePath, err := os.Executable()
	if err != nil {
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
		log.Println("⚠️ 未找到本地服务器文件，放弃启动。")
		return false
	}

	cmd := exec.Command(serverPath)
	cmd.Dir = exeDir
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	if err := cmd.Start(); err != nil {
		log.Printf("❌ 启动失败: %v，转为 [访客]", err)
		return false
	}

	serverCmd = cmd

	// 🟢 2. 验证环节：等待我启动的服务器就绪
	// 这里稍微等久一点，确保是我自己启动成功的
	if waitForHealthCheck() {
		log.Printf("✅ 后端服务启动成功 (PID: %d)，我是 [👑 房主]", cmd.Process.Pid)
		return true
	}

	log.Println("⚠️ 启动超时或失败，转为 [访客]")
	return false
}

// 快速检查端口是否存活 (只查一次)
func checkPortAlive() bool {
	client := http.Client{Timeout: 500 * time.Millisecond}

	// 🟢 优先检查 80 端口 (生产环境)
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

// 循环检查健康状态 (用于等待启动)
func waitForHealthCheck() bool {
	// 尝试 20 次，共 5 秒
	for i := 0; i < 20; i++ {
		if checkPortAlive() {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

func killBackendServer() {
	if serverCmd != nil && serverCmd.Process != nil {
		log.Println("🛑 关闭后台服务...")
		serverCmd.Process.Kill()
		serverCmd = nil
	}
}
