import { reactive } from 'vue'

const STORE_KEY = 'collab_server_host'
const DEFAULT_HOST_IP = '119.29.55.127'
const DEFAULT_REMOTE_IP = '119.29.55.127'
const DEFAULT_DEV_PORT = '8080'
const DEFAULT_PROD_PORT = '80'

/**
 * 检测是否为有效 IP 地址
 */
const isValidIP = (str) => {
    const ipv4Regex = /^(\d{1,3}\.){3}\d{1,3}$/
    return ipv4Regex.test(str)
}

/**
 * 检测是否为局域网 IP
 */
const isLanIP = (ip) => {
    if (!isValidIP(ip)) return false
    return ip.startsWith('192.168.') ||
        ip.startsWith('10.') ||
        ip.startsWith('172.16.') ||
        ip.startsWith('172.17.') ||
        ip.startsWith('172.18.') ||
        ip.startsWith('172.19.') ||
        ip.startsWith('172.20.') ||
        ip.startsWith('172.21.') ||
        ip.startsWith('172.22.') ||
        ip.startsWith('172.23.') ||
        ip.startsWith('172.24.') ||
        ip.startsWith('172.25.') ||
        ip.startsWith('172.26.') ||
        ip.startsWith('172.27.') ||
        ip.startsWith('172.28.') ||
        ip.startsWith('172.29.') ||
        ip.startsWith('172.30.') ||
        ip.startsWith('172.31.')
}

/**
 * 智能端口推断
 * @param {string} host - 主机名或 IP
 * @returns {string} 推荐端口
 */
const inferPort = (host) => {
    // 已包含端口则不添加
    if (host.includes(':')) return ''

    // 本地开发 → 8080
    if (host === 'localhost' || host === '127.0.0.1') {
        return DEFAULT_DEV_PORT
    }

    // 局域网 IP → 8080
    if (isLanIP(host)) {
        return DEFAULT_DEV_PORT
    }

    // 公网 IP 或域名 → 80
    return DEFAULT_PROD_PORT
}

/**
 * 动态寻路逻辑 v3.9.0
 * 智能检测当前环境并返回最佳服务器地址
 */
const getAutoHost = () => {
    const hostname = window.location.hostname

    // 0. Wails 生产环境检测（最高优先级）
    // Wails 构建后，window.location.hostname 返回 'wails.localhost'
    if (hostname === 'wails.localhost' || hostname.endsWith('.wails.localhost')) {
        console.log(`🖥️ [Config] Wails 客户端模式 → 强制指向生产服务器 ${DEFAULT_REMOTE_IP}`)
        return DEFAULT_REMOTE_IP  // 不带端口 = 隐式 80
    }

    // 1. 优先使用环境变量（构建时注入）
    if (import.meta.env.VITE_APP_API_URL) {
        console.log('[Config] Using Env Host:', import.meta.env.VITE_APP_API_URL)
        return import.meta.env.VITE_APP_API_URL
    }

    const locationPort = window.location.port

    // 2. 本地开发环境 (localhost / 127.0.0.1)
    if (hostname === 'localhost' || hostname === '127.0.0.1') {
        const result = `localhost:${DEFAULT_DEV_PORT}`
        console.log(`🔌 [Config] 本地开发模式 → ${result}`)
        return result
    }

    // 3. 局域网 IP 访问
    if (isLanIP(hostname)) {
        const result = `${hostname}:${DEFAULT_DEV_PORT}`
        console.log(`🔌 [Config] 局域网模式 → ${result}`)
        return result
    }

    // 4. 腾讯云公网 IP 显式匹配（强制 80 端口，省略显示）
    if (hostname === DEFAULT_REMOTE_IP || hostname === '119.29.55.127') {
        console.log(`🌐 [Config] 腾讯云公网 IP → ${hostname} (端口 80)`)
        return hostname  // 不带端口 = 隐式 80
    }

    // 5. 其他公网 IP 或域名访问
    if (hostname) {
        // 如果当前页面有端口，使用相同端口；否则公网默认 80
        const port = locationPort || DEFAULT_PROD_PORT
        const result = port === '80' ? hostname : `${hostname}:${port}`
        console.log(`🌐 [Config] 公网/域名模式 → ${result}`)
        return result
    }

    // 6. 兜底默认 IP（不带端口，公网默认80）
    console.log('[Config] Using Default Fallback:', DEFAULT_REMOTE_IP)
    return DEFAULT_REMOTE_IP
}

// 初始化状态
// 如果用户手动设置过 (localStorage)，优先尊重用户选择？
// 题目要求 "动态寻路逻辑... 优先级..."，通常意味着自动检测优先，或者默认值策略。
// 这里采用：如果 LocalStorage 有值且有效，使用它；否则使用自动检测。
const initialHost = localStorage.getItem(STORE_KEY) || getAutoHost()

const state = reactive({
    host: initialHost
})

export const serverConfig = {
    // 获取当前 host
    getHost() {
        return state.host
    },

    // 手动设置 host (例如在设置界面修改)
    setHost(newHost) {
        let cleanHost = newHost.replace(/^https?:\/\//, '').replace(/\/+$/, '')
        if (!cleanHost) cleanHost = getAutoHost()

        state.host = cleanHost
        localStorage.setItem(STORE_KEY, cleanHost)
        console.log(`[Config] Server set to: ${state.host}`)
    },

    // 动态生成 HTTP URL
    getHttpUrl() {
        // 如果 host 已经包含端口，直接使用；否则默认 80 (浏览器会自动隐藏 :80)
        // 注意：getAutoHost 返回的可能是 IP，也可能是 IP:Port
        // 如果是纯 IP，构建时加上 http://
        return `http://${state.host}`
    },

    // 动态生成 WebSocket URL
    getWsUrl() {
        return `ws://${state.host}`
    }
}
