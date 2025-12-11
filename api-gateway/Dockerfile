FROM docker.io/library/golang:1.24-alpine AS builder

# 设置国内 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 设置工作目录
WORKDIR /app

# 换ustc源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖（使用国内代理，避免超时）
RUN go mod download -x

# 复制源代码
COPY . .

# 构建应用（优化二进制）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o main .

# 运行阶段
FROM docker.io/library/alpine:latest

# 换ustc源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装运行时依赖和 Docker CLI（用于日志查看功能）
RUN apk add --no-cache ca-certificates tzdata wget docker-cli

# 设置工作目录
WORKDIR /app

# 复制二进制
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行应用（以 root 运行以便访问 Docker socket）
CMD ["./main"] 