# 统一用户服务 API 文档

## 概述

统一用户服务整合了原有的用户管理、学生信息和教师信息三个服务的功能，提供统一的用户管理 API。

**基础 URL**: `http://localhost:8084/api`

## 认证

所有需要认证的接口都需要在请求头中包含 JWT 令牌：

```
Authorization: Bearer <your-jwt-token>
```

## 响应格式

所有 API 响应都遵循统一的格式：

```json
{
  "code": 0, // 状态码，0表示成功
  "message": "success", // 响应消息
  "data": {
    // 响应数据
    // 具体数据内容
  }
}
```

错误响应格式：

```json
{
  "code": 400, // 错误码
  "message": "错误描述", // 错误消息
  "data": null
}
```

## 用户管理接口

### 1. 用户注册

**POST** `/users/register`

学生用户注册接口。

**请求体**:

```json
{
  "username": "student001",
  "password": "password123",
  "email": "student@example.com",
  "phone": "13800138000",
  "real_name": "张三",
  "user_type": "student",
  "student_id": "2023001",
  "college": "计算机学院",
  "major": "软件工程",
  "class": "软件2301",
  "grade": "2023"
}
```

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "学生注册成功",
    "user": {
      "user_id": "uuid",
      "username": "student001",
      "email": "student@example.com",
      "phone": "13800138000",
      "real_name": "张三",
      "user_type": "student",
      "status": "active",
      "student_id": "2023001",
      "college": "计算机学院",
      "major": "软件工程",
      "class": "软件2301",
      "grade": "2023",
      "register_time": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

### 2. 获取当前用户信息

**GET** `/users/profile`

获取当前登录用户的信息。

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "student001",
    "email": "student@example.com",
    "phone": "13800138000",
    "real_name": "张三",
    "user_type": "student",
    "status": "active",
    "avatar": "",
    "last_login_at": "2024-01-01T00:00:00Z",
    "register_time": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "student_id": "2023001",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023"
  }
}
```

### 3. 更新当前用户信息

**PUT** `/users/profile`

更新当前登录用户的信息。

**请求体**:

```json
{
  "email": "newemail@example.com",
  "phone": "13900139000",
  "real_name": "李四",
  "college": "信息学院",
  "major": "计算机科学",
  "class": "计科2301",
  "grade": "2023"
}
```

### 4. 获取用户统计信息

**GET** `/users/stats`

获取用户统计信息（需要认证）。

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_users": 1000,
    "active_users": 950,
    "suspended_users": 50,
    "student_users": 800,
    "teacher_users": 150,
    "admin_users": 50,
    "new_users_today": 10,
    "new_users_week": 50,
    "new_users_month": 200
  }
}
```

## 管理员接口

### 1. 创建教师

**POST** `/users/teachers`

管理员创建教师用户。

**请求体**:

```json
{
  "username": "teacher001",
  "password": "password123",
  "email": "teacher@example.com",
  "phone": "13800138001",
  "real_name": "王老师",
  "user_type": "teacher",
  "department": "计算机系",
  "title": "副教授",
  "specialty": "人工智能"
}
```

### 2. 创建学生

**POST** `/users/students`

管理员创建学生用户。

**请求体**:

```json
{
  "username": "student002",
  "password": "password123",
  "email": "student2@example.com",
  "phone": "13800138002",
  "real_name": "李同学",
  "user_type": "student",
  "student_id": "2023002",
  "college": "计算机学院",
  "major": "软件工程",
  "class": "软件2301",
  "grade": "2023"
}
```

### 3. 获取所有用户

**GET** `/users`

