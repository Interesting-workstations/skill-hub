# Agent Skills 资源库

发现适用于 Claude Code、Codex 等 AI 编程助手的可复用技能资源库官网。

## 技术栈

- 前端：React 19 + TypeScript + Vite 8 + react-router-dom v7 + GSAP 3
- 后端：Go 1.23（标准库 `net/http`，零第三方依赖）
- 前后端分离：前端经 `services/api` 调用后端 REST API

## 快速开始

需要同时启动后端与前端：

```bash
# 1. 启动后端（:8080）
cd server && go run ./cmd/server

# 2. 启动前端（:5173）
npm install
npm run dev
```

其他命令：

```bash
npm run build    # 前端类型检查 + 生产构建
npm run lint     # oxlint
go -C server test ./...   # 后端测试
```

## 目录结构

```
skill-hub/
├── server/                 # Go 后端（模块化单体，见 server/README.md）
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── domain/         # 领域模型
│   │   ├── skill/          # 技能业务模块（handler/service/repository）
│   │   ├── response/       # 统一响应 {code,message,data}
│   │   ├── middleware/     # RequestID/Recovery/Logger/CORS
│   │   └── router/
│   └── data/skills.json    # 种子数据
└── src/
    ├── app/                # 应用基础设施（App/ErrorBoundary/router）
    ├── layouts/            # MainLayout（Navbar + Outlet + Footer）
    ├── pages/              # 页面按功能分目录
    ├── components/
    │   ├── layout/  sections/  skill/  shared/
    ├── hooks/
    │   ├── usePageAnimation.ts  # 页面入场动画
    │   ├── usePageMeta.ts       # 轻量 SEO
    │   ├── useAsyncData.ts      # 通用异步数据三态
    │   └── useSkillData.ts      # 技能数据 Hooks
    ├── services/
    │   └── api/            # API 客户端与数据获取函数
    ├── data/
    │   └── types.ts        # 领域类型
    ├── styles/             # tokens.css（设计令牌）+ globals.css
    ├── config/site.ts      # 站点元信息
    └── animations/         # GSAP 动画封装层
```

## 数据流

```
Page / Section
   ↓
Hooks (useStats / useSkills / useAuthor …)
   ↓
services/api/skills.ts（fetch 封装）
   ↓
后端 REST API（/api/v1/*）
   ↓
Repository（内存 + JSON 种子数据）
```

## 架构约定

- **依赖方向单向向下**：`app → layouts → pages → components → hooks → services`
- `components/` 不依赖 `pages/`；UI 组件不直接处理数据请求
- 页面查询统一走 `hooks/useSkillData` → `services/api`，组件不直接 fetch
- 颜色一律使用 `styles/tokens.css` 中的设计令牌，禁止硬编码
- 页面路由统一在 `app/router/routes.tsx` 注册，页面使用 `usePageAnimation` + `usePageMeta`
- 数据源在后端 `server/data/skills.json`，前端不再持有静态数据

## API 基础地址

默认 `http://localhost:8080/api/v1`，可通过环境变量 `VITE_API_BASE_URL` 覆盖。

## SEO

每个页面通过 `usePageMeta` 设置独立的 `title` / `description`，站点默认值在 `config/site.ts`。
