package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// HistoryRecord 表示一条历史记录
type HistoryRecord struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ImagePath string    `json:"image_path"`
	Thumbnail string    `json:"thumbnail"`
	Text      string    `json:"text"`
	Answer    string    `json:"answer"`
}

// DBManager 数据库管理器
type DBManager struct {
	db *sql.DB
}

// NewDBManager 创建一个新的数据库管理器
func NewDBManager(dbPath string) (*DBManager, error) {
	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 创建管理器
	manager := &DBManager{db: db}

	// 初始化数据库表
	if err := manager.initTables(); err != nil {
		return nil, err
	}

	return manager, nil
}

// Close 关闭数据库连接
func (m *DBManager) Close() error {
	return m.db.Close()
}

// initTables 初始化数据库表
func (m *DBManager) initTables() error {
	// 创建历史记录表
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		image_path TEXT NOT NULL,
		thumbnail TEXT NOT NULL,
		text TEXT NOT NULL,
		answer TEXT NOT NULL
	);
	`

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("创建历史记录表失败: %v", err)
	}

	log.Println("数据库表初始化成功")
	return nil
}

// AddHistory 添加一条历史记录
func (m *DBManager) AddHistory(record *HistoryRecord) (int64, error) {
	// 准备SQL语句
	query := `
	INSERT INTO history (timestamp, image_path, thumbnail, text, answer)
	VALUES (?, ?, ?, ?, ?);
	`

	// 执行插入
	result, err := m.db.Exec(
		query,
		record.Timestamp,
		record.ImagePath,
		record.Thumbnail,
		record.Text,
		record.Answer,
	)
	if err != nil {
		return 0, fmt.Errorf("插入历史记录失败: %v", err)
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取插入ID失败: %v", err)
	}

	return id, nil
}

// GetHistory 获取历史记录列表
func (m *DBManager) GetHistory(limit int) ([]HistoryRecord, error) {
	// 准备SQL语句
	query := `
	SELECT id, timestamp, image_path, thumbnail, text, answer
	FROM history
	ORDER BY timestamp DESC
	LIMIT ?;
	`

	// 执行查询
	rows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("查询历史记录失败: %v", err)
	}
	defer rows.Close()

	// 解析结果
	var records []HistoryRecord
	for rows.Next() {
		var record HistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.Timestamp,
			&record.ImagePath,
			&record.Thumbnail,
			&record.Text,
			&record.Answer,
		); err != nil {
			return nil, fmt.Errorf("解析历史记录失败: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// GetHistoryByID 根据ID获取历史记录
func (m *DBManager) GetHistoryByID(id int64) (*HistoryRecord, error) {
	// 准备SQL语句
	query := `
	SELECT id, timestamp, image_path, thumbnail, text, answer
	FROM history
	WHERE id = ?;
	`

	// 执行查询
	row := m.db.QueryRow(query, id)

	// 解析结果
	var record HistoryRecord
	if err := row.Scan(
		&record.ID,
		&record.Timestamp,
		&record.ImagePath,
		&record.Thumbnail,
		&record.Text,
		&record.Answer,
	); err != nil {
		return nil, fmt.Errorf("获取历史记录失败: %v", err)
	}

	return &record, nil
}