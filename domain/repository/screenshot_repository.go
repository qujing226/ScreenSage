package repository

import (
	"github.com/qujing226/screen_sage/domain/model"
)

// ScreenshotRepository 定义截图仓库接口
type ScreenshotRepository interface {
	// Save 保存截图
	Save(screenshot *model.Screenshot) (int64, error)

	// FindByID 根据ID查找截图
	FindByID(id int64) (*model.Screenshot, error)

	// FindRecent 查找最近的截图记录
	FindRecent(limit int) ([]*model.Screenshot, error)

	// Delete 删除截图
	Delete(id int64) error
}
