# 权限服务 API 文档

## 概述

权限服务负责处理用户权限管理，包括角色管理、权限管理、用户权限分配等操作。该服务集成在auth-service中，通过 `/api/permissions` 路径提供权限管理功能。

## 基础信息

- **服务名称**: auth-service (权限管理模块)
- **端口**: 8081
- **基础路径**: `/api/permissions`

## 统一返回格式

所有 API 接口都使用统一的返回格式：

### 成功响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体的数据内容
  }
}
```

### 错误响应格式
```json
{
  "code": 400,  // 错误码
  "message": "错误描述",
  "data": null
}
```

### 错误码说明
- `400` - 请求参数错误
- `401` - 未认证
- `403` - 权限不足
- `404` - 资源不存在
- `500` - 服务器内部错误

## 权限要求

所有权限管理接口都需要以下权限：
- **认证**: 需要有效的JWT令牌
- **权限**: 需要 `permission:manage` 权限

## API 接口

### 1. 初始化权限

**POST** `/api/permissions/init`

初始化系统默认权限和角色（无需权限验证）。

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "权限初始化成功",
    "roles_created": 3,
    "permissions_created": 20
  }
}
```

### 2. 角色管理

#### 2.1 创建角色

**POST** `/api/permissions/roles`

创建新的角色。

#### 请求参数

```json
{
  "name": "teacher",
  "description": "教师角色",
  "is_system": false
}
```

#### 响应示例

**成功响应 (201 Created)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "role123",
    "name": "teacher",
    "description": "教师角色",
    "is_system": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 2.2 获取所有角色

**GET** `/api/permissions/roles`

获取所有角色列表。

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "role123",
      "name": "admin",
      "description": "管理员角色",
      "is_system": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "role456",
      "name": "teacher",
      "description": "教师角色",
      "is_system": false,
      "created_at": "2024-01-02T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z"
    }
  ]
}
```

#### 2.3 获取指定角色

**GET** `/api/permissions/roles/{roleID}`

获取指定角色的详细信息。

#### 路径参数

- `roleID` (string, required): 角色ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "role123",
    "name": "admin",
    "description": "管理员角色",
    "is_system": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 2.4 更新角色

**PUT** `/api/permissions/roles/{roleID}`

更新指定角色的信息。

#### 路径参数

- `roleID` (string, required): 角色ID

#### 请求参数

```json
{
  "name": "updated_teacher",
  "description": "更新后的教师角色描述",
  "is_system": false
}
```

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "role456",
    "name": "updated_teacher",
    "description": "更新后的教师角色描述",
    "is_system": false,
    "created_at": "2024-01-02T00:00:00Z",
    "updated_at": "2024-01-02T12:00:00Z"
  }
}
```

#### 2.5 删除角色

**DELETE** `/api/permissions/roles/{roleID}`

删除指定角色（不能删除系统角色）。

#### 路径参数

- `roleID` (string, required): 角色ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**错误响应 (400 Bad Request)**
```json
{
  "code": 400,
  "message": "不能删除系统角色",
  "data": null
}
```

### 3. 权限管理

#### 3.1 创建权限

**POST** `/api/permissions`

创建新的权限。

#### 请求参数

```json
{
  "name": "affair:delete",
  "description": "删除事务权限",
  "resource": "affair",
  "action": "delete"
}
```

#### 响应示例

**成功响应 (201 Created)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "perm123",
    "name": "affair:delete",
    "description": "删除事务权限",
    "resource": "affair",
    "action": "delete",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 3.2 获取所有权限

**GET** `/api/permissions`

获取所有权限列表。

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "perm123",
      "name": "affair:read",
      "description": "查看事务权限",
      "resource": "affair",
      "action": "read",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "perm456",
      "name": "application:write",
      "description": "创建申请权限",
      "resource": "application",
      "action": "write",
      "created_at": "2024-01-02T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z"
    }
  ]
}
```

#### 3.3 获取指定权限

**GET** `/api/permissions/{id}`

获取指定权限的详细信息。

#### 路径参数

- `id` (string, required): 权限ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "perm123",
    "name": "affair:read",
    "description": "查看事务权限",
    "resource": "affair",
    "action": "read",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 3.4 删除权限

**DELETE** `/api/permissions/{id}`

删除指定权限。

#### 路径参数

- `id` (string, required): 权限ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 4. 用户权限分配

#### 4.1 分配角色给用户

**POST** `/api/permissions/users/{userID}/roles`

为用户分配角色。

#### 路径参数

- `userID` (string, required): 用户ID

#### 请求参数

```json
{
  "role_id": "role123"
}
```

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "角色分配成功",
    "role": {
      "id": "role123",
      "name": "teacher",
      "description": "教师角色"
    }
  }
}
```

