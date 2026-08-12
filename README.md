# Agent Skills 资源库

> AI 编程助手的可复用技能资源库 —— 官网 + 后端 API + 运营后台的 Monorepo 项目。

## 📁 项目结构

三个子项目为**平级目录**，共同组成一个 Git 仓库：

```
skill-hub/
├── web/          # 前端官网（React 19 + Vite + TypeScript + GSAP）
├── server/       # 后端（Go 模块化单体 REST API + MySQL + 爬虫）
├── admin/        # 运营后台（React 19 + Vite + TypeScript，单管理员）
└── README.md     # 本文件（仓库总览）
```

| 子项目 | 说明 | 详细文档 |
|---|---|---|
| `web/` | 技能资源库官网（页面/组件/设计令牌/SEO） | [web/README.md](web/README.md) |
| `server/` | 技能数据 REST API（分层架构/统一响应/MySQL/爬虫） | [server/README.md](server/README.md) |
| `admin/` | 单管理员运营后台（工作台/爬虫管理/数据管理/官网内容） | [admin/README.md](admin/README.md) |

## 🚀 快速开始

三个服务需要分别启动。前端官网与运营后台通过 `http://localhost:8080/api/v1` 调用后端。

```bash
# 1. 启动 MySQL（Docker，首次）
docker run -d --name skillhub-mysql \
  -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=skillhub \
  -p 3306:3306 mysql:8.0

# 2. 启动后端（:8080）
cd server
go run ./cmd/server

# 3. 启动官网（:5173）
cd web
npm install        # 首次
npm run dev

# 4. 启动运营后台（:5174）
cd admin
npm install        # 首次
npm run dev
```

> 运营后台演示账号：`admin / admin123`（爬虫任务模块为 Mock 数据，官网数据对接真实后端）。

## 🛠️ 常用命令

| 位置 | 命令 | 说明 |
|---|---|---|
| `web/` | `npm run dev` | 官网开发服务器（:5173） |
| `web/` | `npm run build` | 官网类型检查 + 生产构建 |
| `admin/` | `npm run dev` | 运营后台开发服务器（:5174） |
| `admin/` | `npm run build` | 运营后台类型检查 + 生产构建 |
| `server/` | `go run ./cmd/server` | 启动后端（:8080） |
| `server/` | `go test ./...` | 后端单元测试 |

## 🏗️ 架构总览

```
浏览器（web/ 前端）
   ↓  REST API（/api/v1/*）
server/ 后端
   ↓
Handler → Service → Repository → JSON 种子数据
```

- **前端**：React 19 SPA，路由懒加载 + 代码分割，设计令牌统一视觉，每页独立 SEO。
- **后端**：Go 模块化单体，Handler/Service/Repository 分层，统一响应 `{code, message, data}`，内存 Repository（预留数据库替换点）。

## 📄 许可

本仓库为个人项目，采用 **CC BY-NC 4.0**（署名-非商业性使用）。
