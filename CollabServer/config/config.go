package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// =============================================================================
// LoadConfig 读取或自动生成 .env 文件
// =============================================================================
// 加载优先级：
// 1. 可执行文件同级目录的 .env（兼容被客户端拉起的场景）
// 2. 当前工作目录的 .env（兼容 go run 开发模式）
// 3. 如果都不存在 → 自动生成默认 .env（零配置启动）
// =============================================================================
func LoadConfig() {
	envPath := findOrCreateEnvFile()
	if envPath != "" {
		if err := godotenv.Load(envPath); err == nil {
			fmt.Printf("✅ [Config] 已加载 .env: %s\n", envPath)
		}
	}

	// 确保关键密钥存在（即使 .env 文件中遗漏了）
	ensureSecret("JWT_SECRET", envPath)
}

// GetEnv 获取具体的配置项，如果没找到就用默认值
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// =============================================================================
// findOrCreateEnvFile 查找或自动创建 .env 文件
// =============================================================================
func findOrCreateEnvFile() string {
	// 1. 优先检查可执行文件同级目录
	exePath, err := os.Executable()
	if err == nil {
		envPath := filepath.Join(filepath.Dir(exePath), ".env")
		if _, statErr := os.Stat(envPath); statErr == nil {
			return envPath
		}
		// .env 不存在 → 在 exe 目录自动生成
		if generateDefaultEnv(envPath) == nil {
			fmt.Printf("🆕 [Config] 已自动生成 .env: %s\n", envPath)
			return envPath
		}
	}

	// 2. Fallback: 检查当前工作目录
	if _, err := os.Stat(".env"); err == nil {
		return ".env"
	}

	// 3. 在当前目录生成
	if generateDefaultEnv(".env") == nil {
		fmt.Println("🆕 [Config] 已在工作目录自动生成 .env")
		return ".env"
	}

	// 4. 终极回退：用户配置目录（%AppData%/CollabServer/.env）
	if configDir, err := os.UserConfigDir(); err == nil {
		appConfigDir := filepath.Join(configDir, "CollabServer")
		os.MkdirAll(appConfigDir, 0755)
		envPath := filepath.Join(appConfigDir, ".env")
		if _, statErr := os.Stat(envPath); statErr == nil {
			return envPath
		}
		if generateDefaultEnv(envPath) == nil {
			fmt.Printf("🆕 [Config] 已在用户目录自动生成 .env: %s\n", envPath)
			return envPath
		}
	}

	fmt.Println("⚠️ [Config] 无法创建 .env，使用系统环境变量")
	return ""
}

// =============================================================================
// generateDefaultEnv 生成包含随机密钥的默认 .env 文件
// =============================================================================
func generateDefaultEnv(path string) error {
	jwtSecret := generateRandomKey(32)

	content := fmt.Sprintf(`# =============================================================================
# CollabServer 配置（自动生成）
# =============================================================================

# JWT 密钥（自动生成的随机密钥）
JWT_SECRET=%s

# 服务运行端口
PORT=8080

# CORS 白名单（逗号分隔）
CORS_ORIGINS=http://localhost:5173,http://wails.localhost

# 数据库
DB_NAME=collab.db

# Redis（可选）
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=

# 局域网发现端口
DISCOVERY_PORT=9999

# 上传目录
UPLOAD_DIR=uploads

# 前端静态资源目录
DIST_PATH=./dist
`, jwtSecret)

	return os.WriteFile(path, []byte(content), 0600)
}

// =============================================================================
// ensureSecret 确保某个密钥环境变量已设置，若为空则自动补充
// =============================================================================
func ensureSecret(key, envPath string) {
	if val, exists := os.LookupEnv(key); exists && val != "" && val != "CHANGE_ME_TO_A_RANDOM_SECRET_KEY" && val != "CHANGE_ME_TO_A_STRONG_PASSWORD" {
		return // 已有有效值
	}

	// 生成随机值并设置到环境变量
	newVal := generateRandomKey(32)
	os.Setenv(key, newVal)
	fmt.Printf("🔑 [Config] %s 已自动生成随机值\n", key)

	// 尝试回写到 .env 文件
	if envPath != "" {
		appendToEnvFile(envPath, key, newVal)
	}
}

// generateRandomKey 生成指定字节数的 Base64 随机密钥
func generateRandomKey(bytes int) string {
	b := make([]byte, bytes)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// appendToEnvFile 将 key=value 追加或替换到 .env 文件
func appendToEnvFile(path, key, value string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+"=") {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}
