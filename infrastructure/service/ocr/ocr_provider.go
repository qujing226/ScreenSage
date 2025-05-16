package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/qujing226/screen_sage/internal/config"
)

// DeepSeekOCRProvider 是DeepSeek OCR API的客户端
type DeepSeekOCRProvider struct {
	HTTPClient  *http.Client
	APIEndpoint string
}

// OCRRequest 表示DeepSeek OCR请求的结构
type OCRRequest struct {
	Image string `json:"image"` // Base64编码的图像
}

// OCRResponse 表示DeepSeek OCR响应的结构
type OCRResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

// NewDeepSeekOCRProvider 创建一个新的DeepSeek OCR提供者
func NewDeepSeekOCRProvider() *DeepSeekOCRProvider {
	return &DeepSeekOCRProvider{
		APIEndpoint: "https://api.deepseek.com/v1/ocr",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RecognizeText 实现OCRProvider接口，识别图像中的文本
func (p *DeepSeekOCRProvider) RecognizeText(imageBase64 string) (string, error) {
	// 获取API密钥
	cfg := config.GetConfig()
	apiKey := cfg.DeepSeekAPIKey

	if apiKey == "" {
		return "", fmt.Errorf("DeepSeek API密钥未配置")
	}

	// 构建请求体
	reqBody := OCRRequest{
		Image: imageBase64,
	}

	// 序列化请求体
	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", p.APIEndpoint, bytes.NewBuffer(reqData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送OCR请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR API返回错误: %s, 状态码: %d", string(respBody), resp.StatusCode)
	}

	// 解析响应
	var ocrResp OCRResponse
	if err := json.Unmarshal(respBody, &ocrResp); err != nil {
		return "", fmt.Errorf("解析OCR响应失败: %v", err)
	}

	// 检查错误
	if ocrResp.Error != "" {
		return "", fmt.Errorf("OCR API返回错误: %s", ocrResp.Error)
	}

	return ocrResp.Text, nil
}
