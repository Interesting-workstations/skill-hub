# skill-hub 生产部署（Docker Compose 全容器化）

将 **web（官网）+ server（Go 后端）+ admin（运营后台）+ MySQL** 全部容器化，
一条命令在服务器上启动，适合生产环境。

## 架构

```
                        ┌───────────────────────────────┐
 用户 ──:8080──►  web 容器（Nginx 静态 + /api 反代）    │
                        │        │                      │
 管理员 ─:8081──► admin 容器（Nginx 静态 + /api 反代）  │
                        │        ▼                      │
                        └──► server 容器（Go :8080）    │
                                 │                      │
                                 ▼                      │
                             mysql 容器（MySQL 8.0）     │
                        └───────────────────────────────┘
```

- `web` / `admin` 前端构建为静态文件，由容器内 Nginx 托管；
- `/api/*` 由 Nginx 反向代理到 `server` 容器（含 WebSocket 升级，支持执行进度实时推送）；
- `server` 与 `mysql` 只在 compose 内网通信（`expose`，不暴露到宿主机）。

## 文件说明

| 文件 | 说明 |
|---|---|
| `docker-compose.yml` | 编排：mysql / server / web / admin |
| `.env.example` | 环境配置模板（复制为 `.env` 使用） |
| `deploy.sh` | 一键部署脚本（自动装 Docker、构建、启动） |
| `../server/Dockerfile` | 后端多阶段构建（golang:1.25 → alpine） |
| `../web/Dockerfile` + `nginx/default.conf` | 官网构建 + Nginx 配置 |
| `../admin/Dockerfile` + `nginx/default.conf` | 后台构建 + Nginx 配置 |

## 本地构建验证（可选，需 Docker）

```bash
cd deploy
cp .env.example .env            # 可暂不填 GITHUB_TOKEN
docker compose build            # 先在本地验证镜像可构建
```

## 服务器部署（CentOS 7.9 实测步骤）

1. **本机授权 SSH 公钥**（只需一次）：
   ```bash
   ssh-copy-id -i ~/.ssh/id_ed25519.pub root@服务器IP
   ```
2. **上传代码**（本地执行，排除 node_modules / dist / .git / .env）：
   ```bash
   rsync -av --delete \
     --exclude 'node_modules' --exclude 'dist' --exclude '.git' \
     --exclude '.env' --exclude 'deploy/.env' \
     -e ssh ./skill-hub/ root@服务器IP:/opt/skill-hub/
   ```
3. **服务器上配置并启动**：
   ```bash
   cd /opt/skill-hub/deploy
   cp .env.example .env
   vim .env          # 填入 GITHUB_TOKEN 等
   ./deploy.sh up    # 首次会自动安装 Docker、构建镜像并启动
   ```
4. **访问**：
   - 官网：`http://服务器IP:8080`
   - 运营后台：`http://服务器IP:8081`（登录 `admin / admin123`）
   - 后端健康检查：`http://服务器IP:8080/api/v1/health`（经 Nginx 反代）

> ⚠️ 若从公网访问，请在云厂商安全组/防火墙放行 **8080 / 8081** 端口。

## 常用命令

```bash
./deploy.sh up        # 构建并启动（增量）
./deploy.sh logs      # 跟踪日志
./deploy.sh restart   # 重启容器
./deploy.sh down      # 停止并删除容器（数据卷保留）
./deploy.sh ps        # 查看状态
```

## 数据与升级

- MySQL 数据持久化在 `mysql_data` 卷（`docker volume inspect skillhub_mysql_data`），删容器不丢数据；
- 代码更新后重新上传，然后 `./deploy.sh up`（`--build` 会重建镜像）；
- 数据库结构变更由后端启动时自动迁移（ensure*Column 逻辑），无需手动执行 SQL。

## 安全提示

- `.env` 含 GITHUB_TOKEN / 数据库密码，**不要提交到 git**（已加入忽略）；
- 建议将 3306（MySQL）保持仅内网，不向公网暴露；
- 生产环境建议在 Nginx 前加 HTTPS（Caddy / Nginx + Let's Encrypt）。
