package main

import (
	"collab-server/controllers"
	"collab-server/database"
	"collab-server/models"
	"collab-server/websocket"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors" // 🟢 引入标准 CORS 库
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 1. 初始化数据库
	fmt.Println("⏳ 正在连接数据库...")
	database.Connect()
	// 自动迁移数据库结构
	database.DB.AutoMigrate(&models.User{}, &models.Document{}, &models.Message{}, &models.History{})

	// 2. 初始化 WebSocket 中心
	hub := websocket.NewHub()
	go hub.Run()

	// 3. 设置 Gin 路由
	r := gin.Default()

	// 设置文件上传大小限制 (10MB)
	r.MaxMultipartMemory = 10 << 20 // 10 MiB

	// 🟢 核心修复：使用标准库配置 CORS
	// 这种配置允许所有来源 (*)，但不允许携带 Cookie 凭证 (AllowCredentials: false)
	// 这是解决浏览器拦截 200 OK 响应的唯一正解
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // 允许所有 IP 访问 (包括 Wails 和 局域网)
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}

	r.Use(cors.New(config))

	// 开启静态资源服务
	r.Static("/uploads", "./uploads")

	// 4. 定义路由
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/history", controllers.GetHistory)

	// 图片上传接口
	r.POST("/upload", controllers.UploadImage)

	// WebSocket 接口
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c)
	})

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	fmt.Println("🚀 CollabServer 已启动: http://localhost:8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}
