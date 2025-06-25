# 用户管理服务 API 文档

## 概述

用户管理服务负责用户注册、登录、信息管理等功能。支持学生、教师、管理员三种用户类型。

## 基础信息

- **服务地址**: `http://localhost:8084`
- **API 前缀**: `/api/users`
- **认证方式**: JWT Token

## 权限说明

### 用户角色
- **admin**: 管理员，可以管理所有用户，创建学生和教师
- **teacher**: 教师，可以查看学生信息
- **student**: 学生，只能查看自己的信息

### 数据访问权限
- **管理员**: 可以查看所有用户，创建、更新、删除用户
- **教师**: 可以查看学生信息
- **学生**: 只能查看自己的信息

## API 接口

### 1. 用户注册（仅限学生）

**POST** `/api/users/register`

注册新学生用户。注意：只能注册学生用户，教师用户只能由管理员创建。

#### 请求参数

```json
{
  "username": "zhangsan",
  "password": "password123",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "real_name": "张三",
  "user_type": "student",
  "student_id": "2024001"
}
```

**参数说明**:
- `username`: 用户名（必填，唯一）
- `password`: 密码（必填）
- `email`: 邮箱（必填，唯一）
- `phone`: 手机号（可选）
- `real_name`: 真实姓名（必填）
- `user_type`: 用户类型（必填，只能为 "student"）
- `student_id`: 学号（可选，不传则使用系统生成的UUID）

#### 响应示例

**成功响应 (201 Created)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "学生注册成功",
    "user": {
      "user_id": "user123",
      "username": "zhangsan",
      "email": "zhangsan@example.com",
      "phone": "13800138000",
      "real_name": "张三",
      "user_type": "student",
      
      "status": "active",
      "register_time": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

**错误响应 (400 Bad Request)**
```json
{
  "code": 400,
  "message": "只能注册学生用户",
  "data": null
}
```

**错误响应 (409 Conflict)**
```json
{
  "code": 409,
  "message": "用户名已存在",
  "data": null
}
```

### 2. 管理员创建学生

**POST** `/api/users/students`

管理员创建学生账号。需要管理员权限。

#### 请求头

```
Authorization: Bearer <admin_token>
```

#### 请求参数

```json
{
  "username": "student001",
  "password": "password123",
  "email": "student001@example.com",
  "phone": "13800138001",
  "real_name": "李同学",
  "user_type": "student",
  "student_id": "2024001"
}
```

**参数说明**:
- `username`: 用户名（必填，唯一）
- `password`: 密码（必填）
- `email`: 邮箱（必填，唯一）
- `phone`: 手机号（可选）
- `real_name`: 真实姓名（必填）
- `user_type`: 用户类型（必填，只能为 "student"）
- `student_id`: 学号（可选，不传则使用系统生成的UUID）

#### 响应示例

**成功响应 (201 Created)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "学生创建成功",
    "user": {
      "user_id": "user456",
      "username": "student001",
      "email": "student001@example.com",
      "phone": "13800138001",
      "real_name": "李同学",
      "user_type": "student",
      
      "status": "active",
      "register_time": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

**错误响应 (401 Unauthorized)**
```json
{
  "code": 401,
  "message": "未认证，无法操作",
  "data": null
}
```

**错误响应 (403 Forbidden)**
```json
{
  "code": 403,
  "message": "只有管理员可以创建学生",
  "data": null
}
```

### 3. 管理员创建教师

**POST** `/api/users/teachers`

管理员创建教师账号。需要管理员权限。

#### 请求头

```
Authorization: Bearer <admin_token>
```

#### 请求参数

```json
{
  "username": "teacher001",
  "password": "password123",
  "email": "teacher001@example.com",
  "phone": "13800138001",
  "real_name": "李老师"
}
```

**参数说明**:
- `username`: 用户名（必填，唯一）
- `password`: 密码（必填）
- `email`: 邮箱（必填，唯一）
- `phone`: 手机号（可选）
- `real_name`: 真实姓名（必填）
- `user_type`: 用户类型（必填，只能为 "teacher"）

#### 响应示例

**成功响应 (201 Created)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "教师创建成功",
    "user": {
      "user_id": "user789",
      "username": "teacher001",
      "email": "teacher001@example.com",
      "phone": "13800138001",
      "real_name": "李老师",
      "user_type": "teacher",
      
      "status": "active",
      "register_time": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

**错误响应 (401 Unauthorized)**
```json
{
  "code": 401,
  "message": "未认证，无法操作",
  "data": null
}
```

**错误响应 (403 Forbidden)**
```json
{
  "code": 403,
  "message": "只有管理员可以创建教师",
  "data": null
}
```

### 4. 获取用户统计信息

**GET** `/api/users/stats`

获取用户统计信息，包括各类型用户数量。

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_users": 100,
    "student_users": 80,
    "teacher_users": 15,
    "admin_users": 5,
    "active_users": 95,
    "inactive_users": 5
  }
}
```

### 5. 获取当前用户信息

**GET** `/api/users/profile`

获取当前登录用户的详细信息。

#### 请求头

```
Authorization: Bearer <token>
```

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "user123",
    "username": "zhangsan",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "real_name": "张三",
    "user_type": "student",
    
    "status": "active",
    "avatar": "",
    "last_login_at": null,
    "register_time": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "student_id": "2024001",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软工2024-1班"
  }
}
```

### 6. 更新当前用户信息

**PUT** `/api/users/profile`

更新当前登录用户的信息。

#### 请求头

```
Authorization: Bearer <token>
```

#### 请求参数

```json
{
  "phone": "13800138001",
  "real_name": "张三丰",
  "avatar": "avatar.jpg"
}
```

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "用户信息更新成功"
  }
}
```

### 7. 获取所有用户（管理员）

**GET** `/api/users`

管理员获取所有用户列表。

#### 请求头

```
Authorization: Bearer <admin_token>
```

#### 查询参数

- `page`: 页码（可选，默认1）
- `size`: 每页数量（可选，默认10）
- `user_type`: 用户类型过滤（可选）
- `status`: 状态过滤（可选）

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [
      {
        "id": "user123",
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "real_name": "张三",
        "user_type": "student",
        
        "status": "active",
        "register_time": "2024-01-01T00:00:00Z",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "size": 10
  }
}
```

