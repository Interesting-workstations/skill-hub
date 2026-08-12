# Agent Skills API（后端）

Agent Skills 资源库后端服务（Go），采用**模块化单体（Modular Monolith）** + 分层架构，为前端官网（[skill-hub](https://github.com/Interesting-workstations/skill-hub)）提供技能数据 API。

## 技术栈

- Go 1.23+（标准库 `net/http`）
- MySQL 8.0（`github.com/go-sql-driver/mysql`，首次启动自动建库建表，并从 `data/skills.json` 种子数据初始化）
- 统一响应格式 `{ code, message, data }`

## 快速开始

### 1. 启动 MySQL（Docker）

```bash
docker run -d --name skillhub-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=skillhub \
  -p 3306:3306 mysql:8.0
```

### 2. 启动服务

```bash
go run ./cmd/server          # 默认监听 :8080
# 或指定端口与数据库连接
SERVER_ADDR=:9090 MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/skillhub?charset=utf8mb4&parseTime=true' go run ./cmd/server
```

首次启动时若数据库为空，会自动创建数据表（`categories` / `authors` / `skills`）并从 `data/skills.json` 填充种子数据。

### 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `SERVER_ADDR` | `:8080` | HTTP 监听地址 |
| `MYSQL_DSN` | `root:root@tcp(127.0.0.1:3306)/skillhub?charset=utf8mb4&parseTime=true` | MySQL 连接串 |
| `SEED_PATH` | `data/skills.json` | 种子数据 JSON 路径 |

### 精选技能规则

首页「精选技能」由规则算法生成（`GET /api/v1/skills?featured=true`）：

1. **官方技能优先**：官方出品排在非官方之前；
2. **GitHub 星标降序**：同一优先级下按星标数排序（支持 `861`、`168.1k`、`1.2m` 等格式）；
3. **分类多样性**：每个分类最多入选 2 个，避免首页被单一分类刷屏；
4. 默认共 6 个（`DefaultFeaturedLimit`）。

分类数量与站点统计（`GET /api/v1/stats`）均实时从数据库聚合计算，保证真实准确。

## 🕷️ 技能爬虫（cmd/crawler）

从 GitHub 自动爬取公开的 Agent Skill，输出与数据模型兼容的 JSON。

```bash
# 使用 GitHub Token 可显著提升 API 速率限制
export GITHUB_TOKEN=ghp_xxx

go run ./cmd/crawler -query "claude skills" -limit 50 -output data/crawled-skills.json
```

| 参数 | 默认值 | 说明 |
|---|---|---|
| `-query` | `agent skills` | GitHub 搜索关键词（如 `claude skills`、`codex skills`） |
| `-limit` | `20` | 最多处理的仓库数量 |
| `-per-page` | `10` | 每次搜索请求返回的仓库数（最大 100） |
| `-output` | stdout | 输出 JSON 文件路径 |

**skill 识别规则**：

1. 仓库根目录存在 `SKILL.md` → 整个仓库作为一个技能
2. 存在 `skills/`（或 `skillsets/`）目录 → 其下每个含 `SKILL.md` 的子目录为一个技能

**提取字段**：名称（SKILL.md frontmatter 或目录名）、描述、作者、标签、下载链接（zip）、GitHub 地址、stars、License。

```bash
# 示例：爬取结果合并入种子数据后重新运行服务
go run ./cmd/crawler -query "claude skills" -limit 50 -output data/crawled-skills.json
go run ./cmd/server
```

## 测试

```bash
go test ./...
```

## API 一览

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/health` | 健康检查 |
| GET | `/api/v1/stats` | 站点聚合统计（技能/作者/分类/官方/精选数量） |
| GET | `/api/v1/skills` | 技能列表（`?category=&author=&official=&featured=` 筛选） |
| GET | `/api/v1/skills/{id}` | 技能详情（含 content 内容区块） |
| GET | `/api/v1/authors` | 作者列表 |
| GET | `/api/v1/authors/{slug}` | 作者详情（含其发布的技能） |
| GET | `/api/v1/categories` | 分类列表（含分类下技能） |
| GET | `/api/v1/categories/{slug}` | 分类详情 |

### 统一响应

```json
// 成功
{ "code": 0, "message": "success", "data": { } }
// 错误（不暴露内部细节）
{ "code": 40401, "message": "技能不存在", "data": null }
```

错误码：`0` 成功 · `40001` 参数错误 · `40401` 资源不存在 · `40501` 方法不允许 · `50001` 系统错误

## 目录结构

```
server/
├── cmd/server/main.go        # 启动入口（优雅关闭）
├── internal/
│   ├── domain/model.go       # 领域模型（Skill/Author/Category）
│   ├── skill/                # 技能业务模块
│   │   ├── handler.go        #   HTTP 层：参数解析/统一响应
│   │   ├── service.go        #   业务逻辑层：筛选/统计
│   │   ├── repository.go     #   数据访问层：内存 + JSON 种子
│   │   └── dto.go            #   DTO：请求/响应结构
│   ├── response/response.go  # 统一响应与业务错误码
│   ├── middleware/           # RequestID / Recovery / Logger / CORS
│   └── router/router.go      # 路由组装
└── data/skills.json          # 种子数据
```

## 架构分层

```
HTTP
 ↓
Router + Middleware（RequestID → Recovery → Logger → CORS）
 ↓
Handler        ← 只做参数解析、调用 Service、返回响应
 ↓
Service        ← 业务规则与数据组合，不感知 HTTP
 ↓
Repository     ← 数据访问接口（当前内存实现，可替换为数据库）
 ↓
JSON 种子数据
```

## 设计说明

- **模块化单体**：按后端架构规范分层，但根据项目实际（只读资源站）未引入认证/数据库等当前不需要的能力，避免过度设计
- **Repository 接口**：业务层只依赖接口，未来接入 MySQL/PostgreSQL 时只需新增实现，无需改动 Service/Handler
- **错误处理**：统一错误码，内部错误记录日志，客户端只看到安全通用信息
- **日志**：请求 ID / 方法 / 路径 / 状态码 / 耗时
- **可测试**：Service 依赖 Repository 接口，便于单元测试

## 配置

通过环境变量配置（不硬编码）：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `SERVER_ADDR` | `:8080` | 监听地址 |
| `DATA_PATH` | `data/skills.json` | 种子数据路径 |
