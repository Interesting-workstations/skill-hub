# syntax=docker/dockerfile:1
# ---- 构建阶段：编译 Go 后端（go 1.25）----
FROM golang:1.25-alpine AS builder
WORKDIR /build

# 先复制依赖清单以利用构建缓存
COPY go.mod go.sum ./
# 国内网络加速：使用 goproxy.cn 镜像源
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/skillhub-server ./cmd/server

# ---- 运行阶段：精简 alpine ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/skillhub-server /app/skillhub-server
# 种子数据（首次启动建库填充）
COPY data/skills.json /app/data/skills.json
USER app
EXPOSE 8080
CMD ["/app/skillhub-server"]
