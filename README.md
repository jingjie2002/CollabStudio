# CollabStudio

CollabStudio 是一个基于 `Wails + Vue 3 + Go + Gin + SQLite` 的即时协作聊天与文档编辑系统，支持桌面端和浏览器端两种运行方式。项目当前的核心目标是完成局域网场景下的登录注册、房间协作、实时消息同步、协作文档编辑、图片上传展示以及桌面端交付。

## 主要功能

- 用户注册、登录与 JWT 鉴权
- 房间创建、加入与在线成员同步
- 文本消息实时收发
- 基于 Tiptap / ProseMirror 的富文本编辑
- 图片上传、展示与房间内同步
- 桌面端本地文件导入、保存与窗口退出确认
- AI 总结、翻译、润色等扩展功能

## 技术栈

### 前端

- Vue 3
- Vite
- Tiptap / ProseMirror
- Axios
- Remix Icon

### 桌面端

- Wails v2

### 后端

- Go
- Gin
- GORM
- SQLite
- JWT
- Gorilla WebSocket

## 项目结构

```text
MyGradProject/
├─ CollabClient/                  Wails 桌面客户端
│  ├─ app.go                      桌面端绑定与窗口行为
│  ├─ main.go                     Wails 应用入口
│  ├─ frontend/                   Vue 前端工程
│  └─ wails.json                  Wails 配置
├─ CollabServer/                  Go 后端服务
│  ├─ main.go                     HTTP / WebSocket 服务入口
│  ├─ config/                     配置加载
│  ├─ controllers/                业务接口
│  ├─ database/                   数据库初始化
│  ├─ middleware/                 鉴权中间件
│  ├─ models/                     数据模型
│  └─ websocket/                  协作与消息同步核心
├─ PROJECT_STATUS_REPORT.md       项目状态说明
├─ TECH_INDEX.md                  技术索引
└─ .gitignore                     Git 忽略规则
```

## 本地运行

### 1. 启动后端

```powershell
cd CollabServer
go run .
```

### 2. 启动前端开发模式

```powershell
cd CollabClient/frontend
npm install
npm run dev
```

### 3. 启动桌面端开发模式

```powershell
cd CollabClient
wails dev
```

### 4. 构建桌面端

```powershell
cd CollabClient
wails build
```

## GitHub 提交建议

这个仓库建议只提交源码、配置模板和必要说明文档，不要提交本地运行产物、数据库、日志和个人环境文件。

### 建议提交

- `CollabClient/` 下的源码与配置
- `CollabServer/` 下的源码与配置
- `README.md`
- `.gitignore`
- `PROJECT_STATUS_REPORT.md`
- `TECH_INDEX.md`
- `.env.example`

### 不建议提交

- `.env`
- `node_modules/`
- `dist/`
- `build/`
- `uploads/`
- `*.db`
- `*.log`
- `*.exe`
- `.idea/`
- `.vscode/`
- `.trae/`
- `.worktrees/`

## 当前提交边界

当前仓库根目录是 `MyGradProject/`。论文、模板、截图素材、Codex 过程文件和课程资料不在这个仓库提交范围内。

## 补充说明

- 浏览器端和桌面端共用同一套核心业务逻辑，但桌面端额外封装了本地文件操作和关闭拦截能力。
- `AI` 相关能力依赖外部密钥与网络环境，适合作为扩展能力说明，不建议作为基础运行验证的唯一依据。
- 如果准备公开仓库，提交前建议再次检查 `.env`、数据库、上传图片和构建产物是否已被正确忽略。