#### 4.2 移除用户角色

**DELETE** `/api/permissions/users/{userID}/roles/{roleID}`

移除用户的指定角色。

#### 路径参数

- `userID` (string, required): 用户ID
- `roleID` (string, required): 角色ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

#### 4.3 分配权限给用户

**POST** `/api/permissions/users/{userID}/permissions`

为用户直接分配权限。

#### 路径参数

- `userID` (string, required): 用户ID

#### 请求参数

```json
{
  "permission_id": "perm123"
}
```

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "权限分配成功",
    "permission": {
      "id": "perm123",
      "name": "affair:read",
      "description": "查看事务权限"
    }
  }
}
```

#### 4.4 移除用户权限

**DELETE** `/api/permissions/users/{userID}/permissions/{permissionID}`

移除用户的指定权限。

#### 路径参数

- `userID` (string, required): 用户ID
- `permissionID` (string, required): 权限ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 5. 角色权限管理

#### 5.1 分配权限给角色

**POST** `/api/permissions/roles/{roleID}/permissions`

为角色分配权限。

#### 路径参数

- `roleID` (string, required): 角色ID

#### 请求参数

```json
{
  "permission_id": "perm123"
}
```

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "权限分配成功",
    "permission": {
      "id": "perm123",
      "name": "affair:read",
      "description": "查看事务权限"
    }
  }
}
```

#### 5.2 移除角色权限

**DELETE** `/api/permissions/roles/{roleID}/permissions/{permissionID}`

移除角色的指定权限。

#### 路径参数

- `roleID` (string, required): 角色ID
- `permissionID` (string, required): 权限ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 6. 查询接口

#### 6.1 获取用户角色

**GET** `/api/permissions/users/{userID}/roles`

获取指定用户的所有角色。

#### 路径参数

- `userID` (string, required): 用户ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "role123",
      "name": "teacher",
      "description": "教师角色",
      "is_system": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 6.2 获取用户权限

**GET** `/api/permissions/users/{userID}/permissions`

获取指定用户的所有权限（包括直接分配的权限和通过角色继承的权限）。

#### 路径参数

- `userID` (string, required): 用户ID

#### 响应示例

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "perm123",
      "name": "affair:read",
      "description": "查看事务权限",
      "resource": "affair",
      "action": "read",
      "source": "role",
      "role_name": "teacher"
    },
    {
      "id": "perm456",
      "name": "application:write",
      "description": "创建申请权限",
      "resource": "application",
      "action": "write",
      "source": "direct"
    }
  ]
}
```

## 数据模型

### Role 模型

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "is_system": "boolean",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Permission 模型

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "resource": "string",
  "action": "string",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### RoleRequest 模型

```json
{
  "name": "string",
  "description": "string",
  "is_system": "boolean"
}
```

### RoleUpdateRequest 模型

```json
{
  "name": "string",
  "description": "string",
  "is_system": "boolean"
}
```

### PermissionRequest 模型

```json
{
  "name": "string",
  "description": "string",
  "resource": "string",
  "action": "string"
}
```

## 权限资源

- `affair` - 事务管理
- `application` - 申请管理
- `student` - 学生信息
- `teacher` - 教师信息
- `user` - 用户管理
- `permission` - 权限管理
- `system` - 系统管理

## 权限操作

- `read` - 读取
- `write` - 写入
- `delete` - 删除
- `manage` - 管理

## 系统角色

- `admin` - 管理员（系统角色）
- `teacher` - 教师
- `student` - 学生
- `user` - 普通用户

## 错误处理

所有错误都遵循统一的返回格式，包含错误码和错误描述。常见错误包括：

- 参数验证失败
- 权限不足
- 资源不存在
- 系统角色不能删除
- 服务器内部错误

## 权限继承机制

1. **直接权限**: 直接分配给用户的权限
2. **角色权限**: 通过角色继承的权限
3. **权限合并**: 用户的最终权限是直接权限和角色权限的并集
4. **权限优先级**: 直接权限优先于角色权限

## 使用示例

### 创建教师角色并分配权限

1. 创建教师角色
```bash
POST /api/permissions/roles
{
  "name": "teacher",
  "description": "教师角色",
  "is_system": false
}
```

2. 为教师角色分配权限
```bash
POST /api/permissions/roles/{roleID}/permissions
{
  "permission_id": "affair:read"
}
```

3. 为用户分配教师角色
```bash
POST /api/permissions/users/{userID}/roles
{
  "role_id": "roleID"
}
``` 