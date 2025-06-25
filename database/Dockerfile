# 使用官方PostgreSQL镜像作为基础
FROM postgres:15-alpine

# 设置环境变量
ENV POSTGRES_DB=credit_management
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=password

# 复制初始化脚本
COPY init.sql /docker-entrypoint-initdb.d/01-init.sql

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户（PostgreSQL容器内部已经处理了用户权限）
# 这里主要是为了保持与其他服务的一致性

# 暴露端口
EXPOSE 5432

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pg_isready -U postgres -d credit_management || exit 1

# 使用默认的PostgreSQL启动命令
CMD ["postgres"] 