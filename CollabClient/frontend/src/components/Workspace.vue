<template>
  <div class="workspace-layout fade-in">
    <!-- ================= 1. 顶部导航栏 ================= -->
    <header class="navbar">
      <div class="nav-left">
        <div class="logo">CollabStudio</div>
        <div class="room-pill">
          <i class="ri-hashtag"></i>
          <span>{{ roomID }}</span>
        </div>
        <!-- 🟢 房主标识：只有房主才会显示 -->
        <div v-if="isHost" class="host-badge" title="你是房主，关闭窗口会导致全员掉线">
          <i class="ri-vip-crown-fill"></i> 房主
        </div>
      </div>

      <div class="nav-right">
        <div class="action-group">
          <button @click="handleSaveFile" class="nav-btn" title="保存到本地">
            <i class="ri-save-3-line"></i>
          </button>
          <button @click="handleOpenFile" class="nav-btn" title="导入文件">
            <i class="ri-folder-open-line"></i>
          </button>
        </div>

        <div class="divider-v"></div>

        <div class="connection-status" :class="{ online: isConnected }" :title="isConnected ? '连接正常' : '已离线'">
          <span class="dot"></span>
          {{ isConnected ? '已连接' : '离线' }}
        </div>

        <div class="user-badge">{{ username }}</div>
        <button @click="$emit('logout')" class="nav-btn danger" title="退出">
          <i class="ri-logout-box-r-line"></i>
        </button>
      </div>
    </header>

    <!-- ================= 2. 主体内容区 ================= -->
    <div class="main-content">
      <!-- 左侧：编辑器 -->
      <main class="editor-area">
        <div class="editor-wrapper">
          <Editor
              ref="editorRef"
              @update="handleDocChange"
              @cursor-update="handleCursorMove"
              @check-connection="handleCheckConnection"
          />
        </div>
      </main>

      <!-- 右侧：侧边栏 -->
      <aside class="sidebar">
        <!-- 上半部分：用户列表 -->
        <div class="panel users-panel">
          <div class="panel-header">
            <h3><i class="ri-group-line"></i> 在线成员 ({{ onlineUsers.length }})</h3>
          </div>
          <div class="user-list">
            <div v-for="(user, idx) in onlineUsers" :key="idx" class="user-row">
              <div class="avatar-mini" :style="{ backgroundColor: stringToColor(user) }">
                {{ user.charAt(0).toUpperCase() }}
              </div>
              <span class="username-text">{{ user }}</span>
              <span v-if="user === username" class="me-tag">我</span>
            </div>
          </div>
        </div>

        <!-- 下半部分：聊天室 -->
        <div class="panel chat-panel">
          <div class="panel-header">
            <h3><i class="ri-chat-3-line"></i> 讨论区</h3>
          </div>

          <div class="chat-messages" ref="chatBoxRef">
            <div v-for="(msg, idx) in chatMessages" :key="idx"
                 class="message-bubble"
                 :class="{ 'my-message': msg.sender === username, 'other-message': msg.sender !== username }">
              <div class="msg-meta" v-if="msg.sender !== username">{{ msg.sender }}</div>

              <!-- 图片消息处理 -->
              <div class="msg-content" v-if="msg.text.startsWith('image:')">
                <img
                    :src="getImageUrl(msg.text.substring(6))"
                    class="chat-image"
                    @click="previewImage(msg.text.substring(6))"
                    title="点击在浏览器打开"
                />
              </div>
              <!-- 文本消息处理 -->
              <div class="msg-content" v-else>{{ msg.text }}</div>
            </div>
          </div>

          <!-- 聊天输入框区域 -->
          <div class="chat-footer">
            <!-- 表情选择器弹窗 -->
            <div v-if="showEmojiPicker" class="emoji-picker fade-in">
              <span v-for="emoji in emojiList" :key="emoji" @click="insertEmoji(emoji)">{{ emoji }}</span>
            </div>

            <!-- 工具栏 -->
            <div class="toolbar-row">
              <button class="icon-btn" @click="showEmojiPicker = !showEmojiPicker"><i class="ri-emotion-line"></i></button>
              <button class="icon-btn" @click="triggerChatImage"><i class="ri-image-line"></i></button>
              <!-- 隐藏的文件上传 Input -->
              <input type="file" ref="chatFileInput" style="display:none" accept="image/*" @change="handleChatImageUpload">
            </div>

            <!-- 输入框与发送按钮 -->
            <div class="input-row">
              <input v-model="chatInput" @keyup.enter="sendChatMessage" placeholder="输入消息..." />
              <button @click="sendChatMessage" class="send-btn"><i class="ri-send-plane-fill"></i></button>
            </div>
          </div>
        </div>
      </aside>
    </div>

    <!-- ================= 3. 房主退出警告弹窗 ================= -->
    <div v-if="showExitModal" class="modal-overlay fade-in">
      <div class="modal-content">
        <div class="modal-icon"><i class="ri-alert-fill"></i></div>
        <h3>确定要关闭吗？</h3>
        <p>你是当前的 <strong>房主</strong>。</p>
        <p class="warning-text">如果关闭窗口，房间将解散，所有其他成员都会被迫离线！</p>
        <div class="modal-actions">
          <button @click="showExitModal = false" class="btn cancel">取消</button>
          <button @click="confirmExit" class="btn confirm">解散并退出</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import Editor from './Editor.vue'
