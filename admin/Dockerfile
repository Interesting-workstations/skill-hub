# syntax=docker/dockerfile:1
# ---- 构建阶段：Vite 生产构建 ----
FROM node:22-alpine AS builder
WORKDIR /app

COPY package.json package-lock.json ./
# 国内网络加速：使用 npmmirror 镜像源
RUN npm config set registry https://registry.npmmirror.com && npm ci

COPY . .
# 后台隐藏入口：ADMIN_BASE=访问路径前缀（如 361d130e47），ADMIN_SECRET=API secret 前缀
ARG ADMIN_BASE=/
ARG ADMIN_SECRET=
ENV ADMIN_BASE=${ADMIN_BASE}
ENV VITE_ADMIN_BASE=${ADMIN_BASE}
ENV VITE_ADMIN_SECRET=${ADMIN_SECRET}
RUN npm run build

# ---- 运行阶段：Nginx 托管静态文件 + 反代 API ----
FROM nginx:1.27-alpine
ARG ADMIN_BASE=/
ENV ADMIN_BASE=${ADMIN_BASE}
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template
# 将构建产物放入隐藏路径下（与 nginx try_files 回退路径一致，避免 500 重定向循环）
COPY --from=builder /app/dist /usr/share/nginx/html/${ADMIN_BASE}/
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
