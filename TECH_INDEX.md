# CollabStudio 技术全量索引清单 (TECH_INDEX.md)

> **版本状态**: 6913bd4 (已还原原始界面)  
> **核心目标**: 血肉重塑 —— 从零掌握协作系统的底层实现

---

## 1. 项目概览 (Project Overview)
本项目是一个基于 **Go (Backend)** 和 **Vue 3 (Frontend)** 的实时协同办公系统。它结合了 Wails 的桌面端能力，实现了跨平台的实时文档编辑、多人光标同步和局域网自动发现。

---

## 2. 前端技术栈 (Frontend Architecture - CollabClient)

### 2.1 核心框架
- **Vue 3 (Composition API)**: 响应式界面逻辑。
- **Vite**: 前端构建工具，提供极速的开发体验。
- **Tiptap**: 基于 ProseMirror 的富文本编辑器核心，处理复杂的文档节点。

### 2.2 关键组件清单 (`/frontend/src/components`)
| 组件名 | 核心功能描述 | 技术要点 |
| :--- | :--- | :--- |
| `App.vue` | 根组件 | 负责 `Login`, `Lobby`, `Workspace` 的路由分发与状态流转。 |
| `Login.vue` | 登录/注册 | 服务器动态 IP 配置，调用 Wails 调用局域网扫描。 |
| `Workspace.vue` | 协作主容器 | 整合编辑器、聊天室、侧边栏；管理 WebSocket 连接生命周期。 |
| `Editor.vue` | 协同编辑器 | Tiptap 实例化，接收并应用远程 `content` 与 `cursor` 更新。 |
| `MenuBar.vue` | 编辑器工具栏 | 控制文档格式（加粗、标题、列表、代码块、图片上传）。 |
| `Sidebar.vue` | 资源管理器 | 管理当前房间内的文档树（目前主要用于显示布局）。 |

### 2.3 状态与通信 (`/frontend/src/`)
- **`store.js`**: 动态管理后端服务器地址（HTTP/WS URL）。
- **`wailsjs/`**: 自动生成的 Go 方法绑定，前端通过它调用本地系统功能（如 `ScanLanServers`）。
- **`utils/cursor.js`**: 处理编辑器光标的计算与渲染逻辑。

---

## 3. 后端技术栈 (Backend Architecture - CollabServer)

### 3.1 核心框架
- **Go (Golang)**: 高并发后端引擎。
- **Gin**: 轻量级 HTTP Web 框架。
- **GORM**: 数据库 ORM，负责模型映射与迁移。
- **JWT (golang-jwt)**: 身份验证与令牌管理。

### 3.2 关键模块清单
| 文件夹 | 功能描述 | 技术要点 |
| :--- | :--- | :--- |
| `main.go` | 程序入口 | 初始化数据库/Redis、启动 HTTP 服务 (Port 80)、启动 UDP 广播。 |
| `websocket/` | 实时通信核心 | `hub.go` 管理所有房间和连接；`client.go` 处理单个连接的消息读写。 |
| `controllers/` | 业务逻辑处理 | `auth.go` (用户认证), `upload.go` (文件存储), `user.go` (用户管理)。 |
| `database/` | 持久化与缓存 | `db.go` (SQLite/MySQL 连接), `redis.go` (Redis 缓存初始化)。 |
| `models/` | 数据模型定义 | 定义 User, Document, Message, History 的数据库结构。 |
| `middleware/` | 中间件 | `auth.go` 拦截请求并校验 JWT Token。 |

---

## 4. 核心工作流 (Core Workflows)

### 4.1 实时同步流程 (The Sync Loop)
1. **输入**: 用户在 `Editor.vue` 中修改内容。
2. **发送**: `Workspace.vue` 通过 WebSocket 发送 `type: content` 消息。
3. **分发**: `CollabServer` 的 `Hub` 接收消息，并将其广播给该房间内除发送者外的所有 `Client`。
4. **接收**: 远程前端收到消息，调用 `editor.commands.setContent` (带防抖处理)。

### 4.2 局域网发现 (LAN Discovery)
- **服务端**: `startUDPDiscoveryService` 监听 UDP 9999 端口，收到 `WHOIS_COLLAB_HOST` 后回传主机名。
- **客户端**: `App.go` 中的 `ScanLanServers` 发送 UDP 广播包，收集响应并返回给 Vue 界面。

### 4.3 身份验证 (Auth)
- 登录成功后，后端返回 JWT Token。
- 前端将其存入 `localStorage`。
- 后续 HTTP 请求（如 `/history`）在 Header 中携带 `Authorization: Bearer <token>`。
- WebSocket 连接通过 Query 参数 `?token=<token>` 进行校验。

---

## 5. 反向工程指导 (Reverse Engineering Guidance)

### 5.1 如何脱离 AI 复刻该项目？
1. **理解 WebSocket**: 必须亲手写一个原生的 `net/http` WebSocket 升级程序，理解 `Upgrader` 的原理。
2. **掌握 Gin 中间件**: 理解 `c.Next()` 和 `c.Abort()` 的调用链控制。
3. **编辑器原理**: 研究 Tiptap/ProseMirror 的 `Transaction` 和 `Step` 概念，这是协同编辑的基础。
4. **并发控制**: 深入理解 `Hub` 中 `register`, `unregister`, `broadcast` 三个 channel 的协作模式，避免 Race Condition。

---
*此索引由 AI 技术管家生成，作为“血肉重塑”计划的导航地图。*
