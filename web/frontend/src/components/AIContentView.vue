<template>
  <v-container class="pa-2">
    <v-card class="mb-4 mobile-card">
      <v-card-title class="text-h6">
        AI分析结果
        <v-spacer></v-spacer>
        <v-chip
          v-if="processingStatus"
          color="primary"
          text-color="white"
          class="ml-2"
          small
        >
          {{ processingStatus }}
        </v-chip>
      </v-card-title>
      <v-card-text v-if="currentResult">
        <v-row>
          <v-col cols="12">
            <v-img
              v-if="currentResult.thumbnail"
              :src="`data:image/jpeg;base64,${currentResult.thumbnail}`"
              max-height="200"
              contain
              class="grey lighten-2 rounded mb-3"
            ></v-img>
            <v-skeleton-loader
              v-else
              type="image"
              height="200"
            ></v-skeleton-loader>
          </v-col>
          <v-col cols="12">
            <v-expansion-panels accordion flat>
              <v-expansion-panel>
                <v-expansion-panel-header>
                  <div class="text-subtitle-1 font-weight-medium">OCR识别结果</div>
                </v-expansion-panel-header>
                <v-expansion-panel-content>
                  <v-card outlined class="pa-2 mb-2">
                    <pre class="text-body-2 ocr-text">{{ currentResult.text }}</pre>
                  </v-card>
                </v-expansion-panel-content>
              </v-expansion-panel>
            </v-expansion-panels>
            
            <div class="text-subtitle-1 font-weight-medium mt-3 mb-2">AI分析</div>
            <v-card outlined class="pa-3">
              <div class="text-body-1 answer-text">{{ currentResult.answer }}</div>
            </v-card>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-text v-else>
        <v-alert type="info" dense>
          等待新的OCR处理任务...
          <br>
          您可以点击右上角上传按钮或使用底部导航上传图片进行OCR识别。
        </v-alert>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script>
export default {
  name: 'AIContentView',
  data() {
    return {
      currentResult: null,
      processingStatus: '',
      currentProcessId: null
    }
  },
  mounted() {
    this.setupWebSocketListeners()
  },
  methods: {
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
        console.log('处理开始，ID:', this.currentProcessId)
        
        // 如果没有当前结果，初始化一个空对象
        if (!this.currentResult) {
          this.currentResult = {}
        }
      })

      // OCR完成
      this.$ws.on('ocrComplete', (data) => {
        // 检查ID是否匹配，如果匹配或者当前没有处理中的任务，则更新结果
        if (data.id === this.currentProcessId || !this.currentProcessId) {
          this.processingStatus = data.status || 'OCR识别完成'
          if (!this.currentResult) {
            this.currentResult = {}
            // 更新当前处理ID
            this.currentProcessId = data.id
          }
          this.currentResult.text = data.text
          console.log('OCR识别完成，文本长度:', data.text.length, '当前处理ID:', this.currentProcessId)
        } else {
          console.log('收到OCR完成事件，但ID不匹配。当前ID:', this.currentProcessId, '收到的ID:', data.id)
        }
      })

      // 处理完成
      this.$ws.on('processComplete', (data) => {
        console.log('收到processComplete事件，数据:', data)
        // 检查process_id是否匹配，如果匹配或者当前没有处理中的任务，则更新结果
        if (data.process_id === this.currentProcessId || !this.currentProcessId) {
          this.processingStatus = '处理完成'
          this.currentResult = {
            id: data.id,
            text: data.text,
            answer: data.answer,
            timestamp: typeof data.timestamp === 'string' ? data.timestamp : new Date(data.timestamp).toISOString(),
            thumbnail: data.thumbnail
          }
          // 更新当前处理ID
          this.currentProcessId = data.process_id
          console.log('处理完成，更新结果:', this.currentResult)
        } else {
          console.log('收到处理完成事件，但ID不匹配。当前ID:', this.currentProcessId, '收到的ID:', data.process_id)
        }
      })

      // 处理错误
      this.$ws.on('processError', (data) => {
        if (data.id === this.currentProcessId) {
          this.processingStatus = `错误: ${data.error}`
        }
      })
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
  white-space: pre-line;
  word-break: break-word;
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