// 引入所有需要的后端 Go 方法
import { SaveFile, OpenFile, IsHostUser, ConfirmExit } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime'
import { serverConfig } from '../store'
import 'remixicon/fonts/remixicon.css'

const props = defineProps({
  username: { type: String, default: 'Guest' },
  initialRoom: { type: String, default: 'demo-room' }
})

// --- 状态变量定义 ---
const editorRef = ref(null)
const chatBoxRef = ref(null)
const socket = ref(null)
const isConnected = ref(false)
const roomID = ref(props.initialRoom ? props.initialRoom.trim() : 'demo-room')

const chatInput = ref('')
const chatMessages = ref([])
const onlineUsers = ref([])
const remoteCursors = new Map()

const showEmojiPicker = ref(false)
const chatFileInput = ref(null)
const emojiList = ['😀','😂','😅','🥰','😎','🤔','😐','😭','😱','😡','👍','👎','👋','🙏','🚀','🔥','🎉','❤️','💔','💩']

// 节流控制 (防止打字太快刷屏)
const isThrottled = ref(false)
const pendingUpdate = ref(null)
const THROTTLE_DELAY = 40

// 房主保护状态
const isHost = ref(false)
const showExitModal = ref(false)

// --- 辅助函数 ---
const stringToColor = (str) => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = str.charCodeAt(i) + ((hash << 5) - hash);
  const c = (hash & 0x00ffffff).toString(16).toUpperCase();
  return '#' + '00000'.substring(0, 6 - c.length) + c;
}

const handleCheckConnection = () => {
  if (!socket.value || socket.value.readyState === WebSocket.CLOSED) connectWebSocket()
}

