# Docker 部署指南

## 概述

本指南详细说明了如何使用Docker和Docker Compose部署学分管理系统。

## 系统架构

### 服务组件

1. **API网关** (api-gateway:8080)
   - 统一API入口
   - 路由转发
   - 负载均衡
   - CORS处理

2. **认证服务** (auth-service:8081)
   - 用户认证
   - JWT令牌管理
   - 权限验证

3. **学分活动服务** (credit-activity-service:8083)
   - 学分活动管理
   - 参与者管理
   - 申请管理
   - 自动申请生成

4. **用户管理服务** (user-management-service:8084)
   - 用户信息管理
   - 通知管理

5. **学生信息服务** (student-info-service:8085)
   - 学生信息管理

6. **教师信息服务** (teacher-info-service:8086)
   - 教师信息管理

7. **前端应用** (frontend:3000)
   - React + TypeScript + Tailwind CSS

8. **数据库** (postgres:5432)
   - PostgreSQL数据库

## 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少4GB可用内存
- 至少10GB可用磁盘空间

## 快速部署

### 1. 克隆项目

```bash
git clone <repository-url>
cd credit-management
```

### 2. 环境配置

复制并修改环境变量文件：

```bash
# 创建环境变量文件
cp .env.example .env

# 编辑环境变量
nano .env
```

主要环境变量：

```env
# 数据库配置
POSTGRES_DB=credit_management
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password

# JWT配置
JWT_SECRET=your_jwt_secret_key

# 服务端口
API_GATEWAY_PORT=8080
AUTH_SERVICE_PORT=8081
CREDIT_ACTIVITY_SERVICE_PORT=8083
USER_SERVICE_PORT=8084
STUDENT_SERVICE_PORT=8085
TEACHER_SERVICE_PORT=8086
FRONTEND_PORT=3000
POSTGRES_PORT=5432
```

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f
```

### 4. 验证部署

```bash
# 检查API网关健康状态
curl http://localhost:8080/health

# 检查数据库连接
docker-compose exec postgres pg_isready -U postgres

# 检查前端访问
curl http://localhost:3000
```

## 服务管理

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 启动特定服务
docker-compose up -d api-gateway
docker-compose up -d credit-activity-service
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止特定服务
docker-compose stop credit-activity-service
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart credit-activity-service
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f credit-activity-service

# 查看最近100行日志
docker-compose logs --tail=100 credit-activity-service
```

### 进入容器

```bash
# 进入数据库容器
docker-compose exec postgres psql -U postgres -d credit_management

# 进入服务容器
docker-compose exec credit-activity-service sh
```

## 数据管理

### 数据库备份

```bash
# 创建备份
docker-compose exec postgres pg_dump -U postgres credit_management > backup_$(date +%Y%m%d_%H%M%S).sql

# 备份到指定目录
docker-compose exec postgres pg_dump -U postgres credit_management > /backup/credit_management_backup.sql
```

### 数据库恢复

```bash
# 恢复数据库
docker-compose exec -T postgres psql -U postgres credit_management < backup_file.sql
```

### 数据卷管理

```bash
# 查看数据卷
docker volume ls

# 备份数据卷
docker run --rm -v credit_management_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_data_backup.tar.gz -C /data .

# 恢复数据卷
docker run --rm -v credit_management_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_data_backup.tar.gz -C /data
```

## 监控和维护

### 服务健康检查

```bash
# 检查所有服务状态
docker-compose ps

# 检查服务健康状态
curl http://localhost:8080/health
```

### 资源监控

```bash
# 查看容器资源使用情况
docker stats

# 查看磁盘使用情况
docker system df
```

### 日志分析

```bash
# 查看错误日志
docker-compose logs | grep ERROR

# 查看特定时间段的日志
docker-compose logs --since="2024-01-01T00:00:00" --until="2024-01-02T00:00:00"
```

## 故障排除

### 常见问题

