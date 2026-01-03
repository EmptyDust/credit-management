# 学分管理系统 API 文档

## 📋 目录

- [概述](#概述)
- [认证说明](#认证说明)
- [通用响应格式](#通用响应格式)
- [认证服务 API](#认证服务-api)
- [用户服务 API](#用户服务-api)
  - [公共接口](#公共接口无需认证)
  - [个人信息管理](#个人信息管理需认证)
  - [管理员专用接口](#管理员专用接口)
  - [搜索接口](#搜索接口)
- [学分活动服务 API](#学分活动服务-api)
  - [活动管理](#活动管理)
  - [活动审核](#活动审核)
  - [参与者管理](#参与者管理)
  - [附件管理](#附件管理)
  - [申请管理](#申请管理)
  - [搜索功能](#搜索功能)
- [权限说明](#权限说明)
- [错误码说明](#错误码说明)

---

## 概述

**基础URL：** `http://localhost:8080` (API网关)

**服务端口：**
- API网关: 8080
- 认证服务: 8081
- 用户服务: 8084
- 学分活动服务: 8083

**数据格式：** JSON

**字符编码：** UTF-8

---

## 认证说明

### JWT Token 认证

大部分API需要在请求头中携带JWT Token：

```http
Authorization: Bearer <your_jwt_token>
```

### 获取Token

通过登录接口获取Token：

```bash
POST /api/auth/login
```

Token会在响应的 `data.token` 字段中返回。

### Token存储建议

**前端存储方式：**
- **localStorage**: 持久化存储，刷新页面不丢失
- **sessionStorage**: 会话存储，关闭标签页后清除
- **内存**: 最安全但刷新页面会丢失

**示例代码：**
```javascript
// 登录成功后存储token
localStorage.setItem('token', response.data.token);
localStorage.setItem('refresh_token', response.data.refresh_token);

// 后续请求携带token
const token = localStorage.getItem('token');
fetch('/api/activities', {
    headers: {
        'Authorization': `Bearer ${token}`
    }
});
```

---

## 通用响应格式

所有API响应遵循统一格式：

### 成功响应

```json
{
    "code": 0,
    "message": "success",
    "data": {
        // 具体数据
    }
}
```

### 错误响应

```json
{
    "code": 400,
    "message": "错误描述",
    "data": null
}
```

---

## 认证服务 API

### 1. 用户登录

**接口地址：** `POST /api/auth/login`

**请求头：**
```http
Content-Type: application/json
```

**请求参数：**

支持三种登录方式（三选一）：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 否 | 用户名 |
| student_id | string | 否 | 学号 |
| teacher_id | string | 否 | 工号 |
| password | string | 是 | 密码 |

**请求示例：**

```bash
# 方式1：使用用户名登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "student",
    "password": "adminpassword"
  }'

# 方式2：使用学号登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "20240000",
    "password": "adminpassword"
  }'

# 方式3：使用工号登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "teacher_id": "T0000001",
    "password": "adminpassword"
  }'
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
            "uuid": "33333333-3333-3333-3333-333333333333",
            "username": "student",
            "email": "student@example.com",
            "real_name": "Default Student",
            "user_type": "student",
            "status": "active"
        },
        "message": "登录成功"
    }
}
```

**错误响应：**

```json
// 用户名或密码错误
{
    "code": 401,
    "message": "用户名或密码错误",
    "data": null
}

// 账户未激活
{
    "code": 403,
    "message": "账户未激活",
    "data": null
}
```

### 2. Token 验证

**接口地址：** `POST /api/auth/validate`

**请求头：**
```http
Content-Type: application/json
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | JWT Token |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/auth/validate \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "valid": true,
        "user_id": "33333333-3333-3333-3333-333333333333",
        "username": "student",
        "user_type": "student"
    }
}
```

**Token无效响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "valid": false,
        "message": "token无效或已过期"
    }
}
```

### 3. 用户登出

**接口地址：** `POST /api/auth/logout`

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**成功响应：**

```json
{
    "code": 0,
    "message": "登出成功",
    "data": null
}
```

### 4. 刷新Token

**接口地址：** `POST /api/auth/refresh`

**请求头：**
```http
Content-Type: application/json
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| refresh_token | string | 是 | Refresh Token |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
}
```

---

## 用户服务 API

### 公共接口（无需认证）

#### 1. 获取配置选项

**接口地址：** `GET /api/config/options`

**说明：** 获取系统配置选项（学校、学部、专业、班级等）

**请求示例：**
```bash
curl -X GET http://localhost:8080/api/config/options
```

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "schools": ["大连理工大学"],
        "departments": ["计算机科学与技术学部", "软件学院"],
        "majors": ["软件工程", "计算机科学与技术"],
        "classes": ["2024221", "2024222"]
    }
}
```

#### 2. 获取用户头像

**接口地址：** `GET /api/uploads/avatars/:filename`

**说明：** 获取用户头像文件（静态文件服务）

**请求示例：**
```bash
curl -X GET http://localhost:8080/api/uploads/avatars/avatar_123.jpg
```

#### 3. 学生自助注册

**接口地址：** `POST /api/students/register`

**说明：** 学生自助注册账号

**请求头：**
```http
Content-Type: application/json
```

**请求参数：**
```json
{
    "student_id": "20240001",
    "username": "student_zhang",
    "password": "password123",
    "email": "student@university.edu.cn",
    "phone": "13800138000",
    "real_name": "张三",
    "college": "计算机科学与技术学部",
    "major": "软件工程",
    "class": "2024221",
    "grade": "2024"
}
```

**成功响应：**
```json
{
    "code": 0,
    "message": "注册成功",
    "data": {
        "uuid": "...",
        "username": "student_zhang",
        "student_id": "20240001",
        "status": "active"
    }
}
```

---

### 个人信息管理（需认证）

#### 4. 获取当前用户信息

**接口地址：** `GET /api/users/me`

**请求头：**
```http
Authorization: Bearer <token>
```

**请求示例：**

```bash
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "uuid": "33333333-3333-3333-3333-333333333333",
        "username": "student",
        "email": "student@example.com",
        "phone": "13800000002",
        "real_name": "Default Student",
        "user_type": "student",
        "status": "active",
        "student_id": "20240000",
        "college": "计算机科学与技术学部",
        "major": "软件工程",
        "class": "2024222",
        "grade": "2024",
        "avatar": null,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

#### 5. 获取用户统计信息

**接口地址：** `GET /api/users/stats`

**说明：** 获取当前用户的活动统计、学分统计等

**请求头：**
```http
Authorization: Bearer <token>
```

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total_activities": 15,
        "total_credits": 12.5,
        "pending_applications": 3,
        "approved_activities": 10
    }
}
```

#### 6. 获取指定用户信息

**接口地址：** `GET /api/users/:id`

**权限：** 所有认证用户

**请求头：**
```http
Authorization: Bearer <token>
```

#### 7. 更新用户信息

**接口地址：** `PUT /api/users/me`

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 否 | 邮箱 |
| phone | string | 否 | 手机号 |
| real_name | string | 否 | 真实姓名 |

**请求示例：**

```bash
curl -X PUT http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com",
    "phone": "13900000000"
  }'
```

**成功响应：**

```json
{
    "code": 0,
    "message": "更新成功",
    "data": {
        "uuid": "33333333-3333-3333-3333-333333333333",
        "email": "newemail@example.com",
        "phone": "13900000000",
        ...
    }
}
```

### 3. 修改密码

**接口地址：** `POST /api/users/change-password`

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 旧密码 |
| new_password | string | 是 | 新密码 |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/users/change-password \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "oldpassword",
    "new_password": "newpassword"
  }'
```

**成功响应：**

```json
{
    "code": 0,
    "message": "密码修改成功",
    "data": null
}
```

#### 8. 上传头像

**接口地址：** `POST /api/users/avatar`

**权限：** 所有认证用户

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**请求参数：**
- `avatar`: 图片文件（支持jpg, png, gif）

**请求示例：**
```bash
curl -X POST http://localhost:8080/api/users/avatar \
  -H "Authorization: Bearer <token>" \
  -F "avatar=@/path/to/avatar.jpg"
```

**成功响应：**
```json
{
    "code": 0,
    "message": "头像上传成功",
    "data": {
        "avatar_url": "/api/uploads/avatars/avatar_123.jpg"
    }
}
```

#### 9. 删除头像

**接口地址：** `DELETE /api/users/avatar`

**权限：** 所有认证用户

**请求头：**
```http
Authorization: Bearer <token>
```

**成功响应：**
```json
{
    "code": 0,
    "message": "头像删除成功",
    "data": null
}
```

#### 10. 获取用户活动记录

**接口地址：** `GET /api/users/activity`

**说明：** 获取当前用户的活动记录（操作日志）

**权限：** 所有认证用户

**请求头：**
```http
Authorization: Bearer <token>
```

**接口地址：** `GET /api/users/:id/activity`

**说明：** 获取指定用户的活动记录（管理员/教师可访问）

---

### 管理员专用接口

#### 11. 创建教师账号

**接口地址：** `POST /api/users/teachers`

**权限：** 仅管理员

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求参数：**
```json
{
    "teacher_id": "T0000001",
    "username": "teacher_li",
    "password": "password123",
    "email": "teacher@university.edu.cn",
    "phone": "13800138001",
    "real_name": "李教授",
    "department_id": "uuid",
    "title": "教授"
}
```

**成功响应：**
```json
{
    "code": 0,
    "message": "创建成功",
    "data": {
        "uuid": "...",
        "username": "teacher_li",
        "teacher_id": "T0000001",
        "user_type": "teacher"
    }
}
```

#### 12. 创建学生账号

**接口地址：** `POST /api/users/students`

**权限：** 仅管理员

**请求参数：** 同学生自助注册

#### 13. 批量删除用户

**接口地址：** `POST /api/users/batch_delete`

**权限：** 仅管理员

**请求参数：**
```json
{
    "user_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**成功响应：**
```json
{
    "code": 0,
    "message": "批量删除成功",
    "data": {
        "deleted_count": 3
    }
}
```

#### 14. 批量更新用户状态

**接口地址：** `POST /api/users/batch_status`

**权限：** 仅管理员

**请求参数：**
```json
{
    "user_ids": ["uuid1", "uuid2"],
    "status": "active"
}
```

#### 15. 重置用户密码

**接口地址：** `POST /api/users/reset_password`

**权限：** 仅管理员

**请求参数：**
```json
{
    "user_id": "uuid",
    "new_password": "newpassword123"
}
```

#### 16. 导出用户数据

**接口地址：** `GET /api/users/export`

**权限：** 仅管理员

**查询参数：**
- `format`: 导出格式（csv, excel）
- `user_type`: 用户类型（student, teacher, admin）

**请求示例：**
```bash
curl -X GET "http://localhost:8080/api/users/export?format=excel&user_type=student" \
  -H "Authorization: Bearer <token>" \
  --output users.xlsx
```

#### 17. 从CSV导入用户

**接口地址：** `POST /api/users/import-csv`

**权限：** 仅管理员

**请求头：**
```http
Content-Type: multipart/form-data
```

**请求参数：**
- `file`: CSV文件

#### 18. 获取CSV模板

**接口地址：** `GET /api/users/csv-template`

**权限：** 仅管理员

#### 19. 通用导入接口

**接口地址：** `POST /api/users/import`

**权限：** 仅管理员

**说明：** 支持Excel和CSV格式导入

#### 20. 获取Excel模板

**接口地址：** `GET /api/users/excel-template`

**权限：** 仅管理员

#### 21. 获取学生统计信息

**接口地址：** `GET /api/users/stats/students`

**权限：** 学生、教师或管理员

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total_students": 1500,
        "active_students": 1450,
        "by_grade": {
            "2024": 500,
            "2023": 480
        }
    }
}
```

#### 22. 获取教师统计信息

**接口地址：** `GET /api/users/stats/teachers`

**权限：** 学生、教师或管理员

---

### 搜索接口

#### 23. 搜索用户

**接口地址：** `GET /api/search/users`

**权限：** 所有认证用户

**查询参数：**
- `keyword`: 搜索关键词（用户名、姓名、学号、工号）
- `user_type`: 用户类型
- `status`: 状态
- `page`: 页码
- `page_size`: 每页数量

**请求示例：**
```bash
curl -X GET "http://localhost:8080/api/search/users?keyword=张三&user_type=student" \
  -H "Authorization: Bearer <token>"
```

---

### 4. 获取用户列表（管理员）

**接口地址：** `GET /api/users`

**请求头：**
```http
Authorization: Bearer <token>
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |
| user_type | string | 否 | 用户类型：student/teacher/admin |
| status | string | 否 | 状态：active/inactive |
| keyword | string | 否 | 搜索关键词（用户名、姓名） |

**请求示例：**

```bash
curl -X GET "http://localhost:8080/api/users?page=1&page_size=20&user_type=student" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "page": 1,
        "page_size": 20,
        "users": [
            {
                "uuid": "...",
                "username": "student1",
                "real_name": "张三",
                "user_type": "student",
                "status": "active",
                ...
            }
        ]
    }
}
```

---

## 学分活动服务 API

### 活动管理

#### 1. 获取活动列表

**接口地址：** `GET /api/activities`

**请求头：**
```http
Authorization: Bearer <token>
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |
| category | string | 否 | 活动类别 |
| status | string | 否 | 状态：draft/pending_review/approved/rejected |
| keyword | string | 否 | 搜索关键词（标题） |

**请求示例：**

```bash
curl -X GET "http://localhost:8080/api/activities?page=1&page_size=20&status=approved" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 50,
        "page": 1,
        "page_size": 20,
        "activities": [
            {
                "id": "44444444-4444-4444-4444-444444444444",
                "title": "互联网+创新创业大赛",
                "description": "参加全国互联网+创新创业大赛",
                "category": "创新创业实践活动",
                "status": "approved",
                "start_date": "2024-03-01",
                "end_date": "2024-06-30",
                "details": {
                    "item": "互联网+创新创业大赛",
                    "company": "教育部",
                    "total_hours": 120.00
                },
                "owner": {
                    "uuid": "...",
                    "username": "student1",
                    "real_name": "张三"
                },
                "created_at": "2024-01-01T00:00:00Z"
            }
        ]
    }
}
```

### 2. 获取活动详情

**接口地址：** `GET /api/activities/:id`

**请求头：**
```http
Authorization: Bearer <token>
```

**请求示例：**

```bash
curl -X GET http://localhost:8080/api/activities/44444444-4444-4444-4444-444444444444 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "44444444-4444-4444-4444-444444444444",
        "title": "互联网+创新创业大赛",
        "description": "参加全国互联网+创新创业大赛",
        "category": "创新创业实践活动",
        "status": "approved",
        "start_date": "2024-03-01",
        "end_date": "2024-06-30",
        "details": {
            "item": "互联网+创新创业大赛",
            "company": "教育部",
            "project_no": "INT2024001",
            "total_hours": 120.00
        },
        "owner": {
            "uuid": "33333333-3333-3333-3333-333333333333",
            "username": "student",
            "real_name": "张三"
        },
        "reviewer": {
            "uuid": "22222222-2222-2222-2222-222222222222",
            "username": "teacher",
            "real_name": "李教授"
        },
        "participants": [
            {
                "user_id": "33333333-3333-3333-3333-333333333333",
                "username": "student1",
                "real_name": "张三",
                "credits": 2.0
            }
        ],
        "attachments": [
            {
                "id": "...",
                "file_name": "project_proposal.pdf",
                "original_name": "项目申请书.pdf",
                "file_size": 2048576,
                "file_type": ".pdf",
                "uploaded_at": "2024-01-01T00:00:00Z"
            }
        ],
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

### 3. 创建活动

**接口地址：** `POST /api/activities`

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 活动标题 |
| description | string | 否 | 活动描述 |
| category | string | 是 | 活动类别 |
| start_date | string | 是 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 是 | 结束日期 (YYYY-MM-DD) |
| details | object | 否 | 活动详情（JSONB） |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/activities \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "ACM程序设计竞赛",
    "description": "参加ACM国际大学生程序设计竞赛",
    "category": "学科竞赛",
    "start_date": "2024-10-15",
    "end_date": "2024-11-15",
    "details": {
        "level": "国家级",
        "competition": "ACM程序设计竞赛",
        "award_level": "三等奖"
    }
  }'
```

**成功响应：**

```json
{
    "code": 0,
    "message": "创建成功",
    "data": {
        "id": "...",
        "title": "ACM程序设计竞赛",
        "status": "draft",
        ...
    }
}
```

### 4. 更新活动

**接口地址：** `PUT /api/activities/:id`

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: application/json
```

**请求参数：** 同创建活动

**请求示例：**

```bash
curl -X PUT http://localhost:8080/api/activities/44444444-4444-4444-4444-444444444444 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "更新后的标题",
    "description": "更新后的描述"
  }'
```

**成功响应：**

```json
{
    "code": 0,
    "message": "更新成功",
    "data": {
        "id": "44444444-4444-4444-4444-444444444444",
        "title": "更新后的标题",
        ...
    }
}
```

### 5. 删除活动

**接口地址：** `DELETE /api/activities/:id`

**请求头：**
```http
Authorization: Bearer <token>
```

**请求示例：**

```bash
curl -X DELETE http://localhost:8080/api/activities/44444444-4444-4444-4444-444444444444 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**成功响应：**

```json
{
    "code": 0,
    "message": "删除成功",
    "data": null
}
```

### 6. 获取我的活动

**接口地址：** `GET /api/activities/my`

**请求头：**
```http
Authorization: Bearer <token>
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |
| role | string | 否 | 角色：owner/participant |

**请求示例：**

```bash
curl -X GET "http://localhost:8080/api/activities/my?role=participant" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**成功响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 10,
        "activities": [
            {
                "id": "...",
                "title": "互联网+创新创业大赛",
                "my_credits": 2.0,
                "awarded_credits": 2.0,
                ...
            }
        ]
    }
}
```

#### 7. 获取活动统计信息

**接口地址：** `GET /api/activities/stats`

**权限：** 所有认证用户

**请求头：**
```http
Authorization: Bearer <token>
```

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total_activities": 150,
        "my_activities": 15,
        "pending_review": 5,
        "approved": 120,
        "rejected": 10,
        "by_category": {
            "学科竞赛": 50,
            "创新创业实践活动": 40
        }
    }
}
```

#### 8. 获取活动分类列表

**接口地址：** `GET /api/activities/categories`

**权限：** 所有认证用户

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "categories": [
            "学科竞赛",
            "创新创业实践活动",
            "论文专利",
            "文体活动",
            "社会实践"
        ]
    }
}
```

#### 9. 获取活动模板列表

**接口地址：** `GET /api/activities/templates`

**权限：** 所有认证用户

**说明：** 获取预定义的活动模板，方便快速创建活动

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "templates": [
            {
                "id": "template_1",
                "name": "ACM程序设计竞赛模板",
                "category": "学科竞赛",
                "details": {
                    "level": "国家级",
                    "competition": "ACM程序设计竞赛"
                }
            }
        ]
    }
}
```

#### 10. 提交活动申请

**接口地址：** `POST /api/activities/:id/submit`

**权限：** 所有认证用户

**说明：** 将草稿状态的活动提交审核

**请求示例：**
```bash
curl -X POST http://localhost:8080/api/activities/44444444-4444-4444-4444-444444444444/submit \
  -H "Authorization: Bearer <token>"
```

**成功响应：**
```json
{
    "code": 0,
    "message": "提交成功",
    "data": {
        "id": "44444444-4444-4444-4444-444444444444",
        "status": "pending_review"
    }
}
```

#### 11. 撤回活动申请

**接口地址：** `POST /api/activities/:id/withdraw`

**权限：** 活动创建者

**说明：** 撤回待审核的活动申请

#### 12. 复制活动

**接口地址：** `POST /api/activities/:id/copy`

**权限：** 所有认证用户

**说明：** 复制现有活动创建新活动

**成功响应：**
```json
{
    "code": 0,
    "message": "复制成功",
    "data": {
        "id": "new_activity_uuid",
        "title": "互联网+创新创业大赛 (副本)",
        "status": "draft"
    }
}
```

#### 13. 保存为模板

**接口地址：** `POST /api/activities/:id/save-template`

**权限：** 教师或管理员

**请求参数：**
```json
{
    "template_name": "互联网+竞赛标准模板",
    "description": "适用于互联网+创新创业大赛的标准模板"
}
```

#### 14. 批量创建活动

**接口地址：** `POST /api/activities/batch`

**权限：** 教师或管理员

**请求参数：**
```json
{
    "activities": [
        {
            "title": "活动1",
            "category": "学科竞赛",
            "start_date": "2024-01-01",
            "end_date": "2024-06-30"
        }
    ]
}
```

#### 15. 导入活动

**接口地址：** `POST /api/activities/import`

**权限：** 教师或管理员

**请求头：**
```http
Content-Type: multipart/form-data
```

**请求参数：**
- `file`: Excel或CSV文件

#### 16. 导出活动

**接口地址：** `GET /api/activities/export`

**权限：** 所有认证用户

**查询参数：**
- `format`: 导出格式（csv, excel, pdf）
- `category`: 活动分类
- `status`: 活动状态

**请求示例：**
```bash
curl -X GET "http://localhost:8080/api/activities/export?format=excel&category=学科竞赛" \
  -H "Authorization: Bearer <token>" \
  --output activities.xlsx
```

---

### 活动审核

#### 17. 审核活动

**接口地址：** `POST /api/activities/:id/review`

**权限：** 教师或管理员

**请求参数：**
```json
{
    "action": "approve",
    "comment": "审核意见",
    "awarded_credits": 2.5
}
```

**请求示例：**
```bash
curl -X POST http://localhost:8080/api/activities/44444444-4444-4444-4444-444444444444/review \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "approve",
    "comment": "活动符合要求，批准",
    "awarded_credits": 2.5
  }'
```

#### 18. 获取待审核活动列表

**接口地址：** `GET /api/activities/pending`

**权限：** 教师或管理员

**查询参数：**
- `page`: 页码
- `page_size`: 每页数量

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 15,
        "activities": [
            {
                "id": "uuid",
                "title": "ACM程序设计竞赛",
                "status": "pending_review",
                "owner": {
                    "username": "student1",
                    "real_name": "张三"
                }
            }
        ]
    }
}
```

---

### 参与者管理

#### 19. 添加参与者

**接口地址：** `POST /api/activities/:id/participants`

**权限：** 活动创建者或管理员

**请求参数：**
```json
{
    "user_id": "user_uuid",
    "credits": 2.5,
    "role": "participant"
}
```

#### 20. 移除参与者

**接口地址：** `DELETE /api/activities/:id/participants/:user_id`

**权限：** 活动创建者或管理员

#### 21. 更新参与者信息

**接口地址：** `PUT /api/activities/:id/participants/:user_id`

**权限：** 活动创建者或管理员

**请求参数：**
```json
{
    "credits": 3.0,
    "role": "leader"
}
```

#### 22. 设置参与者学分

**接口地址：** `POST /api/activities/:id/participants/:user_id/credits`

**权限：** 教师或管理员

**请求参数：**
```json
{
    "credits": 2.5
}
```

#### 23. 批量添加参与者

**接口地址：** `POST /api/activities/:id/participants/batch`

**权限：** 活动创建者或管理员

**请求参数：**
```json
{
    "participants": [
        {
            "user_id": "uuid1",
            "credits": 2.5,
            "role": "leader"
        }
    ]
}
```

#### 24. 获取参与者列表

**接口地址：** `GET /api/activities/:id/participants`

**权限：** 所有认证用户

---

### 附件管理

#### 25. 上传附件

**接口地址：** `POST /api/activities/:id/attachments`

**请求头：**
```http
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 文件 |
| description | string | 否 | 文件描述 |

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/activities/44444444-4444-4444-4444-444444444444/attachments \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F "file=@/path/to/file.pdf" \
  -F "description=项目申请书"
```

**成功响应：**

```json
{
    "code": 0,
    "message": "上传成功",
    "data": {
        "id": "...",
        "file_name": "abc123.pdf",
        "original_name": "file.pdf",
        "file_size": 2048576,
        "file_type": ".pdf",
        "description": "项目申请书",
        "uploaded_at": "2024-01-01T00:00:00Z"
    }
}
```

#### 26. 下载附件

**接口地址：** `GET /api/activities/:id/attachments/:attachment_id/download`

**权限：** 所有认证用户

**请求示例：**
```bash
curl -X GET http://localhost:8080/api/activities/44444444-4444-4444-4444-444444444444/attachments/attachment_id/download \
  -H "Authorization: Bearer <token>" \
  --output document.pdf
```

#### 27. 预览附件

**接口地址：** `GET /api/activities/:id/attachments/:attachment_id/preview`

**权限：** 所有认证用户

**说明：** 在线预览附件（支持PDF、图片等）

#### 28. 删除附件

**接口地址：** `DELETE /api/activities/:id/attachments/:attachment_id`

**权限：** 活动创建者或管理员

---

### 申请管理

#### 29. 获取申请列表

**接口地址：** `GET /api/activities/applications`

**权限：** 教师或管理员

**查询参数：**
- `status`: 申请状态（pending, approved, rejected）
- `user_id`: 用户ID
- `activity_id`: 活动ID
- `page`: 页码
- `page_size`: 每页数量

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 50,
        "applications": [
            {
                "id": "uuid",
                "activity": {
                    "id": "activity_uuid",
                    "title": "互联网+创新创业大赛"
                },
                "user": {
                    "username": "student1",
                    "real_name": "张三"
                },
                "status": "pending",
                "applied_credits": 2.5
            }
        ]
    }
}
```

#### 30. 获取申请统计

**接口地址：** `GET /api/activities/applications/stats`

**权限：** 教师或管理员

**成功响应：**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total_applications": 150,
        "pending": 25,
        "approved": 100,
        "rejected": 25
    }
}
```

---

### 搜索功能

#### 31. 搜索活动

**接口地址：** `GET /api/search/activities`

**权限：** 所有认证用户

**查询参数：**
- `keyword`: 搜索关键词（标题、描述）
- `category`: 活动分类
- `status`: 活动状态
- `page`: 页码
- `page_size`: 每页数量

**请求示例：**
```bash
curl -X GET "http://localhost:8080/api/search/activities?keyword=ACM&category=学科竞赛" \
  -H "Authorization: Bearer <token>"
```

#### 32. 搜索申请

**接口地址：** `GET /api/search/applications`

**权限：** 教师或管理员

#### 33. 搜索参与者

**接口地址：** `GET /api/search/participants`

**权限：** 所有认证用户

#### 34. 搜索附件

**接口地址：** `GET /api/search/attachments`

**权限：** 所有认证用户

---

## 权限说明

### 权限级别

1. **公共访问（无需认证）**
   - 配置选项获取
   - 头像访问
   - 学生注册

2. **所有认证用户**
   - 个人信息管理
   - 查看活动列表
   - 创建活动（草稿）
   - 搜索功能

3. **活动创建者**
   - 编辑自己的活动
   - 管理参与者
   - 上传附件

4. **教师**
   - 审核活动
   - 查看所有申请
   - 批量操作
   - 导出数据

5. **管理员**
   - 所有权限
   - 用户管理
   - 系统配置
   - 批量导入导出

### 权限验证

所有需要认证的接口都需要在请求头中携带JWT Token：

```http
Authorization: Bearer <your_jwt_token>
```

权限不足时返回：

```json
{
    "code": 403,
    "message": "权限不足",
    "data": null
}
```

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（未登录或token无效） |
| 403 | 禁止访问（权限不足） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 测试账号

### 管理员
- 用户名: `admin`
- 密码: `adminpassword`

### 教师
- 用户名: `teacher`
- 工号: `T0000001`
- 密码: `adminpassword`

### 学生
- 用户名: `student`
- 学号: `20240000`
- 密码: `adminpassword`

---

## 常见问题

### 1. 如何查看JWT Token？

**方法1：浏览器开发者工具**
- 打开开发者工具 (F12)
- 进入 Application → Local Storage
- 查找 `token` 键

**方法2：Network标签**
- 打开开发者工具 (F12)
- 进入 Network 标签
- 找到登录请求，查看Response

### 2. Token过期怎么办？

使用refresh token刷新：
```bash
POST /api/auth/refresh
{
    "refresh_token": "your_refresh_token"
}
```

### 3. 如何测试API？

**使用curl：**
```bash
# 1. 先登录获取token
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"student","password":"adminpassword"}' \
  | jq -r '.data.token')

# 2. 使用token访问其他接口
curl -X GET http://localhost:8080/api/activities \
  -H "Authorization: Bearer $TOKEN"
```

**使用Postman：**
1. 创建新请求
2. 在Headers中添加：`Authorization: Bearer <token>`
3. 发送请求

### 4. CORS跨域问题

如果前端遇到CORS错误，确保API网关已配置CORS：
```go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

---

## 更新日志

### v2.0.0 (2025-12-22)
- 🎉 **重大更新**：新增54个API接口
- 📝 **用户服务**：新增23个接口
  - 公共接口：配置选项、头像访问、学生注册
  - 个人信息管理：用户统计、头像管理、活动记录
  - 管理员功能：批量操作、导入导出、统计信息
  - 搜索功能：用户搜索
- 📝 **学分活动服务**：新增30个接口
  - 活动管理：统计、分类、模板、提交、撤回、复制、批量操作、导入导出
  - 审核流程：审核活动、待审核列表
  - 参与者管理：添加、移除、更新、批量操作
  - 附件管理：下载、预览、删除
  - 申请管理：列表、统计
  - 搜索功能：活动、申请、参与者、附件搜索
- 🔐 **权限说明**：新增详细的权限级别说明
- 📚 **文档结构**：优化目录结构，按功能模块分类

### v1.0.0 (2025-12-21)
- 初始版本
- 包含认证、用户、学分活动的主要API

---

**文档维护：** 请在API变更时及时更新本文档
**最后更新：** 2025-12-22
