# CollabStudio 技术索引

更新时间：2026-04-04

本文档用于快速定位当前项目的真实技术结构，避免继续参考过期说明。

## 1. 总体架构

当前项目采用：

- 桌面端：`Wails + Vue 3`
- 浏览器端：`Vue 3 + Vite`
- 后端：`Go + Gin + GORM + SQLite`
- 实时通信：`WebSocket`

当前实现不是 `Yjs` 协同版本，而是：

- 文档内容通过 WebSocket 广播 `doc_update`
- 前端做内容去重后更新编辑器
- 用户光标通过独立消息同步

## 2. 关键目录

### 2.1 桌面端与前端

- `CollabClient/main.go`
  - Wails 应用入口
  - 管理窗口尺寸、关闭拦截和资源嵌入

- `CollabClient/app.go`
  - 桌面端启动逻辑
  - 自动拉起同目录后端
  - 桌面文件打开/保存
  - 关闭确认与工作区状态

- `CollabClient/frontend/src/App.vue`
  - 前端根组件
  - 管理 `login / lobby / workspace`
  - 管理多房间标签页

- `CollabClient/frontend/src/components/Login.vue`
  - 登录 / 注册界面
  - 服务器地址手动配置

- `CollabClient/frontend/src/components/Lobby.vue`
  - 大厅界面
  - 最近访问记录与进入房间入口

- `CollabClient/frontend/src/components/Workspace.vue`
  - 协作工作区主容器
  - 管理 WebSocket、聊天、AI 面板与图片上传

- `CollabClient/frontend/src/components/Editor.vue`
  - Tiptap 编辑器实例

- `CollabClient/frontend/src/components/MenuBar.vue`
  - 编辑器格式工具栏

- `CollabClient/frontend/src/components/AiPanel.vue`
  - AI 总结、翻译、润色与自由提问

- `CollabClient/frontend/src/store.js`
  - 服务端地址选择与 HTTP / WS URL 生成

- `CollabClient/frontend/src/utils/auth.js`
  - JWT 存取与旧 token 迁移

## 3. 后端关键模块

- `CollabServer/main.go`
  - 加载配置
  - 初始化数据库
  - 配置 Gin 路由与 CORS
  - 暴露 `/ping`、认证接口、历史记录接口、上传接口和 WebSocket

- `CollabServer/config/config.go`
  - 加载和生成 `.env`

- `CollabServer/database/db.go`
  - 连接 SQLite 并初始化 GORM

- `CollabServer/controllers/auth.go`
  - 注册 / 登录

- `CollabServer/controllers/user.go`
  - 历史访问记录等用户相关接口

- `CollabServer/controllers/upload.go`
  - 图片上传

- `CollabServer/controllers/ai.go`
  - AI 聊天代理与流式响应

- `CollabServer/websocket/hub.go`
  - 房间、连接与广播管理

- `CollabServer/websocket/client.go`
  - 单连接收发与消息处理

## 4. 当前真实功能边界

### 已实现

- 账号注册与登录
- 房间式多人协作
- 在线成员列表
- 聊天消息与图片消息
- 富文本编辑
- AI 面板调用
- 桌面端一键启动同目录后端

### 当前不应写成“已完整交付”

- 浏览器端局域网扫描
- 完整自动化测试体系
- 完整产物归档与仓库净化

## 5. 调试与验证建议

### 后端

```powershell
cd CollabServer
go test ./...
```

### 前端

```powershell
cd CollabClient/frontend
npm run build
```

### 桌面端

```powershell
cd CollabClient
go test ./...
wails build
```

## 6. 使用原则

如果后续文档、论文或答辩说明与当前源码不一致，应优先以以下文件为准：

- `CollabClient/app.go`
- `CollabClient/main.go`
- `CollabClient/frontend/src/`
- `CollabServer/main.go`
- `CollabServer/controllers/`
- `CollabServer/websocket/`
