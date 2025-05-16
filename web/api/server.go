package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/qujing226/screen_sage/domain/model"
	"github.com/qujing226/screen_sage/internal/config"
	"github.com/qujing226/screen_sage/internal/ocr"
	"github.com/qujing226/screen_sage/internal/storage"
)

// Server 表示Web服务器
// 负责处理HTTP请求、WebSocket连接和广播消息
type Server struct {
	Port        int                      // 服务器监听端口
	DBManager   *storage.DBManager       // 数据库管理器
	OCRClient   ocr.OCRClient            // OCR客户端接口
	DeepSeekKey string                   // DeepSeek API密钥
	StaticPath  string                   // 静态文件路径
	Clients     map[*websocket.Conn]bool // 已连接的WebSocket客户端
	Broadcast   chan *BroadcastMessage   // 广播消息通道
	ClientsMux  sync.Mutex               // 客户端列表互斥锁
	Upgrader    websocket.Upgrader       // WebSocket升级器
}

// BroadcastMessage 表示广播消息的结构
type BroadcastMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewServer 创建一个新的Web服务器
func NewServer(port int, dbPath string, baiduAPIKey string, baiduSecretKey string, deepSeekKey string, staticPath string) (*Server, error) {
	// 初始化数据库管理器
	dbManager, err := storage.NewDBManager(dbPath)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %v", err)
	}

	// 创建OCR客户端
	ocrClient := ocr.NewBaiduOCRClient(baiduAPIKey, baiduSecretKey)

	// 创建服务器
	server := &Server{
		Port:        port,
		DBManager:   dbManager,
		OCRClient:   ocrClient,
		DeepSeekKey: deepSeekKey,
		StaticPath:  staticPath,
		Clients:     make(map[*websocket.Conn]bool),
		Broadcast:   make(chan *BroadcastMessage),
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有跨域请求，生产环境中应该更严格
			},
		},
	}

	// 启动广播处理协程
	go server.handleBroadcasts()

	return server, nil
}

// StartServer 启动Web服务器
func StartServer() (*Server, error) {
	// 获取配置
	cfg := config.GetConfig()

	// 创建服务器实例
	server, err := NewServer(
		cfg.Port,           // 端口
		cfg.DBPath,         // 数据库路径
		cfg.BaiduAPIKey,    // 百度API密钥
		cfg.BaiduSecretKey, // 百度Secret密钥
		cfg.DeepSeekAPIKey, // DeepSeek API密钥
		cfg.StaticPath,     // 静态文件路径
	)
	if err != nil {
		return nil, err
	}

	// 确保数据目录存在
	err = os.MkdirAll(filepath.Dir("./data/screensage.db"), 0755)
	if err != nil {
		return nil, err
	}

	// 注册路由
	http.HandleFunc("/api/history", server.handleHistory)
	http.HandleFunc("/api/upload", server.handleUpload)
	http.HandleFunc("/api/exit", server.handleExit)
	http.HandleFunc("/ws", server.handleWebSocket)

	// 静态文件服务
	http.Handle("/", http.FileServer(http.Dir(server.StaticPath)))

	// 启动服务器
	addr := fmt.Sprintf(":%d", server.Port)
	log.Printf("Web服务器启动，监听端口 %s", addr)

	// 启动HTTP服务器
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("HTTP服务器错误: %v", err)
		}
	}()

	return server, nil
}

// handleBroadcasts 处理广播消息
func (s *Server) handleBroadcasts() {
	for {
		msg := <-s.Broadcast

		// 向所有连接的客户端发送消息
		s.ClientsMux.Lock()
		for client := range s.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("发送消息失败: %v", err)
				client.Close()
				delete(s.Clients, client)
			}
		}
		s.ClientsMux.Unlock()
	}
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 注册新客户端
	s.ClientsMux.Lock()
	s.Clients[conn] = true
	s.ClientsMux.Unlock()

	// 处理连接关闭
	defer func() {
		conn.Close()
		s.ClientsMux.Lock()
		delete(s.Clients, conn)
		s.ClientsMux.Unlock()
	}()

	// 发送历史记录
	records, err := s.DBManager.GetHistory(10) // 最近10条记录
	if err == nil && len(records) > 0 {
		conn.WriteJSON(&BroadcastMessage{
			Type:    "history",
			Payload: records,
		})
	}

	// 监听消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// 目前只是保持连接活跃，不处理客户端消息
	}
}

