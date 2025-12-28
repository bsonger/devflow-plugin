# -----------------------------
# Stage 1: Build Go plugin
# -----------------------------
FROM golang:1.24.6-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git bash ca-certificates

WORKDIR /workspace

# 下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译插件二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devflow-plugin main.go

# -----------------------------
# Stage 2: Sidecar 镜像
# -----------------------------
FROM quay.io/argoproj/argocd:v2.12.4

# 创建插件目录
RUN mkdir -p /home/argocd/cmp-server/plugins

# 拷贝插件二进制
COPY --from=builder /workspace/devflow-plugin /home/argocd/cmp-server/plugins/devflow-plugin
RUN chmod +x /home/argocd/cmp-server/plugins/devflow-plugin

# 工作目录
WORKDIR /app

# 默认 CMD，必须保持官方 repo-server 启动入口
CMD ["/var/run/argocd/argocd-cmp-server"]