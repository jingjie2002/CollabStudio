import { defineConfig, searchForWorkspaceRoot } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url' // 🟢 新增

// 🟢 手动定义 __dirname (适配 type: module)
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue()],
    resolve: {
        // 强制别名，解决 "TypeError: reading 'eq'" 崩溃
        alias: {
            'prosemirror-model': path.resolve(__dirname, 'node_modules/prosemirror-model'),
            'prosemirror-state': path.resolve(__dirname, 'node_modules/prosemirror-state'),
            'prosemirror-view': path.resolve(__dirname, 'node_modules/prosemirror-view'),
            'prosemirror-transform': path.resolve(__dirname, 'node_modules/prosemirror-transform'),
            '@tiptap/pm/state': path.resolve(__dirname, 'node_modules/prosemirror-state'),
            '@tiptap/pm/view': path.resolve(__dirname, 'node_modules/prosemirror-view'),
            '@tiptap/pm/model': path.resolve(__dirname, 'node_modules/prosemirror-model'),
        }
    },
    server: {
        fs: {
            allow: [
                searchForWorkspaceRoot(process.cwd()),
                '..'
            ]
        }
    }
})