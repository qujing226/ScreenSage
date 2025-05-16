import { createApp } from 'vue'
import App from './App.vue'
import wsClient from './websocket'

// Vuetify
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary: '#1867C0',
          secondary: '#5CBBF6',
        },
      },
    },
  },
})

// 创建Vue应用实例
const app = createApp(App)

// 使用Vuetify
app.use(vuetify)

// 将WebSocket客户端添加到全局属性
app.config.globalProperties.$ws = wsClient

// 挂载应用
app.mount('#app')

// 连接WebSocket
wsClient.connect()

// 监听WebSocket事件
wsClient.on('connect', () => {
  console.log('WebSocket已连接，可以接收实时通知')
})

wsClient.on('processComplete', (data) => {
  console.log('收到处理完成的结果:', data)
  // 这里可以更新UI或显示通知
})

// 在页面关闭时断开WebSocket连接
window.addEventListener('beforeunload', () => {
  wsClient.disconnect()
})