1. **服务启动失败**
   ```bash
   # 查看详细错误信息
   docker-compose logs service-name
   
   # 检查端口冲突
   netstat -tulpn | grep :8080
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose exec postgres pg_isready -U postgres
   
   # 检查数据库日志
   docker-compose logs postgres
   ```

3. **服务间通信失败**
   ```bash
   # 检查网络连接
   docker network ls
   docker network inspect credit_management_credit_network
   
   # 测试服务间连接
   docker-compose exec api-gateway ping credit-activity-service
   ```

### 性能优化

1. **资源限制**
   ```yaml
   # 在docker-compose.yml中添加资源限制
   services:
     credit-activity-service:
       deploy:
         resources:
           limits:
             memory: 512M
             cpus: '0.5'
   ```

2. **日志轮转**
   ```yaml
   # 配置日志轮转
   services:
     credit-activity-service:
       logging:
         driver: "json-file"
         options:
           max-size: "10m"
           max-file: "3"
   ```

## 安全配置

### 网络安全

```yaml
# 配置网络安全
services:
  postgres:
    networks:
      - internal_network
  
  api-gateway:
    networks:
      - external_network
      - internal_network

networks:
  internal_network:
    internal: true
  external_network:
    driver: bridge
```

### 环境变量安全

```bash
# 使用Docker secrets管理敏感信息
echo "your_secure_password" | docker secret create db_password -

# 在docker-compose.yml中使用secrets
services:
  postgres:
    secrets:
      - db_password
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
```

## 生产环境部署

### 高可用配置

```yaml
# 配置服务副本
services:
  api-gateway:
    deploy:
      replicas: 3
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
```

### 负载均衡

```yaml
# 配置负载均衡
services:
  api-gateway:
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
```

### 监控集成

```yaml
# 集成Prometheus监控
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

## 更新和升级

### 服务更新

```bash
# 拉取最新代码
git pull origin main

# 重新构建镜像
docker-compose build

# 更新服务
docker-compose up -d --force-recreate
```

### 数据库迁移

```bash
# 运行数据库迁移
docker-compose exec credit-activity-service ./migrate up

# 回滚迁移
docker-compose exec credit-activity-service ./migrate down 1
```

## 备份和恢复

### 完整备份

```bash
#!/bin/bash
# 创建完整备份脚本

BACKUP_DIR="/backup/$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR

# 备份数据库
docker-compose exec postgres pg_dump -U postgres credit_management > $BACKUP_DIR/database.sql

# 备份配置文件
cp docker-compose.yml $BACKUP_DIR/
cp .env $BACKUP_DIR/

# 备份数据卷
docker run --rm -v credit_management_postgres_data:/data -v $BACKUP_DIR:/backup alpine tar czf /backup/postgres_data.tar.gz -C /data .

echo "备份完成: $BACKUP_DIR"
```

### 完整恢复

```bash
#!/bin/bash
# 创建完整恢复脚本

BACKUP_DIR="$1"
if [ -z "$BACKUP_DIR" ]; then
    echo "请指定备份目录"
    exit 1
fi

# 停止服务
docker-compose down

# 恢复数据卷
docker run --rm -v credit_management_postgres_data:/data -v $BACKUP_DIR:/backup alpine tar xzf /backup/postgres_data.tar.gz -C /data

# 启动数据库
docker-compose up -d postgres

# 等待数据库启动
sleep 10

# 恢复数据库
docker-compose exec -T postgres psql -U postgres credit_management < $BACKUP_DIR/database.sql

# 启动所有服务
docker-compose up -d

echo "恢复完成"
```

## 总结

本部署指南涵盖了学分管理系统的完整Docker部署流程，包括：

- 快速部署步骤
- 服务管理命令
- 数据管理操作
- 监控和维护
- 故障排除
- 安全配置
- 生产环境部署
- 更新和升级
- 备份和恢复

遵循本指南可以确保系统稳定、安全、高效地运行。 