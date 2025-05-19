package ocr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OCRClient 是OCR API的通用接口
type OCRClient interface {
	RecognizeText(imageBase64 string) (string, error)
}

// BaiduOCRClient 是百度OCR API的客户端
type BaiduOCRClient struct {
	APIKey      string
	SecretKey   string
	AccessToken string
	TokenExpiry time.Time
	APIEndpoint string
	HTTPClient  *http.Client
}

// BaiduOCRRequest 表示百度OCR请求的结构
type BaiduOCRRequest struct {
	Image string `json:"image"` // Base64编码的图像
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

// BaiduTokenResponse 表示百度访问令牌响应的结构
type BaiduTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrorCode   int    `json:"error_code,omitempty"`
	ErrorMsg    string `json:"error_msg,omitempty"`
}

// DeepSeekOCRClient 是DeepSeek OCR API的客户端
type DeepSeekOCRClient struct {
	APIKey      string
	APIEndpoint string
	HTTPClient  *http.Client
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

// NewBaiduOCRClient 创建一个新的百度OCR客户端
func NewBaiduOCRClient(apiKey, secretKey string) *BaiduOCRClient {
	return &BaiduOCRClient{
		APIKey:      apiKey,
		SecretKey:   secretKey,
		APIEndpoint: "https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccessToken 获取百度API访问令牌
func (c *BaiduOCRClient) GetAccessToken() (string, error) {
	// 如果令牌有效且未过期，直接返回
	if c.AccessToken != "" && time.Now().Before(c.TokenExpiry) {
		return c.AccessToken, nil
	}

	// 构建令牌请求URL
	tokenURL := "https://aip.baidubce.com/oauth/2.0/token"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.APIKey)
	data.Set("client_secret", c.SecretKey)

	// 发送请求
	resp, err := c.HTTPClient.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("获取访问令牌失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
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
	if tokenResp.ErrorCode != 0 {
		return "", fmt.Errorf("获取令牌错误: %s (代码: %d)", tokenResp.ErrorMsg, tokenResp.ErrorCode)
	}

	// 保存令牌和过期时间（提前5分钟过期，以确保安全）
	c.AccessToken = tokenResp.AccessToken
	c.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return c.AccessToken, nil
}

// RecognizeText 使用百度OCR API从图像中识别文本
func (c *BaiduOCRClient) RecognizeText(imageBase64 string) (string, error) {
	// 获取访问令牌
	token, err := c.GetAccessToken()
	if err != nil {
		return "", err
	}

	// 如果图像数据包含前缀（如data:image/jpeg;base64,），则去除
	if idx := strings.Index(imageBase64, ","); idx != -1 {
		imageBase64 = imageBase64[idx+1:]
	}

	// 准备请求数据
	data := url.Values{}
	data.Set("image", imageBase64)

	// 创建HTTP请求
	reqURL := fmt.Sprintf("%s?access_token=%s", c.APIEndpoint, token)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送OCR请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var ocrResp BaiduOCRResponse
	if err := json.Unmarshal(respBody, &ocrResp); err != nil {
		return "", fmt.Errorf("解析OCR响应失败: %v", err)
	}

	// 检查API错误
	if ocrResp.ErrorCode != 0 {
		return "", fmt.Errorf("OCR API返回错误: %s (代码: %d)", ocrResp.ErrorMsg, ocrResp.ErrorCode)
	}

	// 合并识别结果
	var result strings.Builder
	for i, word := range ocrResp.WordsResult {
		result.WriteString(word.Words)
		if i < len(ocrResp.WordsResult)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// NewDeepSeekOCRClient 创建一个新的DeepSeek OCR客户端
func NewDeepSeekOCRClient(apiKey string) *DeepSeekOCRClient {
	return &DeepSeekOCRClient{
		APIKey:      apiKey,
		APIEndpoint: "https://api.deepseek.com/v1/ocr", // 假设的API端点
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RecognizeText 从图像中识别文本
func (c *DeepSeekOCRClient) RecognizeText(imageBase64 string) (string, error) {
	// 准备请求数据
	reqData := OCRRequest{
		Image: imageBase64,
	}

	// 将请求数据转换为JSON
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("编码请求数据失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.APIEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送OCR请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var ocrResp OCRResponse
	if err := json.Unmarshal(respBody, &ocrResp); err != nil {
		return "", fmt.Errorf("解析OCR响应失败: %v", err)
	}

	// 检查API错误
	if ocrResp.Error != "" {
		return "", fmt.Errorf("OCR API返回错误: %s", ocrResp.Error)
	}

	return ocrResp.Text, nil
}

// MockRecognizeText 模拟OCR识别，用于开发测试
func MockRecognizeText(imageBase64 string) (string, error) {
	// 返回模拟的OCR结果
	return "这是一个模拟的OCR识别结果，用于开发测试。实际使用时请替换为真实的API调用。", nil
}

// ProcessWithDeepSeek 将OCR识别的文本发送给DeepSeek进行处理
// 此函数负责将OCR识别的文本发送到DeepSeek API进行分析和处理
// 参数:
//   - text: OCR识别出的文本内容
//   - apiKey: DeepSeek API密钥
//
// 返回:
//   - 处理后的分析结果
//   - 错误信息（如果有）
func ProcessWithDeepSeek(text string, apiKey string) (string, error) {
	// 检查输入参数
	if text == "" {
		return "", fmt.Errorf("OCR文本为空，无法处理")
	}

	if apiKey == "" {
		return "", fmt.Errorf("DeepSeek API密钥未提供")
	}

	// DeepSeek API端点
	endpoint := "https://api.deepseek.com/v1/chat/completions"

	// 构建专门的提示词模板
	systemPrompt := "你是一个专业的屏幕内容分析助手。以下是通过OCR技术从屏幕截图中识别出的文本内容。请分析这些文本，找出其中包含的问题或关键信息，然后给出清晰、准确的回答或解释。如果文本中包含代码或错误信息，请特别关注并提供相关的解决方案。"

	userPrompt := fmt.Sprintf("以下是从屏幕截图中识别出的文本内容：\n\n%s\n\n请分析上述内容，找出其中的问题或关键信息，并给出专业的回答。", text)

	// 准备请求数据
	reqData := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
		"max_tokens":  2000, // 设置最大输出长度
	}

	// 将请求数据转换为JSON
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("编码请求数据失败: %v", err)
	}

	// 创建HTTP客户端
	client := &http.Client{Timeout: 60 * time.Second}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查API错误
	if response.Error.Message != "" {
		return "", fmt.Errorf("API返回错误: %s", response.Error.Message)
	}

	// 返回处理结果
	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("API未返回有效响应")
}

// ImageToBase64 将图像文件转换为Base64编码的字符串
func ImageToBase64(imageBytes []byte) string {
	return base64.StdEncoding.EncodeToString(imageBytes)
}
