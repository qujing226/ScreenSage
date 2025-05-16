package ai

import (
	"fmt"
	"strings"

	"github.com/qujing226/screen_sage/internal/config"
)

// SimpleAIProvider 是一个简单的AI服务提供者实现
type SimpleAIProvider struct{}

// NewSimpleAIProvider 创建一个新的简单AI提供者
func NewSimpleAIProvider() *SimpleAIProvider {
	return &SimpleAIProvider{}
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

	// 在实际实现中，这里应该调用DeepSeek或其他AI API
	return fmt.Sprintf("识别到的文本内容：\n%s\n\n这是一个AI生成的回答示例。在实际应用中，这里会调用DeepSeek API生成更有意义的回答。", text), nil
}