### 8. 根据用户类型获取用户（管理员）

**GET** `/api/users/type/:userType`

管理员根据用户类型获取用户列表。

#### 请求头

```
Authorization: Bearer <admin_token>
```

#### 路径参数

- `userType`: 用户类型（student/teacher/admin）

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [
      {
        "id": "user123",
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "real_name": "张三",
        "user_type": "student",
        
        "status": "active",
        "register_time": "2024-01-01T00:00:00Z",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 80
  }
}
```

### 9. 获取指定用户信息（管理员）

**GET** `/api/users/:id`

管理员获取指定用户的详细信息。

#### 请求头

```
Authorization: Bearer <admin_token>
```

#### 路径参数

- `id`: 用户ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "user123",
    "username": "zhangsan",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "real_name": "张三",
    "user_type": "student",
    
    "status": "active",
    "avatar": "",
    "last_login_at": null,
    "register_time": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 10. 更新指定用户信息（管理员）

**PUT** `/api/users/:id`

管理员更新指定用户的信息。

#### 请求头

```
Authorization: Bearer <admin_token>
```

#### 路径参数

- `id`: 用户ID

#### 请求参数

```json
{
  "phone": "13800138001",
  "real_name": "张三丰",
  "status": "inactive"
}
```

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "用户信息更新成功"
  }
}
```

### 11. 删除用户（管理员）

**DELETE** `/api/users/:id`

管理员删除指定用户。此操作会同时删除用户账户和对应的学生/教师档案。

#### 权限要求

- 只有管理员可以执行此操作
- 不能删除自己的账户
- 不能删除最后一个系统管理员

#### 请求头

```
Authorization: Bearer <admin_token>
```

#### 路径参数

- `id`: 用户ID

#### 业务逻辑

1. **权限验证**: 验证当前用户是否为管理员
2. **自删除检查**: 防止管理员删除自己的账户
3. **管理员保护**: 防止删除最后一个系统管理员
4. **关联数据清理**: 
   - 如果删除学生用户，会同时删除学生信息服务中的学生档案
   - 如果删除教师用户，会同时删除教师信息服务中的教师档案
5. **软删除**: 用户记录采用软删除，设置删除时间但不真正删除

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "用户 zhangsan 删除成功",
    "deleted_user": {
      "user_id": "user123",
      "username": "zhangsan",
      "user_type": "student"
    }
  }
}
```

**错误响应 (400 Bad Request)**
```json
{
  "code": 400,
  "message": "不能删除自己的账户",
  "data": null
}
```

**错误响应 (400 Bad Request)**
```json
{
  "code": 400,
  "message": "不能删除最后一个系统管理员",
  "data": null
}
```

**错误响应 (403 Forbidden)**
```json
{
  "code": 403,
  "message": "只有管理员可以删除用户",
  "data": null
}
```

**错误响应 (404 Not Found)**
```json
{
  "code": 404,
  "message": "用户不存在",
  "data": null
}
```

**错误响应 (500 Internal Server Error)**
```json
{
  "code": 500,
  "message": "删除学生档案失败: connection refused",
  "data": null
}
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如用户名已存在） |
| 500 | 服务器内部错误 |

## 权限说明

- **普通用户**: 只能注册学生账号，查看和更新自己的信息
- **管理员**: 可以创建学生和教师账号，管理所有用户信息
- **教师**: 可以查看和更新自己的信息

## 注意事项

1. 用户注册时只能注册学生类型，教师账号只能由管理员创建
2. 管理员创建用户时，系统会自动同步创建对应的学生或教师档案
3. 学号（student_id）支持自定义，如果不提供则使用系统生成的UUID
4. 所有需要认证的接口都需要在请求头中携带有效的JWT Token
5. 用户删除采用软删除，不会真正删除数据库记录 
