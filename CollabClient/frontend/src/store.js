import { reactive } from 'vue'

// 默认服务器地址
const DEFAULT_HOST = 'localhost:8080'
const STORE_KEY = 'collab_server_host'

// 从本地存储读取上次使用的地址，如果没有则用默认值
const savedHost = localStorage.getItem(STORE_KEY) || DEFAULT_HOST

// 使用 reactive 让它具有响应性 (虽然在这里不是必须的，但为了规范)
const state = reactive({
    host: savedHost
})

export const serverConfig = {
    // 获取当前的基础地址
    getHost() {
        return state.host
    },

    // 设置新的地址并保存到本地
    setHost(newHost) {
        // 去掉可能误输入的 http:// 前缀
        let cleanHost = newHost.replace(/^https?:\/\//, '').replace(/\/+$/, '')
        if (!cleanHost) cleanHost = DEFAULT_HOST

        state.host = cleanHost
        localStorage.setItem(STORE_KEY, cleanHost)
        console.log(`[Config] Server set to: ${state.host}`)
    },

    // 动态生成 HTTP URL
    getHttpUrl() {
        return `http://${state.host}`
    },

    // 动态生成 WebSocket URL
    getWsUrl() {
        return `ws://${state.host}`
    }
}