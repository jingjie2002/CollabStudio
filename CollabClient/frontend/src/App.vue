<template>
  <div id="app-root">
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
  </div>
</template>

<script setup>
import { ref } from 'vue'
import Login from './components/Login.vue'
import Lobby from './components/Lobby.vue' // 🟢 确保这个文件还在 components 目录下
import Workspace from './components/Workspace.vue'

// 视图状态：login -> lobby -> workspace
const currentView = ref('login')
const currentUser = ref(null)
const targetRoom = ref('demo-room')

// 处理登录成功
const handleLoginSuccess = (username) => {
  console.log("[App] Login success:", username)
  currentUser.value = username
  // 🟢 关键修正：登录后去大厅，而不是直接去工作台
  currentView.value = 'lobby'
}

// 处理进入房间
const handleEnterRoom = (roomId) => {
  console.log("[App] Entering room:", roomId)
  if (roomId) {
    targetRoom.value = roomId
  }
  // 🟢 关键修正：从大厅跳转到工作台
  currentView.value = 'workspace'
}

// 处理退出登录
const handleLogout = () => {
  console.log("[App] User logged out")
  currentUser.value = null
  currentView.value = 'login'
}
</script>

<style>
/* 全局样式保持不变 */
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
</style>