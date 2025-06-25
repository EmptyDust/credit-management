# 学分活动服务 API 文档

## 概述

学分活动服务提供完整的活动管理、参与者管理和申请管理功能。该服务支持学生和教师创建活动，但只有学生可以参与活动。活动通过审核后会自动生成申请记录。

## 基础信息

- **服务端口**: 8083
- **基础路径**: `/api`
- **认证方式**: JWT Token
- **数据格式**: JSON

## 认证

所有API请求都需要在Header中包含有效的JWT Token：

```
Authorization: Bearer <token>
```

## 通用响应格式

### 成功响应
```json
{
  "success": true,
  "message": "操作成功",
  "data": {}
}
```

### 错误响应
```json
{
  "success": false,
  "message": "错误信息",
  "error": "详细错误描述"
}
```

## 活动管理 API

### 1. 创建活动

**POST** `/api/activities`

创建新的学分活动。

**权限**: 所有认证用户（学生和教师都可以创建活动）

**请求体**:
```json
{
  "title": "学术讲座",
  "description": "关于人工智能的学术讲座",
  "start_date": "2024-01-15T09:00:00Z",
  "end_date": "2024-01-15T11:00:00Z",
  "category": "学术活动",
  "requirements": "需要提前报名"
}
```

**响应**:
```json
{
  "success": true,
  "message": "活动创建成功",
  "data": {
    "id": "uuid",
    "title": "学术讲座",
    "description": "关于人工智能的学术讲座",
    "start_date": "2024-01-15T09:00:00Z",
    "end_date": "2024-01-15T11:00:00Z",
    "status": "draft",
    "category": "学术活动",
    "requirements": "需要提前报名",
    "owner_id": "user-uuid",
    "created_at": "2024-01-10T10:00:00Z",
    "updated_at": "2024-01-10T10:00:00Z"
  }
}
```

### 2. 获取活动列表

**GET** `/api/activities`

获取活动列表，支持分页和筛选。

**权限**: 所有认证用户（学生只能看到自己创建或参与的活动，教师可以看到所有活动）

**查询参数**:
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10
- `status` (可选): 活动状态筛选
- `category` (可选): 活动类别筛选
- `owner_id` (可选): 创建者ID筛选

**响应**:
```json
{
  "success": true,
  "message": "获取活动列表成功",
  "data": {
    "activities": [
      {
        "id": "uuid",
        "title": "学术讲座",
        "description": "关于人工智能的学术讲座",
        "start_date": "2024-01-15T09:00:00Z",
        "end_date": "2024-01-15T11:00:00Z",
        "status": "draft",
        "category": "学术活动",
        "requirements": "需要提前报名",
        "owner_id": "user-uuid",
        "created_at": "2024-01-10T10:00:00Z",
        "updated_at": "2024-01-10T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
}
```

### 3. 获取活动详情

**GET** `/api/activities/{id}`

获取指定活动的详细信息。

**权限**: 所有认证用户（学生只能看到自己创建或参与的活动）

**响应**:
```json
{
  "success": true,
  "message": "获取活动详情成功",
  "data": {
    "id": "uuid",
    "title": "学术讲座",
    "description": "关于人工智能的学术讲座",
    "start_date": "2024-01-15T09:00:00Z",
    "end_date": "2024-01-15T11:00:00Z",
    "status": "draft",
    "category": "学术活动",
    "requirements": "需要提前报名",
    "owner_id": "user-uuid",
    "reviewer_id": null,
    "review_comments": null,
    "reviewed_at": null,
    "created_at": "2024-01-10T10:00:00Z",
    "updated_at": "2024-01-10T10:00:00Z",
    "participants": [
      {
        "user_id": "student-uuid",
        "credits": 2.0,
        "joined_at": "2024-01-10T10:00:00Z",
        "user_info": {
          "id": "student-uuid",
          "name": "张三",
          "student_id": "2021001"
        }
      }
    ],
    "applications": []
  }
}
```

### 4. 更新活动

**PUT** `/api/activities/{id}`

