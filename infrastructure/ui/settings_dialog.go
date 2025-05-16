package ui

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/qujing226/screen_sage/internal/config"
)

// ShowSettingsDialog 显示设置对话框
// 由于跨平台GUI实现复杂，这里使用简单的命令行对话框
// 在实际应用中，应该使用适当的GUI库实现
func ShowSettingsDialog() {
	// 获取当前配置
	cfg := config.GetConfig()

	// 根据操作系统选择不同的对话框实现
	switch runtime.GOOS {
	case "windows":
		showWindowsDialog(cfg)
	case "darwin":
		showMacDialog(cfg)
	default: // linux等
		showLinuxDialog(cfg)
	}
}

// Windows平台使用PowerShell实现简单对话框
func showWindowsDialog(cfg *config.Config) {
	// 构建PowerShell脚本
	script := fmt.Sprintf(`
	$deepseekKey = '%s'
	
	# 创建表单
	Add-Type -AssemblyName System.Windows.Forms
	Add-Type -AssemblyName System.Drawing
	
	$form = New-Object System.Windows.Forms.Form
	$form.Text = 'ScreenSage 设置'
	$form.Size = New-Object System.Drawing.Size(400,250)
	$form.StartPosition = 'CenterScreen'
	
	# DeepSeek API密钥
	$label = New-Object System.Windows.Forms.Label
	$label.Location = New-Object System.Drawing.Point(10,20)
	$label.Size = New-Object System.Drawing.Size(380,20)
	$label.Text = 'DeepSeek API密钥:'
	$form.Controls.Add($label)
	
	$textBox = New-Object System.Windows.Forms.TextBox
	$textBox.Location = New-Object System.Drawing.Point(10,40)
	$textBox.Size = New-Object System.Drawing.Size(360,20)
	$textBox.Text = $deepseekKey
	$form.Controls.Add($textBox)


	// 添加百度API密钥输入框
	$baiduLabel = New-Object System.Windows.Forms.Label
	$baiduLabel.Location = New-Object System.Drawing.Point(10,70)
	$baiduLabel.Size = New-Object System.Drawing.Size(380,20)
	$baiduLabel.Text = '百度OCR API密钥:'
	$form.Controls.Add($baiduLabel)
	
	$baiduTextBox = New-Object System.Windows.Forms.TextBox
	$baiduTextBox.Location = New-Object System.Drawing.Point(10,90)
	$baiduTextBox.Size = New-Object System.Drawing.Size(360,20)
	$baiduTextBox.Text = '%s' # 这里填入当前百度API密钥
	$form.Controls.Add($baiduTextBox)
	
	// 百度SecretKey输入框
	$baiduSecretLabel = New-Object System.Windows.Forms.Label
	$baiduSecretLabel.Location = New-Object System.Drawing.Point(10,120)
	$baiduSecretLabel.Size = New-Object System.Drawing.Size(380,20)
	$baiduSecretLabel.Text = '百度OCR Secret密钥:'
	$form.Controls.Add($baiduSecretLabel)
	
	$baiduSecretTextBox = New-Object System.Windows.Forms.TextBox
	$baiduSecretTextBox.Location = New-Object System.Drawing.Point(10,140)
	$baiduSecretTextBox.Size = New-Object System.Drawing.Size(360,20)
	$baiduSecretTextBox.Text = '%s' # 这里填入当前百度SecretKey
	$form.Controls.Add($baiduSecretTextBox)

	
	# 数据库路径提示
	$dbLabel = New-Object System.Windows.Forms.Label
	$dbLabel.Location = New-Object System.Drawing.Point(10,80)
	$dbLabel.Size = New-Object System.Drawing.Size(380,40)
	$dbLabel.Text = '数据库路径: ' + '%s' + '\n(重启应用后生效)'
	$form.Controls.Add($dbLabel)
	
	# 确定按钮
	$okButton = New-Object System.Windows.Forms.Button
	$okButton.Location = New-Object System.Drawing.Point(200,170)
	$okButton.Size = New-Object System.Drawing.Size(75,23)
	$okButton.Text = '确定'
	$okButton.DialogResult = [System.Windows.Forms.DialogResult]::OK
	$form.AcceptButton = $okButton
	$form.Controls.Add($okButton)
	
	# 取消按钮
	$cancelButton = New-Object System.Windows.Forms.Button
	$cancelButton.Location = New-Object System.Drawing.Point(290,170)
	$cancelButton.Size = New-Object System.Drawing.Size(75,23)
	$cancelButton.Text = '取消'
	$cancelButton.DialogResult = [System.Windows.Forms.DialogResult]::Cancel
	$form.CancelButton = $cancelButton
	$form.Controls.Add($cancelButton)
	
	# 显示对话框
	$result = $form.ShowDialog()
	
	# 处理结果
	if ($result -eq [System.Windows.Forms.DialogResult]::OK) {
		$newKey = $textBox.Text
		$baiduKey = $baiduTextBox.Text
		$baiduSecret = $baiduSecretTextBox.Text
		Write-Output "DEEPSEEK_KEY:$newKey"
		Write-Output "BAIDU_KEY:$baiduKey"
		Write-Output "BAIDU_SECRET:$baiduSecret"
	}
	`, cfg.DeepSeekAPIKey, cfg.BaiduAPIKey, cfg.BaiduSecretKey, cfg.DBPath)

	// 执行PowerShell脚本
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("显示设置对话框失败: %v，详细信息：%v", err, string(output))
		return
	}

	// 解析输出
	outputStr := string(output)
	newConfig := &config.Config{}

	for _, line := range strings.Split(outputStr, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "DEEPSEEK_KEY:") {
			newConfig.DeepSeekAPIKey = strings.TrimPrefix(line, "DEEPSEEK_KEY:")
		} else if strings.HasPrefix(line, "BAIDU_KEY:") {
			newConfig.BaiduAPIKey = strings.TrimPrefix(line, "BAIDU_KEY:")
		} else if strings.HasPrefix(line, "BAIDU_SECRET:") {
			newConfig.BaiduSecretKey = strings.TrimPrefix(line, "BAIDU_SECRET:")
		}
	}

	config.UpdateConfig(newConfig)
	log.Println("API密钥已更新")
}

