<template>
  <div class="editor-container">
    <MenuBar v-if="editor" :editor="editor" />
    <editor-content :editor="editor" class="editor-content" />
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, defineExpose, defineEmits, shallowRef, toRaw, markRaw } from 'vue' // 🟢 引入 markRaw
import { Editor, EditorContent, VueNodeViewRenderer } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import CodeBlock from '@tiptap/extension-code-block'
import CodeBlockComponent from './CodeBlockComponent.vue'
import Image from '@tiptap/extension-image'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Placeholder from '@tiptap/extension-placeholder'
import { RemoteCursor, cursorPluginKey } from '../utils/cursor'
import MenuBar from './MenuBar.vue'

const emit = defineEmits(['update', 'cursor-update', 'check-connection'])

const editor = shallowRef(null)
// 🟢 远程更新锁：防止 setContent 触发 onUpdate 重新发送
let isRemoteUpdating = false

// 🟢 HTML 清洗函数：去除格式差异导致的无效同步
const sanitizeHTML = (html) => {
  if (!html) return ''
  return html
    .replace(/\s+/g, ' ')           // 多个空白符合并为一个
    .replace(/> </g, '><')          // 去除标签之间的空格
    .replace(/<p><\/p>/g, '')       // 去除空段落
    .trim()
}

const stringToColor = (str) => {
  let hash = 0
  for (let i = 0; i < str.length; i++) hash = str.charCodeAt(i) + ((hash << 5) - hash)
  const c = (hash & 0x00ffffff).toString(16).toUpperCase()
  return '#' + '00000'.substring(0, 6 - c.length) + c
}

const triggerConnectionCheck = () => {
  emit('check-connection')
}

onMounted(() => {
  // 🟢 核心修复：使用 markRaw 包裹 Editor 实例
  // 这会给对象打上 skip 标记，Vue 响应式系统见到它会直接绕道走
  const _editor = new Editor({
    content: '',
    extensions: [
      StarterKit.configure({
        codeBlock: false,
        history: true,
      }),
      CodeBlock.extend({
        addNodeView() {
          return VueNodeViewRenderer(CodeBlockComponent)
        },
      }),
      Image.configure({
        inline: true,
        allowBase64: true,
      }),
      TaskList,
      TaskItem.configure({ nested: true }),
      Placeholder.configure({ placeholder: '输入内容，或使用 Markdown 语法...' }),
      RemoteCursor,
    ],
    editorProps: {
      attributes: {
        class: 'tiptap-editor',
      },
    },
    onFocus: () => {
      triggerConnectionCheck()
    },
    onUpdate: ({ editor }) => {
      // 🟢 如果是远程更新引起的变化，严禁再次发送
      if (isRemoteUpdating) {
        return
      }
      emit('update', editor.getHTML())
    },
    onSelectionUpdate: ({ editor }) => {
      const { anchor } = editor.state.selection
      emit('cursor-update', anchor)
    },
  })

  // 赋值给 shallowRef
  editor.value = markRaw(_editor)

  window.addEventListener('focus', triggerConnectionCheck)
})

const setContent = (newContent) => {
  if (editor.value) {
    const rawEditor = toRaw(editor.value)

    // 🟢 使用清洗后的内容进行比较，避免 HTML 格式差异导致无效同步
    const currentSanitized = sanitizeHTML(rawEditor.getHTML())
    const newSanitized = sanitizeHTML(newContent)

    // 检查内容是否变化，避免无效更新
    if (currentSanitized !== newSanitized) {
      // 🟢 设置远程更新锁，防止 onUpdate 重新触发发送
      isRemoteUpdating = true

      // 光标保护：保存当前光标位置
      const { from, to } = rawEditor.state.selection
      const docLength = rawEditor.state.doc.content.size

      // emitUpdate: false 防止 setContent 触发 onUpdate
      rawEditor.commands.setContent(newContent, false)

      // 光标恢复：尝试恢复到原位置
      const newDocLength = rawEditor.state.doc.content.size
      const safeFrom = Math.min(from, newDocLength - 1)
      const safeTo = Math.min(to, newDocLength - 1)
      
      try {
        // 只有当文档长度变化不大时才恢复光标
        if (Math.abs(newDocLength - docLength) < 100) {
          rawEditor.commands.setTextSelection({ from: Math.max(1, safeFrom), to: Math.max(1, safeTo) })
        }
      } catch (e) {
        console.debug('[Editor] 光标恢复跳过:', e.message)
      }

      // 🟢 释放远程更新锁（使用 nextTick 确保所有同步事件处理完毕）
      setTimeout(() => {
        isRemoteUpdating = false
      }, 0)
    }
  }
}