更新活动信息。

**权限**: 活动创建者、管理员（只有草稿状态的活动可以修改）

**请求体**:
```json
{
  "title": "更新后的学术讲座",
  "description": "更新后的描述",
  "start_date": "2024-01-16T09:00:00Z",
  "end_date": "2024-01-16T11:00:00Z",
  "category": "学术活动",
  "requirements": "更新后的要求"
}
```

**响应**:
```json
{
  "success": true,
  "message": "活动更新成功",
  "data": {
    "id": "uuid",
    "title": "更新后的学术讲座",
    "description": "更新后的描述",
    "start_date": "2024-01-16T09:00:00Z",
    "end_date": "2024-01-16T11:00:00Z",
    "status": "draft",
    "category": "学术活动",
    "requirements": "更新后的要求",
    "owner_id": "user-uuid",
    "updated_at": "2024-01-10T11:00:00Z"
  }
}
```

### 5. 删除活动

**DELETE** `/api/activities/{id}`

删除活动。

**权限**: 活动创建者、管理员

**响应**:
```json
{
  "success": true,
  "message": "活动删除成功"
}
```

### 6. 提交活动审核

**POST** `/api/activities/{id}/submit`

提交活动进行审核。

**权限**: 活动创建者（只有草稿状态的活动可以提交）

**响应**:
```json
{
  "success": true,
  "message": "活动提交审核成功",
  "data": {
    "id": "uuid",
    "status": "pending_review",
    "submitted_at": "2024-01-10T12:00:00Z"
  }
}
```

### 7. 撤回活动

**POST** `/api/activities/{id}/withdraw`

撤回活动到草稿状态，同时删除所有相关申请。

**权限**: 活动创建者（在提交后、被拒绝后、已通过后都可以撤回）

**响应**:
```json
{
  "success": true,
  "message": "活动撤回成功",
  "data": {
    "id": "uuid",
    "status": "draft",
    "withdrawn_at": "2024-01-10T13:00:00Z"
  }
}
```

### 8. 审核活动

**POST** `/api/activities/{id}/review`

审核活动。

**权限**: 教师、管理员（只有待审核状态的活动可以审核）

**请求体**:
```json
{
  "status": "approved",
  "review_comments": "活动内容符合要求，审核通过"
}
```

**响应**:
```json
{
  "success": true,
  "message": "审核完成",
  "data": {
    "id": "uuid",
    "status": "approved",
    "reviewer_id": "teacher-uuid",
    "review_comments": "活动内容符合要求，审核通过",
    "reviewed_at": "2024-01-10T14:00:00Z"
  }
}
```

### 9. 获取待审核活动

**GET** `/api/activities/pending`

获取所有待审核的活动。

**权限**: 教师、管理员

**查询参数**:
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10

**响应**:
```json
{
  "success": true,
  "message": "获取待审核活动成功",
  "data": {
    "activities": [
      {
        "id": "uuid",
        "title": "学术讲座",
        "description": "关于人工智能的学术讲座",
        "start_date": "2024-01-15T09:00:00Z",
        "end_date": "2024-01-15T11:00:00Z",
        "status": "pending_review",
        "category": "学术活动",
        "owner_id": "user-uuid",
        "created_at": "2024-01-10T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10
  }
}
```

### 10. 获取活动统计

**GET** `/api/activities/stats`

获取活动统计信息。

**权限**: 所有认证用户

**响应**:
```json
{
  "success": true,
  "message": "获取活动统计成功",
  "data": {
    "total_activities": 10,
    "draft_count": 3,
    "pending_review_count": 2,
    "approved_count": 4,
    "rejected_count": 1,
    "my_activities": 5,
    "my_participations": 3
  }
}
```

## 参与者管理 API

### 1. 添加参与者

**POST** `/api/activities/{id}/participants`

为活动添加参与者。

**权限**: 活动创建者、管理员（只能添加学生用户）

**请求体**:
```json
{
  "user_ids": ["student-uuid-1", "student-uuid-2"],
  "credits": 2.0
}
```

