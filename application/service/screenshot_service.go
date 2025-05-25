package service

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/qujing226/screen_sage/internal/ocr"
	"github.com/qujing226/screen_sage/internal/storage"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qujing226/screen_sage/domain/model"
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
	Db          *storage.DBManager
	OCRProvider OCRProvider
	//AIProvider  AIProvider
	DeepseekKey string
}

// NewScreenshotService 创建截图服务
func NewScreenshotService(
	db *storage.DBManager,
	ocrProvider OCRProvider,
	deepseekKey string,
) *ScreenshotService {
	return &ScreenshotService{
		Db:          db,
		OCRProvider: ocrProvider,
		DeepseekKey: deepseekKey,
	}
}

// ProcessScreenshot 处理截图
func (s *ScreenshotService) ProcessScreenshot(imgBytes []byte) (*model.Screenshot, error) {
	// 保存图片到文件
	timestamp := time.Now().Format("20060102150405")

	// 获取当前可执行文件所在目录
	execDir, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 使用可执行文件所在目录作为基础目录
	baseDir := filepath.Dir(execDir)
	imageDir := filepath.Join(baseDir, "images")
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
	answer, err := ocr.ProcessWithDeepSeek(text, s.DeepseekKey)
	if err != nil {
		log.Printf("生成回答失败: %v", err)
		answer = "无法生成回答"
	}

	// 获取最后一行作为标题
	title := strings.TrimSpace(strings.Split(answer, "\n")[len(strings.Split(answer, "\n"))-1])
	// 剩余内容作为回答回答正文
	answer = strings.Join(strings.Split(answer, "\n")[:len(strings.Split(answer, "\n"))-2], "\n")
	// 创建截图实体
	screenshot := model.NewScreenshot(
		imagePath,
		"data:image/png;base64,"+imageBase64, // 缩略图直接使用Base64
		text,
		answer,
		title,
	)

	// 保存到仓库
	id, err := s.Db.AddHistory(&storage.HistoryRecord{
		Timestamp: screenshot.Timestamp,
		ImagePath: screenshot.ImagePath,
		Thumbnail: screenshot.Thumbnail,
		Text:      screenshot.Text,
		Answer:    screenshot.Answer,
		Title: sql.NullString{
			String: screenshot.Title,
			Valid:  screenshot.Title != "",
		},
	})
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

// GetRecentScreenshots 获取最近的截图
func (s *ScreenshotService) GetRecentScreenshots(limit int) ([]*model.Screenshot, error) {
	res, err := s.Db.GetHistory(limit)
	if err != nil {
		return nil, fmt.Errorf("获取最近截图失败: %v", err)
	}
	screenshots := make([]*model.Screenshot, len(res))
	for i, record := range res {
		screenshots[i] = model.NewScreenshot(
			record.ImagePath,
			record.Thumbnail,
			record.Text,
			record.Answer,
			record.Title.String,
		)
	}
	return screenshots, nil
}
