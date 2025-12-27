<script setup>
import { ref } from 'vue'
import axios from 'axios'

const emit = defineEmits(['register-success', 'switch-to-login'])

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

const handleRegister = async () => {
  // 基础验证
  if (!username.value || !password.value) {
    errorMessage.value = "请输入用户名和密码"
    return
  }
  if (password.value !== confirmPassword.value) {
    errorMessage.value = "两次输入的密码不一致"
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    // 调用后端注册接口
    await axios.post('http://localhost:8080/register', {
      username: username.value,
      password: password.value
    })

    // 注册成功，通知父组件（可以自动填入账号，或者直接跳回登录页）
    alert("注册成功！请登录")
    emit('register-success', username.value)

  } catch (error) {
    if (error.response && error.response.data) {
      errorMessage.value = error.response.data.error || "注册失败"
    } else {
      errorMessage.value = "网络错误，请检查后端服务"
    }
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="register-container">
    <div class="register-box">
      <h2 class="title">Create Account</h2>
      <p class="subtitle">加入 CollabStudio 开始协作</p>

      <div class="input-group">
        <label>Username</label>
        <input v-model="username" type="text" placeholder="设置用户名" />
      </div>

      <div class="input-group">
        <label>Password</label>
        <input v-model="password" type="password" placeholder="设置密码" />
      </div>

      <div class="input-group">
        <label>Confirm Password</label>
        <input v-model="confirmPassword" type="password" placeholder="确认密码" />
      </div>

      <div v-if="errorMessage" class="error-msg">
        {{ errorMessage }}
      </div>

      <button :disabled="isLoading" @click="handleRegister" class="btn-register">
        {{ isLoading ? '注册中...' : '立即注册' }}
      </button>

      <div class="footer-links">
        <span>已有账号? </span>
        <a href="#" @click.prevent="$emit('switch-to-login')">去登录</a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #2c3e50, #000000);
  color: white;
}

.register-box {
  background: rgba(255, 255, 255, 0.1);
  padding: 40px;
  border-radius: 10px;
  backdrop-filter: blur(10px);
  width: 320px;
  box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.37);
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.title {
  margin: 0;
  text-align: center;
  font-size: 24px;
  color: #ecf0f1;
}

.subtitle {
  text-align: center;
  color: #bdc3c7;
  font-size: 14px;
  margin-bottom: 20px;
}

.input-group {
  margin-bottom: 15px;
}

.input-group label {
  display: block;
  font-size: 12px;
  margin-bottom: 5px;
  color: #bdc3c7;
}

.input-group input {
  width: 100%;
  padding: 10px;
  border-radius: 5px;
  border: 1px solid #34495e;
  background-color: #2c3e50;
  color: white;
  outline: none;
  box-sizing: border-box; /* 关键：防止 padding 撑大宽度 */
}

.input-group input:focus {
  border-color: #3498db;
}

.btn-register {
  width: 100%;
  padding: 12px;
  background-color: #27ae60;
  border: none;
  border-radius: 5px;
  color: white;
  font-weight: bold;
  cursor: pointer;
  margin-top: 10px;
  transition: background 0.3s;
}

.btn-register:hover {
  background-color: #2ecc71;
}

.btn-register:disabled {
  background-color: #7f8c8d;
  cursor: not-allowed;
}

.error-msg {
  color: #e74c3c;
  font-size: 12px;
  text-align: center;
  margin-bottom: 10px;
}

.footer-links {
  margin-top: 15px;
  text-align: center;
  font-size: 13px;
}

.footer-links a {
  color: #3498db;
  text-decoration: none;
}

.footer-links a:hover {
  text-decoration: underline;
}
</style>