**响应**:
```json
{
  "success": true,
  "message": "参与者添加成功",
  "data": {
    "added_count": 2,
    "participants": [
      {
        "user_id": "student-uuid-1",
        "credits": 2.0,
        "joined_at": "2024-01-10T15:00:00Z",
        "user_info": {
          "id": "student-uuid-1",
          "name": "张三",
          "student_id": "2021001"
        }
      },
      {
        "user_id": "student-uuid-2",
        "credits": 2.0,
        "joined_at": "2024-01-10T15:00:00Z",
        "user_info": {
          "id": "student-uuid-2",
          "name": "李四",
          "student_id": "2021002"
        }
      }
    ]
  }
}
```

### 2. 批量设置学分

**PUT** `/api/activities/{id}/participants/batch-credits`

批量设置参与者的学分。

**权限**: 活动创建者、管理员

**请求体**:
```json
{
  "credits_map": {
    "student-uuid-1": 3.0,
    "student-uuid-2": 2.5
  }
}
```

**响应**:
```json
{
  "success": true,
  "message": "学分设置成功",
  "data": {
    "updated_count": 2,
    "participants": [
      {
        "user_id": "student-uuid-1",
        "credits": 3.0,
        "user_info": {
          "id": "student-uuid-1",
          "name": "张三",
          "student_id": "2021001"
        }
      },
      {
        "user_id": "student-uuid-2",
        "credits": 2.5,
        "user_info": {
          "id": "student-uuid-2",
          "name": "李四",
          "student_id": "2021002"
        }
      }
    ]
  }
}
```

### 3. 设置单个学分

**PUT** `/api/activities/{id}/participants/{user_id}/credits`

设置单个参与者的学分。

**权限**: 活动创建者、管理员

**请求体**:
```json
{
  "credits": 3.0
}
```

**响应**:
```json
{
  "success": true,
  "message": "学分设置成功",
  "data": {
    "user_id": "student-uuid",
    "credits": 3.0,
    "user_info": {
      "id": "student-uuid",
      "name": "张三",
      "student_id": "2021001"
    }
  }
}
```

### 4. 删除参与者

**DELETE** `/api/activities/{id}/participants/{user_id}`

从活动中删除参与者。

**权限**: 活动创建者、管理员

**响应**:
```json
{
  "success": true,
  "message": "参与者删除成功"
}
```

### 5. 退出活动

**POST** `/api/activities/{id}/leave`

参与者主动退出活动。

**权限**: 活动参与者（学生）

**响应**:
```json
{
  "success": true,
  "message": "退出活动成功"
}
```

### 6. 获取活动参与者列表

**GET** `/api/activities/{id}/participants`

获取活动的参与者列表。

**权限**: 所有认证用户

**响应**:
```json
{
  "success": true,
  "message": "获取参与者列表成功",
  "data": {
    "participants": [
      {
        "user_id": "student-uuid",
        "credits": 2.0,
        "joined_at": "2024-01-10T15:00:00Z",
        "user_info": {
          "id": "student-uuid",
          "name": "张三",
          "student_id": "2021001"
        }
      }
    ],
    "total": 1
  }
}
```

## 申请管理 API

### 1. 获取用户申请列表

**GET** `/api/applications`

获取当前用户的申请列表。

**权限**: 所有认证用户（学生只能看到自己的申请）

**查询参数**:
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10
- `status` (可选): 申请状态筛选

**响应**:
```json
{
  "success": true,
  "message": "获取申请列表成功",
  "data": {
    "applications": [
      {
        "id": "uuid",
        "activity_id": "activity-uuid",
        "user_id": "student-uuid",
        "status": "approved",
        "applied_credits": 2.0,
        "awarded_credits": 2.0,
        "submitted_at": "2024-01-10T16:00:00Z",
        "created_at": "2024-01-10T16:00:00Z",
        "activity": {
          "id": "activity-uuid",
          "title": "学术讲座",
          "description": "关于人工智能的学术讲座",
          "category": "学术活动",
          "start_date": "2024-01-15T09:00:00Z",
          "end_date": "2024-01-15T11:00:00Z"
        }
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
}
```

