<div align="center">
<img src="CollabClient/build/appicon.png" width="120" height="120" alt="CollabStudio Logo" />
<h1>CollabStudio</h1>
<h3>✨ 基于 Go Wails + Vue3 的局域网实时协作工作台 ✨</h3>
<p>
实时文档同步 · 局域网一键联机 · 房主/访客机制 · 现代化 UI
</p>

<p>
<img src="https://www.google.com/search?q=https://img.shields.io/badge/Go-1.25%2B-00ADD8%3Fstyle%3Dflat-square%26logo%3Dgo" alt="Go">
<img src="https://www.google.com/search?q=https://img.shields.io/badge/Wails-v2.9%2B-red%3Fstyle%3Dflat-square%26logo%3Dwails" alt="Wails">
<img src="https://www.google.com/search?q=https://img.shields.io/badge/Vue-3.5%2B-4FC08D%3Fstyle%3Dflat-square%26logo%3Dvue.js" alt="Vue">
<img src="https://www.google.com/search?q=https://img.shields.io/badge/Tiptap-Editor-000000%3Fstyle%3Dflat-square%26logo%3Dtiptap" alt="Tiptap">
<img src="https://www.google.com/search?q=https://img.shields.io/badge/SQLite-Pure%2520Go-003B57%3Fstyle%3Dflat-square%26logo%3Dsqlite" alt="SQLite">
<img src="https://www.google.com/search?q=https://img.shields.io/badge/License-MIT-yellow%3Fstyle%3Dflat-square" alt="License">
</p>
</div>

📖 项目简介 (Introduction)

CollabStudio 是一款专为局域网环境设计的全栈实时协作软件。它采用 CS (Client-Server) 混合架构，但在体验上做到了极致的 “单机化” —— 用户无需手动部署复杂的服务器，只需双击客户端即可自动拉起后台服务，实现“即开即用”的协作体验。

项目解决了传统 Web 协作在数据隐私和部署难度上的痛点，利用 WebSocket 实现毫秒级文档同步，利用 Go Wails 构建高性能桌面客户端，并创新性地引入了**“伴随式”房主机制**，解决了局域网工具的服务器托管难题。

🚀 核心特性 (Key Features)

👑 独创“伴随式”房主机制：

智能启动：首个启动的用户自动成为“房主”，后台静默启动 Server 核心。

无感连接：后续用户自动识别服务，以“访客”身份加入，无需复杂配置。

安全拦截：房主退出时的智能拦截保护，防止误操作导致房间解散。

⚡️ 毫秒级实时同步：

基于 WebSocket 的全双工通信。

支持多人同时编辑同一文档，光标实时跟随，无冲突。

流量控制 (Flow Control)：后端非阻塞广播 + 前端自适应节流，高并发下依然丝滑。

📝 强大的富文本编辑器：

基于 Tiptap (ProseMirror) 内核。

支持 Markdown 语法、代码块高亮、任务列表。

支持图片粘贴上传、表情包发送、图片预览。

💾 本地优先的数据权：

协作内容实时存入房主本地 SQLite 数据库。

支持将协作成果一键 导出/保存 到本地硬盘（txt/md）。

支持从本地 导入 文件内容到协作空间。

🛡️ 工业级稳定性：

解决了 Vite 依赖分身导致的编辑器崩溃问题。

实现了完善的断线重连与健康检查机制。

🏗️ 系统架构 (Architecture)

CollabStudio 采用 混合伴随架构 (Hybrid Sidecar Architecture)，前端负责 UI 与本地系统交互，后端负责数据分发与持久化。

graph TD
subgraph "Host Machine (房主)"
A[CollabClient.exe] -->|Process Spawn| B[CollabServer.exe]
B --> C[(SQLite DB)]
B --> D[Uploads Dir]
A <-->|WebSocket :8080| B
end

    subgraph "Guest Machine (访客)"
        E[CollabClient.exe] 
    end
    
    E <-->|LAN WebSocket| B
    
    note1[Wails 前端: UI / 本地文件IO]
    note2[Gin 后端: HTTP / WS / DB]
    
    A -.-> note1
    E -.-> note1
    B -.-> note2


📦 技术栈与依赖解析 (Tech Stack)

本项目精选了 Go 和 Vue 生态中最高效的库。以下是详细的依赖清单：

1. 核心依赖概览

依赖库

版本

作用与选型理由

wails/v2

v2.9+

核心框架。将 Go 后端与 Web 前端打包成原生 Windows .exe，提供系统级 API。

gin-gonic/gin

v1.11.0

Web 框架。处理 HTTP 接口（登录、注册、图片上传），路由性能极佳。

gorilla/websocket

v1.5.3

实时通讯。实现标准 WebSocket 协议，支撑文档的差异化同步与广播。

glebarez/sqlite

v1.11.0

纯 Go SQLite。关键选型：无需 CGO 即可在 Windows 编译，确保生成单文件 exe，部署极简。

