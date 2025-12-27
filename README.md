<div align="center">
<img src="CollabClient/build/appicon.png" width="120" height="120" alt="CollabStudio Logo" />
<h1>CollabStudio</h1>
<h3>✨ 基于 Go Wails + Vue3 的局域网实时协作工作台 ✨</h3>
<p>
实时文档同步 · 局域网一键联机 · 房主/访客机制 · 现代化 UI
</p>

<p>
<img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go">
<img src="https://img.shields.io/badge/Wails-v2.10+-red?style=flat-square&logo=wails" alt="Wails">
<img src="https://img.shields.io/badge/Vue-3.2+-4FC08D?style=flat-square&logo=vue.js" alt="Vue">
<img src="https://img.shields.io/badge/Tiptap-Editor-000000?style=flat-square" alt="Tiptap">
<img src="https://img.shields.io/badge/SQLite-Pure%20Go-003B57?style=flat-square&logo=sqlite" alt="SQLite">
<img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
</p>
</div>

---

## 📖 项目简介

**CollabStudio** 是一款专为局域网环境设计的**全栈实时协作软件**。它采用 CS (Client-Server) 混合架构，但在体验上做到了极致的 "单机化" —— 用户无需手动部署复杂的服务器，只需**双击客户端即可自动拉起后台服务**，实现"即开即用"的协作体验。

项目解决了传统 Web 协作在数据隐私和部署难度上的痛点：
- 利用 **WebSocket** 实现毫秒级文档同步
- 利用 **Go Wails** 构建高性能桌面客户端
- 创新性地引入了 **"伴随式"房主机制**，解决了局域网工具的服务器托管难题

---

## 🚀 核心特性

### 👑 独创"伴随式"房主机制
| 特性 | 说明 |
|------|------|
| **智能启动** | 首个启动的用户自动成为"房主"，后台静默启动 Server 核心 |
| **无感连接** | 后续用户自动识别服务，以"访客"身份加入，无需复杂配置 |
| **安全拦截** | 房主退出时的智能拦截保护，防止误操作导致房间解散 |
| **双端口兼容** | 自动检测 80/8080 端口，适配生产/开发环境 |

### ⚡️ 毫秒级实时同步
- 基于 **WebSocket** 的全双工通信
- 支持多人同时编辑同一文档，光标实时跟随
- 流量控制 (Flow Control)：后端非阻塞广播 + 前端自适应节流

### 📝 强大的富文本编辑器
- 基于 **Tiptap (ProseMirror)** 内核
- 支持 Markdown 语法、代码块高亮、任务列表
- 支持图片粘贴上传、拖拽上传、图片预览
- 工具栏一键格式化：加粗/斜体/删除线/标题/列表等

### 💾 本地优先的数据权
- 协作内容实时存入房主本地 **SQLite** 数据库
- 支持将协作成果一键 **导出/保存** 到本地硬盘（txt/md）
- 支持从本地 **导入** 文件内容到协作空间

### 🛡️ 工业级稳定性
- JWT 认证 + 密码加密存储
- 完善的断线重连与健康检查机制
- Redis 支持（可选，用于增强缓存）

---

## 🏗️ 系统架构

CollabStudio 采用 **混合伴随架构 (Hybrid Sidecar Architecture)**：

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Host Machine (房主)                             │
│  ┌─────────────────┐         ┌─────────────────────────────────┐   │
│  │ CollabClient.exe│──spawn──▶│ CollabServer.exe               │   │
│  │   (Wails App)   │         │   ├─ HTTP API (:8080)           │   │
│  │                 │◀──WS────│   ├─ WebSocket (/ws)            │   │
│  └─────────────────┘         │   ├─ SQLite (collab.db)         │   │
│                              │   └─ Uploads (./uploads/)       │   │
│                              └─────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                                      ▲
                                      │ LAN WebSocket
                                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Guest Machine (访客)                             │
│  ┌─────────────────┐                                                │
│  │ CollabClient.exe│ ──────────────▶ 连接房主 IP                     │
│  │   (Wails App)   │                                                │
│  └─────────────────┘                                                │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 📦 技术栈与依赖

