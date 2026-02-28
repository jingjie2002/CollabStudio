package router

import (
	"collab-server/config"
	"collab-server/controllers"
	"collab-server/websocket"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// =============================================================================
// SetupRouter 配置并返回 Gin 路由引擎
// =============================================================================
func SetupRouter(hub *websocket.Hub) *gin.Engine {
	// ⚡ v3.8.8 调试锚点：确认代码已生效
	absPath, _ := filepath.Abs(".")
	log.Printf("⚡ [路由激活] v3.8.8 物理路径: %s", absPath)

	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20 // 10MB

	// CORS 配置
	setupCORS(r)

	// 静态资源
	setupStaticFiles(r)

	// API 路由
	setupRoutes(r, hub)

	return r
}

// =============================================================================
// setupCORS v4.0 智能识别模式
// =============================================================================
// 升级逻辑：根据 GIN_MODE 环境变量智能切换 CORS 策略
//   - GIN_MODE=release：生产模式，严格白名单（CORS_ORIGINS + wails://）
//   - 其他值或未设置：开发模式，全放行
//
// =============================================================================
func setupCORS(r *gin.Engine) {
	ginMode := os.Getenv("GIN_MODE")
	isRelease := ginMode == "release"

	if isRelease {
		log.Println("🔒 [CORS] v4.0 生产模式已启用（严格白名单）")
	} else {
		log.Println("🔓 [CORS] v4.0 开发模式已启用（全放行）")
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = nil // 使用 AllowOriginFunc 进行动态判断

	if isRelease {
		// =============================================================================
		// 🔒 v4.0 生产模式：严格白名单
		// =============================================================================
		allowedOrigins := parseOrigins(os.Getenv("CORS_ORIGINS"))
		log.Printf("📋 [CORS] 白名单: %v", allowedOrigins)

		corsConfig.AllowOriginFunc = func(origin string) bool {
			// 允许 wails:// 协议（Wails 桌面客户端旧版）
			if strings.HasPrefix(origin, "wails://") {
				log.Printf("✅ [CORS] 放行 Wails 客户端: %s", origin)
				return true
			}
			// 允许 http://wails.localhost（Wails 2.x 生产环境 WebView）
			if origin == "http://wails.localhost" || strings.HasPrefix(origin, "http://wails.localhost:") {
				log.Printf("✅ [CORS] 放行 Wails WebView: %s", origin)
				return true
			}
			// 检查白名单
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					log.Printf("✅ [CORS] 白名单放行: %s", origin)
					return true
				}
			}
			log.Printf("❌ [CORS] 拒绝未授权来源: %s", origin)
			return false
		}
	} else {
		// =============================================================================
		// 🔓 v4.0 开发模式：全放行
		// =============================================================================
		corsConfig.AllowOriginFunc = func(origin string) bool {
			log.Printf("✅ [CORS] 开发模式放行: %s", origin)
			return true
		}
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true

	r.Use(cors.New(corsConfig))
	log.Println("✅ [CORS] 中间件已挂载")
}

// =============================================================================
// parseOrigins 解析逗号分隔的 CORS_ORIGINS 环境变量
// =============================================================================
func parseOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// =============================================================
// setupStaticFiles 配置静态文件服务
// =============================================================
func setupStaticFiles(r *gin.Engine) {
	// 上传目录配置
	uploadDir := config.GetEnv("UPLOAD_DIR", "uploads")
	os.MkdirAll(uploadDir, 0755)
	log.Printf("📂 [Static] 上传目录已挂载: /uploads -> %s", uploadDir)

	r.GET("/uploads/*filepath", func(c *gin.Context) {
		filePath := c.Param("filepath")
		fullPath := filepath.Join(uploadDir, filePath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			log.Printf("❌ [Static] 文件不存在: %s", fullPath)
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}
		// v3.8.4: 强制添加 CORS 头，防止跨域拦截
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		log.Printf("✅ [Static] 提供文件: %s", fullPath)
		c.File(fullPath)
	})

	// 前端静态资源
	distPath := config.GetEnv("DIST_PATH", "./dist")
	absDistPath, _ := filepath.Abs(distPath)

	if _, err := os.Stat(absDistPath); err == nil {
		r.Static("/assets", filepath.Join(absDistPath, "assets"))
		r.StaticFile("/favicon.ico", filepath.Join(absDistPath, "favicon.ico"))

		indexPath := filepath.Join(absDistPath, "index.html")
		r.GET("/", func(c *gin.Context) {
			c.File(indexPath)
		})

		log.Printf("✅ 前端静态资源已挂载: %s", absDistPath)

		// SPA 回退
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if !strings.HasPrefix(path, "/api") &&
				!strings.HasPrefix(path, "/ws") &&
				!strings.HasPrefix(path, "/uploads") &&
				!strings.HasPrefix(path, "/register") &&
				!strings.HasPrefix(path, "/login") &&
				!strings.HasPrefix(path, "/history") &&
				!strings.HasPrefix(path, "/upload") &&
				!strings.HasPrefix(path, "/ping") {
				c.File(indexPath)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		})
	}
}

// setupRoutes 配置 API 路由
func setupRoutes(r *gin.Engine, hub *websocket.Hub) {
	// 公开路由
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/history", controllers.GetHistory)
	r.POST("/upload", controllers.UploadImage)

	// WebSocket
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c)
	})

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "status": "healthy"})
	})

	// 管理员 API
	admin := r.Group("/api/admin")
	admin.Use(controllers.AdminAuth())
	{
		admin.GET("/rooms", controllers.GetAllRooms(hub))
		admin.GET("/stats", controllers.GetServerStats(hub))
		admin.POST("/clear-room", controllers.ClearRoom(hub))
	}
}
