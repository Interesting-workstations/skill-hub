# Agent Skills 运营后台（admin）

单管理员运营控制台 —— 围绕「官网内容 → 爬虫 → 数据 → 审核 → 发布」业务链设计。
界面强调信息密度、状态反馈与操作效率，不做传统企业 OA 式设计。

## 技术栈

- React 19 + TypeScript + Vite 8 + react-router-dom v7
- 原生 CSS Design Tokens（无 UI 库，风格统一、克制）
- **全部数据由 Go 后端提供**（`/api/v1` + `/api/v1/admin`，经 Vite 代理），无 Mock
- 认证：Bearer Token（登录由后端校验）

## 快速开始

```bash
# 依赖安装（首次）
npm install

# 开发服务器（:5174，/api 代理到后端 :8080）
npm run dev

# 生产构建
npm run build

# 登录账号
admin / admin123（由 Go 后端 admin_users 表校验）
```

> 需先启动 skill-hub 后端（`server/`）与 MySQL。爬虫任务执行会真实调用 GitHub API 抓取数据。

## 模块

| 模块 | 页面 | 数据来源 |
|---|---|---|
| 工作台 | 数据概览 / 爬虫状态 / 待审核 | 后端 `/api/v1/admin/stats` |
| 爬虫管理 | 爬虫任务 / 执行记录 / 执行详情 / 失败任务 / 爬虫配置 | 后端 `/api/v1/admin/*`（真实执行） |
| 数据管理 | 抓取数据 / 数据审核 / 数据导出 | 后端 `/api/v1/admin/data`（skills 表） |
| 官网内容 | 分类管理 / 首页内容 / 文章管理 / SEO 配置 | 后端 `/api/v1` + `/api/v1/admin/*` |
| 系统设置 | 管理员设置 / 网站基础配置 | 后端 `/api/v1/admin/*` |

## 后端接口一览

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/admin/login` | 管理员登录（返回 Token） |
| GET | `/api/v1/admin/stats` | 工作台聚合统计 |
| GET/POST | `/api/v1/admin/tasks` | 爬虫任务列表 / 新建 |
| PUT/DELETE | `/api/v1/admin/tasks/{id}` | 更新 / 删除任务 |
| POST | `/api/v1/admin/tasks/{id}/run` | 执行任务（真实爬虫，异步） |
| POST | `/api/v1/admin/tasks/{id}/stop` | 停止任务 |
| GET | `/api/v1/admin/executions` · `/executions/{id}` | 执行记录（轮询详情） |
| GET | `/api/v1/admin/failures` | 失败任务 |
| GET/PUT | `/api/v1/admin/config` | 爬虫配置 |
| GET | `/api/v1/admin/data?status=` | 抓取数据（含审核状态） |
| PUT | `/api/v1/admin/data/{id}/status` | 数据审核 / 发布 |
| GET/POST/DELETE | `/api/v1/admin/articles` | 文章管理 |
| GET/PUT | `/api/v1/admin/seo` · `/site-config` | SEO / 站点配置 |
| PUT | `/api/v1/admin/password` | 修改密码 |

除登录外均需 `Authorization: Bearer <token>`。

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
