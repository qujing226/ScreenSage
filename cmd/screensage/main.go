package main

import (
	"fmt"
	"github.com/qujing226/screen_sage/internal/storage"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/qujing226/screen_sage/application/service"
	"github.com/qujing226/screen_sage/infrastructure/service/ocr"
	"github.com/qujing226/screen_sage/infrastructure/ui"
	"github.com/qujing226/screen_sage/internal/config"
	"github.com/qujing226/screen_sage/internal/hotkey"
	"github.com/qujing226/screen_sage/internal/screenshot"
	"github.com/qujing226/screen_sage/web/api"
)

// 全局服务实例
var (
	serverInstance    *api.Server
	serverMutex       sync.Mutex
	screenshotService *service.ScreenshotService
)

// 获取服务器实例
func getServerInstance() *api.Server {
	serverMutex.Lock()
	defer serverMutex.Unlock()
	return serverInstance
}

// 设置服务器实例
func setServerInstance(server *api.Server) {
	serverMutex.Lock()
	defer serverMutex.Unlock()
	serverInstance = server
}

func main() {
	// 设置日志
	log.SetOutput(os.Stdout)
	log.SetPrefix("[ScreenSage] ")

	// 初始化配置
	cfg := config.GetConfig()

	// 确保数据库目录存在
	if err := config.EnsureDBPath(); err != nil {
		log.Fatalf("确保数据库目录存在失败: %v", err)
	}

	// 使用配置中的API密钥

	// 初始化仓库
	dbManager, err := storage.NewDBManager(cfg.DBPath)
	if err != nil {
		log.Fatalf("初始化数据库仓库失败: %v", err)
	}
	defer dbManager.Close()

	// 初始化OCR提供者
	ocrProvider := ocr.NewBaiduOCRProvider(cfg.BaiduAPIKey, cfg.BaiduSecretKey)

	deepseekKey := cfg.DeepSeekAPIKey
	// 初始化截图服务
	screenshotService = service.NewScreenshotService(
		dbManager,
		ocrProvider,
		deepseekKey,
	)

	// 启动系统托盘
	go systray.Run(onReady, onExit)

	// 启动Web服务
	server, err := api.StartServer()
	if err != nil {
		log.Fatalf("启动Web服务失败: %v", err)
	}

	// 保存服务器实例
	setServerInstance(server)

	// 等待服务器完全初始化
	time.Sleep(500 * time.Millisecond)

	// 等待信号，用于系统杀死进程时的退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func onReady() {
	// 加载应用图标
	// 使用绝对路径加载资源（适用于 Windows）
	execDir, err := os.Executable()
	if err != nil {
		log.Printf("获取可执行文件路径失败: %v", err)
	}

	// 使用可执行文件所在目录作为基础目录
	baseDir := filepath.Dir(execDir)
	iconPath := filepath.Join(baseDir, "/src/screen_sage.ico")

	// 尝试读取图标文件
	iconData, err := os.ReadFile(iconPath)
	if err != nil {
		log.Printf("加载应用图标失败: %v", err)
		// 尝试从项目根目录读取
		iconPath = filepath.Join(baseDir, "../..", "screen_sage.png")
		iconData, err = os.ReadFile(iconPath)
		if err != nil {
			log.Printf("从备用路径加载应用图标失败: %v", err)
		}
	}

	if iconData != nil {
		systray.SetIcon(iconData)
	}

	// 设置系统托盘图标和菜单
	systray.SetTitle("ScreenSage")
	systray.SetTooltip("屏幕识图+智能问答助手")

	// 添加菜单项
	mHistory := systray.AddMenuItem("历史记录", "查看历史记录")
	mSettings := systray.AddMenuItem("设置", "配置应用")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出应用")

	// 注册全局热键
	if err := hotkey.RegisterHotkey(func() {
		processScreenshot()
	}); err != nil {
		log.Printf("注册热键失败: %v", err)
	}

	// 处理菜单事件
	go func() {
		for {
			select {
			case <-mHistory.ClickedCh:
				// 打开历史记录页面
				fmt.Println("打开历史记录")
			case <-mSettings.ClickedCh:
				// 打开设置页面
				showSettingsDialog()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// 清理资源
	log.Println("应用退出")
	os.Exit(0)
}

// 显示设置对话框
func showSettingsDialog() {
	// 调用UI包中的设置对话框
	ui.ShowSettingsDialog()
}

// 处理截图
func processScreenshot() {
	// 捕获屏幕
	imgBytes, err := screenshot.CaptureScreen()
	if err != nil {
		log.Printf("截图失败: %v", err)
		return
	}

	// 处理截图
	log.Printf("截图成功，大小: %d bytes", len(imgBytes))

	// 使用截图服务处理
	if screenshotService != nil {
		go func() {
			screen, err := screenshotService.ProcessScreenshot(imgBytes)
			if err != nil {
				log.Printf("处理截图失败: %v", err)
				return
			}

			// 广播到客户端
			fmt.Printf("正在将 %+v 广播到客户端\n", screen.Title)
			server := getServerInstance()
			if server != nil {
				server.BroadcastScreenshot(screen)
			}
		}()
	} else {
		log.Printf("截图服务未初始化，无法处理截图")
	}
}
