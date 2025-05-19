package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/qujing226/screen_sage/internal/config"
)

const (
	BaseURL = "https://api.deepseek.com/v1/chat/completions"
)

// SimpleAIProvider 是一个简单的AI服务提供者实现
type SimpleAIProvider struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewSimpleAIProvider 创建一个新的简单AI提供者
func NewSimpleAIProvider() *SimpleAIProvider {
	cfg := config.GetConfig()
	apiKey := cfg.DeepSeekAPIKey
	return &SimpleAIProvider{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

// GenerateAnswer 实现AIProvider接口，根据文本生成回答
func (p *SimpleAIProvider) GenerateAnswer(text string) (string, error) {
	// 获取API密钥
	cfg := config.GetConfig()
	apiKey := cfg.DeepSeekAPIKey

	if apiKey == "" {
		return "AI服务未配置API密钥，请在应用托盘设置中配置。", nil
	}

	// 这里应该调用实际的AI API
	// 由于这只是一个示例，我们返回一个简单的回答
	if strings.TrimSpace(text) == "" {
		return "未能识别到文本内容", nil
	}

	res, err := p.doRequest(text)
	if err != nil {
		return "", err
	}
	// 在实际实现中，这里应该调用DeepSeek或其他AI API
	return fmt.Sprintf("回答: %s", res), nil
}

// doRequest performs an HTTP request to the DeepSeek API
func (p *SimpleAIProvider) doRequest(body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest("POST", BaseURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
