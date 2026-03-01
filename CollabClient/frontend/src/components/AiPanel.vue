<template>
  <div class="ai-panel">
    <div class="panel-header">
      <h3><i class="ri-robot-line"></i> AI 助手</h3>
    </div>

    <!-- 快捷操作 -->
    <div class="quick-actions">
      <button @click="doAction('summarize')" :disabled="loading" class="action-chip">
        <i class="ri-file-list-3-line"></i> 总结
      </button>
      <button @click="doAction('translate')" :disabled="loading" class="action-chip">
        <i class="ri-translate"></i> 翻译
      </button>
      <button @click="doAction('polish')" :disabled="loading" class="action-chip">
        <i class="ri-quill-pen-line"></i> 润色
      </button>
    </div>

    <!-- 对话区 -->
    <div class="ai-messages" ref="messagesRef">
      <div v-if="messages.length === 0" class="empty-hint">
        <i class="ri-sparkling-line"></i>
        <p>选择上方操作或直接提问</p>
      </div>
      <div v-for="(msg, idx) in messages" :key="idx"
           class="ai-msg" :class="msg.role">
        <div class="msg-icon">
          <i :class="msg.role === 'user' ? 'ri-user-line' : 'ri-robot-line'"></i>
        </div>
        <div class="msg-body" v-html="formatContent(msg.content)"></div>
      </div>
      <div v-if="loading" class="ai-msg assistant">
        <div class="msg-icon"><i class="ri-robot-line"></i></div>
        <div class="msg-body typing">
          <span class="dot-pulse"></span> 思考中...
        </div>
      </div>
    </div>

    <!-- 输入区 -->
    <div class="ai-input-area">
      <input
        v-model="userInput"
        @keyup.enter="sendMessage"
        placeholder="向 AI 提问..."
        :disabled="loading"
      />
      <button @click="sendMessage" :disabled="loading || !userInput.trim()" class="send-btn">
        <i class="ri-send-plane-fill"></i>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { serverConfig } from '../store'
import { settings } from '../settings'

const props = defineProps({
  // 获取当前编辑器内容的函数
  getEditorContent: { type: Function, default: () => '' }
})

const messages = ref([])
const userInput = ref('')
const loading = ref(false)
const messagesRef = ref(null)

// 快捷操作 → 自动生成 prompt
const actionPrompts = {
  summarize: (content) => `请对以下文档内容进行结构化总结，提取关键要点：\n\n${content}`,
  translate: (content) => `请将以下内容翻译。如果是中文则翻译为英文，如果是英文则翻译为中文，保持原始格式：\n\n${content}`,
  polish: (content) => `请对以下内容进行润色和优化，使其更加专业流畅，保持原意不变：\n\n${content}`,
}

const doAction = (action) => {
  const content = props.getEditorContent()
  if (!content || content.trim().length < 2) {
    alert('编辑器内容为空，请先输入一些文本')
    return
  }
  const prompt = actionPrompts[action](content)
  userInput.value = ''
  sendToAI(prompt, action)
}

const sendMessage = () => {
  const msg = userInput.value.trim()
  if (!msg) return
  userInput.value = ''
  sendToAI(msg)
}

