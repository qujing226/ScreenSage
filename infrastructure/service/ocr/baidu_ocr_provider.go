package ocr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BaiduOCRProvider 是百度OCR API的客户端
type BaiduOCRProvider struct {
	HTTPClient  *http.Client
	APIKey      string
	SecretKey   string
	APIEndpoint string
	TokenURL    string
	AccessToken string
	ExpiresAt   time.Time
}

// BaiduTokenResponse 表示百度API令牌响应的结构
type BaiduTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// BaiduOCRResponse 表示百度OCR响应的结构
type BaiduOCRResponse struct {
	WordsResult []struct {
		Words string `json:"words"`
	} `json:"words_result"`
	WordsResultNum int    `json:"words_result_num"`
	ErrorCode      int    `json:"error_code,omitempty"`
	ErrorMsg       string `json:"error_msg,omitempty"`
}

// NewBaiduOCRProvider 创建一个新的百度OCR提供者
func NewBaiduOCRProvider(apiKey, secretKey string) *BaiduOCRProvider {
	return &BaiduOCRProvider{
		APIKey:      apiKey,
		SecretKey:   secretKey,
		APIEndpoint: "https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic", // 通用文字识别标准版API
		TokenURL:    "https://aip.baidubce.com/oauth/2.0/token",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getAccessToken 获取百度API访问令牌
func (p *BaiduOCRProvider) getAccessToken() (string, error) {
	// 如果令牌有效且未过期，直接返回
	if p.AccessToken != "" && time.Now().Before(p.ExpiresAt) {
		return p.AccessToken, nil
	}
	// 构建请求参数
	u := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?client_id=%s&client_secret=%s&grant_type=client_credentials", p.APIKey, p.SecretKey)
	payload := strings.NewReader(``)
	client := &http.Client{}
	req, err := http.NewRequest("POST", u, payload)
	if err != nil {
		return "", fmt.Errorf("创建令牌请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送令牌请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取令牌响应失败: %v", err)
	}

	// 解析响应
	var tokenResp BaiduTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("解析令牌响应失败: %v", err)
	}

	// 检查错误
	if tokenResp.Error != "" {
		return "", fmt.Errorf("获取令牌失败: %s - %s", tokenResp.Error, tokenResp.ErrorDescription)
	}

	// 保存令牌和过期时间
	p.AccessToken = tokenResp.AccessToken
	p.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // 提前60秒过期

	return p.AccessToken, nil
}

// RecognizeText 实现OCRProvider接口，识别图像中的文本
func (p *BaiduOCRProvider) RecognizeText(imageBase64 string) (string, error) {
	// 获取访问令牌
	token, err := p.getAccessToken()
	if err != nil {
		return "", err
	}

	// 构建请求URL
	requestURL := fmt.Sprintf("%s?access_token=%s", p.APIEndpoint, token)

	// 构建请求参数
	data := url.Values{}
	// 确保图片数据是正确的Base64格式，不需要额外的URL编码
	data.Set("image", imageBase64)
	// 添加可选参数
	data.Set("language_type", "CHN_ENG") // 中英文混合识别

	// 创建HTTP请求
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建OCR请求失败: %v", err)
	}

	// 设置请求头 - 确保使用正确的Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送OCR请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取OCR响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR API返回错误: %s, 状态码: %d", string(respBody), resp.StatusCode)
	}

	// 解析响应
	var ocrResp BaiduOCRResponse
	if err := json.Unmarshal(respBody, &ocrResp); err != nil {
		return "", fmt.Errorf("解析OCR响应失败: %v", err)
	}

	// 检查错误
	if ocrResp.ErrorCode != 0 {
		return "", fmt.Errorf("OCR API返回错误: %s (错误码: %d)", ocrResp.ErrorMsg, ocrResp.ErrorCode)
	}

	// 合并识别结果
	var textBuilder strings.Builder
	for i, result := range ocrResp.WordsResult {
		if i > 0 {
			textBuilder.WriteString("\n")
		}
		textBuilder.WriteString(result.Words)
	}

	return textBuilder.String(), nil
}
