# 学生信息服务 API 文档

## 概述
学生信息服务负责处理学生信息的创建、查询、更新、删除等操作，支持基于角色的数据访问控制。

## 权限说明

### 用户角色
- **admin**: 管理员，可以查看所有学生信息，管理所有学生
- **teacher**: 教师，可以查看所有学生信息
- **student**: 学生，只能查看自己的信息

### 数据访问权限
- **管理员**: 可以查看所有学生信息，创建、更新、删除学生
- **教师**: 可以查看所有学生信息
- **学生**: 只能查看自己的信息（UUID、姓名、学号）

### 数据过滤
- **管理员/教师**: 可以看到完整的学生信息
- **学生**: 只能看到基本信息（UUID、姓名、学号）

## 认证
所有API都需要在请求头中包含有效的JWT token：
```
Authorization: Bearer <token>
```

## API 端点

### 1. 获取所有学生
**GET** `/api/students`

**权限要求**: 所有认证用户

**查询参数**:
- `college`: 学院
- `major`: 专业
- `class`: 班级
- `grade`: 年级
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "students": [
      {
        "id": "student-uuid",
        "user_id": "user-uuid",
        "student_id": "2021001",
        "name": "张三",
        "college": "计算机学院",
        "major": "软件工程",
        "class": "软件2101",
        "grade": "2021",
        "enrollment_date": "2021-09-01",
        "status": "active"
      }
    ],
    "total": 100,
    "page": 1,
    "limit": 10
  }
}
```

**数据过滤说明**:
- 管理员/教师：看到完整信息
- 学生：只看到 `id`, `name`, `student_id`

### 2. 获取指定学生
**GET** `/api/students/{id}`

**权限要求**: 所有认证用户（学生只能查看自己的信息）

**路径参数**:
- `id`: 学生UUID

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "student-uuid",
    "user_id": "user-uuid",
    "student_id": "2021001",
    "name": "张三",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2101",
    "grade": "2021",
    "enrollment_date": "2021-09-01",
    "status": "active",
    "created_at": "2024-01-01T09:00:00Z",
    "updated_at": "2024-01-01T09:00:00Z"
  }
}
```

### 3. 创建学生
**POST** `/api/students`

**权限要求**: 仅管理员

**请求体**:
```json
{
  "user_id": "user-uuid",
  "student_id": "2021001",
  "college": "计算机学院",
  "major": "软件工程",
  "class": "软件2101",
  "grade": "2021",
  "enrollment_date": "2021-09-01",
  "status": "active"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "student-uuid",
    "user_id": "user-uuid",
    "student_id": "2021001",
    "name": "张三",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2101",
    "grade": "2021",
    "enrollment_date": "2021-09-01",
    "status": "active",
    "created_at": "2024-01-01T09:00:00Z"
  }
}
```

### 4. 更新学生信息
**PUT** `/api/students/{id}`

**权限要求**: 管理员、学生本人

**路径参数**:
- `id`: 学生UUID

**请求体**:
```json
{
  "college": "计算机学院",
  "major": "软件工程",
  "class": "软件2101",
  "grade": "2021",
  "status": "active"
}
```

**说明**: 学生只能更新部分信息，不能修改学号

### 5. 删除学生
**DELETE** `/api/students/{id}`

**权限要求**: 仅管理员

**路径参数**:
- `id`: 学生UUID

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "deleted_student_id": "student-uuid",
    "message": "学生信息已删除"
  }
}
```

### 6. 根据用户ID删除学生
**DELETE** `/api/students/user/{user_id}`

**权限要求**: 仅管理员

**路径参数**:
- `user_id`: 用户UUID

**说明**: 用于删除用户时同步删除学生信息

### 7. 搜索学生
**GET** `/api/students/search`

**权限要求**: 管理员、教师

**查询参数**:
- `q`: 搜索关键词（姓名、学号）
- `college`: 学院
- `major`: 专业
- `class`: 班级
- `page`: 页码
- `limit`: 每页数量

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "students": [
      {
        "id": "student-uuid",
        "student_id": "2021001",
        "name": "张三",
        "college": "计算机学院",
        "major": "软件工程",
        "class": "软件2101"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 10
  }
}
```

### 8. 获取学生统计
**GET** `/api/students/stats`

**权限要求**: 管理员、教师

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_students": 1000,
    "active_students": 980,
    "inactive_students": 20,
    "college_stats": {
      "计算机学院": 200,
      "机械学院": 150,
      "经管学院": 180
    },
    "grade_stats": {
      "2021": 250,
      "2022": 240,
      "2023": 260,
      "2024": 250
    }
  }
}
```

### 9. 批量导入学生
**POST** `/api/students/batch`

**权限要求**: 仅管理员

**请求**: `multipart/form-data`
- `file`: CSV文件

**CSV格式**:
```csv
student_id,name,college,major,class,grade,enrollment_date,status
2021001,张三,计算机学院,软件工程,软件2101,2021,2021-09-01,active
2021002,李四,计算机学院,软件工程,软件2101,2021,2021-09-01,active
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_rows": 100,
    "imported_count": 95,
    "failed_count": 5,
    "errors": [
      {
        "row": 3,
        "error": "学号已存在"
      }
    ]
  }
}
```

## 数据字段说明

### 学生信息字段
- `id`: 学生记录UUID
- `user_id`: 关联用户UUID
- `student_id`: 学号（唯一）
- `name`: 姓名
- `college`: 学院
- `major`: 专业
- `class`: 班级
- `grade`: 年级
- `enrollment_date`: 入学日期
- `status`: 状态（active, inactive, graduated, suspended）

### 数据过滤规则
- **管理员/教师**: 可以看到所有字段
- **学生**: 只能看到 `id`, `name`, `student_id`

## 错误码
- `0`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 冲突（如学号已存在）
- `500`: 服务器内部错误

## 使用说明
1. **权限控制**: 基于角色的数据访问控制
2. **数据过滤**: 学生只能看到基本信息
3. **学生创建**: 只有管理员可以创建学生
4. **信息更新**: 学生可以更新部分信息
5. **学生删除**: 只有管理员可以删除学生
6. **搜索功能**: 支持按关键词搜索
7. **统计功能**: 提供学生数量统计
8. **批量导入**: 支持CSV批量导入
9. **分页查询**: 支持分页获取学生列表
10. **数据同步**: 与用户管理服务同步

## 注意事项
1. 学号必须唯一
2. 学生信息与用户信息关联
3. 删除用户时会同步删除学生信息
4. 学生只能查看自己的信息
5. 教师可以查看所有学生信息
6. 管理员拥有所有权限
7. 支持批量导入学生信息
8. 提供详细的数据统计 