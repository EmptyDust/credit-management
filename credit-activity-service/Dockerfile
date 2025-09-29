# 多阶段构建 - 构建阶段
FROM docker.io/library/golang:1.24.4-alpine AS builder

# 设置国内 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 设置工作目录
WORKDIR /app

# 换ustc源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装必要的构建工具
RUN apk add --no-cache ca-certificates tzdata wget

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖（添加超时和重试）
RUN go mod download -x || (sleep 5 && go mod download -x) || (sleep 10 && go mod download -x)

# 复制源代码
COPY . .

# 构建应用（添加构建参数优化）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o credit-activity-service .

# 运行阶段
FROM docker.io/library/alpine:latest

# 换ustc源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/credit-activity-service .

# 复制配置文件
COPY --from=builder /app/config ./config

# 创建uploads目录
RUN mkdir -p /app/uploads/attachments

# 设置权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8083

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# 设置环境变量
ENV GIN_MODE=release
ENV PORT=8083

# 启动应用
CMD ["./credit-activity-service"] 