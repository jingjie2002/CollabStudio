package controllers

import (
	"collab-server/websocket"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// =============================================================================
// AdminAuth 管理员身份验证中间件
// =============================================================================
// 从 JWT Token 中解析用户角色，仅允许 role == "admin" 的用户访问
// =============================================================================
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization 头"})
			c.Abort()
			return
		}

		// 去掉 "Bearer " 前缀
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// 解析 Token（不验证签名，仅解析 claims）
		// 注意：生产环境应使用完整的签名验证
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 解析失败"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token 格式"})
			c.Abort()
			return
		}

		// 验证管理员角色
		role, exists := claims["role"].(string)
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		// 将用户信息存入 Context
		c.Set("userId", claims["userId"])
		c.Set("username", claims["username"])
		c.Set("role", role)

		c.Next()
	}
}

// =============================================================================
// GetAllRooms 获取所有活跃房间列表
// =============================================================================
// 高阶函数，接收 Hub 实例，返回 Gin 处理函数
// 返回数据：房间 ID、在线用户数
// =============================================================================
func GetAllRooms(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomStats := hub.GetRoomStats()
		c.JSON(http.StatusOK, gin.H{
			"rooms": roomStats,
			"total": len(roomStats),
		})
	}
}

// =============================================================================
// GetServerStats 获取服务器实时统计信息
// =============================================================================
// 返回数据：总连接数、活跃房间数、Go 协程数量
// =============================================================================
func GetServerStats(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := hub.GetStats()
		stats["goroutines"] = runtime.NumGoroutine()

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"stats":  stats,
		})
	}
}

// =============================================================================
// ClearRoom 清空指定房间（预留接口）
// =============================================================================
// 接收 room_id 参数，执行房间清理逻辑
// =============================================================================
func ClearRoom(c *gin.Context) {
	roomID := c.PostForm("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 room_id 参数"})
		return
	}

	// TODO: 实现房间清理逻辑
	// 目前仅返回成功响应，实际清理逻辑需根据业务需求实现
	c.JSON(http.StatusOK, gin.H{
		"message": "房间清理请求已接收",
		"room_id": roomID,
	})
}