### 后端核心（Go Modules）

| 依赖库 | 版本 | 作用 |
|--------|------|------|
| `github.com/wailsapp/wails/v2` | v2.10+ | 桌面应用框架，打包 Go+Web 为原生 exe |
| `github.com/gin-gonic/gin` | v1.11.0 | Web 框架，处理 HTTP API |
| `github.com/gorilla/websocket` | v1.5.3 | WebSocket 协议实现 |
| `github.com/glebarez/sqlite` | v1.11.0 | **纯 Go SQLite 驱动**（无需 CGO） |
| `gorm.io/gorm` | v1.31.x | ORM 数据库映射 |
| `github.com/gin-contrib/cors` | v1.7.6 | 跨域中间件 |
| `github.com/golang-jwt/jwt/v5` | v5.3.0 | JWT 认证 |
| `github.com/redis/go-redis/v9` | v9.17.2 | Redis 客户端（可选） |
| `golang.org/x/crypto` | v0.45.0 | 密码加密 (bcrypt) |

### 前端核心（Vue 3 + Vite）

| 依赖库 | 版本 | 作用 |
|--------|------|------|
| `vue` | ^3.2.37 | 核心 UI 框架 |
| `vue-router` | ^4.6.3 | 路由管理 |
| `vite` | ^3.0.7 | 构建工具 |
| `@tiptap/vue-3` | ^3.11.0 | 富文本编辑器核心 |
| `@tiptap/starter-kit` | ^3.11.0 | 基础功能套件 |
| `@tiptap/extension-image` | ^3.11.0 | 图片插入支持 |
| `@tiptap/extension-code-block` | ^3.11.0 | 代码块高亮 |
| `@tiptap/extension-task-list` | ^3.11.1 | 任务列表 |
| `axios` | ^1.13.2 | HTTP 请求库 |
| `remixicon` | ^4.7.0 | 矢量图标库 |

---

## 📂 项目目录结构

```
MyGradProject/
├── 📁 CollabClient/                 # 🖥️ 客户端 (Wails + Vue3)
│   ├── app.go                       # [Go] 桥接层：文件保存/读取等本地系统调用
│   ├── main.go                      # [Go] 程序入口：自动启动 Server 的逻辑
│   ├── wails.json                   # Wails 项目配置
│   ├── 📁 build/                    # 构建输出目录
│   │   └── bin/                     # 生成的 exe 文件
│   └── 📁 frontend/                 # 🎨 前端源码
│       ├── 📁 src/
│       │   ├── 📁 components/       # Vue 组件
│       │   │   ├── Login.vue        # 登录页面
│       │   │   ├── Register.vue     # 注册页面
│       │   │   ├── Lobby.vue        # 大厅/房间选择
│       │   │   ├── Workspace.vue    # 协作工作区（主界面）
│       │   │   ├── Editor.vue       # Tiptap 富文本编辑器
│       │   │   ├── MenuBar.vue      # 编辑器工具栏
│       │   │   └── CodeBlockComponent.vue  # 代码块组件
│       │   ├── 📁 utils/            # 工具类
│       │   │   └── cursor.js        # 光标同步插件
│       │   ├── store.js             # 全局状态管理（服务器IP配置）
│       │   ├── App.vue              # 根组件
│       │   ├── main.js              # 入口文件
│       │   └── style.css            # 全局样式
│       ├── package.json
│       └── vite.config.js           # Vite 构建配置
│
├── 📁 CollabServer/                 # ⚙️ 服务端 (Go + Gin)
│   ├── main.go                      # 服务器入口：启动 HTTP 和 WebSocket
│   ├── go.mod / go.sum              # Go 模块依赖
│   ├── .env                         # 环境变量配置
│   ├── .env.example                 # 环境变量示例
│   ├── collab.db                    # SQLite 数据库（运行时生成）
│   ├── 📁 uploads/                  # 图片存储目录
│   ├── 📁 config/                   # 配置模块
│   ├── 📁 controllers/              # 业务逻辑控制器
│   │   ├── auth.go                  # 认证（登录/注册）
│   │   ├── upload.go                # 文件上传
│   │   ├── user.go                  # 用户管理
│   │   └── admin.go                 # 管理员功能
│   ├── 📁 database/                 # 数据库连接与初始化
│   ├── 📁 models/                   # 数据模型定义
│   │   ├── user.go                  # 用户模型
│   │   ├── document.go              # 文档模型
│   │   ├── message.go               # 消息模型
│   │   └── history.go               # 历史记录模型
│   ├── 📁 middleware/               # 中间件
│   ├── 📁 router/                   # 路由配置
│   │   └── router.go                # API 路由定义
│   ├── 📁 websocket/                # WebSocket 核心逻辑
│   │   ├── hub.go                   # 连接管理器（广播中心）
│   │   └── client.go                # 客户端连接处理
│   ├── deploy.sh                    # Linux 部署脚本
│   └── collab.service               # Systemd 服务配置
│
├── README.md                        # 📖 本文档
├── TECH_INDEX.md                    # 技术索引
└── PROJECT_STATUS_REPORT.md         # 项目状态报告
```

