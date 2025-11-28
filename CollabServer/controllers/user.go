package controllers

import (
	"collab-server/database"
	"collab-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHistory 获取指定用户的最近 10 条访问记录
func GetHistory(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username required"})
		return
	}

	var histories []models.History
	// 按时间倒序，取前 10 条
	database.DB.Where("username = ?", username).Order("updated_at desc").Limit(10).Find(&histories)

	c.JSON(http.StatusOK, gin.H{"history": histories})
}
