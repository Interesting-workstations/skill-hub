# Agent Skills 运营后台（admin）

单管理员运营控制台 —— 围绕「官网内容 → 爬虫 → 数据 → 审核 → 发布」业务链设计。
界面强调信息密度、状态反馈与操作效率，不做传统企业 OA 式设计。

## 技术栈

- React 19 + TypeScript + Vite 8 + react-router-dom v7
- 原生 CSS Design Tokens（无 UI 库，风格统一、克制）
- 官网数据对接 skill-hub 后端（`/api/v1`，经 Vite 代理）
- 爬虫任务模块为 Mock 数据（后端任务 API 尚未实现，接口已预留）

## 快速开始

```bash
# 依赖安装（首次）
npm install

# 开发服务器（:5174，/api 代理到后端 :8080）
npm run dev

# 生产构建
npm run build

# 演示账号
admin / admin123
```

> 需先启动 skill-hub 后端（`server/`），官网数据模块才能展示真实数据。

## 模块

| 模块 | 页面 | 数据来源 |
|---|---|---|
| 工作台 | 数据概览 / 今日任务 / 爬虫状态 / 最近数据 | 后端 stats + Mock |
| 爬虫管理 | 爬虫任务 / 执行记录 / 执行详情 / 失败任务 / 爬虫配置 | Mock |
| 数据管理 | 抓取数据 / 数据审核 / 数据导出 | Mock |
| 官网内容 | 分类管理 / 首页内容 / 文章管理 / SEO 配置 | 后端 categories/stats/featured + Mock |
| 系统设置 | 管理员设置 / 网站基础配置 | Mock |

## 目录结构

```
admin/
├── src/
│   ├── api/            # 接口层（site 对接后端 / crawler·content Mock）
│   ├── core/           # http 封装 / 认证 / mock 数据
│   ├── components/     # 通用组件（AppTable/AppDialog/TaskStatus/TaskProgress/ExecutionLog/StatCard）
│   ├── layouts/        # AdminLayout（深色侧边栏 + 顶栏）
│   ├── pages/          # 各业务模块页面
│   ├── store/          # 全局状态（管理员 / 侧边栏折叠）
│   ├── styles/         # design tokens + 全局组件样式
│   ├── config/         # 站点与后端配置
│   ├── utils/          # 格式化工具
│   └── types.ts        # 领域类型
└── vite.config.ts      # :5174，/api 代理 → http://localhost:8080
```

## 认证与安全

- 单管理员模型（无 RBAC / 部门 / 组织架构）
- Token 存储于 localStorage，请求头携带 `Authorization: Bearer`
- 登录页含失败提示；管理员设置含登录失败限制 / 会话有效期说明
