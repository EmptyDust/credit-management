# 使用官方Go镜像作为构建环境
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

# 设置Go环境变量
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖（添加超时和重试）
RUN go mod download -x || (sleep 5 && go mod download -x) || (sleep 10 && go mod download -x)

# 复制源代码
COPY . .

# 构建应用（添加构建参数优化）
RUN go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -a -installsuffix cgo \
    -o main .

# 运行阶段
FROM docker.io/library/alpine:latest

# 换ustc源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/config ./config

# 更改文件所有者
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8084

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8084/health || exit 1

# 设置环境变量
ENV GIN_MODE=release
ENV PORT=8084

# 运行应用
CMD ["./main"] 