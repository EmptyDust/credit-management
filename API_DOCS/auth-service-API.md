# 认证服务 API 文档

## 概述
认证服务负责处理用户认证、授权和权限验证，提供JWT token生成、验证和权限管理功能。

## 权限说明

### 用户角色
- **admin**: 管理员，拥有所有权限
- **teacher**: 教师，可以审核事务和申请
- **student**: 学生，可以参与事务和申请

### 权限验证
- 所有API都需要有效的JWT token
- 权限验证基于用户角色和资源访问权限
- 支持细粒度的权限控制

## 认证
除了登录和注册接口，所有API都需要在请求头中包含有效的JWT token：
```
Authorization: Bearer <token>
```

## API 端点

### 1. 用户登录
**POST** `/api/auth/login`

**权限要求**: 无需认证（公开接口）

**请求体**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "user-uuid",
      "username": "admin",
      "email": "admin@example.com",
      "name": "管理员",
      "role": "admin"
    },
    "expires_in": 3600
  }
}
```

### 2. 用户注册
**POST** `/api/auth/register`

**权限要求**: 无需认证（公开接口）

**请求体**:
```json
{
  "username": "student001",
  "password": "password123",
  "email": "student001@example.com",
  "name": "张三",
  "role": "student"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "user-uuid",
    "username": "student001",
    "email": "student001@example.com",
    "name": "张三",
    "role": "student",
    "created_at": "2024-01-01T09:00:00Z"
  }
}
```

**说明**: 只允许注册学生角色

### 3. 验证Token
**POST** `/api/auth/verify`

**权限要求**: 需要有效token

**请求体**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "user": {
      "id": "user-uuid",
      "username": "admin",
      "email": "admin@example.com",
      "name": "管理员",
      "role": "admin"
    },
    "expires_at": "2024-01-01T10:00:00Z"
  }
}
```

### 4. 刷新Token
**POST** `/api/auth/refresh`

**权限要求**: 需要有效token

**请求体**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

### 5. 权限验证
**POST** `/api/auth/validate-permission`

**权限要求**: 需要有效token

**请求体**:
```json
{
  "resource": "students",
  "action": "read",
  "target_user_id": "user-uuid"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "has_permission": true,
    "user_role": "admin",
    "resource": "students",
    "action": "read",
    "message": "权限验证通过"
  }
}
```

### 6. 获取用户权限
**GET** `/api/auth/permissions`

**权限要求**: 需要有效token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "user-uuid",
    
    "permissions": [
      {
        "resource": "users",
        "actions": ["create", "read", "update", "delete"]
      },
      {
        "resource": "students",
        "actions": ["create", "read", "update", "delete"]
      },
      {
        "resource": "teachers",
        "actions": ["create", "read", "update", "delete"]
      },
      {
        "resource": "affairs",
        "actions": ["create", "read", "update", "delete", "review"]
      },
      {
        "resource": "applications",
        "actions": ["create", "read", "update", "delete", "review"]
      }
    ]
  }
}
```

### 7. 登出
**POST** `/api/auth/logout`

**权限要求**: 需要有效token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "登出成功"
  }
}
```

### 8. 重置密码
**POST** `/api/auth/reset-password`

**权限要求**: 需要有效token

**请求体**:
```json
{
  "old_password": "oldpassword123",
  "new_password": "newpassword123"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "密码重置成功"
  }
}
```

### 9. 忘记密码
**POST** `/api/auth/forgot-password`

**权限要求**: 无需认证（公开接口）

**请求体**:
```json
{
  "email": "user@example.com"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "重置密码邮件已发送"
  }
}
```

### 10. 获取当前用户信息
**GET** `/api/auth/me`

**权限要求**: 需要有效token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "user-uuid",
    "username": "admin",
    "email": "admin@example.com",
    "name": "管理员",
    
    "created_at": "2024-01-01T09:00:00Z",
    "last_login": "2024-01-01T10:00:00Z"
  }
}
```

## 权限模型

### 资源权限
- **users**: 用户管理
- **students**: 学生信息
- **teachers**: 教师信息
- **affairs**: 事务管理
- **applications**: 申请管理

### 操作权限
- **create**: 创建
- **read**: 读取
- **update**: 更新
- **delete**: 删除
- **review**: 审核

### 角色权限矩阵

| 角色 | users | students | teachers | affairs | applications |
|------|-------|----------|----------|---------|--------------|
| admin | CRUD | CRUD | CRUD | CRUD+Review | CRUD+Review |
| teacher | R | R | R | CRUD+Review | CRUD+Review |
| student | R(own) | R(own) | R | CRUD | CRUD |

## JWT Token 结构

### Header
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### Payload
```json
{
  "user_id": "user-uuid",
  "username": "admin",
  
  "exp": 1704096000,
  "iat": 1704092400
}
```

### 签名
使用密钥对header和payload进行签名

## 错误码
- `0`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 冲突（如用户名已存在）
- `500`: 服务器内部错误

## 使用说明
1. **登录认证**: 使用用户名和密码登录获取token
2. **Token验证**: 所有API调用都需要在header中携带token
3. **权限验证**: 通过权限验证接口检查用户权限
4. **角色控制**: 基于用户角色的权限控制
5. **Token刷新**: 支持token刷新机制
6. **密码管理**: 支持密码重置和忘记密码
7. **用户信息**: 提供当前用户信息查询
8. **登出功能**: 支持用户登出
9. **权限查询**: 可以查询用户的所有权限
10. **安全机制**: 使用JWT确保安全性

## 安全注意事项
1. Token有过期时间，需要定期刷新
2. 密码使用bcrypt加密存储
3. 敏感操作需要验证权限
4. 支持token黑名单机制
5. 记录用户登录日志
6. 防止暴力破解攻击
7. 支持多设备登录控制
8. 提供会话管理功能

## 集成说明
1. **前端集成**: 在请求头中添加Authorization
2. **服务间调用**: 传递用户token进行权限验证
3. **权限中间件**: 使用权限验证接口检查权限
4. **错误处理**: 统一处理认证和授权错误
5. **日志记录**: 记录认证和授权相关日志 
5. 用户登出时调用 `/api/auth/logout` 接口 
