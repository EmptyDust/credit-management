# 认证服务 (Auth Service)

这是学分管理系统的认证服务，负责用户登录、JWT token 生成、验证和登出等功能。

## 功能特性

- **用户登录** - 支持用户名/密码登录，生成 JWT token
- **Token 验证** - 验证 JWT token 的有效性和合法性
- **Token 刷新** - 支持刷新过期的 token
- **用户登出** - 将 token 加入黑名单，实现安全的登出
- **Redis 集成** - 使用 Redis 管理 token 黑名单

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7.2+

### 本地运行

```bash
# 安装依赖
go mod download

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=credit_management
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=password
export JWT_SECRET=your-secret-key
export PORT=8081

# 运行服务
go run main.go
```

### Docker 运行

```bash
# 构建镜像
docker build -t auth-service .

# 运行容器
docker run -d \
  --name auth-service \
  -p 8081:8081 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  -e REDIS_HOST=your-redis-host \
  -e REDIS_PASSWORD=your-redis-password \
  -e JWT_SECRET=your-jwt-secret \
  auth-service
```

## API 文档

### 认证相关

```http
POST /api/auth/login                    # 用户登录
POST /api/auth/validate-token           # 验证 token
POST /api/auth/validate-token-with-claims # 验证 token 并返回用户信息
POST /api/auth/refresh-token            # 刷新 token
POST /api/auth/logout                   # 用户登出
```

### 健康检查

```http
GET  /health                            # 健康检查
```

## 环境变量

| 变量名           | 说明            | 默认值              |
| ---------------- | --------------- | ------------------- |
| `DB_HOST`        | 数据库主机      | `localhost`         |
| `DB_PORT`        | 数据库端口      | `5432`              |
| `DB_USER`        | 数据库用户名    | `postgres`          |
| `DB_PASSWORD`    | 数据库密码      | `password`          |
| `DB_NAME`        | 数据库名称      | `credit_management` |
| `DB_SSLMODE`     | 数据库 SSL 模式 | `disable`           |
| `REDIS_HOST`     | Redis 主机      | `localhost`         |
| `REDIS_PORT`     | Redis 端口      | `6379`              |
| `REDIS_PASSWORD` | Redis 密码      | `password`          |
| `JWT_SECRET`     | JWT 密钥        | `your-secret-key`   |
| `PORT`           | 服务端口        | `8081`              |

## JWT Token 管理

### Token 生成

登录成功后，服务会生成包含以下信息的 JWT token：

- `user_id`: 用户 UUID
- `username`: 用户名
- `user_type`: 用户类型（student/teacher/admin）
- `exp`: 过期时间（默认 24 小时）
- `iat`: 签发时间

### Token 验证

系统通过以下方式验证 token：

1. 解析 token 并验证签名
2. 检查 token 是否过期
3. 检查 token 是否在黑名单中（Redis）

### Token 黑名单

登出时，token 会被添加到 Redis 黑名单中，有效期与 token 过期时间相同。这样可以：

- 立即使 token 失效
- 防止 token 在过期前被继续使用
- 支持多服务器部署

## Redis 使用

### 黑名单键格式

```
blacklist:{token}
```

### 会话管理（预留）

```
session:{user_id}
```

## 数据库依赖

认证服务依赖以下数据库表：

- `users`: 用户表，用于验证用户名和密码

## 安全考虑

1. **密码加密**: 使用 bcrypt 加密存储密码
2. **Token 过期**: JWT token 设置合理的过期时间
3. **黑名单机制**: 登出时立即撤销 token
4. **HTTPS**: 生产环境建议使用 HTTPS
5. **密钥管理**: JWT_SECRET 应使用强随机字符串，并妥善保管

## 故障排除

### 常见问题

1. **Redis 连接失败**

   - 检查 Redis 服务是否运行
   - 确认 Redis 连接配置正确
   - 检查网络连接

2. **数据库连接失败**

   - 检查数据库服务是否运行
   - 确认数据库连接配置正确
   - 检查数据库用户权限

3. **Token 验证失败**
   - 确认 JWT_SECRET 配置一致
   - 检查 token 是否过期
   - 验证 token 格式是否正确