管理员获取所有用户列表。

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "user_id": "uuid1",
      "username": "student001",
      "email": "student@example.com",
      "real_name": "张三",
      "user_type": "student",
      "status": "active",
      "student_id": "2023001",
      "college": "计算机学院"
    },
    {
      "user_id": "uuid2",
      "username": "teacher001",
      "email": "teacher@example.com",
      "real_name": "王老师",
      "user_type": "teacher",
      "status": "active",
      "department": "计算机系",
      "title": "副教授"
    }
  ]
}
```

### 4. 根据用户类型获取用户

**GET** `/users/type/{userType}`

获取指定类型的用户列表。

**路径参数**:

- `userType`: 用户类型 (`student`, `teacher`, `admin`)

### 5. 获取指定用户信息

**GET** `/users/{id}`

获取指定用户 ID 的用户信息。

**权限说明**：

- 所有认证用户都可以访问该接口。
- 用户自己可以获取自己的详细信息。
- 管理员可以获取所有用户的详细信息。
- 教师可以获取学生的详细信息，获取其他教师时仅返回基本信息。
- 学生可以获取其他学生和教师的基本信息，不能获取详细信息。

**响应示例（学生获取其他学生）**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "student002",
    "real_name": "李同学",
    "student_id": "2023002",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023",
    "status": "active",
    "avatar": "",
    "register_time": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**响应示例（用户获取自己/管理员获取详细信息）**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "student002",
    "email": "student2@example.com",
    "phone": "13800138002",
    "real_name": "李同学",
    "user_type": "student",
    "status": "active",
    "avatar": "",
    "last_login_at": "2024-01-01T00:00:00Z",
    "register_time": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "student_id": "2023002",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023"
  }
}
```

### 6. 更新指定用户信息

**PUT** `/users/{id}`

更新指定用户的信息。

### 7. 删除用户

**DELETE** `/users/{id}`

删除指定用户（软删除）。

## 学生管理接口

### 1. 获取所有学生

**GET** `/students`

获取所有学生列表。

**查询参数**:

- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 10）
- `status`: 状态筛选
- `college`: 学院筛选
- `major`: 专业筛选
- `class`: 班级筛选
- `grade`: 年级筛选

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "user_id": "uuid",
      "username": "student001",
      "email": "student@example.com",
      "real_name": "张三",
      "user_type": "student",
      "status": "active",
      "student_id": "2023001",
      "college": "计算机学院",
      "major": "软件工程",
      "class": "软件2301",
      "grade": "2023"
    }
  ]
}
```

### 2. 获取学生统计信息

**GET** `/students/stats`

获取学生统计信息。

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_students": 800,
    "active_students": 750,
    "graduated_students": 50,
    "students_by_college": {
      "计算机学院": 300,
      "信息学院": 250,
      "机械学院": 250
    },
    "students_by_major": {
      "软件工程": 200,
      "计算机科学": 150,
      "信息安全": 100
    },
    "students_by_grade": {
      "2023": 300,
      "2022": 250,
      "2021": 250
    }
  }
}
```

## 教师管理接口

### 1. 获取所有教师

**GET** `/teachers`

获取所有教师列表。

**查询参数**:

- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 10）
- `status`: 状态筛选
- `department`: 院系筛选
- `title`: 职称筛选

### 2. 获取教师统计信息

**GET** `/teachers/stats`

获取教师统计信息。

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_teachers": 150,
    "active_teachers": 140,
    "retired_teachers": 10,
    "teachers_by_department": {
      "计算机系": 50,
      "软件工程系": 40,
      "信息安全系": 30
    },
    "teachers_by_title": {
      "教授": 20,
      "副教授": 50,
      "讲师": 80
    }
  }
}
```

## 搜索接口

### 通用用户搜索

**GET** `/search/users`

通用用户搜索接口。

**查询参数**:

- `query`: 搜索关键词
- `user_type`: 用户类型筛选
- `college`: 学院筛选（学生）
- `major`: 专业筛选（学生）
- `class`: 班级筛选（学生）
- `grade`: 年级筛选（学生）
- `department`: 部门筛选（教师）
- `title`: 职称筛选（教师）
- `status`: 状态筛选
- `page`: 页码
- `page_size`: 每页数量

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [
      {
        "user_id": "uuid",
        "username": "student001",
        "email": "student@example.com",
        "real_name": "张三",
        "user_type": "student",
        "status": "active",
        "student_id": "2023001",
        "college": "计算机学院"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10
  }
}
```

## 健康检查

### 服务健康检查

**GET** `/health`

检查服务状态。

**响应**:

```json
{
  "status": "ok",
  "service": "user-service"
}
```

## 错误码说明

| 错误码 | 说明                       |
| ------ | -------------------------- |
| 200    | 成功                       |
| 400    | 请求参数错误               |
| 401    | 未认证                     |
| 403    | 权限不足                   |
| 404    | 资源不存在                 |
| 409    | 资源冲突（如用户名已存在） |
| 500    | 服务器内部错误             |

## 权限说明

### 用户类型权限

- **student**: 学生用户

  - 可以注册（仅限学生）
  - 可以查看和更新自己的信息
  - 可以查看学生列表和统计信息

