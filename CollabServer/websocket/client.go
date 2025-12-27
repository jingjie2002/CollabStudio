package websocket

import (
	"collab-server/config"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// =============================================================================
// WebSocket 连接配置常量
// =============================================================================
// 这些常量定义了 WebSocket 连接的超时和缓冲区大小。
// 在高并发场景下，合理的配置能防止资源耗尽。
// =============================================================================
const (
	writeWait      = 10 * time.Second    // 写操作超时：超过此时间写入失败则断开
	pongWait       = 60 * time.Second    // Pong 等待：客户端必须在此时间内响应 Ping
	pingPeriod     = (pongWait * 9) / 10 // Ping 周期：略小于 Pong 等待，确保及时检测
	maxMessageSize = 10 * 1024 * 1024    // 最大消息：10MB，支持大型富文本文档
)

// =============================================================================
// WebSocket Upgrader 配置
// =============================================================================
// Upgrader 负责将 HTTP 连接升级为 WebSocket 连接。
// CheckOrigin 是防止 CSRF 攻击的关键：它验证请求来源是否在白名单中。
// =============================================================================
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// 🔐 安全校验：只允许白名单中的域名建立 WebSocket 连接
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		// 如果没有 Origin 头（某些特殊情况），允许连接
		// 这主要是为了支持 Wails 桌面客户端（它可能不发送 Origin）
		if origin == "" {
			log.Println("📡 WebSocket 连接无 Origin 头，允许 (可能是桌面客户端)")
			return true
		}

		// 从环境变量读取白名单
		allowedOrigins := config.GetEnv("CORS_ORIGINS", "")
		if allowedOrigins == "" {
			// 未配置时，使用开发默认值
			allowedOrigins = "http://localhost:5173,http://localhost:8080,http://127.0.0.1:5173,wails://wails"
		}

		// 🟢 调试日志：显示当前检查的 Origin 和白名单
		log.Printf("📡 WebSocket Origin 检查: origin=%s, 白名单=%s", origin, allowedOrigins)

		// 检查 origin 是否在白名单中
		for _, allowed := range strings.Split(allowedOrigins, ",") {
			allowed = strings.TrimSpace(allowed)
			// 精确匹配
			if allowed == origin {
				log.Printf("✅ WebSocket Origin 匹配: %s", origin)
				return true
			}
			// 🟢 前缀匹配：支持 http://119.29.55.127 匹配 http://119.29.55.127:80
			if strings.HasPrefix(origin, allowed) || strings.HasPrefix(allowed, origin) {
				log.Printf("✅ WebSocket Origin 前缀匹配: origin=%s, allowed=%s", origin, allowed)
				return true
			}
		}

		// 不在白名单中，记录日志并拒绝
		log.Printf("⚠️ WebSocket 拒绝非法来源: %s (请检查 .env 中的 CORS_ORIGINS)", origin)
		return false
	},
}

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	RoomID   string
	Username string
	UUID     string // 🟢 唯一客户端标识，用于防止消息反射
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.Hub.broadcast <- BroadcastMessage{
			RoomID:  c.RoomID,
			Message: message,
			Sender:  c,
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 保持你的修复：确保每次只写一条消息，避免粘包
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, c *gin.Context) {
	roomID := c.Query("room")
	if roomID == "" {
		roomID = "lobby"
	}
	username := c.Query("username")
	if username == "" {
		username = "Anonymous"
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// 🟢 生成唯一客户端 UUID
	clientUUID := uuid.New().String()

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		RoomID:   roomID,
		Username: username,
		UUID:     clientUUID,
	}

	// 🟢 立即发送 client_id 给前端，用于消息隔离
	welcomeMsg := []byte(`{"type":"client_id","uuid":"` + clientUUID + `"}`)
	conn.WriteMessage(websocket.TextMessage, welcomeMsg)

	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}
