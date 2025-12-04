# API 网关 (API Gateway)

这是学分管理系统的 API 网关服务，作为系统的统一入口，负责路由转发、认证验证和权限控制。

## 功能特性

- **统一入口** - 所有 API 请求的统一入口点
- **路由转发** - 根据请求路径转发到相应的微服务
- **JWT 验证** - 在网关层验证 JWT token
- **权限控制** - 基础的权限中间件检查
- **CORS 支持** - 跨域资源共享支持
- **服务发现** - 动态配置各微服务地址

## 快速开始

### 环境要求

- Go 1.21+

### 本地运行

```bash
# 安装依赖
go mod download

# 设置环境变量
export AUTH_SERVICE_URL=http://localhost:8081
export USER_SERVICE_URL=http://localhost:8084
export CREDIT_ACTIVITY_SERVICE_URL=http://localhost:8083
export JWT_SECRET=your-secret-key
export PORT=8080

# 运行服务
go run main.go
```

### Docker 运行

```bash
# 构建镜像
docker build -t api-gateway .

# 运行容器
docker run -d \
  --name api-gateway \
  -p 8080:8080 \
  -e AUTH_SERVICE_URL=http://auth-service:8081 \
  -e USER_SERVICE_URL=http://user-service:8084 \
  -e CREDIT_ACTIVITY_SERVICE_URL=http://credit-activity-service:8083 \
  -e JWT_SECRET=your-jwt-secret \
  api-gateway
```

## 架构设计

```
客户端请求
    ↓
API Gateway (认证和权限检查)
    ↓
路由转发
    ├─→ Auth Service (认证服务)
    ├─→ User Service (用户服务)
    └─→ Credit Activity Service (学分活动服务)
```

## 路由规则

### 认证服务路由

```http
POST /api/auth/login                    # 登录（无需认证）
POST /api/auth/validate-token           # 验证 token（无需认证）
POST /api/auth/refresh-token            # 刷新 token（无需认证）
POST /api/auth/logout                   # 登出（需要认证）
```

### 用户服务路由

```http
GET    /api/users/*                     # 用户管理相关
POST   /api/users/*                     # 用户管理相关
PUT    /api/users/*                     # 用户管理相关
DELETE /api/users/*                     # 用户管理相关
GET    /api/students/*                  # 学生管理相关
GET    /api/teachers/*                  # 教师管理相关
GET    /api/search/*                    # 搜索功能
```

### 学分活动服务路由

```http
GET    /api/activities/*                # 活动管理相关
POST   /api/activities/*                # 活动管理相关
PUT    /api/activities/*                # 活动管理相关
DELETE /api/activities/*                # 活动管理相关
GET    /api/applications/*              # 申请管理相关
GET    /api/search/*                    # 搜索功能
```

### 权限管理路由（预留）

```http
GET    /api/permissions/*               # 权限管理相关
POST   /api/permissions/*               # 权限管理相关
```

## 环境变量

| 变量名                        | 说明                 | 默认值                                |
| ----------------------------- | -------------------- | ------------------------------------- |
| `AUTH_SERVICE_URL`            | 认证服务地址         | `http://auth-service:8081`            |
| `USER_SERVICE_URL`            | 用户服务地址         | `http://user-service:8084`            |
| `CREDIT_ACTIVITY_SERVICE_URL` | 学分活动服务地址     | `http://credit-activity-service:8083` |
| `JWT_SECRET`                  | JWT 密钥             | `your-secret-key`                     |
| `PORT`                        | 网关端口             | `8080`                                |
| `TEST_DATA_MODE`              | 测试数据模式（可选） | `disabled`                            |

## 权限控制

网关层进行基础的权限检查，细粒度权限控制在各微服务内部实现：

### 权限中间件

- `AllUsers()` - 所有认证用户可访问
- `StudentOnly()` - 仅学生可访问
- `TeacherOrAdmin()` - 教师或管理员可访问
- `AdminOnly()` - 仅管理员可访问

### 认证流程

1. 客户端请求携带 JWT token（Authorization header 或 X-User-ID header）
2. 网关验证 token 并提取用户信息
3. 根据路由配置的权限要求进行权限检查
4. 通过后转发请求到对应微服务
5. 微服务接收请求时已包含用户信息（通过 header 传递）

## 用户信息传递

网关通过以下 HTTP headers 向微服务传递用户信息：

- `X-User-ID`: 用户 UUID
- `X-Username`: 用户名
- `X-User-Type`: 用户类型（student/teacher/admin）

微服务通过中间件从这些 headers 中提取用户信息。

## 健康检查

```http
GET /health
```

返回示例：

```json
{
  "status": "ok",
  "service": "api-gateway",
  "version": "3.0.0"
}
```

## API 信息

访问根路径可获取 API 信息：

```http
GET /
```

返回示例：

```json
{
  "message": "Credit Management API Gateway",
  "version": "3.0.0",
  "services": {
    "auth_service": "http://auth-service:8081",
    "user_service": "http://user-service:8084",
    "credit_activity_service": "http://credit-activity-service:8083"
  },
  "endpoints": {
    "auth": "/api/auth",
    "permissions": "/api/permissions",
    "users": "/api/users",
    "students": "/api/students",
    "teachers": "/api/teachers",
    "search": "/api/search",
    "activities": "/api/activities",
    "health": "/health"
  }
}
```

## 错误处理

网关统一处理以下错误：

- `401 Unauthorized` - 未认证或 token 无效
- `403 Forbidden` - 权限不足
- `404 Not Found` - 路由不存在
- `500 Internal Server Error` - 网关内部错误
- `502 Bad Gateway` - 后端服务不可用

## 故障排除

### 常见问题

1. **服务无法连接**

   - 检查各微服务是否正常运行
   - 确认服务地址配置正确
   - 检查网络连接和防火墙设置

2. **认证失败**

   - 确认 JWT_SECRET 与认证服务一致
   - 检查 token 格式和有效性
   - 查看网关日志了解详细错误

3. **路由转发失败**
   - 检查目标服务健康状态
   - 确认服务地址和端口正确
   - 查看目标服务的日志

## 性能优化

1. **连接池** - 使用 HTTP 客户端连接池复用连接
2. **超时设置** - 合理设置请求超时时间
3. **负载均衡** - 可配合负载均衡器实现高可用
4. **缓存** - 可添加响应缓存减少后端压力
