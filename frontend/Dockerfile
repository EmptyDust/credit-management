# 构建阶段
FROM docker.io/library/node:18-alpine AS builder

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
FROM docker.io/library/nginx:alpine

# 1. 彻底删除所有 entrypoint 相关文件
RUN rm -rf /docker-entrypoint.sh /docker-entrypoint.d

# 2. 删除所有默认配置
RUN rm -rf /etc/nginx/conf.d/* /etc/nginx/nginx.conf

# 3. 复制你的完整配置
COPY nginx.conf /etc/nginx/nginx.conf

# 复制前端构建产物到 nginx 的 html 目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 4. 修正权限
RUN chown -R nginx:nginx /var/cache/nginx /var/log/nginx /etc/nginx /usr/share/nginx/html /tmp

# 5. 非 root 用户
USER nginx
EXPOSE 80

# 6. 直接启动 nginx，跳过 entrypoint
ENTRYPOINT ["nginx", "-g", "daemon off;"]