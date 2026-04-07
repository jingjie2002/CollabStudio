import { reactive } from 'vue'

const STORE_KEY = 'collab_server_host'
const DEFAULT_SERVER_IP = import.meta.env.VITE_DEFAULT_SERVER_IP || 'localhost'
const DEFAULT_DEV_PORT = '8080'
const DEFAULT_PROD_PORT = '80'

const getHttpProtocol = () => {
    if (window.location.protocol === 'https:') {
        return 'https'
    }
    return 'http'
}

const getWsProtocol = () => {
    return getHttpProtocol() === 'https' ? 'wss' : 'ws'
}

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
        /^172\.(1[6-9]|2[0-9]|3[01])\./.test(ip)
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
 * 动态寻路逻辑 v4.0
 * 智能检测当前环境并返回最佳服务器地址
 */
const getAutoHost = () => {
    const hostname = window.location.hostname

    // 0. Wails 生产环境检测（最高优先级）
    // Wails 构建后，window.location.hostname 返回 'wails.localhost'
    if (hostname === 'wails.localhost' || hostname.endsWith('.wails.localhost')) {
        console.log(`🖥️ [Config] Wails 客户端模式 → 指向服务器 ${DEFAULT_SERVER_IP}`)
        // 本地 Wails 模式下优先连本机
        if (DEFAULT_SERVER_IP === 'localhost') {
            return `localhost:${DEFAULT_DEV_PORT}`
        }
        return DEFAULT_SERVER_IP  // 不带端口 = 隐式 80
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

    // 4. 公网 IP 或域名访问
    if (hostname) {
        // 如果当前页面有端口，使用相同端口；否则公网默认 80
        const port = locationPort || DEFAULT_PROD_PORT
        const result = port === '80' ? hostname : `${hostname}:${port}`
        console.log(`🌐 [Config] 公网/域名模式 → ${result}`)
        return result
    }

    // 5. 兜底默认
    console.log('[Config] Using Default Fallback:', DEFAULT_SERVER_IP)
    return DEFAULT_SERVER_IP
}

// 初始化状态
let initialHost = localStorage.getItem(STORE_KEY)
const hostname = window.location.hostname

// 防御性清理：如果当前页面在公网/局域网，但LocalStorage仍缓存了之前的 localhost，强行清除它，防止浏览器拦截
if (initialHost && (initialHost.includes('localhost') || initialHost.includes('127.0.0.1')) && hostname !== 'localhost' && hostname !== '127.0.0.1' && hostname !== 'wails.localhost') {
    console.warn(`[Config] 侦测到 LocalStorage 缓存了失效的本地地址 ${initialHost}，已自动清除并重新寻路！`)
    initialHost = null
    localStorage.removeItem(STORE_KEY)
}

if (!initialHost) {
    initialHost = getAutoHost()
}

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
        return `${getHttpProtocol()}://${state.host}`
    },

    // 动态生成 WebSocket URL
    getWsUrl() {
        return `${getWsProtocol()}://${state.host}`
    }
}

