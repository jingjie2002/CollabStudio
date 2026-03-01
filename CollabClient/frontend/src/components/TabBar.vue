<template>
  <div class="tab-bar" v-if="tabs.length > 0">
    <div class="tabs-scroll">
      <div
        v-for="tab in tabs"
        :key="tab.roomId"
        class="tab-item"
        :class="{ active: tab.roomId === activeRoom }"
        @click="$emit('switch', tab.roomId)"
      >
        <i class="ri-file-text-line"></i>
        <span class="tab-title">{{ tab.roomId }}</span>
        <button
          class="tab-close"
          @click.stop="$emit('close', tab.roomId)"
          title="关闭标签"
        >
          <i class="ri-close-line"></i>
        </button>
      </div>
    </div>
    <button class="tab-add" @click="$emit('add')" title="返回大厅加入新房间">
      <i class="ri-add-line"></i>
    </button>
  </div>
</template>

<script setup>
defineProps({
  tabs: { type: Array, default: () => [] },
  activeRoom: { type: String, default: '' }
})

defineEmits(['switch', 'close', 'add'])
</script>

<style scoped>
.tab-bar {
  display: flex;
  align-items: center;
  background: var(--bg-panel);
  border-bottom: 1px solid var(--border-color);
  height: 36px;
  flex-shrink: 0;
  padding: 0 4px;
  overflow: hidden;
}

.tabs-scroll {
  display: flex;
  flex: 1;
  overflow-x: auto;
  gap: 2px;
}

.tabs-scroll::-webkit-scrollbar { height: 0; }

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  height: 32px;
  border-radius: 6px 6px 0 0;
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--text-muted);
  background: transparent;
  border: 1px solid transparent;
  border-bottom: none;
  white-space: nowrap;
  transition: all 0.15s;
  flex-shrink: 0;
  max-width: 180px;
}

.tab-item:hover {
  color: var(--text-main);
  background: var(--bg-hover);
}

.tab-item.active {
  color: var(--text-main);
  background: var(--bg-main);
  border-color: var(--border-color);
}

.tab-title {
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: monospace;
}

.tab-close {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.85rem;
  padding: 0;
  line-height: 1;
  border-radius: 3px;
  opacity: 0;
  transition: all 0.15s;
}

.tab-item:hover .tab-close,
.tab-item.active .tab-close {
  opacity: 1;
}

.tab-close:hover {
  color: var(--danger-color);
  background: rgba(239, 68, 68, 0.1);
}

.tab-add {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 1rem;
  transition: all 0.15s;
  flex-shrink: 0;
}

.tab-add:hover {
  color: var(--text-main);
  background: var(--bg-hover);
}
</style>
