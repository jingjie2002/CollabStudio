package controllers

import (
	"collab-server/config"
	"collab-server/database"
	"collab-server/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInput 定义前端传过来的数据格式
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 处理注册请求
func Register(c *gin.Context) {
	var input RegisterInput

	// 1. 检查前端传的数据对不对
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. 密码加密
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	// 3. 创建用户对象
	user := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Role:     "user",
	}

	// 4. 存入数据库
	// 注意：这里假设你的 Username 字段在数据库是 UNIQUE 的，如果重复会报错
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在或数据库错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功，请登录", "userId": user.ID})
}

// LoginInput 登录的数据格式
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 处理登录请求 (严格模式：不自动注册)
func Login(c *gin.Context) {
	var input LoginInput
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 查找用户
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	// 2. 比对密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 3. 生成 JWT Token
	// ==========================================================================
	// JWT (JSON Web Token) 是一种无状态的身份验证方案。
	// 它由三部分组成：Header.Payload.Signature
	// - Header: 声明算法 (HS256)
	// - Payload: 存放用户信息 (userId, username, role, exp)
	// - Signature: 用 JWT_SECRET 对前两部分签名，防止篡改
	//
	// 🔐 安全要点：
	// 1. JWT_SECRET 必须足够复杂（至少32字符）
	// 2. 绝对不能硬编码在代码中
	// 3. 生产环境需要定期轮换
	// ==========================================================================
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	})

	// 🔐 核心安全点：JWT_SECRET 不再有默认值
	// 这强制运维人员必须配置 .env 文件，否则 main.go 会拒绝启动
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token 生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    tokenString,
		"username": user.Username,
		"userId":   user.ID,
	})
}
