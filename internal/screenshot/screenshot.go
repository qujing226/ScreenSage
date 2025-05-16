package screenshot

import (
	"bytes"
	"fmt"
	"image/png"

	"github.com/kbinani/screenshot"
)

// CaptureScreen 捕获当前屏幕并返回PNG格式的图像数据
func CaptureScreen() ([]byte, error) {
	// 获取活跃显示器数量
	n := screenshot.NumActiveDisplays()
	if n < 1 {
		return nil, fmt.Errorf("no active display")
	}

	// 获取主显示器边界
	bounds := screenshot.GetDisplayBounds(0)

	// 捕获屏幕区域
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}

	// 编码为PNG格式
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CaptureScreenToBase64 捕获屏幕并返回Base64编码的图像数据
func CaptureScreenToBase64() (string, error) {
	// 捕获屏幕
	imgBytes, err := CaptureScreen()
	if err != nil {
		return "", err
	}

	// 转换为Base64
	return bytesToBase64(imgBytes), nil
}

// 将字节数组转换为Base64字符串
func bytesToBase64(data []byte) string {
	return fmt.Sprintf("data:image/png;base64,%s", string(data))
}