package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	isHost     bool // 标记是否为房主
	forceClose bool // 标记是否确认强制退出
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
