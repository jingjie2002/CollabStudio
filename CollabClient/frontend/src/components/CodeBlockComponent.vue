<template>
  <node-view-wrapper class="code-block-wrapper">
    <div class="code-block-header">
      <span class="lang-label">{{ selectedLanguage || 'text' }}</span>

      <button contenteditable="false" @click="copyCode" class="copy-btn">
        {{ copyText }}
      </button>
    </div>

    <pre><node-view-content as="code" /></pre>
  </node-view-wrapper>
</template>

<script setup>
import { ref, computed } from 'vue'
import { NodeViewWrapper, NodeViewContent, nodeViewProps } from '@tiptap/vue-3'

const props = defineProps(nodeViewProps)

const selectedLanguage = computed(() => {
  return props.node.attrs.language
})

// 复制功能
const copyText = ref('Copy')
const copyCode = () => {
  const code = props.node.textContent
  navigator.clipboard.writeText(code).then(() => {
    copyText.value = 'Copied!'
    setTimeout(() => { copyText.value = 'Copy' }, 2000)
  })
}
</script>

<style scoped>
.code-block-wrapper {
  background: #282c34; /* 深色背景 */
  border-radius: 8px;
  margin: 1rem 0;
  overflow: hidden;
  position: relative;
  font-family: 'Consolas', 'Fira Code', monospace;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
  border: 1px solid #3e4451;
}

.code-block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background: #21252b; /* 头部颜色更深 */
  border-bottom: 1px solid #181a1f;
  user-select: none;
}

.lang-label {
  color: #abb2bf;
  font-size: 0.85rem;
  font-weight: bold;
  text-transform: uppercase;
}

.copy-btn {
  background: transparent;
  border: 1px solid #3d4547;
  color: #abb2bf;
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}
.copy-btn:hover { background: #3d4547; color: white; }

/* 覆盖 Tiptap 默认 pre 样式 */
pre {
  margin: 0;
  padding: 1rem;
  overflow-x: auto;
  color: #abb2bf; /* 代码文字颜色 */
  background: transparent;
  font-family: inherit;
}
</style>