const sendToAI = async (prompt, actionLabel) => {
  // 添加用户消息到对话
  const displayMsg = actionLabel
    ? `[${actionLabel === 'summarize' ? '总结' : actionLabel === 'translate' ? '翻译' : '润色'}] 正在处理文档内容...`
    : prompt
  messages.value.push({ role: 'user', content: displayMsg })
  scrollToBottom()

  loading.value = true
  // 预先创建 assistant 消息占位，逐步填充
  const assistantMsg = { role: 'assistant', content: '' }
  messages.value.push(assistantMsg)

  try {
    const aiMessages = [
      { role: 'system', content: '你是 CollabStudio 的 AI 助手，帮助用户处理文档编辑、翻译和总结任务。回答要简洁专业。' },
      { role: 'user', content: prompt }
    ]

    const baseUrl = serverConfig.getHttpUrl()
    const resp = await fetch(`${baseUrl}/api/ai/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('jwt_token')}`
      },
      body: JSON.stringify({
        messages: aiMessages,
        apiUrl: settings.ai.apiUrl || '',
        apiKey: settings.ai.apiKey || '',
        model: settings.ai.model || '',
      })
    })

    if (!resp.ok) {
      const errData = await resp.json().catch(() => ({}))
      assistantMsg.content = `❌ ${errData.error || '请求失败'}`
      return
    }

    // SSE 流式读取
    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() // 保留未完整的行

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6)
        if (data === '[DONE]') break
        try {
          const chunk = JSON.parse(data)
          if (chunk.content) {
            assistantMsg.content += chunk.content
            scrollToBottom()
          }
        } catch (e) { /* 跳过无法解析的行 */ }
      }
    }

    if (!assistantMsg.content) {
      assistantMsg.content = '❌ AI 未返回有效回复'
    }
  } catch (e) {
    assistantMsg.content = `❌ 连接 AI 服务失败: ${e.message}`
  } finally {
    loading.value = false
    scrollToBottom()
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) messagesRef.value.scrollTop = messagesRef.value.scrollHeight
  })
}

// 简单的 markdown 格式化
const formatContent = (text) => {
  if (!text) return ''
  return text
    .replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br>')
}
</script>

<style scoped>
.ai-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-panel);
}

.panel-header {
  padding: 12px 16px;
  background: rgba(0,0,0,0.1);
  border-bottom: 1px solid var(--border-color);
}

.panel-header h3 {
  margin: 0;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 8px;
  text-transform: uppercase;
}

.quick-actions {
  display: flex;
  gap: 6px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-color);
  flex-wrap: wrap;
}

.action-chip {
  background: var(--bg-hover);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 5px 10px;
  border-radius: 16px;
  font-size: 0.78rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.2s;
}

.action-chip:hover:not(:disabled) {
  background: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

.action-chip:disabled { opacity: 0.5; cursor: not-allowed; }

.ai-messages {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  opacity: 0.5;
  font-size: 0.85rem;
}

.empty-hint i { font-size: 2rem; margin-bottom: 8px; }

.ai-msg {
  display: flex;
  gap: 8px;
  animation: fadeIn 0.3s ease;
}

.msg-icon {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  flex-shrink: 0;
  margin-top: 2px;
}

.ai-msg.user .msg-icon { background: var(--primary-color); color: white; }
.ai-msg.assistant .msg-icon { background: #8b5cf6; color: white; }

.msg-body {
  flex: 1;
  font-size: 0.85rem;
  line-height: 1.5;
  color: var(--text-main);
  word-break: break-word;
}

.msg-body :deep(pre) {
  background: var(--bg-main);
  padding: 8px 10px;
  border-radius: 6px;
  overflow-x: auto;
  font-size: 0.8rem;
  margin: 6px 0;
}

.msg-body :deep(code) {
  background: var(--bg-hover);
  padding: 1px 4px;
  border-radius: 3px;
  font-family: monospace;
  font-size: 0.82rem;
}

.msg-body.typing {
  color: var(--text-muted);
  font-style: italic;
}

.dot-pulse {
  display: inline-block;
  width: 6px;
  height: 6px;
  background: var(--primary-color);
  border-radius: 50%;
  animation: pulse 1s infinite;
  margin-right: 4px;
}

@keyframes pulse {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

.ai-input-area {
  display: flex;
  gap: 6px;
  padding: 10px 12px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-panel);
}

.ai-input-area input {
  flex: 1;
  background: var(--bg-main);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 8px 12px;
  border-radius: 6px;
  outline: none;
  font-size: 0.85rem;
}

.ai-input-area input:focus { border-color: var(--primary-color); }

.ai-input-area .send-btn {
  background: var(--primary-color);
  color: white;
  border: none;
  width: 36px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.ai-input-area .send-btn:hover:not(:disabled) { background: var(--primary-hover); }
.ai-input-area .send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
