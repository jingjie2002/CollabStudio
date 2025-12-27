package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadImage 处理图片上传
func UploadImage(c *gin.Context) {
	// 1. 获取文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}

	// 2. 简单的文件校验 (可选: 检查后缀名)
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持 jpg, png, gif 格式"})
		return
	}

	// 3. 准备保存目录
	uploadPath := "./uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.Mkdir(uploadPath, 0755) // 如果目录不存在则创建
	}

	// 4. 生成唯一文件名 (时间戳 + 原始文件名) 以防覆盖
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadPath, filename)

	// 5. 保存文件到磁盘
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 6. 返回可访问的 URL
	// 注意: 这里的 host 最好根据实际情况动态获取，目前先用相对路径或固定 localhost
	protocol := "http://"
	if c.Request.TLS != nil {
		protocol = "https://"
	}
	host := c.Request.Host
	fileURL := fmt.Sprintf("%s%s/uploads/%s", protocol, host, filename)

	c.JSON(http.StatusOK, gin.H{
		"url": fileURL,
	})
}
