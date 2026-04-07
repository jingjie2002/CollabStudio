package controllers

import (
	"bufio"
	"bytes"
	"collab-server/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
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

func parseAllowedAIHosts() map[string]struct{} {
	raw := config.GetEnv("AI_ALLOWED_HOSTS", "api.deepseek.com,api.openai.com,dashscope.aliyuncs.com")
	allowed := map[string]struct{}{}
	for _, host := range strings.Split(raw, ",") {
		host = strings.TrimSpace(strings.ToLower(host))
		if host != "" {
			allowed[host] = struct{}{}
		}
	}
	return allowed
}

func normalizeAIBaseURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("AI API URL 非法")
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("AI API URL 不完整")
	}

	if !strings.EqualFold(u.Scheme, "https") {
		return "", fmt.Errorf("仅允许 HTTPS 的 AI API URL")
	}

	hostname := strings.ToLower(u.Hostname())
	if hostname == "" {
		return "", fmt.Errorf("AI API URL Host 非法")
	}

	if hostname == "localhost" || strings.HasSuffix(hostname, ".local") || strings.HasSuffix(hostname, ".internal") {
		return "", fmt.Errorf("禁止访问本地或内网主机")
	}
	if ip := net.ParseIP(hostname); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return "", fmt.Errorf("禁止访问本地或内网 IP")
		}
	}

	if _, ok := parseAllowedAIHosts()[hostname]; !ok {
		return "", fmt.Errorf("该 AI API Host 不在允许列表")
	}

	basePath := strings.TrimRight(u.Path, "/")
	if basePath == "" {
		basePath = "/v1"
	}

	return fmt.Sprintf("%s://%s%s", strings.ToLower(u.Scheme), u.Host, basePath), nil
}

// AIChat 处理 AI 对话请求（SSE 流式输出）
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
		apiUrl = strings.TrimRight(apiUrl, "/")
	} else {
		normalized, normalizeErr := normalizeAIBaseURL(apiUrl)
		if normalizeErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": normalizeErr.Error()})
			return
		}
		apiUrl = normalized
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

	// 构建 OpenAI 兼容请求（流式）
	openAIReq := openAIRequest{
		Model:    model,
		Messages: req.Messages,
		Stream:   true,
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
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("X-Accel-Buffering", "no")

	// 配置 HTTP 客户端：必须继承系统代理环境 (http.ProxyFromEnvironment)
	// 这样当用户开启科学上网代理时，后端的请求才能透传
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	client := &http.Client{
		Timeout:   120 * time.Second,
		Transport: transport,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("❌ [AI] API 请求失败: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI 服务连接失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("❌ [AI] API 返回异常 (%d): %s", resp.StatusCode, string(respBody))
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("AI 服务返回错误 (%d)", resp.StatusCode), "detail": string(respBody)})
		return
	}

	// SSE 流式输出到前端
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, _ := c.Writer.(http.Flusher)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			break
		}

		// 解析 streaming chunk
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(data), &chunk) == nil && len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				chunkJSON, _ := json.Marshal(map[string]string{"content": content})
				fmt.Fprintf(c.Writer, "data: %s\n\n", chunkJSON)
				if flusher != nil {
					flusher.Flush()
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("⚠️ [AI] SSE 读取异常: %v", err)
	}
}
