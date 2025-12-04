# Redis 配置和使用指南

## 概述

本项目使用 Redis 作为缓存和会话存储，主要用于：

- **JWT Token 黑名单管理** - 存储已撤销的 JWT token，实现登出功能
- **用户会话缓存** - 存储用户会话信息（可选）
- **系统数据缓存** - 缓存常用数据（预留功能）

**实际使用情况**：目前 Redis 主要由 `auth-service` 使用，用于 JWT token 黑名单管理。

## 快速开始

### 1. 启动 Redis 服务

```bash
# 使用Docker Compose启动
docker-compose up -d redis

# 或者使用启动脚本
./redis/start-redis.sh
```

### 2. 连接 Redis

```bash
# 使用Docker命令
docker exec -it credit_management_redis redis-cli

# 使用脚本
./redis/redis-cli.sh

# 直接连接（如果Redis在本地运行）
redis-cli -h localhost -p 6379
```

### 3. 基本操作

```bash
# 查看所有键
KEYS *

# 查看黑名单中的token
KEYS blacklist:*

# 查看用户会话
KEYS session:*

# 查看缓存数据
KEYS cache:*

# 删除所有数据（谨慎使用）
FLUSHALL
```

## 配置说明

### Redis 配置文件 (redis.conf)

- **端口**: 6379
- **内存限制**: 256MB
- **持久化**: RDB + AOF
- **安全**: 可配置密码保护

### 环境变量

```bash
REDIS_HOST=localhost      # Redis主机地址
REDIS_PORT=6379          # Redis端口
REDIS_PASSWORD=          # Redis密码（可选）
```

## 功能模块

### 1. JWT Token 黑名单

用于管理已撤销的 JWT token：

```bash
# 查看黑名单中的token
KEYS blacklist:*

# 查看特定token是否在黑名单中
EXISTS blacklist:your-token-here
```

### 2. 用户会话管理

存储用户会话信息：

```bash
# 查看用户会话
HGETALL session:user-id

# 查看所有会话
KEYS session:*
```

### 3. 系统缓存

缓存常用数据：

```bash
# 查看缓存数据
KEYS cache:*

# 获取缓存值
GET cache:key-name
```

## 监控和维护

### 1. 查看 Redis 状态

```bash
# 查看Redis信息
INFO

# 查看内存使用
INFO memory

# 查看连接数
INFO clients
```

### 2. 性能监控

```bash
# 查看慢查询
SLOWLOG GET 10

# 查看命令统计
INFO commandstats
```

### 3. 数据备份

```bash
# 手动触发RDB备份
BGSAVE

# 查看备份文件
docker exec -it credit_management_redis ls -la /data/
```

## 故障排除

### 1. 连接问题

```bash
# 检查Redis是否运行
docker-compose ps redis

# 查看Redis日志
docker-compose logs redis

# 测试连接
redis-cli -h localhost -p 6379 PING
```

### 2. 内存问题

```bash
# 查看内存使用情况
INFO memory

# 清理过期数据
redis-cli FLUSHDB

# 查看内存策略
CONFIG GET maxmemory-policy
```

### 3. 性能问题

```bash
# 查看慢查询
SLOWLOG GET 10

# 查看命令执行统计
INFO commandstats

# 监控实时命令
MONITOR
```

## 安全建议

1. **设置强密码**: 在生产环境中设置 Redis 密码
2. **网络隔离**: 限制 Redis 只接受内部网络连接
3. **定期备份**: 定期备份 Redis 数据
4. **监控告警**: 设置内存使用和连接数监控
5. **访问控制**: 限制 Redis 访问权限

## 开发集成

### Go 代码中使用 Redis

在 `auth-service` 中使用 Redis 客户端：

```go
import "credit-management/auth-service/utils"

// 创建 Redis 客户端
redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
redisClient := utils.NewRedisClient(redisAddr, redisPassword, 0)
if redisClient == nil {
    log.Fatal("Failed to connect to Redis")
}
defer redisClient.Close()

// 添加到黑名单（登出时）
ctx := context.Background()
expiration := 24 * time.Hour // token 过期时间
err := redisClient.AddToBlacklist(ctx, token, expiration)

// 检查黑名单（验证 token 时）
blacklisted, err := redisClient.IsBlacklisted(ctx, token)
if blacklisted {
    // token 已被撤销
}

// 设置用户会话（可选）
sessionData := map[string]interface{}{
    "user_id": userID,
    "username": username,
}
err := redisClient.SetUserSession(ctx, userID, sessionData, time.Hour)

// 获取用户会话（可选）
session, err := redisClient.GetUserSession(ctx, userID)

// 删除用户会话（登出时）
err := redisClient.DeleteUserSession(ctx, userID)

// 设置缓存（预留功能）
err := redisClient.SetCache(ctx, "cache:key", "value", time.Hour)

// 获取缓存（预留功能）
value, err := redisClient.GetCache(ctx, "cache:key")
```

## 生产环境部署

1. **高可用**: 使用 Redis Sentinel 或 Redis Cluster
2. **持久化**: 确保 RDB 和 AOF 配置正确
3. **监控**: 集成监控系统（如 Prometheus + Grafana）
4. **备份**: 设置自动备份策略
5. **安全**: 启用密码认证和网络隔离
