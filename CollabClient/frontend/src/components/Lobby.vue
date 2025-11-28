<script setup>
import { ref, onMounted } from 'vue'
import { serverConfig } from '../store'

const props = defineProps({
  user: { type: String, default: 'Guest' }
})

const emit = defineEmits(['enter-room', 'logout'])
const roomIdInput = ref('demo-room')
const historyList = ref([])
const isLoading = ref(false)

const fetchHistory = async () => {
  if (!props.user) return
  isLoading.value = true
  try {
    const targetUrl = `${serverConfig.getHttpUrl()}/history?username=${encodeURIComponent(props.user)}`
    const res = await fetch(targetUrl)
    const data = await res.json()
    if (data.history) {
      historyList.value = data.history
    }
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
}

const joinFromHistory = (roomID) => {
  roomIdInput.value = roomID
  handleEnter()
}

const formatTime = (isoString) => {
  const date = new Date(isoString)
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`
}

const handleEnter = () => {
  // 🟢 修复：自动去除首尾空格，防止 "room " 和 "room" 不匹配
  const cleanRoomId = roomIdInput.value.trim()

  if (!cleanRoomId) return

  // 发送处理后的房间号
  emit('enter-room', cleanRoomId)
}

onMounted(() => { fetchHistory() })
</script>

<template>
  <div class="lobby-layout fade-in">
    <!-- 左侧导航 -->
    <aside class="lobby-sidebar">
      <div class="user-card">
        <div class="avatar-box">
          <div class="avatar">{{ user.charAt(0).toUpperCase() }}</div>
          <div class="status-indicator" title="已连接"></div>
        </div>
        <div class="user-info">
          <h3>{{ user }}</h3>
          <span class="user-role">Developer</span>
        </div>
      </div>

      <div class="nav-section">
        <div class="section-title">最近访问</div>
        <div v-if="isLoading" class="state-text">加载中...</div>
        <div v-else-if="historyList.length === 0" class="state-text">暂无记录</div>

        <ul v-else class="history-list">
          <li v-for="item in historyList" :key="item.id" @click="joinFromHistory(item.room_id)" class="history-item">
            <i class="ri-file-text-line icon"></i>
            <div class="meta">
              <span class="room-name">{{ item.room_id }}</span>
              <span class="room-time">{{ formatTime(item.last_visited) }}</span>
            </div>
          </li>
        </ul>
      </div>

      <div class="sidebar-footer">
        <button class="btn-logout" @click="$emit('logout')">
          <i class="ri-logout-box-line"></i> 退出登录
        </button>
      </div>
    </aside>

    <!-- 右侧主舞台 -->
    <main class="lobby-stage">
      <div class="stage-content">
        <div class="welcome-header">
          <h1>开始协作</h1>
          <p>输入房间号，立即开启灵感同步之旅</p>
        </div>

        <div class="room-entry-card">
          <div class="input-group">
            <i class="ri-hashtag prefix-icon"></i>
            <input
                v-model="roomIdInput"
                type="text"
                placeholder="输入房间 ID..."
                @keyup.enter="handleEnter"
            />
          </div>
          <button class="btn-enter" @click="handleEnter">
            进入房间 <i class="ri-arrow-right-line"></i>
          </button>
        </div>

        <div class="tips">
          <p><i class="ri-lightbulb-line"></i> 提示：相同的房间号将连接到同一个工作区。</p>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.lobby-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background: var(--bg-main);
}

/* 左侧边栏 */
.lobby-sidebar {
  width: var(--sidebar-width);
  background: var(--bg-panel);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  padding: 24px;
  flex-shrink: 0; /* 防止被挤压 */
}

.user-card {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
}

.avatar-box { position: relative; }
.avatar {
  width: 48px; height: 48px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.2rem; font-weight: bold; color: white;
  box-shadow: 0 4px 6px rgba(0,0,0,0.2);
}
.status-indicator {
  position: absolute; bottom: -2px; right: -2px;
  width: 12px; height: 12px;
  background: var(--success-color);
  border: 2px solid var(--bg-panel);
  border-radius: 50%;
}

.user-info h3 { margin: 0; font-size: 1rem; color: var(--text-main); }
.user-role { font-size: 0.75rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; }

.nav-section { flex: 1; overflow-y: auto; }
.section-title {
  font-size: 0.75rem; color: var(--text-muted);
  text-transform: uppercase; margin-bottom: 12px; font-weight: 600;
}

.history-list { list-style: none; padding: 0; margin: 0; }
.history-item {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 12px; border-radius: 8px;
  cursor: pointer; transition: all 0.2s;
  color: var(--text-muted);
}
.history-item:hover { background: var(--bg-hover); color: var(--text-main); transform: translateX(4px); }
.history-item .icon { font-size: 1.1rem; }
.meta { display: flex; flex-direction: column; }
.room-name { font-size: 0.9rem; font-weight: 500; }
.room-time { font-size: 0.75rem; opacity: 0.7; }

.state-text { font-size: 0.85rem; color: var(--text-muted); font-style: italic; padding: 10px 0; }

.sidebar-footer { margin-top: auto; padding-top: 20px; border-top: 1px solid var(--border-color); }
.btn-logout {
  width: 100%; padding: 10px; background: transparent; border: 1px solid var(--border-color);
  color: var(--danger-color); border-radius: 6px; cursor: pointer;
  display: flex; align-items: center; justify-content: center; gap: 8px;
  transition: all 0.2s;
}
.btn-logout:hover { background: rgba(239, 68, 68, 0.1); border-color: var(--danger-color); }

/* 右侧主区域 */
.lobby-stage {
  flex: 1;
  display: flex; align-items: center; justify-content: center;
  background-image: radial-gradient(var(--bg-hover) 1px, transparent 1px);
  background-size: 30px 30px; /* 点阵背景 */
}

.stage-content { width: 100%; max-width: 480px; text-align: center; }

.welcome-header h1 { font-size: 2.5rem; margin: 0 0 8px; background: linear-gradient(to right, white, #94a3b8); -webkit-background-clip: text; color: transparent; }
.welcome-header p { color: var(--text-muted); margin-bottom: 40px; font-size: 1.1rem; }

.room-entry-card {
  display: flex; gap: 12px;
  background: var(--bg-panel); padding: 8px; border-radius: 12px;
  border: 1px solid var(--border-color);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.input-group {
  flex: 1; position: relative; display: flex; align-items: center;
}
.prefix-icon { position: absolute; left: 16px; color: var(--text-muted); font-size: 1.2rem; }
.input-group input {
  width: 100%; padding: 16px 16px 16px 44px; border: none; background: transparent;
  color: white; font-size: 1.1rem; outline: none;
}

.btn-enter {
  padding: 0 32px; border: none; border-radius: 8px;
  background: var(--primary-color); color: white; font-weight: 600; font-size: 1rem;
  cursor: pointer; display: flex; align-items: center; gap: 8px;
  transition: background 0.2s;
}
.btn-enter:hover { background: var(--primary-hover); }

.tips { margin-top: 24px; font-size: 0.85rem; color: var(--text-muted); opacity: 0.7; }
</style>