---

## ⚡️ 快速开始

### 1. 环境要求

| 工具 | 版本要求 | 安装命令/说明 |
|------|----------|---------------|
| **Go** | v1.20+ | [下载地址](https://go.dev/dl/) |
| **Node.js** | v16+ | [下载地址](https://nodejs.org/) |
| **Wails CLI** | v2.9+ | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |

> 💡 验证 Wails 安装：`wails doctor`

### 2. 克隆项目

```bash
git clone <your-repo-url>
cd MyGradProject
```

### 3. 开发环境运行

#### 方式一：分离启动（推荐调试时使用）

**终端 1 - 启动后端服务：**
```bash
cd CollabServer
go run main.go
```
> 默认监听 `:8080`，可通过 `.env` 修改端口

**终端 2 - 启动客户端（热重载）：**
```bash
cd CollabClient
wails dev
```
> 自动打开 Wails 开发窗口，支持前端热更新

#### 方式二：完整客户端启动

```bash
cd CollabClient
wails dev
```
> 客户端会自动检测并启动本地 Server（如果 Server 未运行）

### 4. 生产环境打包

#### 步骤一：编译后端

```bash
cd CollabServer

# Windows
go build -ldflags "-s -w" -o CollabServer.exe main.go

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o collab_server main.go
```

#### 步骤二：编译客户端

```bash
cd CollabClient
wails build -clean
```

> 生成文件位置：`CollabClient/build/bin/CollabClient.exe`

#### 步骤三：部署合体

将 `CollabServer.exe` 复制到 `CollabClient/build/bin/` 目录下，确保与 `CollabClient.exe` **同级**：

```
CollabClient/build/bin/
├── CollabClient.exe      # 客户端主程序
└── CollabServer.exe      # 后端服务（房主自动启动）
```

---

## 🤝 使用说明

### 🏠 房主（Host）模式

1. **双击** `CollabClient.exe`
2. 软件自动在后台启动服务器
3. 左上角显示 **[👑 房主]** 标识
4. 在登录页注册/登录账号
5. 进入大厅，创建或加入文档空间

> ⚠️ **重要提示**：房主关闭窗口会导致房间解散，所有访客断开连接！

### 👤 访客（Guest）模式

1. 获取房主的局域网 IP 地址（如 `192.168.1.5`）
2. 打开 `CollabClient.exe`
3. 点击登录页左下角 **⚙️ 服务器设置**
4. 输入房主 IP：`192.168.1.5:8080`（或端口 80）
5. 保存后注册/登录即可加入协作

### ⌨️ 编辑器操作

| 功能 | 操作方式 |
|------|----------|
| **粘贴图片** | Ctrl+V 直接粘贴剪贴板图片 |
| **拖拽上传** | 将图片文件拖入编辑区域 |
| **Markdown 语法** | 输入 `# ` 自动转标题、`- ` 转列表等 |
| **代码块** | 点击工具栏代码图标或输入 ``` |
| **任务列表** | 点击工具栏任务列表图标 |
| **保存到本地** | 点击菜单栏 📥 导出按钮 |
| **导入本地文件** | 点击菜单栏 📤 导入按钮 |

---

## 🔧 配置说明

### 服务端环境变量 (.env)

```bash
# 服务器配置
PORT=8080                          # HTTP 监听端口
JWT_SECRET=your-secret-key         # JWT 密钥

# 数据库配置
DB_PATH=./collab.db                # SQLite 数据库路径

# Redis 配置（可选）
REDIS_ENABLED=false
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# 上传配置
UPLOAD_PATH=./uploads              # 图片存储路径
MAX_UPLOAD_SIZE=10485760           # 最大上传大小 (10MB)
```

### 客户端连接配置

前端通过 `store.js` 中的 `serverConfig` 管理服务器地址：
- **自动检测**：首次启动自动检测 localhost
- **手动配置**：通过登录页设置按钮修改
- **端口优先级**：优先检测 80 端口，fallback 到 8080

---

## 🐧 Linux 服务器部署

### 使用 Systemd 管理

1. **上传文件到服务器**
```bash
scp collab_server ubuntu@your-server:/home/ubuntu/graduation_project/
scp collab.service ubuntu@your-server:/etc/systemd/system/
```

2. **配置并启动服务**
```bash
sudo systemctl daemon-reload
sudo systemctl enable collab.service
sudo systemctl start collab.service
```

3. **查看运行状态**
```bash
sudo systemctl status collab.service
journalctl -u collab.service -f
```

### 一键部署脚本

```bash
cd CollabServer
chmod +x deploy.sh
./deploy.sh
```

---

## 🔍 常见问题

### Q: 访客无法连接房主？

1. 确保房主和访客在**同一局域网**
2. 检查房主防火墙是否放行 8080（或 80）端口
3. 使用 `ping` 命令测试网络连通性
4. 确认输入的 IP 地址和端口正确

### Q: 编译时报 `undefined: err` 或 `undefined: resp`？

检查 `CollabClient/main.go` 中的 `checkPortAlive` 函数是否完整，参考最新代码修复。

### Q: 图片上传后显示为破损？

1. 确保 `CollabServer/uploads/` 目录存在且有写权限
2. 检查 `.env` 中的 `UPLOAD_PATH` 配置正确
3. 确认服务器 CORS 配置允许图片请求

### Q: 如何更换监听端口？

修改 `CollabServer/.env` 中的 `PORT` 变量，重启服务即可。

---

## 📄 API 接口文档

### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录 |

### 文件接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/upload` | 图片上传 |
| GET | `/uploads/:filename` | 获取上传的图片 |

### WebSocket 接口

| 路径 | 说明 |
|------|------|
| `/ws` | 实时协作通信通道 |

#### WebSocket 消息类型

| 类型 | 方向 | 说明 |
|------|------|------|
| `join` | C→S | 加入协作房间 |
| `leave` | C→S | 离开房间 |
| `doc_update` | C↔S | 文档内容更新 |
| `img-insert` | C↔S | 图片 URL 同步 |
| `cursor` | C↔S | 光标位置同步 |
| `user_list` | S→C | 在线用户列表 |

---

## 🛠️ 开发指南

### 添加新的编辑器扩展

1. 安装 Tiptap 扩展：
```bash
cd CollabClient/frontend
npm install @tiptap/extension-xxx
```

2. 在 `Editor.vue` 中引入并注册：
```javascript
import Xxx from '@tiptap/extension-xxx'

const editor = useEditor({
  extensions: [
    StarterKit,
    Xxx,  // 添加新扩展
  ],
})
```

### 添加新的 API 接口

1. 在 `CollabServer/controllers/` 创建控制器
2. 在 `CollabServer/router/router.go` 注册路由
3. 如需数据库模型，在 `CollabServer/models/` 中定义

---

## 📜 开源协议

本项目采用 **MIT License** 开源协议。

---

<p align="center">
Developed with ❤️ using <b>Go</b> & <b>Vue3</b>
<br>
<sub>Author: jingjie2002 | Email: 2636832425@qq.com</sub>
</p>