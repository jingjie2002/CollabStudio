package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "CollabStudio",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},

		// 生命周期：启动时自动拉起后端
		OnStartup: app.startup,

		// 生命周期：拦截关闭（房主保护）
		OnBeforeClose: func(ctx context.Context) bool {
			// 只有处于工作区且是当前房间房主时才拦截关闭；大厅允许直接关闭。
			if app.loggedIn && app.workspaceActive && app.roomHostActive && !app.forceClose {
				wailsRuntime.EventsEmit(ctx, "show-exit-warning")
				return true // 阻止关闭
			}
			return false
		},

		// 生命周期：关闭时杀死后端子进程
		OnShutdown: app.shutdown,

		Bind: []interface{}{
			app,
		},
	})

	// 双重保险：Wails Run 退出后再次确保后端子进程被清理
	app.killBackendServer()

	if err != nil {
		println("Error:", err.Error())
	}
}