- **teacher**: 教师用户

  - 可以查看和更新自己的信息
  - 可以查看学生和教师列表
  - 可以查看统计信息

- **admin**: 管理员用户
  - 拥有所有权限
  - 可以创建、更新、删除用户
  - 可以查看所有用户信息

### 接口权限

| 接口                   | 权限要求           |
| ---------------------- | ------------------ |
| `POST /users/register` | 公开               |
| `GET /users/profile`   | 认证用户           |
| `PUT /users/profile`   | 认证用户           |
| `GET /users/stats`     | 认证用户           |
| `POST /users/teachers` | 管理员             |
| `POST /users/students` | 管理员             |
| `GET /users`           | 管理员             |
| `GET /users/:id`       | 认证用户           |
| `PUT /users/:id`       | 管理员             |
| `DELETE /users/:id`    | 管理员             |
| `GET /students`        | 学生、教师、管理员 |
| `GET /students/stats`  | 学生、教师、管理员 |
| `GET /teachers`        | 学生、教师、管理员 |
| `GET /teachers/stats`  | 学生、教师、管理员 |
| `GET /search/users`    | 认证用户           |

## 注意事项

1. 所有时间字段使用 ISO 8601 格式
2. 密码字段在响应中不会返回
3. 删除操作采用软删除方式
4. 学号具有唯一性约束
5. 邮箱和用户名具有唯一性约束
6. 分页查询默认每页 10 条记录
7. 搜索支持模糊匹配

## 批量用户管理接口

### 1. 批量删除用户

**POST** `/users/batch_delete`

批量软删除用户。

**权限**: 管理员

**请求体**:

```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "批量删除用户成功",
  "data": {
    "deleted_count": 3,
    "total_count": 3,
    "deleted_users": [
      {
        "user_id": "uuid1",
        "username": "student001",
        "real_name": "张三",
        "deleted_at": "2024-01-01T10:00:00Z"
      },
      {
        "user_id": "uuid2",
        "username": "student002",
        "real_name": "李四",
        "deleted_at": "2024-01-01T10:00:00Z"
      },
      {
        "user_id": "uuid3",
        "username": "teacher001",
        "real_name": "王老师",
        "deleted_at": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
```

**错误处理**:

```json
{
  "code": 400,
  "message": "批量删除用户失败",
  "data": {
    "deleted_count": 2,
    "total_count": 3,
    "deleted_users": [
      {
        "user_id": "uuid1",
        "status": "success"
      },
      {
        "user_id": "uuid2",
        "status": "success"
      }
    ],
    "errors": [
      {
        "user_id": "uuid3",
        "error": "用户不存在或无权限删除"
      }
    ]
  }
}
```

### 2. 批量更新用户状态

**POST** `/users/batch_status`

批量更新用户状态。

**权限**: 管理员

**请求体**:

```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"],
  "status": "active"
}
```

**状态选项**:

- `active`: 激活状态
- `inactive`: 非激活状态
- `suspended`: 暂停状态

**响应示例**:

```json
{
  "code": 0,
  "message": "批量更新用户状态成功",
  "data": {
    "updated_count": 3,
    "total_count": 3,
    "updated_users": [
      {
        "user_id": "uuid1",
        "username": "student001",
        "real_name": "张三",
        "status": "active",
        "updated_at": "2024-01-01T10:00:00Z"
      },
      {
        "user_id": "uuid2",
        "username": "student002",
        "real_name": "李四",
        "status": "active",
        "updated_at": "2024-01-01T10:00:00Z"
      },
      {
        "user_id": "uuid3",
        "username": "teacher001",
        "real_name": "王老师",
        "status": "active",
        "updated_at": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
```

### 3. 重置用户密码

**POST** `/users/reset_password`

重置指定用户的密码。

**权限**: 管理员

**请求体**:

```json
{
  "user_id": "uuid",
  "new_password": "newPassword123"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "密码重置成功",
  "data": {
    "user_id": "uuid",
    "username": "student001",
    "real_name": "张三",
    "password_reset_at": "2024-01-01T10:00:00Z"
  }
}
```

### 4. 修改自己密码

**POST** `/users/change_password`

用户修改自己的密码。

**权限**: 认证用户

**请求体**:

```json
{
  "old_password": "oldPassword123",
  "new_password": "newPassword123"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "密码修改成功",
  "data": {
    "user_id": "uuid",
    "username": "student001",
    "password_changed_at": "2024-01-01T10:00:00Z"
  }
}
```

