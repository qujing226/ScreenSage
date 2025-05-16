<template>
  <v-app>
    <v-app-bar app color="primary" dark>
      <v-app-bar-title>ScreenSage - 屏幕识图+智能问答助手</v-app-bar-title>
      <v-spacer></v-spacer>
      <v-btn icon @click="showUploadDialog = true">
        <v-icon>mdi-upload</v-icon>
      </v-btn>
    </v-app-bar>

    <v-main>
      <v-container class="mobile-container pa-0">
        <!-- 移动端标签页导航 -->
        <v-tabs
          v-model="activeTab"
          background-color="primary"
          dark
          grow
          v-if="$vuetify.display.mobile"
        >
          <v-tab value="ai">AI内容</v-tab>
          <v-tab value="history">历史记录</v-tab>
        </v-tabs>
        
        <v-tabs-items v-model="activeTab" v-if="$vuetify.display.mobile">
          <v-tab-item value="ai">
            <AIContentView ref="aiContentView" />
          </v-tab-item>
          <v-tab-item value="history">
            <HistoryView ref="historyView" />
          </v-tab-item>
        </v-tabs-items>
        
        <!-- 桌面端视图 -->
        <div v-if="!$vuetify.display.mobile">
          <OCRResultView ref="ocrResultView" />
        </div>
      </v-container>
    </v-main>

    <!-- 图片上传对话框 -->
    <v-dialog v-model="showUploadDialog" max-width="500px">
      <v-card>
        <v-card-title>上传图片进行OCR识别</v-card-title>
        <v-card-text>
          <v-file-input
            v-model="imageFile"
            accept="image/*"
            label="选择图片"
            prepend-icon="mdi-camera"
            show-size
            truncate-length="15"
          ></v-file-input>
          <v-alert v-if="uploadError" type="error" dense>
            {{ uploadError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" text @click="showUploadDialog = false">取消</v-btn>
          <v-btn 
            color="primary" 
            :loading="uploading" 
            :disabled="!imageFile" 
            @click="uploadImage"
          >
            上传并识别
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 移动端底部导航 -->
    <v-bottom-navigation v-if="$vuetify.display.mobile" app color="primary" grow>
      <v-btn value="ai" @click="activeTab = 'ai'">
        <v-icon>mdi-brain</v-icon>
        AI内容
      </v-btn>
      <v-btn value="upload" @click="showUploadDialog = true">
        <v-icon>mdi-upload</v-icon>
        上传
      </v-btn>
      <v-btn value="history" @click="activeTab = 'history'">
        <v-icon>mdi-history</v-icon>
        历史
      </v-btn>
    </v-bottom-navigation>

    <v-footer app v-if="!$vuetify.display.mobile">
      <div class="text-caption text-center w-100">
        ScreenSage &copy; {{ new Date().getFullYear() }}
      </div>
    </v-footer>
  </v-app>
</template>

<script>
import OCRResultView from './components/OCRResultView.vue';
import AIContentView from './components/AIContentView.vue';
import HistoryView from './components/HistoryView.vue';

export default {
  name: 'App',
  components: {
    OCRResultView,
    AIContentView,
    HistoryView
  },
  data() {
    return {
      activeTab: 'ai',
      showUploadDialog: false,
      imageFile: null,
      uploading: false,
      uploadError: null
    };
  },
  methods: {
    async uploadImage() {
      if (!this.imageFile) return;
      
      this.uploading = true;
      this.uploadError = null;
      
      try {
        // 读取文件为Base64
        const base64Image = await this.readFileAsBase64(this.imageFile);
        
        // 发送到服务器
        const response = await fetch('/api/upload', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            image: base64Image
          })
        });
        
        if (!response.ok) {
          throw new Error(`服务器返回错误: ${response.status}`);
        }
        
        // 关闭对话框
        this.showUploadDialog = false;
        this.imageFile = null;
      } catch (error) {
        console.error('上传图片失败:', error);
        this.uploadError = `上传失败: ${error.message}`;
      } finally {
        this.uploading = false;
      }
    },
    readFileAsBase64(file) {
      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
          // 获取Base64字符串，去除前缀
          const base64String = reader.result.split(',')[1];
          resolve(base64String);
        };
        reader.onerror = error => reject(error);
        reader.readAsDataURL(file);
      });
    },
    refreshData() {
      // 根据当前激活的标签页刷新相应的组件数据
      if (this.$vuetify.display.mobile) {
        if (this.activeTab === 'ai') {
          // AI内容视图不需要主动刷新，它会通过WebSocket自动更新
        } else if (this.activeTab === 'history') {
          this.$refs.historyView && this.$refs.historyView.loadHistory();
        }
      } else {
        // 桌面端刷新OCRResultView
        this.$refs.ocrResultView && this.$refs.ocrResultView.loadHistory();
      }
    }
  }
};
</script>

<style>
.mobile-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 16px;
}

@media (max-width: 768px) {
  .mobile-container {
    padding: 8px;
  }
  .mobile-card {
    border-radius: 0;
    box-shadow: none;
  }

  /* 确保内容区域在移动设备上有足够的底部边距，避免被底部导航遮挡 */
  .v-main {
    padding-bottom: 56px;
  }
 }

.timeline-content {
  padding: 8px;
}

.answer-text {
  margin-top: 8px;
  white-space: pre-wrap;
}
</style>