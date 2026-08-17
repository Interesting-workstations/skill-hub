# syntax=docker/dockerfile:1
# ---- 构建阶段：Vite 生产构建 ----
FROM node:22-alpine AS builder
WORKDIR /app

COPY package.json package-lock.json ./
# 国内网络加速：使用 npmmirror 镜像源
RUN npm config set registry https://registry.npmmirror.com && npm ci

COPY . .
# 生产环境 API 走相对路径（经 Nginx 反代到后端）
ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN npm run build

# ---- 运行阶段：Nginx 托管静态文件 + 反代 API ----
FROM nginx:1.27-alpine
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
