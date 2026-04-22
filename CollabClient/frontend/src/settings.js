// =============================================================================
// settings.js — 全局设置状态管理
// =============================================================================
// 所有设置存入 localStorage，支持读取/保存/恢复默认
// =============================================================================

import { reactive, watch } from 'vue'

const STORAGE_KEY = 'collab_settings'

// 默认设置
const DEFAULT_SETTINGS = {
    // 编辑器
    fontSize: 15,
    fontFamily: 'Consolas',

    // 外观
    theme: 'dark', // 'dark' | 'light'

    // AI 配置
    ai: {
        apiUrl: '',
        apiKey: '',
        model: 'deepseek-chat',
    }
}

// 从 localStorage 加载设置
function loadSettings() {
    try {
        const stored = localStorage.getItem(STORAGE_KEY)
        if (stored) {
            const parsed = JSON.parse(stored)
            // 深度合并（防止旧存储缺少新字段）
            return deepMerge(structuredClone(DEFAULT_SETTINGS), parsed)
        }
    } catch (e) {
        console.warn('[Settings] 加载设置失败:', e)
    }
    return structuredClone(DEFAULT_SETTINGS)
}

// 深度合并
function deepMerge(target, source) {
    for (const key in source) {
        if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
            if (!target[key]) target[key] = {}
            deepMerge(target[key], source[key])
        } else {
            target[key] = source[key]
        }
    }
    return target
}

// 响应式设置对象
export const settings = reactive(loadSettings())

// 自动保存到 localStorage
watch(settings, (newVal) => {
    try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(newVal))
    } catch (e) {
        console.warn('[Settings] 保存设置失败:', e)
    }
}, { deep: true })

// 监听字体变化 → 全局应用
watch(() => [settings.fontSize, settings.fontFamily], () => {
    applyFont()
})

// 恢复默认设置
export function resetSettings() {
    const defaults = structuredClone(DEFAULT_SETTINGS)
    Object.assign(settings, defaults)
    applyTheme(settings.theme)
    applyFont()
}

// 应用主题到 DOM
export function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme)
    settings.theme = theme
}

// 应用字体到全局
export function applyFont() {
    const fontSize = settings.fontSize + 'px'
    const fontFamily = settings.fontFamily + ', sans-serif'
    const documentRoot = document.documentElement
    if (documentRoot) {
        documentRoot.style.fontSize = fontSize
        documentRoot.style.setProperty('--app-font-size', fontSize)
        documentRoot.style.setProperty('--app-font-family', fontFamily)
    }

    if (document.body) {
        document.body.style.fontFamily = fontFamily
    }

    const root = document.getElementById('app-root')
    if (root) {
        root.style.fontSize = '1rem'
        root.style.fontFamily = fontFamily
    }
}

// 切换主题
export function toggleTheme() {
    const next = settings.theme === 'dark' ? 'light' : 'dark'
    applyTheme(next)
}

// 初始化（在 App.vue onMounted 中调用）
export function initSettings() {
    applyTheme(settings.theme)
    applyFont()
}
