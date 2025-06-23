# Docker构建优化指南 (14核32G配置)

## 概述

本文档描述了为14核32G内存配置优化的Docker构建和运行配置，以防止编译时死机并提供最佳性能。

## 硬件配置

- **CPU**: 14核心
- **内存**: 32GB
- **推荐磁盘**: SSD 100GB+

## 主要优化措施

### 1. 统一Go版本
- 所有Go服务统一使用 `golang:1.24.4-alpine`
- 避免版本不一致导致的构建问题

### 2. 构建阶段优化
- 使用多阶段构建减少镜像大小
- 添加构建参数优化：
  ```dockerfile
  RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags="-w -s" \
      -a -installsuffix cgo \
      -o main .
  ```

### 3. 依赖下载优化
- 添加重试机制防止网络问题：
  ```dockerfile
  RUN go mod download -x || (sleep 5 && go mod download -x) || (sleep 10 && go mod download -x)
  ```

### 4. 资源限制 (14核32G配置)

#### 构建时资源限制
- **API网关**: 4GB内存, 3.0CPU
- **用户管理服务**: 4GB内存, 3.0CPU
- **认证服务**: 4GB内存, 3.0CPU
- **申请管理服务**: 6GB内存, 4.0CPU (文件处理需求)
- **事务管理服务**: 4GB内存, 3.0CPU
- **学生信息服务**: 4GB内存, 3.0CPU
- **教师信息服务**: 4GB内存, 3.0CPU
- **前端应用**: 3GB内存, 2.0CPU

#### 运行时资源限制
- **PostgreSQL**: 4GB内存, 4.0CPU (数据库核心)
- **API网关**: 2GB内存, 2.0CPU
- **用户管理服务**: 2GB内存, 2.0CPU
- **认证服务**: 2GB内存, 2.0CPU
- **申请管理服务**: 4GB内存, 3.0CPU (文件处理)
- **事务管理服务**: 2GB内存, 2.0CPU
- **学生信息服务**: 2GB内存, 2.0CPU
- **教师信息服务**: 2GB内存, 2.0CPU
- **前端应用**: 1GB内存, 1.0CPU

### 5. 安全优化
- 创建非root用户运行应用
- 最小化运行时依赖
- 使用Alpine Linux基础镜像

### 6. 健康检查
- 为所有服务添加健康检查
- 检查间隔：30秒
- 超时时间：3秒
- 重试次数：3次

### 7. 构建顺序优化
- 避免并行构建导致的资源竞争
- 按依赖关系顺序构建服务

## 构建脚本

### Windows版本 (build-optimized.bat)
```batch
# 设置构建参数
set DOCKER_BUILDKIT=1
set COMPOSE_DOCKER_CLI_BUILD=1

# 逐个构建服务 (14核32G配置)
docker build --no-cache --memory=4g --cpus=3.0 -t service_name ./service_path
docker build --no-cache --memory=6g --cpus=4.0 -t application_service ./application-management-service
```

### Linux版本 (build-optimized.sh)
```bash
#!/bin/bash
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# 逐个构建服务 (14核32G配置)
docker build --no-cache --memory=4g --cpus=3.0 -t service_name ./service_path
docker build --no-cache --memory=6g --cpus=4.0 -t application_service ./application-management-service
```

## 故障排除

### 1. 构建超时
- 检查网络连接
- 增加构建超时时间
- 使用国内镜像源

### 2. 内存不足
- 关闭其他应用程序
- 增加Docker内存限制到16GB+
- 使用交换分区

### 3. 依赖下载失败
- 检查网络代理设置
- 使用国内Go模块代理
- 重试构建

### 4. 端口冲突
- 检查端口占用情况
- 修改docker-compose.yml中的端口映射

## 性能监控

### 构建性能指标 (14核32G)
- 构建时间：每个服务约1-3分钟
- 内存使用：峰值约6GB
- CPU使用：峰值约4核心

### 运行时性能指标 (14核32G)
- 内存使用：总计约20GB
- CPU使用：总计约18核心
- 磁盘使用：约2GB

## 最佳实践

1. **构建前准备**
   - 清理Docker缓存：`docker system prune -f`
   - 确保有足够的磁盘空间 (50GB+)
   - 关闭不必要的应用程序

2. **构建过程**
   - 使用提供的构建脚本
   - 监控系统资源使用
   - 不要中断构建过程

3. **构建后检查**
   - 验证所有服务状态：`docker-compose ps`
   - 检查服务日志：`docker-compose logs`
   - 测试API端点

## 环境要求

- Docker Desktop 4.0+
- 内存：32GB (已满足)
- 磁盘空间：100GB+ (推荐SSD)
- CPU：14核心 (已满足)

## 资源分配策略

### 构建阶段 (总计约6GB内存, 4核心)
- 为每个服务分配足够的构建资源
- 申请管理服务获得更多资源 (文件处理)
- 前端构建使用中等资源

### 运行阶段 (总计约20GB内存, 18核心)
- PostgreSQL获得最多资源 (数据库核心)
- 申请管理服务获得较多资源 (文件处理)
- 其他微服务平均分配资源
- 前端使用最少资源

## 联系支持

如果遇到构建问题，请：
1. 检查本文档的故障排除部分
2. 查看Docker日志
3. 确认系统资源充足
4. 联系技术支持团队 