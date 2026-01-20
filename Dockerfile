# -----------------------------
# Stage 1: Build Go plugin
# -----------------------------
FROM registry.cn-hangzhou.aliyuncs.com/devflow/golang:1.25.6 AS builder

# 安装必要工具
#RUN apk add --no-cache git bash ca-certificates
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /workspace

# 下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译插件二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devflow-plugin main.go
RUN chmod +x devflow-plugin

# -----------------------------
# Stage 2: Sidecar 镜像
# -----------------------------
FROM quay.io/argoproj/argocd:v3.2.5

# 拷贝插件二进制
COPY --from=builder /workspace/devflow-plugin /home/argocd/devflow-plugin

# 工作目录
WORKDIR /home/argocd

# 默认 CMD，必须保持官方 repo-server 启动入口
CMD ["/var/run/argocd/argocd-cmp-server"]