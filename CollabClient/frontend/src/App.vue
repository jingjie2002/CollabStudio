<template>
  <div id="app-root">
    <!-- 服务可用性检查：加载动画 -->
    <div v-if="isConnecting" class="loading-overlay">
      <div class="loading-card">
        <div class="spinner"></div>
        <h2>正在连接服务器...</h2>
        <p class="loading-hint">{{ loadingHint }}</p>
        <div v-if="connectionFailed" class="retry-area">
          <p class="error-text">⚠️ 无法连接到后端服务</p>
          <button @click="retryConnection" class="retry-btn">
            <i class="ri-refresh-line"></i> 重试连接
          </button>
        </div>
      </div>
    </div>

    <!-- 主应用视图 -->
    <template v-else>
      <Login
          v-if="currentView === 'login'"
          @login="handleLoginSuccess"
      />

      <Lobby
          v-else-if="currentView === 'lobby'"
          :user="currentUser"
          @enter-room="handleEnterRoom"
          @logout="handleLogout"
      />

      <Workspace
          v-else-if="currentView === 'workspace'"
          :username="currentUser"
          :initial-room="targetRoom"
          @logout="handleLogout"
      />
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Login from './components/Login.vue'
import Lobby from './components/Lobby.vue'
import Workspace from './components/Workspace.vue'
import { initSettings } from './settings'

// 视图状态：login -> lobby -> workspace
const currentView = ref('login')
const currentUser = ref(null)
const targetRoom = ref('demo-room')

// 连接状态
const isConnecting = ref(true)
const connectionFailed = ref(false)
const loadingHint = ref('正在检测后端服务...')

// 服务可用性检查
// 使用 Wails Go 绑定调用 Go 端的健康检查，绕过 WebView CORS 限制
const checkServerAvailability = async () => {
  const maxRetries = 10
  const retryInterval = 500

  for (let i = 0; i < maxRetries; i++) {
    loadingHint.value = `正在检测后端服务... (${i + 1}/${maxRetries})`
    try {
      // 通过 Wails Go 绑定调用 Go 端的 HTTP 健康检查
      const ok = await window.go.main.App.CheckServerHealth()
      if (ok) {
        console.log('✅ [App] 后端服务已就绪 (via Go binding)')
        isConnecting.value = false
        connectionFailed.value = false
        return
      }
    } catch (e) {
      console.log(`[App] 健康检查第 ${i + 1} 次失败:`, e)
    }
    await new Promise(resolve => setTimeout(resolve, retryInterval))
  }

  // 超时
  loadingHint.value = '连接超时'
  connectionFailed.value = true
}

// 重试连接
const retryConnection = () => {
  connectionFailed.value = false
  checkServerAvailability()
}

onMounted(() => {
  initSettings()
  checkServerAvailability()
})

// 处理登录成功
const handleLoginSuccess = (username) => {
  console.log("[App] Login success:", username)
  currentUser.value = username
  currentView.value = 'lobby'
  // 通知 Go 后端已登录，启用关闭拦截
  try { window.go.main.App.SetLoggedIn(true) } catch(e) {}
}

// 处理进入房间
const handleEnterRoom = (roomId) => {
  console.log("[App] Entering room:", roomId)
  if (roomId) {
    targetRoom.value = roomId
  }
  currentView.value = 'workspace'
}

// 处理退出登录
const handleLogout = () => {
  console.log("[App] User logged out")
  currentUser.value = null
  currentView.value = 'login'
  // 退出登录后取消关闭拦截
  try { window.go.main.App.SetLoggedIn(false) } catch(e) {}
}
</script>

<style>
/* 全局样式 */
body, html {
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  font-family: 'Nunito', sans-serif;
  background-color: #1e1e2e;
}

#app-root {
  width: 100vw;
  height: 100vh;
}

/* 滚动条美化 */
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.2); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover { background: rgba(255, 255, 255, 0.3); }

/* 加载覆盖层 */
.loading-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1e1e2e 0%, #181825 100%);
  z-index: 9999;
}

.loading-card {
  text-align: center;
  color: #cdd6f4;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid rgba(137, 180, 250, 0.2);
  border-top-color: #89b4fa;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1.5rem;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}

.loading-card h2 {
  font-size: 1.4rem;
  margin: 0 0 0.5rem;
  background: linear-gradient(to right, #89b4fa, #cba6f7);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.loading-hint {
  color: #a6adc8;
  font-size: 0.9rem;
  margin: 0;
}

.retry-area {
  margin-top: 1.5rem;
}

.error-text {
  color: #f38ba8;
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.retry-btn {
  background: #89b4fa;
  color: #1e1e2e;
  border: none;
  padding: 10px 24px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: transform 0.1s, background 0.3s;
}

.retry-btn:hover { background: #b4befe; transform: translateY(-1px); }
.retry-btn:active { transform: translateY(1px); }
</style>