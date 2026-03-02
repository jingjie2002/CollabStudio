<div align="center">
<img src="CollabClient/build/appicon.png" width="120" height="120" alt="CollabStudio Logo" />
<h1>CollabStudio</h1>
<h3>✨ 基于 Go Wails + Vue3 的局域网实时协作与 AIGC 智能编辑工作台 ✨</h3>
<p>
实时文档同步 · 纯净 Web/桌面 双端支持 · 深度集成 AI 助手 · JWT 工业级鉴权
</p>

<p>
<img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go">
<img src="https://img.shields.io/badge/Wails-v2.10+-red?style=flat-square&logo=wails" alt="Wails">
<img src="https://img.shields.io/badge/Vue-3.2+-4FC08D?style=flat-square&logo=vue.js" alt="Vue">
<img src="https://img.shields.io/badge/Tiptap-Editor-000000?style=flat-square" alt="Tiptap">
<img src="https://img.shields.io/badge/SQLite-Pure%20Go-003B57?style=flat-square&logo=sqlite" alt="SQLite">
<img src="https://img.shields.io/badge/Gemini-AI-orange?style=flat-square" alt="Gemini AI">
<img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
</p>
</div>

---

## 📖 项目简介

**CollabStudio** 是一款专为高效团队设计的**全栈实时协作与 AIGC 智能排版软件**。它突破了传统协同软件的边界，创新性地采用 CS/BS (Client-Server / Browser-Server) 混合架构体验——用户既可以直接运行极速的桌面端独立执行程序，也可以在任意公网浏览器中打开无缝网页版。

项目在实现毫秒级 WebSocket 实时同步的基础上，深度植入了基于 Google Gemini 大模型的生产力 AI 体系，成为您的第二大脑：
- 利用 **WebSocket** 与 **Yjs** 算法实现无冲突的多人实时文档光标同步编辑
- 利用 **Go Wails** 构建高性能本地化的桌面客户端
- 彻底打通的全域 **@AI 唤醒系统**，提供无死角的智能问答、润色、写作辅助
- 从 "伴随式" 局域网开黑到 **公网级 Ubuntu 守护进程部署** 的完整工业闭环

---

## 🚀 核心飞跃特性

### 🤖 全域 AIGC 深度集成 (AI Assistant)
| 特性 | 说明 |
|------|------|
| **极致流式输出** | 打通底层 HTTP Chunk 传输，无惧反向代理缓存，享受与 ChatGPT 相同的毫秒级逐字打字机效果 |
| **全环境 @AI 唤醒** | 聊天大厅自动拦截带有 `@AI` 的指令，可智能总结团队群组聊天记录，随时待命 |
| **文档一键无缝注入** | 面板与 Tiptap 渲染引擎深度桥接，AI 生成的内容一键转换为富文本 Block 瞬间插入到光标处，无需痛苦地复制粘贴 |

### 👑 灵活的 Web / Desktop 双轨引擎
- **智能环境探针**：前端自动侦测运行环境(`window.go`)。在 Wails 桌面端调用底层操作 API，在纯净 Web 浏览器时自动平滑回落为 HTTP/WebSocket 标准接口
- **自动 PNA 破壁路由**：防御 Chrome Private Network Access 跨域拦截，检测到公网部署后自动清空本地局域网（Localhost）的有毒缓存并重新寻路

### ⚡️ 毫秒级实时同步 (Yjs 驱动)
- 基于 **WebSocket** 的全双工低延迟通信机制
- 支持多人同时编辑同一文档，光标实时漂移跟随
- 后端非阻塞多房间广播引擎 + 前端 Vue 响应式自适应节流

### 🔐 工业级安全与用户体系
- **全局 JWT (JSON Web Token) 拦截器**：严格守护 HTTP/WS 两条通路，彻底封锁越权删除、越权访问等漏洞
- **完善的基础 CRUD**：不仅可以协同编辑，团队成员的对话、上传的图片资源均自动云端留存，并以安全的鉴权体系供个人查阅与管理
- 跨域 CORS 白名单防护，杜绝 CSRF/XSS 风险

### 📝 强大的全格式富文本编辑器
- 基于 **Tiptap (ProseMirror)** 内核驱动
- 支持 Markdown 实时语法（输入 `# ` 自动转标题，代码块高亮等）
- 支持直接拖放与 Ctrl+V 剪贴板上传图片，直接上云。编辑器协同呈现！

---

## 🏗️ 系统架构

CollabStudio 采用 **自适应全场景架构**，可在局域网 "免服开黑" 与 公网级部署 间无缝切换：