// 获取完整图片路径 (自动补全服务器 IP)
const getImageUrl = (path) => {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${serverConfig.getHttpUrl()}${path}`
}

// 在浏览器中预览图片
const previewImage = (path) => {
  const fullUrl = getImageUrl(path)
  window.open(fullUrl, '_blank')
}

// 智能 JSON 解析 (处理粘包)
const smartJSONParse = (str) => {
  const results = []
  let depth = 0
  let start = 0
  try { results.push(JSON.parse(str)); return results } catch (e) { }
  for (let i = 0; i < str.length; i++) {
    if (str[i] === '{') { if (depth === 0) start = i; depth++ }
    else if (str[i] === '}') { depth--; if (depth === 0) { try { results.push(JSON.parse(str.substring(start, i + 1))) } catch (e) {} } }
  }
  return results
}

// --- WebSocket 核心逻辑 ---
const connectWebSocket = () => {
  if (socket.value && socket.value.readyState === WebSocket.CONNECTING) return
  if (socket.value) socket.value.close()

  const safeUsername = encodeURIComponent(props.username)
  const wsUrl = `${serverConfig.getWsUrl()}/ws?room=${roomID.value}&username=${safeUsername}`

  console.log(`[WS] Connecting: ${wsUrl}`)
  socket.value = new WebSocket(wsUrl)

  socket.value.onopen = () => {
    isConnected.value = true
    chatMessages.value.push({ sender: 'System', text: `已连接到房间: ${roomID.value}` })
  }

  socket.value.onmessage = (event) => {
    const payloads = smartJSONParse(event.data)
    payloads.forEach(payload => {
      try {
        if (payload.type === 'user_list') {
          onlineUsers.value = payload.users || []
          const currentUsers = new Set(onlineUsers.value)
          for (const key of remoteCursors.keys()) {
            if (!currentUsers.has(key)) remoteCursors.delete(key)
          }
          flushCursors()
        }
        else if (payload.type === 'doc_update') {
          if (payload.sender === props.username) return
          if (editorRef.value) {
            editorRef.value.setContent(payload.content)
          }
        }
        else if (payload.type === 'chat') {
          chatMessages.value.push({ sender: payload.sender, text: payload.message })
          scrollToBottom()
        }
        else if (payload.type === 'chat_history') {
          if (chatMessages.value.length <= 1 && payload.history) {
            const history = payload.history.map(msg => ({
              sender: msg.sender || msg.Sender,
              text: msg.content || msg.Content || msg.message || ""
            }))
            chatMessages.value = [...chatMessages.value, ...history]
            scrollToBottom()
          }
        }
        else if (payload.type === 'cursor_update') {
          if (payload.sender === props.username) return
          remoteCursors.set(payload.sender, payload.cursor)
          flushCursors()
        }
      } catch (e) { console.error('Payload Error:', e) }
    })
  }

  socket.value.onclose = () => { isConnected.value = false }
}

const flushCursors = () => {
  if (!editorRef.value) return
  const list = []
  remoteCursors.forEach((pos, name) => list.push({ username: name, cursorVal: pos }))
  editorRef.value.updateCursors(list)
}

// --- 文档同步 (带节流) ---
const handleDocChange = (content) => {
  if (!socket.value || !isConnected.value) return
  if (!isThrottled.value) {
    sendDocUpdate(content)
    enterThrottle()
  } else {
    pendingUpdate.value = content
  }
}

const sendDocUpdate = (content) => {
  try {
    socket.value.send(JSON.stringify({ type: 'doc_update', content, sender: props.username }))
  } catch (e) { console.error(e) }
}

const enterThrottle = () => {
  isThrottled.value = true
  setTimeout(() => {
    isThrottled.value = false
    if (pendingUpdate.value !== null) {
      const content = pendingUpdate.value
      pendingUpdate.value = null
      sendDocUpdate(content)
      enterThrottle()
    }
  }, THROTTLE_DELAY)
}

const handleCursorMove = (cursorPos) => {
  if (socket.value && isConnected.value) {
    socket.value.send(JSON.stringify({ type: 'cursor_update', cursor: cursorPos, sender: props.username }))
  }
}

// --- 文件操作 ---
const handleOpenFile = async () => {
  try {
    const content = await OpenFile()
    if (content !== 'CANCELLED') {
      editorRef.value.setContent(content)
      handleDocChange(content)
    }
  } catch(e) {}
}

const handleSaveFile = async () => {
  if (editorRef.value) await SaveFile(editorRef.value.getText())
}

// --- 聊天功能 ---
const sendChatMessage = () => {
  if (!chatInput.value.trim() || !socket.value) return
  socket.value.send(JSON.stringify({ type: 'chat', message: chatInput.value, sender: props.username }))
  chatInput.value = ''
  showEmojiPicker.value = false
}

const insertEmoji = (emoji) => { chatInput.value += emoji }
const triggerChatImage = () => { chatFileInput.value.click() }

const handleChatImageUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return
  const formData = new FormData()
  formData.append('image', file)
  try {
    const response = await fetch(`${serverConfig.getHttpUrl()}/upload`, { method: 'POST', body: formData })
    const data = await response.json()
    if (data.url) {
      const msgContent = `image:${data.url}`
      socket.value.send(JSON.stringify({ type: 'chat', message: msgContent, sender: props.username }))
    }
  } catch (e) { alert("图片发送失败") }
  event.target.value = ''
}

const scrollToBottom = () => { nextTick(() => { if (chatBoxRef.value) chatBoxRef.value.scrollTop = chatBoxRef.value.scrollHeight }) }

// --- 生命周期 & 房主逻辑 ---
onMounted(async () => {
  connectWebSocket()

  // 初始化：检查房主身份
  try {
    isHost.value = await IsHostUser()
    console.log("[Workspace] Is Host?", isHost.value)
  } catch (e) { console.error("Check Host failed:", e) }

  // 监听退出警告
  EventsOn("show-exit-warning", () => {
    showExitModal.value = true
  })
})

// 确认退出
const confirmExit = () => {
  ConfirmExit()
}

onUnmounted(() => { if (socket.value) socket.value.close() })
</script>

<style scoped>
/* 基础布局 */
.workspace-layout { display: flex; flex-direction: column; height: 100vh; background: var(--bg-main); color: var(--text-main); overflow: hidden; }

/* 顶部导航栏 */
.navbar { height: var(--header-height); background: var(--bg-panel); border-bottom: 1px solid var(--border-color); display: flex; align-items: center; justify-content: space-between; padding: 0 16px; flex-shrink: 0; }
.nav-left { display: flex; align-items: center; gap: 16px; }
.logo { font-weight: 800; font-size: 1.1rem; color: var(--text-main); letter-spacing: -0.5px; }
.room-pill { display: flex; align-items: center; gap: 6px; background: var(--bg-main); padding: 4px 10px; border-radius: 6px; border: 1px solid var(--border-color); color: var(--text-muted); font-size: 0.85rem; font-family: monospace; }

/* 房主标识样式 */
.host-badge {
  background: linear-gradient(45deg, #f59e0b, #d97706);
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: bold;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: 8px;
  cursor: help;
}

.nav-right { display: flex; align-items: center; gap: 12px; }
.action-group { display: flex; gap: 4px; }
.divider-v { width: 1px; height: 20px; background: var(--border-color); margin: 0 4px; }

/* 按钮通用样式 */
.nav-btn { background: transparent; border: none; color: var(--text-muted); cursor: pointer; padding: 6px 8px; border-radius: 6px; font-size: 1.2rem; transition: all 0.2s; display: flex; align-items: center; justify-content: center; }
.nav-btn:hover { background: var(--bg-hover); color: var(--text-main); }
.nav-btn.danger:hover { background: rgba(239, 68, 68, 0.15); color: var(--danger-color); }

/* 连接状态指示器 */
.connection-status { display: flex; align-items: center; gap: 6px; font-size: 0.8rem; color: var(--danger-color); }
.connection-status.online { color: var(--success-color); }
.dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }

/* 主布局 */
.main-content { flex: 1; display: flex; min-height: 0; }
.editor-area { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.editor-wrapper { flex: 1; overflow: hidden; position: relative; }
.sidebar { width: var(--sidebar-width); background: var(--bg-panel); border-left: 1px solid var(--border-color); display: flex; flex-direction: column; flex-shrink: 0; }

/* 面板通用 */
.panel { display: flex; flex-direction: column; }
.panel-header { padding: 12px 16px; background: rgba(0,0,0,0.1); border-bottom: 1px solid var(--border-color); }
.panel-header h3 { margin: 0; font-size: 0.8rem; font-weight: 600; color: var(--text-muted); display: flex; align-items: center; gap: 8px; text-transform: uppercase; }

/* 用户列表 */
.users-panel { height: 35%; border-bottom: 1px solid var(--border-color); }
.user-list { flex: 1; overflow-y: auto; padding: 12px; }
.user-row { display: flex; align-items: center; gap: 10px; padding: 6px; border-radius: 6px; font-size: 0.9rem; }
.avatar-mini { width: 24px; height: 24px; border-radius: 6px; display: flex; align-items: center; justify-content: center; font-size: 0.75rem; font-weight: bold; color: white; }
.me-tag { margin-left: auto; font-size: 0.7rem; background: var(--bg-hover); padding: 2px 6px; border-radius: 4px; color: var(--text-muted); }

/* 聊天面板 */
.chat-panel { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.chat-messages { flex: 1; overflow-y: auto; padding: 16px; display: flex; flex-direction: column; gap: 12px; }

/* 消息气泡 */
.message-bubble { max-width: 85%; padding: 8px 12px; border-radius: 8px; font-size: 0.9rem; word-break: break-word; line-height: 1.4; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.my-message { align-self: flex-end; background: var(--primary-color); color: white; border-bottom-right-radius: 2px; }
.other-message { align-self: flex-start; background: var(--bg-hover); color: var(--text-main); border-bottom-left-radius: 2px; }
.msg-meta { font-size: 0.7rem; color: rgba(255,255,255,0.7); margin-bottom: 4px; }

.chat-image { max-width: 100%; border-radius: 4px; cursor: pointer; margin-top: 4px; }

/* 聊天输入区 */
.chat-footer { padding: 10px; border-top: 1px solid var(--border-color); background: var(--bg-panel); position: relative; }
.toolbar-row { display: flex; gap: 8px; margin-bottom: 8px; }
.icon-btn { background: none; border: none; color: var(--text-muted); cursor: pointer; font-size: 1.1rem; padding: 4px; transition: color 0.2s; }
.icon-btn:hover { color: var(--primary-color); }

.input-row { display: flex; gap: 8px; }
.input-row input { flex: 1; background: var(--bg-main); border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 12px; border-radius: 6px; outline: none; }
.input-row input:focus { border-color: var(--primary-color); }
.send-btn { background: var(--primary-color); color: white; border: none; width: 36px; border-radius: 6px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: background 0.2s; }
.send-btn:hover { background: var(--primary-hover); }

.emoji-picker { position: absolute; bottom: 100%; left: 10px; background: var(--bg-panel); border: 1px solid var(--border-color); border-radius: 8px; padding: 10px; width: 240px; display: flex; flex-wrap: wrap; gap: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.3); z-index: 10; max-height: 200px; overflow-y: auto; }
.emoji-picker span { font-size: 1.4rem; cursor: pointer; transition: transform 0.1s; padding: 4px; border-radius: 4px; }
.emoji-picker span:hover { background: var(--bg-hover); transform: scale(1.1); }

/* 退出警告弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-content {
  background: #2a2a3e;
  padding: 2rem;
  border-radius: 12px;
  width: 90%;
  max-width: 360px;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 10px 30px rgba(0,0,0,0.5);
}

.modal-icon {
  font-size: 3rem;
  color: #ef4444;
  margin-bottom: 1rem;
}

.modal-content h3 { margin: 0 0 0.5rem 0; color: #fff; }
.modal-content p { color: #a6adc8; font-size: 0.9rem; margin: 0.5rem 0; }
.warning-text { color: #ef4444 !important; font-weight: bold; }

.modal-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
  justify-content: center;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  transition: opacity 0.2s;
}

.btn.cancel { background: rgba(255,255,255,0.1); color: #fff; }
.btn.confirm { background: #ef4444; color: #fff; }
.btn:hover { opacity: 0.9; }
</style>