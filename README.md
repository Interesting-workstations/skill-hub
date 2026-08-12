# Agent Skills 资源库（前端）

发现适用于 Claude Code、Codex 等 AI 编程助手的可复用技能资源库官网。

## 技术栈

- 前端：React 19 + TypeScript + Vite 8 + react-router-dom v7 + GSAP 3
- 后端：Go 1.23（独立仓库，见 [skill-hub-server](https://github.com/Interesting-workstations/skill-hub-server)）
- 前后端分离：前端经 `services/api` 调用后端 REST API

## 快速开始

前后端为两个平级独立仓库，需分别启动：

```bash
# 1. 启动后端（:8080）— 独立仓库 skill-hub-server
git clone https://github.com/Interesting-workstations/skill-hub-server.git
cd skill-hub-server && go run ./cmd/server

# 2. 启动前端（:5173）
npm install
npm run dev
```

其他命令：

```bash
npm run build    # 类型检查 + 生产构建
npm run lint     # oxlint
```

## 目录结构

```
skill-hub/                    # 前端仓库
├── src/
│   ├── app/                  # 应用基础设施（App/ErrorBoundary/router）
│   ├── layouts/              # MainLayout（Navbar + Outlet + Footer）
│   ├── pages/                # 页面按功能分目录
│   ├── components/
│   │   ├── layout/  sections/  skill/  shared/
│   ├── hooks/
│   │   ├── usePageAnimation.ts  # 页面入场动画
│   │   ├── usePageMeta.ts       # 轻量 SEO
│   │   ├── useAsyncData.ts      # 通用异步数据三态
│   │   └── useSkillData.ts      # 技能数据 Hooks
│   ├── services/
│   │   └── api/              # API 客户端与数据获取函数
│   ├── data/
│   │   └── types.ts          # 领域类型
│   ├── styles/               # tokens.css（设计令牌）+ globals.css
│   ├── config/site.ts        # 站点元信息
│   └── animations/           # GSAP 动画封装层
└── index.html / vite.config.ts / package.json …
```

> 后端代码与种子数据位于独立仓库 [skill-hub-server](https://github.com/Interesting-workstations/skill-hub-server)。

## 数据流

```
Page / Section
   ↓
Hooks (useStats / useSkills / useAuthor …)
   ↓
services/api/skills.ts（fetch 封装）
   ↓
后端 REST API（/api/v1/*，skill-hub-server 提供）
```

## 架构约定

- **依赖方向单向向下**：`app → layouts → pages → components → hooks → services`
- `components/` 不依赖 `pages/`；UI 组件不直接处理数据请求
- 页面查询统一走 `hooks/useSkillData` → `services/api`，组件不直接 fetch
- 颜色一律使用 `styles/tokens.css` 中的设计令牌，禁止硬编码
- 页面路由统一在 `app/router/routes.tsx` 注册，页面使用 `usePageAnimation` + `usePageMeta`
- 数据源在后端仓库，前端不持有静态数据

## API 基础地址

默认 `http://localhost:8080/api/v1`，可通过环境变量 `VITE_API_BASE_URL` 覆盖。

## SEO

每个页面通过 `usePageMeta` 设置独立的 `title` / `description`，站点默认值在 `config/site.ts`。