// Mac平台使用AppleScript实现简单对话框
func showMacDialog(cfg *config.Config) {
	// 构建AppleScript
	script := fmt.Sprintf(`
	set deepseekKey to "%s"
	
	tell application "System Events"
		activate
		display dialog "请输入DeepSeek API密钥:" default answer deepseekKey buttons {"取消", "确定"} default button 2
		if button returned of result is "确定" then
			set newKey to text returned of result
			return "DEEPSEEK_KEY:" & newKey
		end if
	end tell
	`, cfg.DeepSeekAPIKey)

	// 执行AppleScript
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 用户取消不算错误
		if !strings.Contains(string(output), "User canceled") {
			log.Printf("显示设置对话框失败: %v", err)
		}
		return
	}

	// 解析输出
	outputStr := string(output)
	if strings.Contains(outputStr, "DEEPSEEK_KEY:") {
		parts := strings.Split(outputStr, "DEEPSEEK_KEY:")
		if len(parts) > 1 {
			newKey := strings.TrimSpace(parts[1])
			// 更新配置
			newConfig := &config.Config{
				DeepSeekAPIKey: newKey,
			}
			config.UpdateConfig(newConfig)
			log.Println("API密钥已更新")
		}
	}
}

// Linux平台使用zenity实现简单对话框
func showLinuxDialog(cfg *config.Config) {
	// 检查是否安装了zenity
	_, err := exec.LookPath("zenity")
	if err != nil {
		log.Println("未找到zenity，无法显示图形设置对话框。请安装zenity或使用配置文件设置API密钥。")
		return
	}

	// 构建zenity命令
	cmd := exec.Command("zenity", "--entry",
		"--title=ScreenSage 设置",
		"--text=请输入DeepSeek API密钥:",
		"--entry-text="+cfg.DeepSeekAPIKey)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 用户取消不算错误
		if len(output) == 0 {
			return
		}
		log.Printf("显示设置对话框失败: %v", err)
		return
	}

	// 解析输出
	newKey := strings.TrimSpace(string(output))
	if newKey != "" && newKey != cfg.DeepSeekAPIKey {
		// 更新配置
		newConfig := &config.Config{
			DeepSeekAPIKey: newKey,
		}
		config.UpdateConfig(newConfig)
		log.Println("API密钥已更新")
	}
}
