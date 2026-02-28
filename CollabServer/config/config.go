package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadConfig 读取 .env 文件
// 优先从可执行文件同级目录加载，兼容被其他程序拉起（CWD 不等于 exe 目录）的场景
func LoadConfig() {
	// 1. 尝试从可执行文件同级目录加载 .env
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		envPath := filepath.Join(exeDir, ".env")
		if _, statErr := os.Stat(envPath); statErr == nil {
			if loadErr := godotenv.Load(envPath); loadErr == nil {
				fmt.Printf("✅ [Config] 已从可执行文件目录加载 .env: %s\n", envPath)
				return
			}
		}
	}

	// 2. Fallback: 从当前工作目录加载（兼容 go run / 开发模式）
	if loadErr := godotenv.Load(); loadErr != nil {
		fmt.Println("⚠️ [Config] .env 文件未找到，使用系统环境变量")
	} else {
		fmt.Println("✅ [Config] 已从工作目录加载 .env")
	}
}

// GetEnv 获取具体的配置项，如果没找到就用默认值
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
