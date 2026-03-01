package middleware

import (
	"collab-server/config"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth 是 JWT 鉴权中间件
// 用于保护需要登录才能访问的接口（如 /upload, /api/history 等）
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取 Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授权令牌"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求头格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 2. 解析并验证 Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 校验使用的签名算法是否是我们指定的 HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("非法的签名算法: %v", token.Header["alg"])
			}
			secret := config.GetEnv("JWT_SECRET", "")
			if secret == "" {
				return nil, fmt.Errorf("服务器未配置密钥")
			}
			return []byte(secret), nil
		})

		// 3. 提取信息并向后传递
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 将解析出的 user 信息存入上下文，供后续的 Controller 使用
			c.Set("userId", claims["userId"])
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无法解析 Token 详情"})
			c.Abort()
			return
		}
	}
}
