<template>
  <v-container class="pa-2">
    <v-card class="mb-4 mobile-card">
      <v-card-title class="text-h6">
        历史记录
        <v-spacer></v-spacer>
        <v-btn
          color="primary"
          @click="loadHistory"
          :loading="loadingHistory"
          icon
          small
        >
          <v-icon>mdi-refresh</v-icon>
        </v-btn>
      </v-card-title>
      <v-card-text>
        <v-alert v-if="historyRecords.length === 0 && !loadingHistory" type="info" dense>
          暂无历史记录
        </v-alert>
        
        <v-list v-else two-line class="pa-0">
          <v-list-item
            v-for="item in historyRecords"
            :key="item.id"
            @click="viewHistoryItem(item)"
            class="mb-2"
          >
            <v-list-item-avatar tile size="60">
              <v-img
                :src="`data:image/jpeg;base64,${item.thumbnail}`"
                contain
                class="grey lighten-2"
              ></v-img>
            </v-list-item-avatar>
            
            <v-list-item-content>
              <v-list-item-title class="text-truncate">
                {{ extractTitle(item) }}
              </v-list-item-title>
              <v-list-item-subtitle>
                {{ new Date(item.timestamp).toLocaleString() }}
              </v-list-item-subtitle>
            </v-list-item-content>
            
            <v-list-item-action>
              <v-icon>mdi-chevron-right</v-icon>
            </v-list-item-action>
          </v-list-item>
        </v-list>
      </v-card-text>
    </v-card>
    
    <!-- 历史记录详情对话框 -->
    <v-dialog v-model="showHistoryDetail" fullscreen transition="dialog-bottom-transition">
      <v-card>
        <v-toolbar dark color="primary">
          <v-btn icon dark @click="showHistoryDetail = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
          <v-toolbar-title>历史记录详情</v-toolbar-title>
        </v-toolbar>
        
        <v-card-text class="pt-4" v-if="selectedItem">
          <v-row>
            <v-col cols="12">
              <v-img
                :src="`data:image/jpeg;base64,${selectedItem.thumbnail}`"
                max-height="200"
                contain
                class="grey lighten-2 rounded mb-3"
              ></v-img>
            </v-col>
            
            <v-col cols="12">
              <v-expansion-panels accordion flat>
                <v-expansion-panel>
                  <v-expansion-panel-header>
                    <div class="text-subtitle-1 font-weight-medium">OCR识别结果</div>
                  </v-expansion-panel-header>
                  <v-expansion-panel-content>
                    <v-card outlined class="pa-2 mb-2">
                      <pre class="text-body-2 ocr-text">{{ selectedItem.text }}</pre>
                    </v-card>
                  </v-expansion-panel-content>
                </v-expansion-panel>
              </v-expansion-panels>
              
              <div class="text-subtitle-1 font-weight-medium mt-3 mb-2">AI分析</div>
              <v-card outlined class="pa-3">
                <div class="text-body-1 answer-text" v-html="renderMarkdown(selectedItem.answer)"></div>
              </v-card>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script>
import MarkdownIt from 'markdown-it';

export default {
  name: 'HistoryView',
  data() {
    return {
      historyRecords: [],
      loadingHistory: false,
      showHistoryDetail: false,
      selectedItem: null,
      md: new MarkdownIt()
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

      // 历史记录
      this.$ws.on('history', (records) => {
        if (Array.isArray(records) && records.length > 0) {
          this.historyRecords = records
        }
      })
      
      // 处理完成后更新历史记录
      this.$ws.on('processComplete', () => {
        this.loadHistory()
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
      this.selectedItem = item
      this.showHistoryDetail = true
    }
  }
}
</script>

<style scoped>
.ocr-text {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.875rem;
  max-height: 200px;
  overflow-y: auto;
}

.answer-text {
  word-break: break-word;
}

.answer-text :deep(pre) {
  background-color: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}

.answer-text :deep(code) {
  background-color: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
}

.answer-text :deep(ul), .answer-text :deep(ol) {
  padding-left: 20px;
}

.answer-text :deep(blockquote) {
  border-left: 4px solid #ccc;
  padding-left: 16px;
  margin-left: 0;
  color: #666;
}

.answer-text :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin-bottom: 16px;
}

.answer-text :deep(th), .answer-text :deep(td) {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

.answer-text :deep(th) {
  background-color: #f2f2f2;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .mobile-card {
    margin: 0 -12px;
    border-radius: 0;
  }
  
  .v-card__title {
    padding: 12px 16px;
  }
  
  .v-card__text {
    padding: 12px 16px;
  }
}
</style>