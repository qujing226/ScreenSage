# ScreenSage

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 项目简介

ScreenSage 是一款屏幕识图+智能问答助手，通过截取屏幕内容并使用 OCR 技术识别文字，结合 AI 模型提供智能问答服务。

## 最近修复的问题

1. **WebSocket通信问题** - 修复了前端无法接收截图内容的问题
2. **OCR处理流程** - 确保OCR识别内容正确发送至DeepSeek并等待响应
3. **项目结构优化** - 重构了项目结构，删除了不必要的代码，并添加了关键注释

### 核心特性

- **一键截屏**：通过全局热键（Ctrl+Alt+Q）快速截取屏幕内容
- **智能识别**：使用 DeepSeek OCR API 进行高精度文字识别
- **实时问答**：将识别结果发送至 AI 模型获取即时回答
- **历史记录**：保存所有截图和问答记录，方便回顾和查询

## 安装说明

### 系统要求

- 支持 Windows 10/11 64位系统
- 浏览器要求：Chrome 90+ / Safari 15+
- 屏幕分辨率：建议缩放比例≤125%

### 推荐配置

- CPU：4核以上
- 内存：8GB+
- 存储：SSD预留500MB空间
- 需要 VC++ Redistributable 运行时（Windows）

### 安装步骤

1. 下载最新版本的安装包
2. 运行安装程序，按照向导完成安装
3. 首次运行时配置 DeepSeek API 密钥

## 使用方法

1. 启动 ScreenSage 应用，程序将在系统托盘中运行
2. 使用全局热键 `Ctrl+Alt+Q` 进行屏幕截图
3. 系统自动识别截图中的文字并发送至 AI 模型
4. 在应用界面查看 AI 回答结果
5. 通过历史记录时间轴查看之前的问答记录

## 技术架构

### 后台服务模块（Go）

- **系统托盘管理**：使用 github.com/getlantern/systray 实现常驻系统托盘
- **全局热键监听**：采用 github.com/micmonay/keybd_event 实现 Ctrl+Alt+Q 组合键监听
- **静默截屏功能**：通过 github.com/kbinani/screenshot 实现无界面截屏
- **图片处理流水线**：
  - PNG 格式转换
  - Base64 编码
  - 调用 DeepSeek OCR API
- **数据存储**：使用 SQLite 存储历史记录（github.com/mattn/go-sqlite3）

### Web 服务模块（Go）

- **RESTful API 设计**：
  - GET /api/history - 获取答题历史
  - POST /api/upload - 处理截图上传
  - GET /api/exit - 安全退出程序
- **静态文件服务**：嵌入打包 Vue 编译产物
- **CORS 配置**：允许跨域访问

### 前端模块（Vue 3）

- **响应式布局**：使用 Vuetify 实现移动端优先设计
- **核心功能组件**：
  - 实时结果流式展示
  - 历史记录时间轴
  - 屏幕适配检测面板
- **数据交互**：axios 配合长轮询（30秒间隔）

## 技术实现细节

### 静默截屏实现

```go
func captureScreen() ([]byte, error) {
    n := screenshot.NumActiveDisplays()
    if n < 1 {
        return nil, fmt.Errorf("no active display")
    }
    
    bounds := screenshot.GetDisplayBounds(0)
    img, err := screenshot.CaptureRect(bounds)
    if err != nil {
        return nil, err
    }
    
    var buf bytes.Buffer
    if err := png.Encode(&buf, img); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
```

### 全局热键监听方案

```go
func registerHotkey() {
    kb, _ := keybd_event.NewKeyBonding()
    kb.SetKeys(keybd_event.VK_Q)
    kb.HasCTRL(true)
    kb.HasALT(true)
    
    go func() {
        for {
            time.Sleep(100 * time.Millisecond)
            if err := kb.Launching(); err == nil {
                processScreenshot()
            }
        }
    }()
}
```

### 移动端适配方案（Vue）

```vue
<template>
  <v-container class="mobile-container">
    <v-card :class="{ 'mobile-card': $vuetify.display.mobile }">
      <v-timeline density="compact" align="start">
        <v-timeline-item
          v-for="(item, i) in history"
          :key="i"
          dot-color="primary"
          size="x-small"
        >
          <div class="timeline-content">
            <div class="text-caption">{{ formatTime(item.time) }}</div>
            <v-img :src="item.thumbnail" max-width="120" />
            <div class="answer-text">{{ item.answer }}</div>
          </div>
        </v-timeline-item>
      </v-timeline>
    </v-card>
  </v-container>
</template>

<style>
@media (max-width: 768px) {
  .mobile-card {
    border-radius: 0;
    box-shadow: none;
  }
  .timeline-content {
    font-size: 0.875rem;
  }
}
</style>
```

## 注意事项

### 安全注意事项

- 建议将 DeepSeek API 密钥加密存储（使用 AES-GCM）
- 启用 HTTPS 时需要有效证书（推荐使用 Let's Encrypt）
- 设置每日 API 调用限额

### 性能优化项

- 图片压缩处理：将截图分辨率限制在 1920x1080 以内
- 结果缓存机制：相同截图 MD5 值重复使用历史答案
- 连接池配置：数据库和 HTTP 客户端都需要

## 扩展能力设计

- 插件系统架构预留
- 多 OCR 服务商支持接口
- 题目分类标记系统
- 错题本功能数据结构

## 开发指南

### 项目启动建议

1. 先实现核心链路：截屏→OCR→结果展示
2. 使用 Mock 数据开发前端界面
3. 分阶段集成各模块功能
4. 重点测试 Windows 权限问题

### 周边工具

- 一键打包脚本（包含 UPX 压缩）
- 自动更新检查模块
- 诊断日志生成工具
- 安装程序制作（NSIS）

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进项目。

## 许可证

本项目采用 MIT 许可证。