# 用户管理微服务

## 概述

用户管理微服务是创新创业学分管理平台的核心服务之一，负责用户基本信息管理、用户注册、用户信息查询和更新等功能。

## 功能特性

### 1. 用户管理功能
- **用户注册**: 支持学生、教师、管理员用户注册
- **用户信息查询**: 支持按用户名查询用户信息
- **用户信息更新**: 支持更新用户基本信息
- **用户删除**: 支持软删除用户
- **用户统计**: 提供用户数量、类型、状态等统计信息

### 2. 数据统计
- **用户统计**: 用户数量、类型、状态等统计
- **实时数据**: 支持今日、本周、本月等时间维度统计
- **分页查询**: 支持用户列表分页查询

### 3. 安全特性
- **密码加密**: 使用bcrypt加密存储密码
- **JWT认证**: 支持JWT token认证
- **数据验证**: 输入参数验证和过滤

## API接口

### 用户管理
```
POST   /api/users/register          # 用户注册
GET    /api/users/:username         # 获取用户信息
PUT    /api/users/:username         # 更新用户信息
DELETE /api/users/:username         # 删除用户
GET    /api/users                   # 获取所有用户（分页）
GET    /api/users/type/:userType    # 根据用户类型获取用户
GET    /api/users/stats             # 获取用户统计信息
```

## 数据库设计

### 用户表 (users)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20),
    real_name VARCHAR(100),
    user_type VARCHAR(20) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    avatar VARCHAR(255),
    last_login_at TIMESTAMP,
    register_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## 配置说明

### 环境变量
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=credit_management
DB_SSLMODE=disable

# JWT配置
JWT_SECRET=your-secret-key

# 服务配置
PORT=8080
```

## 部署说明

### Docker部署
```dockerfile
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### K8s部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-management-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-management-service
  template:
    metadata:
      labels:
        app: user-management-service
    spec:
      containers:
      - name: user-management-service
        image: user-management-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
```

## 测试

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v -run TestUserRegister

# 运行性能测试
go test -bench=.
```

### API测试
```bash
# 用户注册
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "email": "test@example.com",
    "phone": "13800138000",
    "real_name": "测试用户",
    "user_type": "student"
  }'

# 获取用户信息
curl -X GET http://localhost:8080/api/users/testuser \
  -H "Authorization: Bearer YOUR_TOKEN"

# 更新用户信息
curl -X PUT http://localhost:8080/api/users/testuser \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com",
    "phone": "13900139000",
    "real_name": "新姓名"
  }'

# 获取用户统计
curl -X GET http://localhost:8080/api/users/stats
```

## 性能优化建议

1. **数据库优化**
   - 为常用查询字段添加索引
   - 使用连接池管理数据库连接
   - 定期清理过期数据

2. **缓存优化**
   - 使用Redis缓存热点数据
   - 缓存用户基本信息
   - 实现查询结果缓存

3. **并发优化**
   - 使用goroutine处理并发请求
   - 实现请求限流和熔断
   - 优化数据库连接池配置

## 监控和日志

### 健康检查
```bash
curl http://localhost:8080/health
```

### 日志配置
```go
// 配置结构化日志
log.SetFormatter(&log.JSONFormatter{})
log.SetLevel(log.InfoLevel)
```

### 监控指标
- 请求响应时间
- 错误率
- 并发用户数
- 数据库连接数
- 用户注册成功率

## 安全考虑

1. **认证安全**
   - 使用JWT进行身份认证
   - 密码加密存储
   - 输入验证和过滤

2. **API安全**
   - 输入验证和过滤
   - SQL注入防护
   - XSS攻击防护

3. **数据安全**
   - 敏感信息脱敏
   - 数据访问权限控制
   - 定期数据备份

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接参数
   - 检查网络连通性

2. **用户注册失败**
   - 检查用户名和邮箱唯一性
   - 验证密码强度
   - 检查数据库权限

3. **认证验证失败**
   - 检查JWT token有效性
   - 验证用户状态
   - 检查中间件配置

### 日志分析
```bash
# 查看错误日志
grep "ERROR" logs/app.log

# 查看访问日志
tail -f logs/access.log
``` 