### 5. 获取用户活动

**GET** `/users/activity`

获取当前用户的活动信息。

**权限**: 认证用户

**查询参数**:

- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 10）
- `activity_type`: 活动类型筛选（created, participated）

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "activities": [
      {
        "activity_id": "activity-uuid",
        "title": "创新创业实践活动",
        "category": "创新创业",
        "status": "approved",
        "role": "creator",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
}
```

### 6. 获取指定用户活动

**GET** `/users/{id}/activity`

获取指定用户的活动信息。

**权限**: 管理员、教师

**查询参数**:

- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 10）
- `activity_type`: 活动类型筛选（created, participated）

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "student001",
    "real_name": "张三",
    "activities": [
      {
        "activity_id": "activity-uuid",
        "title": "创新创业实践活动",
        "category": "创新创业",
        "status": "approved",
        "role": "participant",
        "credits": 2.0,
        "joined_at": "2024-01-01T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
}
```

### 7. 导出用户数据

**GET** `/users/export`

导出用户数据。

**权限**: 管理员

**查询参数**:

- `format`: 导出格式（json, csv, excel）
- `user_type`: 用户类型筛选（student, teacher, all）
- `status`: 状态筛选（active, inactive, suspended, all）
- `college`: 学院筛选（仅学生）
- `department`: 部门筛选（仅教师）

**响应示例**:

```json
{
  "code": 0,
  "message": "导出成功",
  "data": {
    "download_url": "/downloads/users_20240101_100000.csv",
    "file_name": "users_20240101_100000.csv",
    "file_size": 1024000,
    "exported_at": "2024-01-01T10:00:00Z",
    "total_records": 1000
  }
}
```

### 8. CSV 批量导入用户

**POST** `/users/import-csv`

从 CSV 文件批量导入用户。

**权限**: 管理员

**请求体**: multipart/form-data

- `file`: CSV 文件
- `user_type`: 用户类型（student, teacher）
- `update_existing`: 是否更新已存在的用户（true, false）

**CSV 格式示例**:

```csv
username,email,real_name,password,student_id,college,major,class,grade
student001,student001@example.com,张三,password123,2023001,计算机学院,软件工程,软件2301,2023
student002,student002@example.com,李四,password123,2023002,计算机学院,计算机科学,计科2301,2023
```

**响应示例**:

```json
{
  "code": 0,
  "message": "CSV导入成功",
  "data": {
    "imported_count": 100,
    "updated_count": 10,
    "failed_count": 2,
    "total_count": 112,
    "imported_users": [
      {
        "username": "student001",
        "real_name": "张三",
        "status": "imported"
      }
    ],
    "errors": [
      {
        "row": 5,
        "error": "用户名已存在"
      }
    ]
  }
}
```

### 9. 获取 CSV 模板

**GET** `/users/csv-template`

获取 CSV 导入模板。

**权限**: 管理员

**查询参数**:

- `user_type`: 用户类型（student, teacher）

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "template_url": "/downloads/templates/student_import_template.csv",
    "file_name": "student_import_template.csv",
    "headers": [
      "username",
      "email",
      "real_name",
      "password",
      "student_id",
      "college",
      "major",
      "class",
      "grade"
    ],
    "sample_data": [
      {
        "username": "student001",
        "email": "student001@example.com",
        "real_name": "张三",
        "password": "password123",
        "student_id": "2023001",
        "college": "计算机学院",
        "major": "软件工程",
        "class": "软件2301",
        "grade": "2023"
      }
    ]
  }
}
```

## 批量操作说明

### 批量操作特性

1. **事务安全**: 所有批量操作都使用数据库事务确保数据一致性
2. **错误处理**: 支持部分成功，返回详细的成功和失败信息
3. **权限验证**: 每个操作都会验证用户权限
4. **软删除**: 删除操作采用软删除方式，数据不会真正删除
5. **日志记录**: 所有批量操作都会记录详细的操作日志

### 使用建议

1. **批量大小**: 建议单次批量操作不超过 100 个用户
2. **权限检查**: 确保有足够的权限执行批量操作
3. **数据备份**: 执行重要批量操作前建议备份数据
4. **错误处理**: 仔细检查返回的错误信息，及时处理失败的操作
5. **性能考虑**: 大量数据操作时注意系统性能影响
