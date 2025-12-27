package main

import (
	"collab-server/config"
	"collab-server/controllers"
	"collab-server/database"
	"collab-server/models"
	"collab-server/websocket"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// =============================================================================
// 🔧 环境自动判定：选择服务器端口
// =============================================================================
// 规则：
//   - GIN_MODE=release 或 Linux 系统 → 80 端口（生产环境）
//   - 其他情况（Windows 本地开发）→ 8080 端口
//
// =============================================================================
func getServerPort() string {
	// 优先检测 GIN_MODE 环境变量
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		log.Println("🌐 [环境判定] GIN_MODE=release → 使用生产端口 80")
		return "80"
	}

	// 检测操作系统
	if runtime.GOOS == "linux" {
		log.Println("🌐 [环境判定] Linux 系统 → 使用生产端口 80")
		return "80"
	}

	// 默认本地开发环境
	log.Println("💻 [环境判定] 本地开发环境 → 使用开发端口 8080")
	return "8080"
}

// =============================================================================
// 🔥 CollabServer 主入口
// =============================================================================
// 本文件是整个后端服务的"大脑"。它的职责是：
// 1. 加载配置 (.env)
// 2. 初始化数据库连接
// 3. 启动 WebSocket Hub (实时协作的心脏)
// 4. 配置 HTTP 路由 (Gin 框架)
// 5. 监听系统信号，实现优雅停机
// =============================================================================

// hub 是全局的 WebSocket 中心，需要在优雅停机时访问
var hub *websocket.Hub

