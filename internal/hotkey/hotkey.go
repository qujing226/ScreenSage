package hotkey

import (
	"log"

	"golang.design/x/hotkey"
)

// KeyCallback 定义热键触发时的回调函数类型
type KeyCallback func()

// RegisterHotkey 注册全局热键 Ctrl+Shift+Q
func RegisterHotkey(callback KeyCallback) error {
	// 创建热键组合 Ctrl+Shift+Q
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyQ)

	// 注册热键
	err := hk.Register()
	if err != nil {
		return err
	}

	// 启动监听循环
	go func() {
		for {
			// 等待热键按下事件
			<-hk.Keydown()
			// 热键被触发，执行回调
			callback()
			// 等待热键释放，准备下一次触发
			<-hk.Keyup()
		}
	}()

	log.Println("已注册全局热键: Ctrl+Shift+Q")
	return nil
}
