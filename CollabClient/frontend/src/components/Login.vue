<template>
  <div class="login-container fade-in">
    <div class="login-card">
      <div class="logo-area">
        <h1>CollabStudio</h1>
        <p>实时协作 · 无界沟通</p>
      </div>

      <div class="auth-tabs">
        <span :class="{ active: !isRegister }" @click="isRegister = false">登录</span>
        <span :class="{ active: isRegister }" @click="isRegister = true">注册</span>
      </div>

      <div class="input-group">
        <div class="input-wrapper">
          <i class="ri-user-line icon"></i>
          <input
              v-model="username"
              placeholder="用户名"
              type="text"
              class="login-input"
          />
        </div>

        <div class="input-wrapper">
          <i class="ri-lock-line icon"></i>
          <input
              v-model="password"
              @keyup.enter="handleAuth"
              placeholder="密码"
              type="password"
              class="login-input"
          />
        </div>

        <button @click="handleAuth" class="login-btn" :disabled="loading">
          <span v-if="loading"><i class="ri-loader-4-line spinning"></i> 处理中...</span>
          <span v-else>
            {{ isRegister ? '立即注册' : '进入工作台' }}
            <i class="ri-arrow-right-line"></i>
          </span>
        </button>
      </div>

      <div class="server-config">
        <div class="config-header" @click="showConfig = !showConfig">
          <i class="ri-settings-3-line"></i>
          <span>{{ showConfig ? '收起配置' : '服务器设置' }}</span>
        </div>

        <div v-if="showConfig" class="config-body fade-in">
          <label>服务器地址 (IP:端口)</label>
          <input
              v-model="serverAddress"
              placeholder="例如 192.168.1.5:8080"
              class="config-input"
          />
          <small>默认: localhost:8080 (本机)</small>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { serverConfig } from '../store'

const emit = defineEmits(['login'])

const isRegister = ref(false) // false=登录, true=注册
const username = ref('')
const password = ref('')
const serverAddress = ref('')
const showConfig = ref(false)
const loading = ref(false)

onMounted(() => {
  serverAddress.value = serverConfig.getHost()
})

const handleAuth = async () => {
  if (!username.value.trim() || !password.value.trim()) {
    alert("请输入用户名和密码")
    return
  }

  // 1. 先更新服务器地址配置
  serverConfig.setHost(serverAddress.value)

  loading.value = true
  const baseUrl = serverConfig.getHttpUrl()
  const endpoint = isRegister.value ? '/register' : '/login'

  try {
    // 2. 发起真实的 HTTP 请求验证账号
    const response = await fetch(`${baseUrl}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: username.value,
        password: password.value
      })
    })

    const data = await response.json()

    if (response.ok) {
      // 成功！
      if (isRegister.value) {
        alert("注册成功，请登录！")
        isRegister.value = false // 切换回登录页
      } else {
        // 登录成功，通知父组件进入 Workspace
        emit('login', username.value)
      }
    } else {
      // 失败 (如密码错误)
      alert(data.error || "操作失败，请检查网络或服务器")
    }
  } catch (e) {
    console.error(e)
    alert("连接服务器失败，请检查 IP 配置是否正确。\n当前地址: " + baseUrl)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1e1e2e 0%, #181825 100%);
  color: #cdd6f4;
}

.login-card {
  background: rgba(255, 255, 255, 0.05);
  padding: 2.5rem;
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}

.logo-area {
  text-align: center;
  margin-bottom: 1.5rem;
}

.logo-area h1 {
  font-size: 2rem;
  margin: 0;
  background: linear-gradient(to right, #89b4fa, #cba6f7);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-weight: 800;
}

.logo-area p { color: #a6adc8; margin-top: 0.5rem; font-size: 0.9rem; }

/* Tabs */
.auth-tabs {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.auth-tabs span {
  padding-bottom: 8px;
  cursor: pointer;
  color: #a6adc8;
  font-weight: 600;
  transition: all 0.3s;
  border-bottom: 2px solid transparent;
}

.auth-tabs span:hover { color: #cdd6f4; }

.auth-tabs span.active {
  color: #89b4fa;
  border-bottom-color: #89b4fa;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.icon {
  position: absolute;
  left: 12px;
  color: #a6adc8;
  font-size: 1.1rem;
}

.login-input {
  width: 100%;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 12px 12px 12px 40px; /* Space for icon */
  border-radius: 8px;
  color: white;
  font-size: 1rem;
  outline: none;
  transition: border-color 0.3s;
}

.login-input:focus { border-color: #89b4fa; }

.login-btn {
  background: #89b4fa;
  color: #1e1e2e;
  border: none;
  padding: 12px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: transform 0.1s, background 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 0.5rem;
}

.login-btn:hover:not(:disabled) { background: #b4befe; transform: translateY(-1px); }
.login-btn:active:not(:disabled) { transform: translateY(1px); }
.login-btn:disabled { opacity: 0.7; cursor: not-allowed; }

/* Server Config */
.server-config {
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}
.config-header {
  display: flex; align-items: center; gap: 6px;
  color: #6c7086; font-size: 0.85rem; cursor: pointer; justify-content: center;
}
.config-header:hover { color: #a6adc8; }
.config-body { margin-top: 1rem; display: flex; flex-direction: column; gap: 6px; }
.config-body label { font-size: 0.8rem; color: #a6adc8; }
.config-input {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.05);
  padding: 8px 12px; border-radius: 6px;
  color: #cdd6f4; font-family: monospace; font-size: 0.9rem;
}
.config-body small { font-size: 0.75rem; color: #45475a; }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { 100% { transform: rotate(360deg); } }
.fade-in { animation: fadeIn 0.3s ease-out; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
</style>