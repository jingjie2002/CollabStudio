package controllers

import (
	"bytes"
	"collab-server/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// AI 代理控制器
// =============================================================================
// 将前端的 AI 请求转发到 OpenAI 兼容的 LLM API
// 支持 DeepSeek、通义千问、OpenAI 等所有兼容接口
//
// API Key 优先级：
// 1. 请求中的 apiKey（用户自带 Key）
// 2. .env 中的 AI_API_KEY（服务端共享 Key）
// =============================================================================

// AIChatRequest 前端发来的 AI 请求
type AIChatRequest struct {
	Messages []AIMessage `json:"messages"`
	// 用户自带的 API 配置（可选，覆盖服务端配置）
	APIUrl string `json:"apiUrl,omitempty"`
	APIKey string `json:"apiKey,omitempty"`
	Model  string `json:"model,omitempty"`
}

// AIMessage 单条消息
type AIMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// OpenAI 兼容格式的请求体
type openAIRequest struct {
	Model    string      `json:"model"`
	Messages []AIMessage `json:"messages"`
	Stream   bool        `json:"stream"`
}

// OpenAI 兼容格式的响应体
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AIChat 处理 AI 对话请求
func AIChat(c *gin.Context) {
	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息不能为空"})
		return
	}

	// 确定 API 配置（用户配置 > 服务端配置）
	apiUrl := req.APIUrl
	if apiUrl == "" {
		apiUrl = config.GetEnv("AI_API_URL", "https://api.deepseek.com/v1")
	}

	apiKey := req.APIKey
	if apiKey == "" {
		apiKey = config.GetEnv("AI_API_KEY", "")
	}

	model := req.Model
	if model == "" {
		model = config.GetEnv("AI_MODEL", "deepseek-chat")
	}

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未配置 AI API Key，请在设置面板中填入"})
		return
	}

	// 构建 OpenAI 兼容请求
	openAIReq := openAIRequest{
		Model:    model,
		Messages: req.Messages,
		Stream:   false,
	}

	body, _ := json.Marshal(openAIReq)

	// 发送到 LLM API
	chatEndpoint := apiUrl + "/chat/completions"
	httpReq, err := http.NewRequest("POST", chatEndpoint, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("构建请求失败: %v", err)})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("❌ [AI] API 请求失败: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI 服务连接失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [AI] API 返回异常 (%d): %s", resp.StatusCode, string(respBody))
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("AI 服务返回错误 (%d)", resp.StatusCode), "detail": string(respBody)})
		return
	}

	// 解析响应
	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析 AI 响应失败"})
		return
	}

	if openAIResp.Error != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": openAIResp.Error.Message})
		return
	}

	if len(openAIResp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 未返回有效回复"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": openAIResp.Choices[0].Message.Content,
	})
}
