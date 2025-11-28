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
    // 即使加了 markRaw，使用 toRaw 也是个好习惯，双重保险
    const rawEditor = toRaw(editor.value)

    // 检查内容是否变化，避免死循环
    if (rawEditor.getHTML() !== newContent) {
      // 这里的 emitUpdate: false 很重要，防止 setContent 又触发 onUpdate 发送回 WebSocket
      rawEditor.commands.setContent(newContent, false)
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