### 2. 获取申请详情

**GET** `/api/applications/{id}`

获取申请详情。

**权限**: 所有认证用户（学生只能看到自己的申请）

**响应**:
```json
{
  "success": true,
  "message": "获取申请详情成功",
  "data": {
    "id": "uuid",
    "activity_id": "activity-uuid",
    "user_id": "student-uuid",
    "status": "approved",
    "applied_credits": 2.0,
    "awarded_credits": 2.0,
    "submitted_at": "2024-01-10T16:00:00Z",
    "created_at": "2024-01-10T16:00:00Z",
    "activity": {
      "id": "activity-uuid",
      "title": "学术讲座",
      "description": "关于人工智能的学术讲座",
      "category": "学术活动",
      "start_date": "2024-01-15T09:00:00Z",
      "end_date": "2024-01-15T11:00:00Z"
    }
  }
}
```

### 3. 获取所有申请

**GET** `/api/applications/all`

获取所有申请（教师/管理员功能）。

**权限**: 教师、管理员

**查询参数**:
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10
- `status` (可选): 申请状态筛选
- `activity_id` (可选): 活动ID筛选
- `user_id` (可选): 用户ID筛选

**响应**:
```json
{
  "success": true,
  "message": "获取所有申请成功",
  "data": {
    "applications": [
      {
        "id": "uuid",
        "activity_id": "activity-uuid",
        "user_id": "student-uuid",
        "status": "approved",
        "applied_credits": 2.0,
        "awarded_credits": 2.0,
        "submitted_at": "2024-01-10T16:00:00Z",
        "created_at": "2024-01-10T16:00:00Z",
        "activity": {
          "id": "activity-uuid",
          "title": "学术讲座",
          "description": "关于人工智能的学术讲座",
          "category": "学术活动"
        },
        "user_info": {
          "id": "student-uuid",
          "name": "张三",
          "student_id": "2021001"
        }
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10
  }
}
```

### 4. 导出申请数据

**GET** `/api/applications/export`

导出申请数据为CSV格式。

**权限**: 所有认证用户（学生只能导出自己的申请，教师/管理员可以导出所有申请）

**查询参数**:
- `format` (可选): 导出格式，支持 "csv", "excel"，默认 "csv"
- `activity_id` (可选): 活动ID筛选
- `status` (可选): 申请状态筛选
- `start_date` (可选): 开始日期筛选
- `end_date` (可选): 结束日期筛选

**响应**: 文件下载（CSV或Excel格式）

### 5. 获取申请统计

**GET** `/api/applications/stats`

获取申请统计信息。

**权限**: 所有认证用户

**响应**:
```json
{
  "success": true,
  "message": "获取申请统计成功",
  "data": {
    "total_applications": 10,
    "approved_count": 8,
    "total_credits": 20.0,
    "awarded_credits": 16.0,
    "my_applications": 5,
    "my_awarded_credits": 8.0
  }
}
```

## 健康检查

### 健康检查

**GET** `/health`

检查服务健康状态。

**响应**:
```json
{
  "status": "ok",
  "timestamp": "2024-01-10T10:00:00Z",
  "service": "credit-activity-service",
  "version": "1.0.0"
}
```

## 错误码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 业务逻辑错误 |
| 500 | 服务器内部错误 |

## 常见错误响应

### 权限不足
```json
{
  "success": false,
  "message": "权限不足",
  "error": "您没有权限执行此操作"
}
```

### 资源不存在
```json
{
  "success": false,
  "message": "资源不存在",
  "error": "指定的活动不存在"
}
```

### 业务逻辑错误
```json
{
  "success": false,
  "message": "操作失败",
  "error": "只有草稿状态的活动可以修改"
}
```

### 参数验证错误
```json
{
  "success": false,
  "message": "参数错误",
  "error": "标题不能为空"
}
``` 