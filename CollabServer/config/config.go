package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig 读取 .env 文件
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}
}

// GetEnv 获取具体的配置项，如果没找到就用默认值
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
