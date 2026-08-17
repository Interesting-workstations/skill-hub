# 用 Termius Beta 远程连接并部署 skill-hub

> 服务器：`47.94.5.117`（root + 密码登录，SSH 端口 22）
> 部署方案：Docker Compose 全容器化（web / server / admin / mysql）

## 一、Termius 添加主机

1. 打开 **Termius Beta**，左侧栏点 **Hosts**（或首页 "+"）。
2. 点 **New Host**，填写：
   | 字段 | 值 |
   |---|---|
   | Address | `47.94.5.117` |
   | Label | 任意（如 `skillhub`） |
   | Username | `root` |
   | Password | 输入 root 密码（不会保存到任何代码/聊天，仅存 Termius 本机） |
   | Port | `22`（默认） |
3. 保存后双击该主机即可建立 SSH 连接。

## 二、上传部署包（SFTP）

1. 在 Termius 里打开该主机，右侧切到 **SFTP** 面板。
2. 远程路径进入 `/opt`，本地选择本机文件：
   `/Users/coder-zjc/AI/项目实战/skill-hub.tar.gz`
3. 直接拖拽 / 上传到 `/opt`。

> 备选：也可在 Termius 终端里用 `scp` / `rsync` 从本机推，但需本机也装密钥或输密码；SFTP 拖拽最省事。

## 三、服务器上解压并部署

在 Termius 终端（已连上 47.94.5.117）依次执行：

```bash
cd /opt
tar -xzf skill-hub.tar.gz
cd /opt/skill-hub/deploy

# 首次：生成并编辑 .env，填入 GITHUB_TOKEN（爬虫必需）
cp .env.example .env
vim .env

# 一键部署（首次自动安装 Docker，构建镜像并启动，耗时较长）
./deploy.sh up
```

## 四、验证

```bash
./deploy.sh ps        # 四个容器都应 Up
curl http://127.0.0.1:8080/api/v1/health   # 后端健康检查
```

浏览器访问：

- 官网：`http://47.94.5.117:8080`
- 运营后台：`http://47.94.5.117:8081`（登录 `admin / admin123`）

> ⚠️ 公网访问需在阿里云安全组放行 **8080 / 8081** 端口（CentOS 7.9 若开了 firewalld 也需放行）。

## 五、常用维护命令（服务器上执行）

```bash
cd /opt/skill-hub/deploy
./deploy.sh logs      # 跟踪日志
./deploy.sh restart   # 重启容器
./deploy.sh down      # 停止（数据卷保留）
```

## 六、本机重新打包（代码有改动时）

```bash
cd /Users/coder-zjc/AI/项目实战
tar --exclude='skill-hub/web/node_modules' --exclude='skill-hub/web/dist' \
    --exclude='skill-hub/admin/node_modules' --exclude='skill-hub/admin/dist' \
    --exclude='skill-hub/.git' --exclude='skill-hub/server/.env' \
    --exclude='skill-hub/deploy/.env' \
    -czf skill-hub.tar.gz skill-hub
```

再通过 SFTP 覆盖上传到 `/opt`，服务器上 `cd /opt && tar -xzf skill-hub.tar.gz && cd skill-hub/deploy && ./deploy.sh up`。
