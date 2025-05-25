package model

import (
	"time"
)

// Screenshot 表示一个截图实体
type Screenshot struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ImagePath string    `json:"image_path"`
	Thumbnail string    `json:"thumbnail"`
	Text      string    `json:"text"`
	Answer    string    `json:"answer"`
	Title     string    `json:"title"`
}

// NewScreenshot 创建一个新的截图实体
func NewScreenshot(imagePath, thumbnail, text, answer, title string) *Screenshot {
	return &Screenshot{
		Timestamp: time.Now(),
		ImagePath: imagePath,
		Thumbnail: thumbnail,
		Text:      text,
		Answer:    answer,
		Title:     title,
	}
}