func main() {
	// 🔧 设置日志格式：包含时间戳和文件名行号，方便调试定位
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// ==========================================================================
	// 阶段 0：加载配置
	// ==========================================================================
	config.LoadConfig()

	// 🔐 安全检查：JWT_SECRET 必须配置，否则拒绝启动
	// 这是防止生产环境用默认密钥的硬性保障
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("❌ 致命错误: JWT_SECRET 未配置！请在 .env 文件中设置此项。拒绝启动。")
	}

	// ==========================================================================
	// 阶段 1：初始化数据库
	// ==========================================================================
	fmt.Println("⏳ 正在连接数据库...")
	database.Connect()
	// AutoMigrate 会自动创建或更新表结构，非常适合快速迭代
	database.DB.AutoMigrate(&models.User{}, &models.Document{}, &models.Message{}, &models.History{})

	// ==========================================================================
	// 阶段 2：初始化 WebSocket Hub
	// ==========================================================================
	// Hub 是协作系统的心脏，它管理所有房间和客户端连接
	// 使用全局变量是为了让优雅停机逻辑能够访问它
	hub = websocket.NewHub()
	go hub.Run() // 在独立 goroutine 中运行 Hub 的事件循环

	// ==========================================================================
	// 阶段 3：配置 Gin 路由引擎
	// ==========================================================================
	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20 // 限制上传文件大小为 10MB

	// -------------------------------------------------------------------------
	// 🔐 CORS 安全配置（核心加固点）
	// -------------------------------------------------------------------------
	// CORS (跨域资源共享) 决定了哪些域名可以访问你的 API
	// 生产环境绝不能用 AllowAllOrigins: true，否则任何网站都能调用你的接口
	corsConfig := cors.DefaultConfig()

	// 从 .env 读取白名单，格式：CORS_ORIGINS=http://localhost:5173,http://119.29.55.127
	corsOrigins := config.GetEnv("CORS_ORIGINS", "")
	if corsOrigins == "" {
		// 如果未配置，给予合理的开发默认值
		log.Println("⚠️ CORS_ORIGINS 未配置，使用默认值: http://localhost:5173")
		corsConfig.AllowOrigins = []string{"http://localhost:5173", "http://localhost:8080", "http://127.0.0.1:5173"}
	} else {
		// 解析逗号分隔的域名列表
		origins := strings.Split(corsOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		corsConfig.AllowOrigins = origins
		log.Printf("🔐 CORS 白名单已加载: %v", origins)
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true // 允许携带 Cookie/Token

	r.Use(cors.New(corsConfig))

	// -------------------------------------------------------------------------
	// 🎯 启动辅助服务
	// -------------------------------------------------------------------------
	go startUDPDiscoveryService() // 局域网发现服务

	// -------------------------------------------------------------------------
	// 📂 静态资源服务配置
	// -------------------------------------------------------------------------
	// 🔧 UPLOAD_DIR: 用户上传文件存储目录
	// 使用相对路径时，基于程序执行位置（通常是 CollabServer 目录）
	uploadDir := config.GetEnv("UPLOAD_DIR", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("⚠️ 无法创建 uploads 目录: %v", err)
	} else {
		log.Printf("📂 上传目录已就绪: %s", uploadDir)
	}
	r.Static("/uploads", uploadDir)

	// 🔧 DIST_PATH: 前端打包产物目录
	// 部署时需确保 dist 文件夹与 collab_server 在正确的相对位置
	// 或者配置绝对路径
	distPath := config.GetEnv("DIST_PATH", "./dist")

	// 🔧 将相对路径转换为绝对路径，确保 Linux 环境下路径识别万无一失
	absoluteDistPath, err := filepath.Abs(distPath)
	if err != nil {
		log.Printf("⚠️ 无法解析 DIST_PATH 绝对路径: %v", err)
		absoluteDistPath = distPath // 降级使用原始路径
	}
	log.Printf("📂 正在尝试挂载静态资源目录: %s", absoluteDistPath)

	if absoluteDistPath != "" {
		if _, err := os.Stat(absoluteDistPath); err == nil {
			// 挂载静态文件目录
			r.Static("/assets", filepath.Join(absoluteDistPath, "assets"))
			r.StaticFile("/favicon.ico", filepath.Join(absoluteDistPath, "favicon.ico"))

			// 🏠 根路径返回 index.html
			indexPath := filepath.Join(absoluteDistPath, "index.html")
			r.GET("/", func(c *gin.Context) {
				c.File(indexPath)
			})

			log.Printf("✅ 前端静态资源已挂载: %s", absoluteDistPath)

			// =========================================================================
			// 🔥 SPA 回退逻辑（关键！）
			// =========================================================================
			// 问题：Vue/React 等 SPA 使用前端路由，当用户直接访问 /login 或 /room/123 时，
			//       后端找不到对应文件会返回 404
			// 解决：对于非 API/WebSocket 请求，统一返回 index.html，让前端路由接管
			// =========================================================================
			r.NoRoute(func(c *gin.Context) {
				path := c.Request.URL.Path
				// 如果请求的不是 API 或 WebSocket 端点，返回 index.html
				if !strings.HasPrefix(path, "/api") &&
					!strings.HasPrefix(path, "/ws") &&
					!strings.HasPrefix(path, "/uploads") &&
					!strings.HasPrefix(path, "/register") &&
					!strings.HasPrefix(path, "/login") &&
					!strings.HasPrefix(path, "/history") &&
					!strings.HasPrefix(path, "/upload") &&
					!strings.HasPrefix(path, "/ping") {
					c.File(filepath.Join(absoluteDistPath, "index.html"))
					return
				}
				// 对于 API 请求但路由不存在的情况，返回 404 JSON
				c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
			})
		} else {
			log.Printf("⚠️ DIST_PATH 指定的目录不存在: %s (原始路径: %s)", absoluteDistPath, distPath)
		}
	}

	// -------------------------------------------------------------------------
	// 🛤️ 路由定义
	// -------------------------------------------------------------------------
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/history", controllers.GetHistory)
	r.POST("/upload", controllers.UploadImage)

	// WebSocket 端点
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c)
	})

	// 健康检查端点（用于负载均衡器或监控系统）
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "status": "healthy"})
	})

	// ==========================================================================
	// 阶段 4：启动 HTTP 服务器（环境自动判定端口）
	// ==========================================================================
	port := getServerPort()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 在独立 goroutine 中启动服务器，这样主 goroutine 可以监听停机信号
	go func() {
		fmt.Printf("🚀 CollabServer 已启动: http://0.0.0.0:%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %s", err)
		}
	}()

	// ==========================================================================
	// 阶段 5：优雅停机 (Graceful Shutdown)
	// ==========================================================================
	// 这是生产级服务的必备能力。当收到 Ctrl+C 或 kill 信号时：
	// 1. 停止接受新请求
	// 2. 等待现有请求处理完毕
	// 3. 刷新内存中的数据到磁盘
	// 4. 关闭数据库连接
	// 5. 优雅退出
	gracefulShutdown(srv)
}

