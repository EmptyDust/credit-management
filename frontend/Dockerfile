# 构建阶段
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装pnpm
RUN npm install -g pnpm

# 复制package文件
COPY package*.json ./
COPY pnpm-lock.yaml ./

# 安装依赖（添加缓存优化）
RUN pnpm install --frozen-lockfile --prefer-offline

# 复制源代码
COPY . .

# 构建应用（添加构建优化）
RUN pnpm run build

# 运行阶段
FROM nginx:alpine

# 安装必要的工具
RUN apk add --no-cache curl

# 复制构建产物到nginx目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置
COPY nginx.conf /etc/nginx/nginx.conf

# 创建非root用户
RUN addgroup -g 1001 -S nginxgroup && \
    adduser -u 1001 -S nginxuser -G nginxgroup

# 更改nginx配置目录权限
RUN chown -R nginxuser:nginxgroup /var/cache/nginx && \
    chown -R nginxuser:nginxgroup /var/log/nginx && \
    chown -R nginxuser:nginxgroup /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R nginxuser:nginxgroup /var/run/nginx.pid

# 切换到非root用户
USER nginxuser:nginxgroup

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"] 