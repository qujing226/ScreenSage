package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config 表示应用程序配置
type Config struct {
	// API密钥配置
	BaiduAPIKey    string `json:"baidu_api_key"`
	BaiduSecretKey string `json:"baidu_secret_key"`
	DeepSeekAPIKey string `json:"deepseek_api_key"`

	// 数据库配置
	DBPath string `json:"db_path"`

	// 服务器配置
	Port       int    `json:"port"`
	StaticPath string `json:"static_path"`
}

var (
	instance *Config
	once     sync.Once
	mutex    sync.RWMutex
)

// GetConfig 获取配置单例
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			// 默认配置
			Port:       8081,
			StaticPath: "./web/frontend/dist",
			DBPath:     getDefaultDBPath(),
		}
		// 尝试加载配置文件
		loadConfig()
	})

	mutex.RLock()
	defer mutex.RUnlock()
	return instance
}

// 获取默认数据库路径
func getDefaultDBPath() string {
	// 获取当前可执行文件所在目录
	execDir, err := os.Executable()
	if err != nil {
		// 如果无法获取可执行文件路径，使用当前目录
		return "./data/screensage.db"
	}

	// 使用可执行文件所在目录作为基础目录
	baseDir := filepath.Dir(execDir)
	dataDir := filepath.Join(baseDir, "data")
	return filepath.Join(dataDir, "screensage.db")
}

// 获取配置文件路径
func getConfigFilePath() string {
	// 获取当前可执行文件所在目录
	execDir, err := os.Executable()
	if err != nil {
		// 如果无法获取可执行文件路径，使用当前目录
		return "./config.json"
	}

	// 使用可执行文件所在目录作为基础目录
	baseDir := filepath.Dir(execDir)
	configDir := filepath.Join(baseDir, "config")
	return filepath.Join(configDir, "config.json")
}

// 加载配置文件
func loadConfig() {
	configPath := getConfigFilePath()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		saveConfig()
		return
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		return
	}

	// 解析配置文件
	mutex.Lock()
	defer mutex.Unlock()

	err = json.Unmarshal(data, instance)
	if err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
	}
}

// SaveConfig 保存配置到文件
func SaveConfig() error {
	return saveConfig()
}

// 保存配置到文件
func saveConfig() error {
	configPath := getConfigFilePath()

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化配置
	mutex.RLock()
	data, err := json.MarshalIndent(instance, "", "  ")
	mutex.RUnlock()

	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	fmt.Println("配置已保存")
	return nil
}

// UpdateConfig 更新配置
func UpdateConfig(newConfig *Config) {
	mutex.Lock()
	// 更新配置
	if newConfig.BaiduAPIKey != "" {
		instance.BaiduAPIKey = newConfig.BaiduAPIKey
	}
	if newConfig.BaiduSecretKey != "" {
		instance.BaiduSecretKey = newConfig.BaiduSecretKey
	}
	if newConfig.DeepSeekAPIKey != "" {
		instance.DeepSeekAPIKey = newConfig.DeepSeekAPIKey
	}
	if newConfig.DBPath != "" {
		instance.DBPath = newConfig.DBPath
	}
	if newConfig.Port != 0 {
		instance.Port = newConfig.Port
	}
	if newConfig.StaticPath != "" {
		instance.StaticPath = newConfig.StaticPath
	}
	mutex.Unlock()
	// 保存配置
	err := saveConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
}

// EnsureDBPath 确保数据库路径存在
func EnsureDBPath() error {
	mutex.RLock()
	dbPath := instance.DBPath
	mutex.RUnlock()

	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	return os.MkdirAll(dbDir, 0755)
}