// handleHistory 处理获取历史记录的请求
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取历史记录
	records, err := s.DBManager.GetHistory(50) // 最多返回50条记录
	if err != nil {
		log.Printf("获取历史记录失败: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// handleUpload 处理上传截图的请求
// 此函数处理从前端上传的截图，执行OCR识别，然后将结果发送给DeepSeek进行分析
// 整个处理过程是异步的，通过WebSocket向客户端发送进度更新
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var request struct {
		Image string `json:"image"` // Base64编码的图像
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("解析请求失败: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 验证图像数据
	if request.Image == "" {
		log.Printf("请求中没有图像数据")
		http.Error(w, "No image data", http.StatusBadRequest)
		return
	}

	// 创建处理ID
	processID := fmt.Sprintf("proc_%d", time.Now().UnixNano())

	// 通知客户端处理开始
	s.Broadcast <- &BroadcastMessage{
		Type: "process_start",
		Payload: map[string]string{
			"id":     processID,
			"status": "开始处理图像...",
		},
	}

	// 异步处理图像
	go func() {
		// 调用OCR服务
		log.Printf("开始OCR识别，处理ID: %s", processID)
		text, err := s.OCRClient.RecognizeText(request.Image)
		if err != nil {
			log.Printf("OCR识别失败: %v", err)

			// 通知客户端处理失败
			s.Broadcast <- &BroadcastMessage{
				Type: "process_error",
				Payload: map[string]string{
					"id":    processID,
					"error": fmt.Sprintf("OCR识别失败: %v", err),
				},
			}
			return
		}

		// 通知客户端OCR完成
		log.Printf("OCR识别完成，处理ID: %s，文本长度: %d", processID, len(text))
		s.Broadcast <- &BroadcastMessage{
			Type: "ocr_complete",
			Payload: map[string]string{
				"id":     processID,
				"text":   text,
				"status": "OCR识别完成，正在处理内容...",
			},
		}

		// 调用DeepSeek处理文本
		log.Printf("开始调用DeepSeek处理文本，处理ID: %s", processID)
		answer, err := ocr.ProcessWithDeepSeek(text, s.DeepSeekKey)
		if err != nil {
			log.Printf("DeepSeek处理失败: %v", err)
			answer = "AI处理失败，但您仍然可以查看OCR识别的文本。"
		} else {
			log.Printf("DeepSeek处理成功，处理ID: %s，回答长度: %d", processID, len(answer))
		}

		// 保存历史记录
		record := &storage.HistoryRecord{
			Timestamp: time.Now(),
			ImagePath: "", // TODO: 保存图像文件
			Thumbnail: request.Image,
			Text:      text,
			Answer:    answer,
		}

		id, err := s.DBManager.AddHistory(record)
		if err != nil {
			log.Printf("保存历史记录失败: %v", err)
		}

		// 通知客户端处理完成
		s.Broadcast <- &BroadcastMessage{
			Type: "process_complete",
			Payload: map[string]interface{}{
				"id":         id,
				"process_id": processID,
				"text":       text,
				"answer":     answer,
				"timestamp":  time.Now(),
				"thumbnail":  request.Image, // 确保缩略图被传递到前端
			},
		}
	}()

	// 立即返回处理ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     processID,
		"status": "处理中",
	})
}

// handleExit 处理退出应用的请求
func (s *Server) handleExit(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})

	// 延迟退出，确保响应已发送
	go func() {
		time.Sleep(100 * time.Millisecond)
		log.Println("收到退出请求，应用即将关闭")
		os.Exit(0)
	}()
}

// BroadcastScreenshot 广播截图到客户端
func (s *Server) BroadcastScreenshot(screenshot *model.Screenshot) {
	// 创建处理ID
	processID := fmt.Sprintf("screenshot_%d", time.Now().UnixNano())

	// 广播到客户端
	s.Broadcast <- &BroadcastMessage{
		Type: "process_start",
		Payload: map[string]string{
			"id":     processID,
			"status": "开始处理截图...",
		},
	}

	// 异步处理图像
	go func() {
		// 将图像转换为Base64
		imageBase64 := screenshot.Thumbnail

		// 调用OCR服务
		text, err := s.OCRClient.RecognizeText(imageBase64)
		if err != nil {
			log.Printf("OCR识别失败: %v", err)

			// 通知客户端处理失败
			s.Broadcast <- &BroadcastMessage{
				Type: "process_error",
				Payload: map[string]string{
					"id":    processID,
					"error": fmt.Sprintf("OCR识别失败: %v", err),
				},
			}
			return
		}

		// 通知客户端OCR完成
		s.Broadcast <- &BroadcastMessage{
			Type: "ocr_complete",
			Payload: map[string]string{
				"id":     processID,
				"text":   text,
				"status": "OCR识别完成，正在处理内容...",
			},
		}

		// 调用DeepSeek处理文本
		answer, err := ocr.ProcessWithDeepSeek(text, s.DeepSeekKey)
		if err != nil {
			log.Printf("DeepSeek处理失败: %v", err)
			answer = "AI处理失败，但您仍然可以查看OCR识别的文本。"
		}

		// 保存历史记录
		record := &storage.HistoryRecord{
			Timestamp: time.Now(),
			ImagePath: "", // TODO: 保存图像文件
			Thumbnail: imageBase64,
			Text:      text,
			Answer:    answer,
		}

		id, err := s.DBManager.AddHistory(record)
		if err != nil {
			log.Printf("保存历史记录失败: %v", err)
		}

		// 通知客户端处理完成
		s.Broadcast <- &BroadcastMessage{
			Type: "process_complete",
			Payload: map[string]interface{}{
				"id":         id,
				"process_id": processID,
				"text":       text,
				"answer":     answer,
				"timestamp":  time.Now(),
				"thumbnail":  imageBase64,
			},
		}
	}()
}