// =============================================================================
// gracefulShutdown 实现优雅停机逻辑
// =============================================================================
// 为什么需要优雅停机？
// 1. 直接 kill 进程会导致正在处理的请求中断，用户体验差
// 2. 内存中的数据可能丢失（如 Hub 中的文档内容）
// 3. 数据库连接被强制关闭可能导致数据损坏
// =============================================================================
func gracefulShutdown(srv *http.Server) {
	// 创建一个通道来接收系统信号
	quit := make(chan os.Signal, 1)

	// signal.Notify 告诉 Go 运行时：把这些信号发送到 quit 通道
	// SIGINT = Ctrl+C
	// SIGTERM = kill 命令 (Kubernetes/Docker 默认发送此信号)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号。这行代码会一直等着，直到收到 SIGINT 或 SIGTERM
	sig := <-quit
	log.Printf("🛑 收到停机信号: %v，开始优雅关闭...", sig)

	// -------------------------------------------------------------------------
	// 步骤 1：刷新 Hub 中的热数据到数据库
	// -------------------------------------------------------------------------
	log.Println("📝 正在保存所有房间的文档数据...")
	if hub != nil {
		hub.FlushAllRoomsToDB()
	}

	// -------------------------------------------------------------------------
	// 步骤 2：关闭 HTTP 服务器（给 5 秒时间处理剩余请求）
	// -------------------------------------------------------------------------
	log.Println("⏳ 正在关闭 HTTP 服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ HTTP 服务关闭异常: %v", err)
	}

	// -------------------------------------------------------------------------
	// 步骤 3：关闭数据库连接
	// -------------------------------------------------------------------------
	log.Println("🔌 正在关闭数据库连接...")
	database.Close()

	log.Println("✅ CollabServer 已安全停止。再见！")
}

// =============================================================================
// startUDPDiscoveryService 局域网自动发现服务
// =============================================================================
// 原理：监听 UDP 9999 端口，当收到 "WHOIS_COLLAB_HOST" 暗号时，
// 回复 "IAM_HOST|主机名"，让同一局域网内的客户端能找到服务器。
// =============================================================================
func startUDPDiscoveryService() {
	discoveryPortStr := config.GetEnv("DISCOVERY_PORT", "9999")
	discoveryPort, _ := strconv.Atoi(discoveryPortStr)

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", discoveryPort))
	if err != nil {
		log.Println("❌ UDP 地址解析失败:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("❌ 无法启动局域网发现服务 (可能是端口占用):", err)
		return
	}
	defer conn.Close()

	log.Printf("📡 局域网广播服务已启动 (UDP:%d)，等待被发现...", discoveryPort)

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := string(buf[:n])
		if strings.TrimSpace(msg) == "WHOIS_COLLAB_HOST" {
			hostname, _ := os.Hostname()
			if hostname == "" {
				hostname = "Unknown-Host"
			}
			reply := fmt.Sprintf("IAM_HOST|%s", hostname)
			conn.WriteToUDP([]byte(reply), remoteAddr)
		}
	}
}