```text
                        ┌────────────────────────────────────────────────────────┐
                        │                Collab Server (Go + Gin)                │
                        │                                                        │
                        │   1. HTTP API (:8080)   [JWT Auth | File Uploads]      │
                        │   2. WebSocket (/ws)    [Yjs Collaboration | Cursors]  │
                        │   3. SQLite Database    [Users | History | Chats]      │
                        │   4. AI Gateway         [LLM Streaming Bridge]         │
                        └──────────────────────────┬─────────────────────────────┘
                                                   │
                ┌──────────────────────────────────┴────────────────────────────────┐
                │                                  │                                │
        ┌───────▼──────────┐              ┌────────▼─────────┐             ┌────────▼─────────┐
        │ 💻 桌面端房主本人    │              │ 💻 桌面端局域网访客   │             │ 🌐 普通 Web 浏览器访客 |
        │ (CollabClient.exe)│              │ (CollabClient.exe)│             │ (Chrome / Edge 等) |
        │   隐式启动 Server   │              │     通过 IP 接入     │             │    通过域名或公网IP   |
        └───────────────────┘              └───────────────────┘             └───────────────────┘
```

---

## 📦 技术栈核心速览

- **后端引擎**: Go 1.25+, Gin (Web框架), Gorilla WebSocket, **纯 Go 版 SQLite**, JWT (v5认证), CORS 安全策略
- **前端页面**: Vue 3.2+ (Composition API), Vite, Tiptap (编辑器), axios, RemixIcon
- **桌面包装**: Go Wails v2 (构建跨平台轻量级可执行文件)
- **大语言模型**: 适配 OpenAI 标准接口的底层逻辑，当前演示对接 Google Gemini 核心

---

## 📂 生产环境 Linux (Ubuntu) 部署指南

对于希望将项目作为真正 SaaS 级工具独立托管在云上的用户，本机构建了极度友好的交叉编译与静态分发方案。

### 1. 本地极速编译生成
不需要在 Linux 上配置庞大的 Go 环境，只需在 Windows/Mac 本地直接构建：
```bash
# 构建 Linux 后端
cd CollabServer
$Env:GOOS="linux"; $Env:GOARCH="amd64"; go build -o ../deploy/collab-server-linux-amd64 main.go

# 构建前端静态包
cd CollabClient/frontend
npm run build
# 将 /dist 拷贝进 deploy 文件夹
```

### 2. 服务器目录配置
通过 Tabby 等 SFTP 软件将 `deploy` 文件夹放置在 `/opt/collabstudio`：
```text
/opt/collabstudio/
├── collab-server-linux-amd64  # Linux 二进制文件
├── dist/                      # 前端静态包
├── .env                       # 你的服务器密钥和跨域配置 (CORS_ORIGINS=http://你的IP)
└── collab-server.service      # Systemd 脚本
```

### 3. 后台长期守护 (Systemd Daemon)
```bash
sudo chmod +x /opt/collabstudio/collab-server-linux-amd64
sudo cp /opt/collabstudio/collab-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable collab-server
sudo systemctl start collab-server
```
绿灯亮起，您的私人协作大本营即刻在全球上线！

---

## 🔍 常见使用排错

### Q: 【Web端报错】页面一直卡在 "正在连接服务器..."？
**A**: 这 100% 是浏览器触发了私有网络跨域保护 (CORS 或 PNA 拦截)：
1. 检查服务器端 `/opt/collabstudio/.env` 里的 `CORS_ORIGINS` 是否正确填写了你浏览器地址栏里的那个公网 IP 兼端口！
2. 确保你的腾讯云/阿里云防火墙安全组里放行了对应的 HTTP 协议端口（例如 8080 或 80）。
3. 使用 `Ctrl + F5` 强制刷新浏览器来激活项目内置的 **"本地毒缓存自动清洗补丁"**。

### Q: 【桌面端开发】编译时报 Wails bindings 找不到？
**A**: `wails build` 或 `wails dev` 命令会自动扫描 `app.go` 并生成前端依赖。前端任何强行调用 `window.go.main.App.XXX` 的地方都必须确保其身处在包裹了 Wails 壳的客户端内。项目中采用纯净 Web 回退模式的地方请勿改动检测器代码。

### Q: 【后端报错】Systemd 无法启动 (Status=203/EXEC)？
**A**: 这是 Linux 系统典型的“文件可执行权限不足”或者“架构类型不符合”报错。使用 `sudo chmod +x collab-server-linux-amd64` 赋予可执行权限后重启即可。

---

## 📜 开发者与开源声明

本项目完全采用 **MIT License** 开源协议，欢迎二次自由开垦与改造！

<p align="center">
Developed with ❤️ using <b>Go</b> & <b>Vue3</b>
<br>
<sub>Author: jingjie2002 | Email: 2636832425@qq.com</sub>
</p>