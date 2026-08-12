# Agent Skills 资源库

发现适用于 Claude Code、Codex 等 AI 编程助手的可复用技能资源库官网。

## 技术栈

- React 19 + TypeScript
- Vite 8（代码分割 + 生产构建）
- react-router-dom v7（集中式路由 + 按页面懒加载）
- GSAP 3（入场 / hover 微交互，支持 `prefers-reduced-motion`）
- oxlint

## 快速开始

```bash
npm install
npm run dev      # 开发
npm run build    # 类型检查 + 生产构建
npm run lint     # oxlint
npm run preview  # 预览构建产物
```

## 目录结构

```
src/
├── app/                    # 应用基础设施
│   ├── App.tsx             # ErrorBoundary + Router + Suspense 组合
│   ├── ErrorBoundary.tsx   # 全局错误边界
│   └── router/routes.tsx   # 集中式路由表（新增页面在此注册）
├── layouts/
│   └── MainLayout.tsx      # 站点主布局（Navbar + Outlet + Footer）
├── pages/                  # 页面按功能分目录
│   ├── home/  skill/  author/  category/  categories/
│   ├── official/  featured/  submit/  not-found/
├── components/
│   ├── layout/             # 跨页面布局组件（Navbar/Footer/ScrollToTop）
│   ├── sections/           # 首页/复用 Section（Hero/SkillSection/...）
│   ├── skill/              # 技能业务组件（SkillCard）
│   └── shared/             # 跨页面共享组件（PageContainer/Breadcrumb）
├── hooks/
│   ├── usePageAnimation.ts # 页面入场动画（统一清理）
│   └── usePageMeta.ts      # 轻量 SEO（title/description）
├── data/
│   ├── types.ts            # 领域类型
│   ├── skills.ts           # 静态数据
│   └── queries.ts          # 数据查询函数
├── styles/
│   ├── tokens.css          # 设计令牌（颜色/圆角 CSS 变量）
│   └── globals.css         # 全局 reset
├── config/
│   └── site.ts             # 站点元信息（名称/SEO）
└── animations/             # GSAP 动画封装层
```

## 架构约定

- **依赖方向单向向下**：`app → layouts → pages → components → hooks → data`
- `components/` 不依赖 `pages/`；UI 组件不直接处理数据请求
- 页面查询统一走 `data/queries.ts`，组件不直接操作原始数组
- 颜色一律使用 `styles/tokens.css` 中的设计令牌，禁止硬编码
- 页面路由统一在 `app/router/routes.tsx` 注册，页面使用 `usePageAnimation` + `usePageMeta`

## SEO

每个页面通过 `usePageMeta` 设置独立的 `title` / `description`，站点默认值在 `config/site.ts`。
