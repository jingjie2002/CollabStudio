<template>
  <div class="menu-bar">
    <!-- 🟢 隐藏的文件输入框，用于触发系统文件选择 -->
    <input
        type="file"
        ref="fileInput"
        style="display: none"
        accept="image/png, image/jpeg, image/gif"
        @change="handleImageUpload"
    />

    <div class="button-group">
      <button @click="editor.chain().focus().toggleBold().run()"
              :disabled="!editor.can().chain().focus().toggleBold().run()"
              :class="{ 'is-active': editor.isActive('bold') }"
              title="加粗 (Ctrl+B)">
        <i class="ri-bold"></i>
      </button>
      <button @click="editor.chain().focus().toggleItalic().run()"
              :disabled="!editor.can().chain().focus().toggleItalic().run()"
              :class="{ 'is-active': editor.isActive('italic') }"
              title="斜体 (Ctrl+I)">
        <i class="ri-italic"></i>
      </button>
      <button @click="editor.chain().focus().toggleStrike().run()"
              :disabled="!editor.can().chain().focus().toggleStrike().run()"
              :class="{ 'is-active': editor.isActive('strike') }"
              title="删除线 (Ctrl+Shift+X)">
        <i class="ri-strikethrough"></i>
      </button>
      <button @click="editor.chain().focus().toggleCode().run()"
              :disabled="!editor.can().chain().focus().toggleCode().run()"
              :class="{ 'is-active': editor.isActive('code') }"
              title="行内代码 (Ctrl+E)">
        <i class="ri-code-line"></i>
      </button>
    </div>

    <div class="divider"></div>

    <div class="button-group">
      <button @click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
              :class="{ 'is-active': editor.isActive('heading', { level: 1 }) }"
              title="标题 1">
        <i class="ri-h-1"></i>
      </button>
      <button @click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
              :class="{ 'is-active': editor.isActive('heading', { level: 2 }) }"
              title="标题 2">
        <i class="ri-h-2"></i>
      </button>
      <button @click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
              :class="{ 'is-active': editor.isActive('heading', { level: 3 }) }"
              title="标题 3">
        <i class="ri-h-3"></i>
      </button>
    </div>

    <div class="divider"></div>

    <div class="button-group">
      <button @click="editor.chain().focus().toggleBulletList().run()"
              :class="{ 'is-active': editor.isActive('bulletList') }"
              title="无序列表">
        <i class="ri-list-unordered"></i>
      </button>
      <button @click="editor.chain().focus().toggleOrderedList().run()"
              :class="{ 'is-active': editor.isActive('orderedList') }"
              title="有序列表">
        <i class="ri-list-ordered"></i>
      </button>

      <button @click="editor.chain().focus().toggleTaskList().run()"
              :class="{ 'is-active': editor.isActive('taskList') }"
              title="待办事项">
        <i class="ri-task-line"></i>
      </button>

      <button @click="editor.chain().focus().toggleBlockquote().run()"
              :class="{ 'is-active': editor.isActive('blockquote') }"
              title="引用块">
        <i class="ri-double-quotes-l"></i>
      </button>

      <button @click="editor.chain().focus().toggleCodeBlock().run()"
              :class="{ 'is-active': editor.isActive('codeBlock') }"
              title="代码块">
        <i class="ri-code-box-line"></i>
      </button>

      <!-- 🟢 图片上传按钮 -->
      <button @click="triggerFileSelect" title="插入图片">
        <i class="ri-image-add-line"></i>
      </button>
    </div>

    <div class="divider"></div>

    <div class="button-group">
      <button @click="editor.chain().focus().undo().run()"
              :disabled="!editor.can().chain().focus().undo().run()"
              title="撤销 (Ctrl+Z)">
        <i class="ri-arrow-go-back-line"></i>
      </button>
      <button @click="editor.chain().focus().redo().run()"
              :disabled="!editor.can().chain().focus().redo().run()"
              title="重做 (Ctrl+Shift+Z)">
        <i class="ri-arrow-go-forward-line"></i>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
// 🟢 引入全局配置
import { serverConfig } from '../store'

const props = defineProps({
  editor: {
    type: Object,
    required: true,
  },
})

// 🟢 图片上传逻辑
const fileInput = ref(null)

const triggerFileSelect = () => {
  fileInput.value.click()
}

const handleImageUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  // 构造表单数据
  const formData = new FormData()
  formData.append('image', file)

  try {
    // 🟢 动态地址请求后端
    const response = await fetch(`${serverConfig.getHttpUrl()}/upload`, {
      method: 'POST',
      body: formData
    })

    const data = await response.json()

    if (data.url) {
      // 成功！将图片插入编辑器
      props.editor.chain().focus().setImage({ src: data.url }).run()
    } else {
      alert('图片上传失败: ' + (data.error || '未知错误'))
    }
  } catch (e) {
    console.error(e)
    alert(`上传出错，请检查服务器连接 (${serverConfig.getHttpUrl()})`)
  }

  // 清空 input，防止无法重复选择同一张图片
  event.target.value = ''
}
</script>

<style scoped>
.menu-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  padding: 8px 12px;
  background-color: #2b2b2b; /* 稍微比背景亮一点 */
  border-bottom: 1px solid #333;
}

.button-group {
  display: flex;
  gap: 2px;
}

.divider {
  width: 1px;
  height: 20px;
  background-color: #444;
  margin: 0 8px;
}

button {
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #a0a0a0;
  cursor: pointer;
  padding: 4px 6px;
  font-size: 1.1rem;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

button:hover:not(:disabled) {
  background-color: #3f3f3f;
  color: #fff;
}

button.is-active {
  background-color: #3b82f6; /* 激活色使用主题蓝 */
  color: #fff;
}

button:disabled {
  color: #444;
  cursor: not-allowed;
}

/* 适配 Remix Icon 的微调 */
i {
  line-height: 1;
}
</style>