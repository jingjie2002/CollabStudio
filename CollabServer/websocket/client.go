package websocket

import (
	"collab-server/config"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

		// 生产环境收紧空 Origin。仅允许本机/Wails 桌面场景。
		if origin == "" {
			if os.Getenv("GIN_MODE") != "release" {
				log.Println("📡 WebSocket 连接无 Origin 头，开发模式允许")
				return true
			}

			remoteHost := r.RemoteAddr
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				remoteHost = host
			}
			ua := strings.ToLower(r.UserAgent())
			ip := net.ParseIP(remoteHost)
			if (ip != nil && ip.IsLoopback()) || strings.Contains(ua, "wails") {
				log.Println("📡 WebSocket 连接无 Origin 头，生产模式按本机/Wails 允许")
				return true
			}

			log.Println("⚠️ WebSocket 拒绝空 Origin 连接")
			return false
		}

		// 从环境变量读取白名单
		allowedOrigins := config.GetEnv("CORS_ORIGINS", "")
		if allowedOrigins == "" {
			// 生产模式下，空白名单 → 拒绝（防止部署时忘配）
			if os.Getenv("GIN_MODE") == "release" {
				log.Printf("❌ WebSocket 生产模式拒绝: CORS_ORIGINS 未配置, origin=%s", origin)
				return false
			}
			// 开发模式全放行
			log.Println("📡 WebSocket CORS 未配置，开发模式放行:", origin)
			return true
		}

		// 🟢 调试日志：显示当前检查的 Origin 和白名单
		log.Printf("📡 WebSocket Origin 检查: origin=%s, 白名单=%s", origin, allowedOrigins)

		originURL, err := url.Parse(origin)
		if err != nil || originURL.Scheme == "" || originURL.Hostname() == "" {
			log.Printf("⚠️ WebSocket 非法 Origin 格式: %s", origin)
			return false
		}

		// 检查 origin 是否在白名单中（协议+主机严格匹配，端口可选匹配）
		for _, allowed := range strings.Split(allowedOrigins, ",") {
			allowed = strings.TrimSpace(allowed)
			if allowed == "" {
				continue
			}

			allowedURL, parseErr := url.Parse(allowed)
			if parseErr != nil || allowedURL.Scheme == "" || allowedURL.Hostname() == "" {
				continue
			}

			sameScheme := strings.EqualFold(originURL.Scheme, allowedURL.Scheme)
			sameHost := strings.EqualFold(originURL.Hostname(), allowedURL.Hostname())
			portAllowed := allowedURL.Port() == "" || originURL.Port() == allowedURL.Port()
			if sameScheme && sameHost && portAllowed {
				log.Printf("✅ WebSocket Origin 匹配: origin=%s, allowed=%s", origin, allowed)
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
	UserID   uint
	UUID     string // 🟢 唯一客户端标识，用于防止消息反射
}

func extractTokenFromRequest(c *gin.Context) string {
	if token := strings.TrimSpace(c.Query("token")); token != "" {
		return token
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func authenticateWS(c *gin.Context) (string, uint, error) {
	tokenString := extractTokenFromRequest(c)
	if tokenString == "" {
		return "", 0, fmt.Errorf("缺少 token")
	}

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return "", 0, fmt.Errorf("服务器密钥未配置")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非法签名算法")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", 0, fmt.Errorf("token 无效")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", 0, fmt.Errorf("无法解析 token claims")
	}

	usernameVal, ok := claims["username"]
	if !ok {
		return "", 0, fmt.Errorf("缺少 username claims")
	}
	username, ok := usernameVal.(string)
	if !ok || strings.TrimSpace(username) == "" {
		return "", 0, fmt.Errorf("username claims 非法")
	}

	var userID uint
	if userIDRaw, exists := claims["userId"]; exists {
		switch v := userIDRaw.(type) {
		case float64:
			if v > 0 {
				userID = uint(v)
			}
		case int:
			if v > 0 {
				userID = uint(v)
			}
		case int64:
			if v > 0 {
				userID = uint(v)
			}
		}
	}

	return strings.TrimSpace(username), userID, nil
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
	username, userID, err := authenticateWS(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "WebSocket 鉴权失败"})
		return
	}

	roomID := c.Query("room")
	if roomID == "" {
		roomID = "lobby"
	}
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		roomID = "lobby"
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
		UserID:   userID,
		UUID:     clientUUID,
	}

	// 🟢 立即发送 client_id 给前端，用于消息隔离
	welcomeMsg := []byte(`{"type":"client_id","uuid":"` + clientUUID + `"}`)
	conn.WriteMessage(websocket.TextMessage, welcomeMsg)

	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}
