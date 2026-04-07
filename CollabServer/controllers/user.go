package controllers

import (
	"collab-server/database"
	"collab-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getAuthUsername(c *gin.Context) (string, bool) {
	usernameRaw, exists := c.Get("username")
	if !exists {
		return "", false
	}

	username, ok := usernameRaw.(string)
	if !ok || username == "" {
		return "", false
	}

	return username, true
}

// GetHistory 获取当前登录用户的最近 10 条访问记录
func GetHistory(c *gin.Context) {
	username, ok := getAuthUsername(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证用户"})
		return
	}

	var histories []models.History
	// 按时间倒序，取前 10 条
	database.DB.Where("username = ?", username).Order("updated_at desc").Limit(10).Find(&histories)

	c.JSON(http.StatusOK, gin.H{"history": histories})
}

// DeleteHistory 删除当前登录用户的指定访问记录
func DeleteHistory(c *gin.Context) {
	username, ok := getAuthUsername(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证用户"})
		return
	}

	id := c.Param("id")
	result := database.DB.Where("id = ? AND username = ?", id, username).Delete(&models.History{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
