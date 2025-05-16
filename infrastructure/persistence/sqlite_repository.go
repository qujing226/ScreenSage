package persistence

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/qujing226/screen_sage/domain/model"
	"github.com/qujing226/screen_sage/internal/config"
)

// SQLiteRepository SQLite实现的截图仓库
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository 创建一个新的SQLite仓库
func NewSQLiteRepository() (*SQLiteRepository, error) {
	// 获取配置
	cfg := config.GetConfig()

	// 确保数据库目录存在
	if err := config.EnsureDBPath(); err != nil {
		return nil, fmt.Errorf("确保数据库目录存在失败: %v", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 创建仓库
	repo := &SQLiteRepository{db: db}

	// 初始化数据库表
	if err := repo.initTables(); err != nil {
		return nil, err
	}

	return repo, nil
}

// Close 关闭数据库连接
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// initTables 初始化数据库表
func (r *SQLiteRepository) initTables() error {
	// 创建历史记录表
	query := `
	CREATE TABLE IF NOT EXISTS screenshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		image_path TEXT NOT NULL,
		thumbnail TEXT NOT NULL,
		text TEXT NOT NULL,
		answer TEXT NOT NULL
	);
	`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("创建截图表失败: %v", err)
	}

	log.Println("数据库表初始化成功")
	return nil
}

// Save 保存截图
func (r *SQLiteRepository) Save(screenshot *model.Screenshot) (int64, error) {
	// 准备SQL语句
	query := `
	INSERT INTO screenshots (timestamp, image_path, thumbnail, text, answer)
	VALUES (?, ?, ?, ?, ?);
	`

	// 执行插入
	result, err := r.db.Exec(
		query,
		screenshot.Timestamp,
		screenshot.ImagePath,
		screenshot.Thumbnail,
		screenshot.Text,
		screenshot.Answer,
	)
	if err != nil {
		return 0, fmt.Errorf("插入截图记录失败: %v", err)
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取插入ID失败: %v", err)
	}

	return id, nil
}

// FindByID 根据ID查找截图
func (r *SQLiteRepository) FindByID(id int64) (*model.Screenshot, error) {
	// 准备SQL语句
	query := `
	SELECT id, timestamp, image_path, thumbnail, text, answer
	FROM screenshots
	WHERE id = ?;
	`

	// 执行查询
	row := r.db.QueryRow(query, id)

	// 解析结果
	screenshot := &model.Screenshot{}
	err := row.Scan(
		&screenshot.ID,
		&screenshot.Timestamp,
		&screenshot.ImagePath,
		&screenshot.Thumbnail,
		&screenshot.Text,
		&screenshot.Answer,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到ID为%d的截图", id)
		}
		return nil, fmt.Errorf("查询截图失败: %v", err)
	}

	return screenshot, nil
}

// FindRecent 查找最近的截图记录
func (r *SQLiteRepository) FindRecent(limit int) ([]*model.Screenshot, error) {
	// 准备SQL语句
	query := `
	SELECT id, timestamp, image_path, thumbnail, text, answer
	FROM screenshots
	ORDER BY timestamp DESC
	LIMIT ?;
	`

	// 执行查询
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("查询最近截图失败: %v", err)
	}
	defer rows.Close()

	// 解析结果
	var screenshots []*model.Screenshot
	for rows.Next() {
		screenshot := &model.Screenshot{}
		err := rows.Scan(
			&screenshot.ID,
			&screenshot.Timestamp,
			&screenshot.ImagePath,
			&screenshot.Thumbnail,
			&screenshot.Text,
			&screenshot.Answer,
		)
		if err != nil {
			return nil, fmt.Errorf("解析截图记录失败: %v", err)
		}
		screenshots = append(screenshots, screenshot)
	}

	return screenshots, nil
}

// Delete 删除截图
func (r *SQLiteRepository) Delete(id int64) error {
	// 准备SQL语句
	query := `DELETE FROM screenshots WHERE id = ?;`

	// 执行删除
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除截图失败: %v", err)
	}

	return nil
}
