package middleware

import (
	"collab-server/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware 是一个 Gin 中间件，用于校验 JWT Token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 如果 Header 为空，尝试从 Query 参数获取 (用于 WebSocket 连接)
			authHeader = c.Query("token")
		}

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未携带访问令牌"})
			c.Abort()
			return
		}

		// 2. 解析 Token (支持 Bearer 格式)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 校验签名方法，只接受 HMAC 算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// 🔐 使用环境变量中的密钥，不再有默认值
			return []byte(config.GetEnv("JWT_SECRET", "")), nil
		})

		// 3. 校验 Token 是否有效
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌已过期或无效"})
			c.Abort()
			return
		}

		// 4. 将用户信息存入上下文，方便后续逻辑使用
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			c.Set("userId", claims["userId"])
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])
		}

		c.Next()
	}
}