@tiptap/vue-3

v3.11.1

编辑器核心。基于 ProseMirror，提供 Headless 编辑能力，高度可定制。

<details>
<summary>📦 <strong>点击展开完整依赖清单 (Dependencies Tree)</strong></summary>

Frontend (Vue 3 + Vite)

基于 npm list --depth=0

frontend
├── @tiptap/extension-code-block@3.11.1  # 代码块高亮
├── @tiptap/extension-image@3.11.1       # 图片插入支持
├── @tiptap/extension-placeholder@3.11.1 # 占位提示符
├── @tiptap/extension-task-item@3.11.1   # 任务列表子项
├── @tiptap/extension-task-list@3.11.1   # 任务列表容器
├── @tiptap/pm@3.11.1                    # ProseMirror 包管理 (防冲突)
├── @tiptap/starter-kit@3.11.1           # 基础套件 (段落, 标题, 列表等)
├── @tiptap/vue-3@3.11.1                 # Vue3 适配层
├── @vitejs/plugin-vue@3.2.0             # Vite Vue 插件
├── axios@1.13.2                         # HTTP 请求库
├── remixicon@4.7.0                      # 矢量图标库
├── vite@3.2.11                          # 构建工具
├── vue-router@4.6.3                     # 路由管理 (Login/Workspace)
└── vue@3.5.25                           # 核心 UI 框架


Backend (Go Modules)

基于 go list -m all (精选核心)

collab-server
├── [github.com/wailsapp/wails/v2](https://github.com/wailsapp/wails/v2) v2.9.2    # 桌面应用框架
├── [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) v1.11.0       # HTTP Web 框架
├── [github.com/gorilla/websocket](https://github.com/gorilla/websocket) v1.5.3    # WebSocket 协议实现
├── [github.com/glebarez/sqlite](https://github.com/glebarez/sqlite) v1.11.0     # 纯 Go SQLite 驱动
├── gorm.io/gorm v1.25.x                   # ORM 数据库映射
├── [github.com/gin-contrib/cors](https://github.com/gin-contrib/cors) v1.7.6     # 跨域中间件 (解决 Wails/Gin 通信)
├── [github.com/google/uuid](https://github.com/google/uuid) v1.3.0          # UUID 生成 (文件名/用户ID)
└── [github.com/joho/godotenv](https://github.com/joho/godotenv) v1.5.1        # 环境变量管理


</details>

📂 项目目录结构 (Project Structure)

CollabStudio/
├── CollabClient/                # 🖥️ 客户端 (Wails + Vue3)
│   ├── app.go                   # [Go] 桥接层：负责文件保存/读取等本地系统调用
│   ├── main.go                  # [Go] 程序入口：包含自动启动 Server 的逻辑
│   ├── wails.json               # Wails 项目配置
│   └── frontend/                # 🎨 前端源码
│       ├── src/
│       │   ├── components/      # Vue 组件 (Login, Workspace, Editor...)
│       │   ├── utils/           # 工具类 (cursor.js 光标同步插件)
│       │   ├── store.js         # 全局状态 (服务器 IP 配置)
│       │   └── App.vue          # 根组件
│       └── vite.config.js       # 构建配置 (包含 Alias 别名修复)
│
└── CollabServer/                # ⚙️ 服务端 (Go + Gin)
├── main.go                  # [Go] 服务器入口：启动 HTTP 和 WebSocket
├── collab.db                # [数据] SQLite 数据库 (运行时自动生成)
├── uploads/                 # [数据] 图片存储目录 (运行时自动生成)
├── config/                  # 配置模块
├── controllers/             # 业务逻辑 (Auth, Upload)
├── database/                # 数据库连接与初始化
├── models/                  # 数据库模型定义 (User, Document)
└── websocket/               # 核心同步逻辑 (Hub, Client)


⚡️ 快速开始 (Quick Start)

1. 环境要求

Go: v1.20+

Node.js: v16+

Wails: v2.9+

2. 开发环境运行

启动后端 (Server):

cd CollabServer
go run main.go


启动客户端 (Client):

cd CollabClient
wails dev


3. 生产环境打包

步骤一：编译后端

cd CollabServer
go build -ldflags "-s -w" -o CollabServer.exe main.go


步骤二：编译客户端

cd CollabClient
wails build -clean


步骤三：部署合体
将生成的 CollabServer.exe 复制到 CollabClient/build/bin/ 目录下，确保与 CollabClient.exe 同级。

🤝 使用说明

房主（Host）：

双击 CollabClient.exe，软件会自动在后台启动服务器。

左上角会显示 [👑 房主] 标识。

注意：房主关闭窗口会导致房间解散，所有人离线。

访客（Guest）：

获取房主的局域网 IP 地址（如 192.168.1.5:8080）。

打开 CollabClient.exe -> 点击下方“服务器设置” -> 输入房主 IP。

注册/登录即可加入协作。

<p align="center">Developed with ❤️ using Go & Vue3</p>