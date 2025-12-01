<template>
  <node-view-wrapper as="span" class="image-component">
    <img
        :src="currentSrc"
        :alt="node.attrs.alt"
        :title="node.attrs.title"
        :class="{ 'selected': selected }"
        draggable="false"
        @error="handleImageError"
    />
  </node-view-wrapper>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { nodeViewProps, NodeViewWrapper } from '@tiptap/vue-3'
import { serverConfig } from '../store'

const props = defineProps(nodeViewProps)
const currentSrc = ref('')

// 🟢 计算完整的 URL (自动补全服务器 IP)
const getFullUrl = (src) => {
  if (!src) return ''
  // 如果是 base64 或 http 开头的完整链接，直接用
  if (src.startsWith('data:') || src.startsWith('http')) return src

  // 否则补全服务器地址
  // 例如: "/uploads/abc.png" -> "http://192.168.1.5:8080/uploads/abc.png"
  const baseUrl = serverConfig.getHttpUrl()
  // 防止重复拼接 (虽然一般不会发生，但为了保险)
  if (src.startsWith(baseUrl)) return src

  return `${baseUrl}${src}`
}

// 初始化
onMounted(() => {
  currentSrc.value = getFullUrl(props.node.attrs.src)
})

// 监听属性变化 (比如别人修改了图片)
watch(() => props.node.attrs.src, (newSrc) => {
  currentSrc.value = getFullUrl(newSrc)
})

// 🟢 容错处理：如果图片加载失败，尝试再次强制补全
const handleImageError = (e) => {
  const src = props.node.attrs.src
  if (src && !src.startsWith('http') && !src.startsWith('data:')) {
    // 如果之前没补全成功，这里再强制试一次
    e.target.src = `${serverConfig.getHttpUrl()}${src}`
  }
}
</script>

<style scoped>
.image-component {
  display: inline-block;
  line-height: 0; /* 防止图片底部有缝隙 */
  vertical-align: baseline;
}

.image-component img {
  max-width: 100%;
  height: auto;
  border-radius: 6px;
  transition: all 0.2s;
  cursor: default;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

/* 选中时的样式 (Tiptap 选中) */
.image-component img.selected {
  outline: 3px solid #89b4fa;
  box-shadow: 0 4px 12px rgba(137, 180, 250, 0.4);
}
</style>