const getText = () => {
  return editor.value ? editor.value.getHTML() : ''
}

const updateCursors = (users) => {
  if (!editor.value) return

  const rawEditor = toRaw(editor.value)
  if (!rawEditor || !rawEditor.state) return

  // 准备纯数据 (Pure Data)
  const cursorData = users
      .filter(u => u && u.username)
      .map(u => {
        let safePos = u.cursorVal
        if (typeof safePos !== 'number' || isNaN(safePos)) safePos = 0

        return {
          id: u.username,
          name: u.username,
          pos: safePos,
          color: stringToColor(u.username)
        }
      })

  try {
    const tr = rawEditor.state.tr
    // 我们只传递纯 JSON 数据给插件，不再传递 Decoration 对象
    tr.setMeta(cursorPluginKey, { type: 'update', cursors: cursorData })
    rawEditor.view.dispatch(tr)
  } catch (e) {
    console.error("Cursor update failed:", e)
  }
}

defineExpose({ setContent, getText, updateCursors })

onBeforeUnmount(() => {
  if (editor.value) {
    editor.value.destroy()
  }
  window.removeEventListener('focus', triggerConnectionCheck)
})
</script>

<style>
/* CSS 保持不变 */
.editor-container { height: 100%; width: 100%; display: flex; flex-direction: column; }
.editor-content { flex: 1; overflow-y: auto; padding: 10px 20px; color: #e5e7eb; font-family: 'Consolas', monospace, sans-serif; }
.tiptap-editor { outline: none; min-height: 100%; position: relative; }
.tiptap h1 { font-size: 1.8rem; font-weight: bold; margin: 0.5em 0; color: #fff; }
.tiptap h2 { font-size: 1.4rem; font-weight: bold; margin: 0.5em 0; color: #ddd; }
.tiptap ul, .tiptap ol { padding-left: 1.2em; }
.tiptap blockquote { border-left: 3px solid #3b82f6; padding-left: 1rem; color: #9ca3af; font-style: italic; }
.tiptap code { background-color: #374151; padding: 0.2em 0.4em; border-radius: 4px; font-family: monospace; }
.tiptap p { margin: 0.5em 0; line-height: 1.6; }
.tiptap img { max-width: 100%; height: auto; border-radius: 8px; margin: 10px 0; display: block; box-shadow: 0 4px 6px rgba(0,0,0,0.3); }
.tiptap img.ProseMirror-selectednode { outline: 3px solid #3b82f6; }
ul[data-type="taskList"] { list-style: none; padding: 0; }
ul[data-type="taskList"] li { display: flex; gap: 0.5rem; }
ul[data-type="taskList"] input[type="checkbox"] { cursor: pointer; accent-color: #3b82f6; }
.tiptap p.is-editor-empty:first-child::before { color: #555; content: attr(data-placeholder); float: left; height: 0; pointer-events: none; }
.remote-cursor { position: absolute; border-left: 2px solid; height: 1.2em; margin-top: -0.1em; pointer-events: none; z-index: 10; }
.remote-cursor-label { position: absolute; top: -1.4em; left: -2px; font-size: 10px; padding: 2px 5px; border-radius: 3px; color: white; white-space: nowrap; }
</style>