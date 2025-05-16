// WebSocket客户端实现

class WebSocketClient {
  constructor(url) {
    this.url = url || `ws://${window.location.host}/ws`;
    this.socket = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectInterval = 3000; // 3秒
    this.listeners = {
      message: [],
      connect: [],
      disconnect: [],
      error: [],
      history: [],
      processStart: [],
      processComplete: [],
      processError: [],
      ocrComplete: []
    };
  }

  // 连接WebSocket
  connect() {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    try {
      this.socket = new WebSocket(this.url);

      this.socket.onopen = () => {
        console.log('WebSocket连接已建立');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this._trigger('connect');
      };

      this.socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this._trigger('message', data);

          // 根据消息类型触发特定事件
          switch (data.type) {
            case 'history':
              this._trigger('history', data.payload);
              break;
            case 'process_start':
              this._trigger('processStart', data.payload);
              break;
            case 'process_complete':
              console.log('收到process_complete消息:', data.payload);
              this._trigger('processComplete', data.payload);
              break;
            case 'process_error':
              this._trigger('processError', data.payload);
              break;
            case 'ocr_complete':
              console.log('收到ocr_complete消息:', data.payload);
              this._trigger('ocrComplete', data.payload);
              break;
            case 'screenshot':
              // 处理截图消息
              this._trigger('processStart', {
                id: 'screenshot_' + Date.now(),
                status: '收到新截图，正在处理...'
              });
              break;
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error);
        }
      };

      this.socket.onclose = () => {
        console.log('WebSocket连接已关闭');
        this.isConnected = false;
        this._trigger('disconnect');
        this._attemptReconnect();
      };

      this.socket.onerror = (error) => {
        console.error('WebSocket错误:', error);
        this._trigger('error', error);
      };
    } catch (error) {
      console.error('创建WebSocket连接失败:', error);
      this._attemptReconnect();
    }
  }

  // 尝试重新连接
  _attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`尝试重新连接 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      setTimeout(() => this.connect(), this.reconnectInterval);
    } else {
      console.error('达到最大重连次数，放弃重连');
    }
  }

  // 关闭连接
  disconnect() {
    if (this.socket) {
      this.socket.close();
    }
  }

  // 发送消息
  send(data) {
    if (!this.isConnected) {
      console.error('WebSocket未连接，无法发送消息');
      return false;
    }

    try {
      const message = typeof data === 'string' ? data : JSON.stringify(data);
      this.socket.send(message);
      return true;
    } catch (error) {
      console.error('发送WebSocket消息失败:', error);
      return false;
    }
  }

  // 添加事件监听器
  on(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event].push(callback);
    }
    return this;
  }

  // 移除事件监听器
  off(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback);
    }
    return this;
  }

  // 触发事件
  _trigger(event, data) {
    if (this.listeners[event]) {
      this.listeners[event].forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`执行${event}事件回调时出错:`, error);
        }
      });
    }
  }
}

// 创建单例实例
const wsClient = new WebSocketClient();

export default wsClient;