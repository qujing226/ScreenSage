<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <v-card class="mb-4">
          <v-card-title class="text-h5">
            实时OCR处理
            <v-spacer></v-spacer>
            <v-chip
              v-if="processingStatus"
              color="primary"
              text-color="white"
              class="ml-2"
            >
              {{ processingStatus }}
            </v-chip>
          </v-card-title>
          <v-card-text v-if="currentResult">
            <v-row>
              <v-col cols="12" md="4">
                <v-img
                  v-if="currentResult.thumbnail"
                  :src="`data:image/jpeg;base64,${currentResult.thumbnail}`"
                  max-height="300"
                  contain
                  class="grey lighten-2"
                ></v-img>
                <v-skeleton-loader
                  v-else
                  type="image"
                  height="300"
                ></v-skeleton-loader>
              </v-col>
              <v-col cols="12" md="8">
                <h3>OCR识别结果</h3>
                <v-card outlined class="pa-2 mb-4">
                  <pre class="text-body-1">{{ currentResult.text }}</pre>
                </v-card>
                
                <h3>AI处理结果</h3>
                <v-card outlined class="pa-2">
                  <div class="text-body-1" v-html="renderMarkdown(currentResult.answer)"></div>
                </v-card>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-text v-else>
            <v-alert type="info">
              等待新的OCR处理任务...
              <br>
              您可以使用快捷键截图或上传图片进行OCR识别。
            </v-alert>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title class="text-h5">
            历史记录
            <v-spacer></v-spacer>
            <v-btn
              color="primary"
              @click="loadHistory"
              :loading="loadingHistory"
              icon
            >
              <v-icon>mdi-refresh</v-icon>
            </v-btn>
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="headers"
              :items="historyRecords"
              :items-per-page="5"
              class="elevation-1"
            >
              <template v-slot:item.thumbnail="{ item }">
                <v-img
                  :src="`data:image/jpeg;base64,${item.thumbnail}`"
                  max-width="100"
                  max-height="60"
                  contain
                  class="grey lighten-2"
                ></v-img>
              </template>
              <template v-slot:item.title="{ item }">
                {{ extractTitle(item) }}
              </template>
              <template v-slot:item.timestamp="{ item }">
                {{ new Date(item.timestamp).toLocaleString() }}
              </template>
              <template v-slot:item.actions="{ item }">
                <v-btn
                  icon
                  small
                  @click="viewHistoryItem(item)"
                >
                  <v-icon>mdi-eye</v-icon>
                </v-btn>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import MarkdownIt from 'markdown-it';

export default {
  name: 'OCRResultView',
  data() {
    return {
      currentResult: null,
      processingStatus: '',
      currentProcessId: null,
      historyRecords: [],
      loadingHistory: false,
      md: new MarkdownIt(),
      headers: [
        { text: '缩略图', value: 'thumbnail', sortable: false },
        { text: '标题', value: 'title', sortable: true },
        { text: '时间', value: 'timestamp' },
        { text: '操作', value: 'actions', sortable: false }
      ]
    }
  },
  mounted() {
    this.setupWebSocketListeners()
    this.loadHistory()
  },
  methods: {
    renderMarkdown(text) {
      if (!text) return '';
      return this.md.render(text);
    },
    extractTitle(item) {
      // 首先检查是否有服务器提供的标题
      if (item.title) {
        return item.title.trim();
      }
      // 其次尝试从AI回答中提取【标题】格式的内容
      if (item.answer) {
        const titleMatch = item.answer.match(/【标题】([^\n]+)/);
        if (titleMatch && titleMatch[1]) {
          return titleMatch[1].trim();
        }
      }
      // 如果没有找到标题，则使用OCR文本的第一行
      return item.text ? (item.text.split('\n')[0] || '无文本内容').substring(0, 30) : '无文本内容';
    },
    setupWebSocketListeners() {
      // 确保WebSocket客户端已经初始化
      if (!this.$ws) {
        console.error('WebSocket客户端未初始化')
        return
      }

      // 处理开始
      this.$ws.on('processStart', (data) => {
        this.currentProcessId = data.id
        this.processingStatus = data.status || '处理中...'
      })

      // OCR完成
      this.$ws.on('ocrComplete', (data) => {
        if (data.id === this.currentProcessId) {
          this.processingStatus = data.status || 'OCR识别完成'
          if (!this.currentResult) {
            this.currentResult = {}
          }
          this.currentResult.text = data.text
        }
      })

      // 处理完成
      this.$ws.on('processComplete', (data) => {
        if (data.process_id === this.currentProcessId) {
          this.processingStatus = '处理完成'
          this.currentResult = {
            id: data.id,
            text: data.text,
            answer: data.answer,
            title: data.title,
            timestamp: data.timestamp,
            thumbnail: data.thumbnail
          }
          
          // 刷新历史记录
          this.loadHistory()
        }
      })

      // 处理错误
      this.$ws.on('processError', (data) => {
        if (data.id === this.currentProcessId) {
          this.processingStatus = `错误: ${data.error}`
        }
      })

      // 历史记录
      this.$ws.on('history', (records) => {
        if (Array.isArray(records) && records.length > 0) {
          this.historyRecords = records
        }
      })
    },
    
    async loadHistory() {
      this.loadingHistory = true
      try {
        const response = await fetch('/api/history')
        if (response.ok) {
          const data = await response.json()
          this.historyRecords = data
        } else {
          console.error('加载历史记录失败:', response.statusText)
        }
      } catch (error) {
        console.error('加载历史记录出错:', error)
      } finally {
        this.loadingHistory = false
      }
    },
    
    viewHistoryItem(item) {
      this.currentResult = item
      this.processingStatus = '历史记录'
      this.currentProcessId = null
    }
  }
}
</script>

<style scoped>
pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 200px;
  overflow-y: auto;
}

.text-body-1 :deep(pre) {
  background-color: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}

.text-body-1 :deep(code) {
  background-color: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
}

.text-body-1 :deep(ul), .text-body-1 :deep(ol) {
  padding-left: 20px;
}

.text-body-1 :deep(blockquote) {
  border-left: 4px solid #ccc;
  padding-left: 16px;
  margin-left: 0;
  color: #666;
}

.text-body-1 :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin-bottom: 16px;
}

.text-body-1 :deep(th), .text-body-1 :deep(td) {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

.text-body-1 :deep(th) {
  background-color: #f2f2f2;
}
</style>