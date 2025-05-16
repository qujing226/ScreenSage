package service

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/qujing226/screen_sage/domain/model"
	"github.com/qujing226/screen_sage/domain/repository"
)

// OCRProvider 定义OCR服务提供者接口
type OCRProvider interface {
	RecognizeText(imageBase64 string) (string, error)
}

// AIProvider 定义AI服务提供者接口
type AIProvider interface {
	GenerateAnswer(text string) (string, error)
}

// ScreenshotService 截图服务
type ScreenshotService struct {
	Repository  repository.ScreenshotRepository
	OCRProvider OCRProvider
	AIProvider  AIProvider
}

// NewScreenshotService 创建截图服务
func NewScreenshotService(
	repo repository.ScreenshotRepository,
	ocrProvider OCRProvider,
	aiProvider AIProvider,
) *ScreenshotService {
	return &ScreenshotService{
		Repository:  repo,
		OCRProvider: ocrProvider,
		AIProvider:  aiProvider,
	}
}

// ProcessScreenshot 处理截图
func (s *ScreenshotService) ProcessScreenshot(imgBytes []byte) (*model.Screenshot, error) {
	// 保存图片到文件
	timestamp := time.Now().Format("20060102150405")
	imageDir := filepath.Join(os.Getenv("HOME"), ".screensage", "images")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return nil, fmt.Errorf("创建图片目录失败: %v", err)
	}

	// 生成文件名
	imageFilename := fmt.Sprintf("%s.png", timestamp)
	imagePath := filepath.Join(imageDir, imageFilename)

	// 写入文件
	if err := os.WriteFile(imagePath, imgBytes, 0644); err != nil {
		return nil, fmt.Errorf("保存图片失败: %v", err)
	}

	// 转换为Base64
	imageBase64 := base64.StdEncoding.EncodeToString(imgBytes)

	// OCR识别
	text, err := s.OcrRecognize(imageBase64)
	if err != nil {
		log.Printf("OCR识别失败: %v", err)
		text = "OCR识别失败"
	}

	// AI生成回答
	answer, err := s.GenerateAnswer(text)
	if err != nil {
		log.Printf("生成回答失败: %v", err)
		answer = "无法生成回答"
	}

	// 创建截图实体
	screenshot := model.NewScreenshot(
		imagePath,
		"data:image/png;base64,"+imageBase64, // 缩略图直接使用Base64
		text,
		answer,
	)

	// 保存到仓库
	id, err := s.Repository.Save(screenshot)
	if err != nil {
		return nil, fmt.Errorf("保存截图记录失败: %v", err)
	}

	screenshot.ID = id
	return screenshot, nil
}

// OcrRecognize 执行OCR识别
func (s *ScreenshotService) OcrRecognize(imageBase64 string) (string, error) {
	return s.OCRProvider.RecognizeText(imageBase64)
}

// GenerateAnswer 生成回答
func (s *ScreenshotService) GenerateAnswer(text string) (string, error) {
	return s.AIProvider.GenerateAnswer(text)
}

// GetRecentScreenshots 获取最近的截图
func (s *ScreenshotService) GetRecentScreenshots(limit int) ([]*model.Screenshot, error) {
	return s.Repository.FindRecent(limit)
}
