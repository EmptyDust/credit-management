# 教师信息服务 API 文档

## 概述
教师信息服务负责处理教师信息的创建、查询、更新、删除等操作，支持基于角色的数据访问控制。

## 权限说明

### 用户角色
- **admin**: 管理员，可以查看所有教师信息，管理所有教师
- **teacher**: 教师，可以查看其他教师的基本信息
- **student**: 学生，只能查看教师的基本信息

### 数据访问权限
- **管理员**: 可以查看所有教师信息，创建、更新、删除教师
- **教师**: 可以查看其他教师的基本信息
- **学生**: 只能查看教师的基本信息（UUID、姓名、部门、职称）

### 数据过滤
- **管理员**: 可以看到完整的教师信息
- **教师/学生**: 只能看到基本信息（UUID、姓名、部门、职称）

## 认证
所有API都需要在请求头中包含有效的JWT token：
```
Authorization: Bearer <token>
```

## API 端点

### 1. 获取所有教师
**GET** `/api/teachers`

**权限要求**: 所有认证用户

**查询参数**:
- `department`: 部门
- `title`: 职称
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "teachers": [
      {
        "id": "teacher-uuid",
        "user_id": "user-uuid",
        "employee_id": "T2021001",
        "name": "王老师",
        "department": "计算机学院",
        "title": "副教授",
        "phone": "13800138000",
        "email": "teacher@example.com",
        "hire_date": "2021-09-01",
        "status": "active"
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 10
  }
}
```

**数据过滤说明**:
- 管理员：看到完整信息
- 教师/学生：只看到 `id`, `name`, `department`, `title`

### 2. 获取指定教师
**GET** `/api/teachers/{id}`

**权限要求**: 所有认证用户

**路径参数**:
- `id`: 教师UUID

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "teacher-uuid",
    "user_id": "user-uuid",
    "employee_id": "T2021001",
    "name": "王老师",
    "department": "计算机学院",
    "title": "副教授",
    "phone": "13800138000",
    "email": "teacher@example.com",
    "hire_date": "2021-09-01",
    "status": "active",
    "created_at": "2024-01-01T09:00:00Z",
    "updated_at": "2024-01-01T09:00:00Z"
  }
}
```

### 3. 创建教师
**POST** `/api/teachers`

**权限要求**: 仅管理员

**请求体**:
```json
{
  "user_id": "user-uuid",
  "employee_id": "T2021001",
  "department": "计算机学院",
  "title": "副教授",
  "phone": "13800138000",
  "email": "teacher@example.com",
  "hire_date": "2021-09-01",
  "status": "active"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "teacher-uuid",
    "user_id": "user-uuid",
    "employee_id": "T2021001",
    "name": "王老师",
    "department": "计算机学院",
    "title": "副教授",
    "phone": "13800138000",
    "email": "teacher@example.com",
    "hire_date": "2021-09-01",
    "status": "active",
    "created_at": "2024-01-01T09:00:00Z"
  }
}
```

### 4. 更新教师信息
**PUT** `/api/teachers/{id}`

**权限要求**: 管理员、教师本人

**路径参数**:
- `id`: 教师UUID

**请求体**:
```json
{
  "department": "计算机学院",
  "title": "教授",
  "phone": "13800138001",
  "email": "newemail@example.com",
  "status": "active"
}
```

**说明**: 教师只能更新部分信息，不能修改工号

### 5. 删除教师
**DELETE** `/api/teachers/{id}`

**权限要求**: 仅管理员

**路径参数**:
- `id`: 教师UUID

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "deleted_teacher_id": "teacher-uuid",
    "message": "教师信息已删除"
  }
}
```

### 6. 根据用户ID删除教师
**DELETE** `/api/teachers/user/{user_id}`

**权限要求**: 仅管理员

**路径参数**:
- `user_id`: 用户UUID

**说明**: 用于删除用户时同步删除教师信息

### 7. 搜索教师
**GET** `/api/teachers/search`

**权限要求**: 管理员、教师

**查询参数**:
- `q`: 搜索关键词（姓名、工号）
- `department`: 部门
- `title`: 职称
- `page`: 页码
- `limit`: 每页数量

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "teachers": [
      {
        "id": "teacher-uuid",
        "employee_id": "T2021001",
        "name": "王老师",
        "department": "计算机学院",
        "title": "副教授"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 10
  }
}
```

### 8. 获取教师统计
**GET** `/api/teachers/stats`

**权限要求**: 管理员、教师

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_teachers": 50,
    "active_teachers": 48,
    "inactive_teachers": 2,
    "department_stats": {
      "计算机学院": 15,
      "机械学院": 12,
      "经管学院": 10,
      "文学院": 8,
      "理学院": 5
    },
    "title_stats": {
      "教授": 10,
      "副教授": 20,
      "讲师": 15,
      "助教": 5
    }
  }
}
```

### 9. 批量导入教师
**POST** `/api/teachers/batch`

**权限要求**: 仅管理员

**请求**: `multipart/form-data`
- `file`: CSV文件

**CSV格式**:
```csv
employee_id,name,department,title,phone,email,hire_date,status
T2021001,王老师,计算机学院,副教授,13800138000,teacher1@example.com,2021-09-01,active
T2021002,李老师,机械学院,教授,13800138001,teacher2@example.com,2020-09-01,active
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_rows": 50,
    "imported_count": 48,
    "failed_count": 2,
    "errors": [
      {
        "row": 3,
        "error": "工号已存在"
      }
    ]
  }
}
```

### 10. 获取部门教师列表
**GET** `/api/teachers/department/{department}`

**权限要求**: 所有认证用户

**路径参数**:
- `department`: 部门名称

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "department": "计算机学院",
    "teachers": [
      {
        "id": "teacher-uuid",
        "name": "王老师",
        "title": "副教授"
      }
    ],
    "total": 15
  }
}
```

## 数据字段说明

### 教师信息字段
- `id`: 教师记录UUID
- `user_id`: 关联用户UUID
- `employee_id`: 工号（唯一）
- `name`: 姓名
- `department`: 部门
- `title`: 职称
- `phone`: 联系电话
- `email`: 邮箱
- `hire_date`: 入职日期
- `status`: 状态（active, inactive, retired, suspended）

### 数据过滤规则
- **管理员**: 可以看到所有字段
- **教师/学生**: 只能看到 `id`, `name`, `department`, `title`

## 错误码
- `0`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 冲突（如工号已存在）
- `500`: 服务器内部错误

## 使用说明
1. **权限控制**: 基于角色的数据访问控制
2. **数据过滤**: 教师和学生只能看到基本信息
3. **教师创建**: 只有管理员可以创建教师
4. **信息更新**: 教师可以更新部分信息
5. **教师删除**: 只有管理员可以删除教师
6. **搜索功能**: 支持按关键词搜索
7. **统计功能**: 提供教师数量统计
8. **批量导入**: 支持CSV批量导入
9. **分页查询**: 支持分页获取教师列表
10. **数据同步**: 与用户管理服务同步

## 注意事项
1. 工号必须唯一
2. 教师信息与用户信息关联
3. 删除用户时会同步删除教师信息
4. 教师和学生只能查看基本信息
5. 管理员拥有所有权限
6. 支持批量导入教师信息
7. 提供详细的数据统计
8. 支持按部门查询教师 