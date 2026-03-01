<template>
  <Transition name="slide">
    <div v-if="visible" class="settings-overlay" @click.self="$emit('close')">
      <div class="settings-drawer">
        <!-- 头部 -->
        <div class="drawer-header">
          <h2><i class="ri-settings-3-line"></i> 设置</h2>
          <button @click="$emit('close')" class="close-btn">
            <i class="ri-close-line"></i>
          </button>
        </div>

        <div class="drawer-body">
          <!-- ====== 外观 ====== -->
          <section class="settings-section">
            <h3><i class="ri-palette-line"></i> 外观</h3>

            <div class="setting-row">
              <label>主题</label>
              <div class="theme-switch" @click="handleToggleTheme">
                <div class="switch-track" :class="{ light: settings.theme === 'light' }">
                  <div class="switch-thumb">
                    <i :class="settings.theme === 'dark' ? 'ri-moon-line' : 'ri-sun-line'"></i>
                  </div>
                </div>
                <span>{{ settings.theme === 'dark' ? '深色' : '浅色' }}</span>
              </div>
            </div>
          </section>

          <!-- ====== 编辑器 ====== -->
          <section class="settings-section">
            <h3><i class="ri-edit-line"></i> 编辑器</h3>

            <div class="setting-row">
              <label>字体大小</label>
              <div class="stepper">
                <button @click="changeFontSize(-1)">−</button>
                <span class="stepper-value">{{ settings.fontSize }}px</span>
                <button @click="changeFontSize(1)">+</button>
              </div>
            </div>

            <div class="setting-row">
              <label>字体</label>
              <select v-model="settings.fontFamily" class="select-input">
                <option value="Consolas">Consolas</option>
                <option value="'Fira Code'">Fira Code</option>
                <option value="'JetBrains Mono'">JetBrains Mono</option>
                <option value="monospace">monospace</option>
                <option value="'Microsoft YaHei'">微软雅黑</option>
                <option value="'Segoe UI'">Segoe UI</option>
              </select>
            </div>
          </section>

          <!-- ====== AI 配置 ====== -->
          <section class="settings-section">
            <h3><i class="ri-robot-line"></i> AI 助手</h3>

            <div class="setting-row vertical">
              <label>API 地址</label>
              <input
                v-model="settings.ai.apiUrl"
                placeholder="https://api.deepseek.com/v1"
                class="text-input"
              />
            </div>

            <div class="setting-row vertical">
              <label>API 密钥</label>
              <input
                v-model="settings.ai.apiKey"
                type="password"
                placeholder="sk-xxxxxxxx"
                class="text-input"
              />
            </div>

            <div class="setting-row vertical">
              <label>模型名称</label>
              <input
                v-model="settings.ai.model"
                placeholder="deepseek-chat"
                class="text-input"
              />
            </div>
          </section>
        </div>

        <!-- 底部操作 -->
        <div class="drawer-footer">
          <button @click="handleReset" class="reset-btn">
            <i class="ri-refresh-line"></i> 恢复默认设置
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { settings, toggleTheme, resetSettings } from '../settings'

defineProps({ visible: Boolean })
defineEmits(['close'])

const handleToggleTheme = () => {
  toggleTheme()
}

const changeFontSize = (delta) => {
  const newSize = settings.fontSize + delta
  if (newSize >= 10 && newSize <= 28) {
    settings.fontSize = newSize
  }
}

const handleReset = () => {
  if (confirm('确定恢复所有设置为默认值？')) {
    resetSettings()
  }
}
</script>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  z-index: 9000;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.settings-drawer {
  position: absolute;
  top: 0;
  right: 0;
  width: 360px;
  height: 100vh;
  background: var(--bg-panel);
  border-left: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  box-shadow: -8px 0 30px rgba(0, 0, 0, 0.3);
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
}

.drawer-header h2 {
  margin: 0;
  font-size: 1.15rem;
  color: var(--text-main);
  display: flex;
  align-items: center;
  gap: 8px;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.3rem;
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  transition: all 0.2s;
}
.close-btn:hover { background: var(--bg-hover); color: var(--text-main); }

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 24px;
}

.settings-section {
  margin-bottom: 28px;
}

.settings-section h3 {
  font-size: 0.8rem;
  text-transform: uppercase;
  color: var(--text-muted);
  margin: 0 0 14px;
  display: flex;
  align-items: center;
  gap: 6px;
  letter-spacing: 0.5px;
  font-weight: 600;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid rgba(255,255,255,0.04);
}

.setting-row.vertical {
  flex-direction: column;
  align-items: stretch;
  gap: 6px;
}

.setting-row label {
  font-size: 0.9rem;
  color: var(--text-main);
}

/* 主题切换 */
.theme-switch {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.theme-switch span {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.switch-track {
  width: 48px;
  height: 26px;
  border-radius: 13px;
  background: #3f3f46;
  position: relative;
  transition: background 0.3s;
}

.switch-track.light {
  background: #60a5fa;
}

.switch-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #1e1e2e;
  transition: transform 0.3s;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

.switch-track.light .switch-thumb {
  transform: translateX(22px);
}

/* Stepper */
.stepper {
  display: flex;
  align-items: center;
  gap: 0;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.stepper button {
  width: 32px;
  height: 32px;
  background: var(--bg-hover);
  border: none;
  color: var(--text-main);
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.2s;
}
.stepper button:hover { background: var(--primary-color); color: white; }

.stepper-value {
  width: 50px;
  text-align: center;
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--text-main);
  background: var(--bg-main);
}

/* Select */
.select-input {
  background: var(--bg-main);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  outline: none;
}
.select-input:focus { border-color: var(--primary-color); }

/* Text Input */
.text-input {
  background: var(--bg-main);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.9rem;
  outline: none;
  font-family: monospace;
  transition: border-color 0.2s;
}
.text-input:focus { border-color: var(--primary-color); }
.text-input::placeholder { color: var(--text-muted); opacity: 0.5; }

/* Footer */
.drawer-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
}

.reset-btn {
  width: 100%;
  padding: 10px;
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--danger-color);
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s;
}
.reset-btn:hover { background: rgba(239, 68, 68, 0.1); border-color: var(--danger-color); }

/* Slide transition */
.slide-enter-active, .slide-leave-active {
  transition: opacity 0.3s ease;
}
.slide-enter-active .settings-drawer, .slide-leave-active .settings-drawer {
  transition: transform 0.3s ease;
}
.slide-enter-from, .slide-leave-to {
  opacity: 0;
}
.slide-enter-from .settings-drawer, .slide-leave-to .settings-drawer {
  transform: translateX(100%);
}
</style>
