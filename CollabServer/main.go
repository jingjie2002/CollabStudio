package main

import (
	"collab-server/controllers"
	"collab-server/database"
	"collab-server/models"
	"collab-server/websocket"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 🟢 定义发现端口
const DiscoveryPort = 9999

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 1. 初始化数据库
	fmt.Println("⏳ 正在连接数据库...")
	database.Connect()
	database.DB.AutoMigrate(&models.User{}, &models.Document{}, &models.Message{}, &models.History{})

	// 2. 初始化 WebSocket 中心
	hub := websocket.NewHub()
	go hub.Run()

	// 🟢 3. 启动 UDP 广播服务 (让别人能搜到我)
	go startUDPDiscoveryService()

	// 4. 设置 Gin 路由
	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20 // 10 MiB

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	r.Use(cors.New(config))

	r.Static("/uploads", "./uploads")

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/history", controllers.GetHistory)
	r.POST("/upload", controllers.UploadImage)
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c)
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	fmt.Println("🚀 CollabServer 已启动: http://localhost:8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}

// 🟢 UDP 发现服务：监听 9999 端口，收到暗号就回复自己的主机名
func startUDPDiscoveryService() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", DiscoveryPort))
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

	log.Println("📡 局域网广播服务已启动，等待被发现...")

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := string(buf[:n])
		// 收到暗号 "WHOIS_COLLAB_HOST"
		if strings.TrimSpace(msg) == "WHOIS_COLLAB_HOST" {
			// 获取本机主机名，方便对方识别
			hostname, _ := os.Hostname()
			if hostname == "" {
				hostname = "Unknown-Host"
			}

			// 回复格式: "IAM_HOST|主机名"
			reply := fmt.Sprintf("IAM_HOST|%s", hostname)
			conn.WriteToUDP([]byte(reply), remoteAddr)
			// log.Printf("📡 回应了来自 %s 的搜索请求", remoteAddr.String())
		}
	}
}
