#!/usr/bin/env bash
# skill-hub 一键部署脚本（在服务器上运行）
# 用法：./deploy.sh [up|build|down|restart|logs|ps]
set -euo pipefail

cd "$(dirname "$0")"
CMD="${1:-up}"

# ---- 1. 检测 / 安装 Docker ----
if ! command -v docker >/dev/null 2>&1; then
  echo "==> 未检测到 Docker，开始安装…"
  if [ -f /etc/redhat-release ]; then
    # CentOS / RHEL 系
    yum install -y yum-utils || true
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo || true
    yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin || {
      echo "!! Docker CE 安装失败，请手动安装 Docker 后重试"; exit 1; }
    systemctl enable --now docker
  else
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker || true
  fi
fi

# ---- 2. 探测 compose 命令 ----
COMPOSE=""
if docker compose version >/dev/null 2>&1; then COMPOSE="docker compose"; fi
if [ -z "$COMPOSE" ] && command -v docker-compose >/dev/null 2>&1; then COMPOSE="docker-compose"; fi
if [ -z "$COMPOSE" ]; then
  echo "!! 未找到 docker compose（需要 docker-compose-plugin 或 docker-compose v1）"
  exit 1
fi
echo "==> 使用: $COMPOSE"

# ---- 3. 环境配置 ----
if [ ! -f .env ]; then
  cp .env.example .env
  echo "==> 已生成 .env，请编辑填入 GITHUB_TOKEN 等配置后重新运行 ./deploy.sh"
  exit 1
fi

# ---- 4. 执行 ----
case "$CMD" in
  up)     $COMPOSE up -d --build ;;
  build)  $COMPOSE build ;;
  down)   $COMPOSE down ;;
  restart) $COMPOSE restart ;;
  logs)   $COMPOSE logs -f --tail=100 ;;
  ps)     $COMPOSE ps ;;
  *) echo "用法: $0 [up|build|down|restart|logs|ps]"; exit 1 ;;
esac
