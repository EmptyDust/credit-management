# 学分活动服务 API 文档

## 概述

学分活动服务提供完整的学分活动管理功能，包括活动创建、参与者管理、申请处理和附件管理。

**基础 URL**: `http://localhost:8083`
**API 前缀**: `/api`

## 认证

所有需要认证的接口都需要在请求头中包含 `Authorization` 字段：

```
Authorization: Bearer <token>
```

## 通用响应格式

所有 API 响应都使用统一的格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

**响应码说明**:

- `0`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

## 1. 活动管理 API

### 1.1 获取活动类别

**GET** `/api/activities/categories`

获取所有可用的活动类别。

**权限**: 无需认证

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "categories": ["创新创业", "学科竞赛", "志愿服务", "学术研究", "文体活动"],
    "count": 5,
    "description": "活动类别列表"
  }
}
```

### 1.2 获取活动模板

**GET** `/api/activities/templates`

获取预定义的活动模板，方便用户快速创建活动。

**权限**: 无需认证

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "创新创业活动",
      "category": "创新创业",
      "title": "创新创业实践活动",
      "description": "参与创新创业项目，提升创新能力和实践技能",
      "requirements": "需要提交项目计划书和成果展示"
    },
    {
      "name": "学科竞赛",
      "category": "学科竞赛",
      "title": "学科竞赛活动",
      "description": "参加各类学科竞赛，提升专业能力和竞争意识",
      "requirements": "需要获得竞赛证书或奖项证明"
    }
  ]
}
```

### 1.3 创建活动

**POST** `/api/activities`

创建单个活动。

**权限**: 所有认证用户

**请求体**:

```json
{
  "title": "创新创业实践活动",
  "description": "参与创新创业项目，提升创新能力和实践技能",
  "start_date": "2024-12-01",
  "end_date": "2024-12-31",
  "category": "创新创业",
  "requirements": "需要提交项目计划书和成果展示"
}
```

**支持的日期格式**:

- `YYYY-MM-DD` (如: 2024-12-01)
- `YYYY-MM-DD HH:mm:ss` (如: 2024-12-01 10:00:00)
- `YYYY-MM-DDTHH:mm:ss` (如: 2024-12-01T10:00:00)
- `YYYY-MM-DDTHH:mm:ssZ` (如: 2024-12-01T10:00:00Z)

**响应示例**:

```json
{
  "code": 0,
  "message": "活动创建成功",
  "data": {
    "id": "uuid",
    "title": "创新创业实践活动",
    "description": "参与创新创业项目，提升创新能力和实践技能",
    "start_date": "2024-12-01T00:00:00Z",
    "end_date": "2024-12-31T00:00:00Z",
    "status": "draft",
    "category": "创新创业",
    "requirements": "需要提交项目计划书和成果展示",
    "owner_id": "user-uuid",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 1.4 批量创建活动

**POST** `/api/activities/batch`

批量创建多个活动（最多 10 个）。

**权限**: 所有认证用户

**请求体**:

```json
{
  "activities": [
    {
      "title": "学科竞赛活动1",
      "description": "第一个学科竞赛活动",
      "start_date": "2024-12-01",
      "end_date": "2024-12-15",
      "category": "学科竞赛",
      "requirements": "需要获得竞赛证书"
    },
    {
      "title": "志愿服务活动1",
      "description": "第一个志愿服务活动",
      "start_date": "2024-12-16",
      "end_date": "2024-12-31",
      "category": "志愿服务",
      "requirements": "需要志愿服务时长证明"
    }
  ]
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "批量创建活动成功",
  "data": {
    "created_count": 2,
    "total_count": 2,
    "created_activities": [
      {
        "id": "uuid1",
        "title": "学科竞赛活动1",
        "status": "draft",
        "created_at": "2024-01-01T00:00:00Z"
      },
      {
        "id": "uuid2",
        "title": "志愿服务活动1",
        "status": "draft",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 1.5 批量更新活动

**PUT** `/api/activities/batch`

批量更新多个活动，支持主表和详情表的批量更新。

**权限**: 所有认证用户（只能更新自己创建的活动）

**请求体**:

```json
{
  "activities": [
    {
      "id": "uuid1",
      "title": "更新后的标题1",
      "description": "更新后的描述1",
      "start_date": "2024-12-01",
      "end_date": "2024-12-15",
      "category": "学科竞赛",
      "requirements": "更新后的要求1",
      "detail": {
        "competition_level": "省级",
        "award_level": "一等奖",
        "certificate_number": "CERT001"
      }
    },
    {
      "id": "uuid2",
      "title": "更新后的标题2",
      "description": "更新后的描述2",
      "start_date": "2024-12-16",
      "end_date": "2024-12-31",
      "category": "创新创业",
      "requirements": "更新后的要求2",
      "detail": {
        "project_type": "创新项目",
        "team_size": 5,
        "mentor_name": "李老师"
      }
    }
  ]
}
```

**字段说明**:

- `id`: 活动 ID（必填）
- 其他字段都是可选的，支持部分更新
- `detail`: 根据活动类别包含不同的详情字段
  - 学科竞赛: `competition_level`, `award_level`, `certificate_number`
  - 创新创业: `project_type`, `team_size`, `mentor_name`
  - 志愿服务: `service_hours`, `service_location`, `service_type`
  - 学术研究: `research_field`, `publication_type`, `co_authors`
  - 文体活动: `activity_type`, `performance_level`, `venue`

**响应示例**:

```json
{
  "code": 0,
  "message": "批量更新活动成功",
  "data": {
    "updated_count": 2,
    "total_count": 2,
    "updated_activities": [
      {
        "id": "uuid1",
        "title": "更新后的标题1",
        "status": "success",
        "updated_at": "2024-01-01T10:00:00Z"
      },
      {
        "id": "uuid2",
        "title": "更新后的标题2",
        "status": "success",
        "updated_at": "2024-01-01T10:00:00Z"
      }
    ],
    "errors": []
  }
}
```

**错误处理**:

```json
{
  "code": 400,
  "message": "批量更新活动失败",
  "data": {
    "updated_count": 1,
    "total_count": 2,
    "updated_activities": [
      {
        "id": "uuid1",
        "status": "success"
      }
    ],
    "errors": [
      {
        "id": "uuid2",
        "error": "活动不存在或无权限更新"
      }
    ]
  }
}
```

### 1.6 获取活动列表

**GET** `/api/activities`

获取活动列表，支持搜索、筛选和分页。

**权限**: 所有认证用户

**查询参数**:

- `query` (可选): 搜索关键词，支持标题、描述、类别、要求的模糊搜索
- `status` (可选): 活动状态筛选 (draft, pending_review, approved, rejected)
- `category` (可选): 活动类别筛选
- `owner_id` (可选): 创建者 ID 筛选
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 10

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": "uuid",
        "title": "创新创业实践活动",
        "description": "参与创新创业项目，提升创新能力和实践技能",
        "start_date": "2024-12-01T00:00:00Z",
        "end_date": "2024-12-31T00:00:00Z",
        "status": "draft",
        "category": "创新创业",
        "requirements": "需要提交项目计划书和成果展示",
        "owner_id": "user-uuid",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "participants": [],
        "applications": []
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
}
```

### 1.7 获取活动详情

**GET** `/api/activities/{id}`

获取指定活动的详细信息。

**权限**: 所有认证用户（学生只能看到自己创建或参与的活动）

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "创新创业实践活动",
    "description": "参与创新创业项目，提升创新能力和实践技能",
    "start_date": "2024-12-01T00:00:00Z",
    "end_date": "2024-12-31T00:00:00Z",
    "status": "draft",
    "category": "创新创业",
    "requirements": "需要提交项目计划书和成果展示",
    "owner_id": "user-uuid",
    "reviewer_id": null,
    "review_comments": null,
    "reviewed_at": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "participants": [
      {
        "user_id": "student-uuid",
        "credits": 2.0,
        "joined_at": "2024-01-01T00:00:00Z",
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

### 1.8 更新活动

**PUT** `/api/activities/{id}`

更新活动信息。只有活动创建者和管理员可以更新活动，且只有草稿状态的活动可以被普通用户修改。

**权限**: 活动创建者、管理员

**请求体** (支持部分更新):

```json
{
  "title": "更新后的活动标题",
  "description": "更新后的活动描述",
  "start_date": "2024-12-01",
  "end_date": "2024-12-31",
  "category": "学科竞赛",
  "requirements": "更新后的要求"
}
```

**字段说明**:

- 所有字段都是可选的，支持部分更新
- 使用指针类型，支持清空字段（传 null）
- 日期格式支持：`YYYY-MM-DD`、`YYYY-MM-DD HH:mm:ss`、`YYYY-MM-DDTHH:mm:ss`、`YYYY-MM-DDTHH:mm:ssZ`

**响应示例**:

```json
{
  "code": 0,
  "message": "活动更新成功",
  "data": {
    "id": "uuid",
    "title": "更新后的活动标题",
    "description": "更新后的活动描述",
    "start_date": "2024-12-01T00:00:00Z",
    "end_date": "2024-12-31T00:00:00Z",
    "status": "draft",
    "category": "学科竞赛",
    "requirements": "更新后的要求",
    "owner_id": "user-uuid",
    "updated_at": "2024-01-01T11:00:00Z"
  }
}
```

### 1.9 删除活动

**DELETE** `/api/activities/{id}`

删除活动。只有活动创建者和管理员可以删除活动。

**权限**: 活动创建者、管理员

**响应示例**:

```json
{
  "code": 0,
  "message": "活动删除成功",
  "data": {
    "activity_id": "uuid",
    "deleted_at": "2024-01-01T12:00:00Z"
  }
}
```

### 1.10 提交活动审核

**POST** `/api/activities/{id}/submit`

提交活动进行审核。只有草稿状态的活动可以提交。

**权限**: 活动创建者

**响应示例**:

```json
{
  "code": 0,
  "message": "活动提交审核成功",
  "data": {
    "id": "uuid",
    "status": "pending_review",
    "submitted_at": "2024-01-01T12:00:00Z"
  }
}
```

### 1.11 撤回活动

**POST** `/api/activities/{id}/withdraw`

撤回活动到草稿状态，同时删除所有相关申请。在提交后、被拒绝后、已通过后都可以撤回。

**权限**: 活动创建者

**响应示例**:

```json
{
  "code": 0,
  "message": "活动撤回成功",
  "data": {
    "id": "uuid",
    "status": "draft",
    "withdrawn_at": "2024-01-01T13:00:00Z"
  }
}
```

### 1.12 审核活动

**POST** `/api/activities/{id}/review`

审核活动。可以审核待审核状态的活动，也可以修改已通过或已拒绝状态的活动。

**权限**: 教师、管理员

**请求体**:

```json
{
  "status": "approved",
  "review_comments": "活动内容符合要求，同意通过"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "审核完成",
  "data": {
    "id": "uuid",
    "status": "approved",
    "reviewer_id": "teacher-uuid",
    "review_comments": "活动内容符合要求，同意通过",
    "reviewed_at": "2024-01-01T14:00:00Z"
  }
}
```

**说明**:

- 可以审核待审核状态的活动
- 可以修改已通过状态的活动为拒绝状态
- 可以修改已拒绝状态的活动为通过状态
- 每次修改都会更新审核人、审核评语和审核时间

### 1.13 获取待审核活动

**GET** `/api/activities/pending`

获取所有待审核的活动列表。

**权限**: 教师、管理员

**查询参数**:

- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 10

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": "uuid",
        "title": "创新创业实践活动",
        "description": "参与创新创业项目，提升创新能力和实践技能",
        "status": "pending_review",
        "category": "创新创业",
        "owner_id": "user-uuid",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
}
```

### 1.14 批量删除活动

**POST** `/api/activities/batch-delete`

批量删除多个活动。

**权限**: 教师、管理员

**请求体**:

```json
{
  "activity_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "批量删除活动成功",
  "data": {
    "deleted_count": 3,
    "total_count": 3,
    "deleted_at": "2024-01-01T15:00:00Z"
  }
}
```

### 1.15 获取可删除的活动列表

**GET** `/api/activities/deletable`

获取用户可删除的活动列表。

**权限**: 所有认证用户

**响应示例**:

```json
{
  "code": 0,
  "message": "获取可删除活动列表成功",
  "data": {
    "activities": [
      {
        "activity_id": "uuid",
        "title": "创新创业实践活动",
        "description": "参与创新创业项目，提升创新能力和实践技能",
        "status": "draft",
        "category": "创新创业",
        "owner_id": "user-uuid",
        "created_at": "2024-01-01T00:00:00Z",
        "can_delete": true
      }
    ],
    "total": 1
  }
}
```

### 1.16 获取活动统计

**GET** `/api/activities/stats`

获取活动统计信息。

**权限**: 所有认证用户

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_activities": 10,
    "draft_count": 3,
    "pending_count": 2,
    "approved_count": 4,
    "rejected_count": 1,
    "total_participants": 25,
    "total_credits": 50.0
  }
}
```

## 2. 参与者管理 API

### 2.1 添加参与者

**POST** `/api/activities/{id}/participants`

为活动添加参与者。只能添加学生用户。

**权限**: 活动创建者、管理员

**请求体**:

```json
{
  "user_ids": ["student-uuid1", "student-uuid2"],
  "credits": 2.0
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "参与者添加成功",
  "data": {
    "added_count": 2,
    "total_count": 2,
    "participants": [
      {
        "user_id": "student-uuid1",
        "credits": 2.0,
        "joined_at": "2024-01-01T16:00:00Z"
      },
      {
        "user_id": "student-uuid2",
        "credits": 2.0,
        "joined_at": "2024-01-01T16:00:00Z"
      }
    ]
  }
}
```

### 2.2 批量设置学分

**PUT** `/api/activities/{id}/participants/batch-credits`

批量设置参与者的学分。

**权限**: 活动创建者、管理员

**请求体**:

```json
{
  "credits_map": {
    "student-uuid1": 2.0,
    "student-uuid2": 1.5,
    "student-uuid3": 3.0
  }
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "批量设置学分成功",
  "data": {
    "updated_count": 3,
    "credits_map": {
      "student-uuid1": 2.0,
      "student-uuid2": 1.5,
      "student-uuid3": 3.0
    }
  }
}
```

### 2.3 设置单个学分

**PUT** `/api/activities/{id}/participants/{user_id}/credits`

设置单个参与者的学分。

**权限**: 活动创建者、管理员

**请求体**:

```json
{
  "credits": 2.5
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "学分设置成功",
  "data": {
    "user_id": "student-uuid",
    "credits": 2.5,
    "updated_at": "2024-01-01T17:00:00Z"
  }
}
```

### 2.4 删除参与者

**DELETE** `/api/activities/{id}/participants/{user_id}`

从活动中删除参与者。

**权限**: 活动创建者、管理员

**响应示例**:

```json
{
  "code": 0,
  "message": "参与者删除成功",
  "data": {
    "user_id": "student-uuid",
    "removed_at": "2024-01-01T18:00:00Z"
  }
}
```

### 2.5 获取参与者列表

**GET** `/api/activities/{id}/participants`

获取活动的参与者列表。

**权限**: 活动创建者、管理员

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "participants": [
      {
        "user_id": "student-uuid",
        "credits": 2.0,
        "joined_at": "2024-01-01T16:00:00Z",
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

### 2.6 退出活动

**POST** `/api/activities/{id}/leave`

学生退出活动。

**权限**: 活动参与者（学生）

**响应示例**:

```json
{
  "code": 0,
  "message": "退出活动成功",
  "data": {
    "user_id": "student-uuid",
    "left_at": "2024-01-01T19:00:00Z"
  }
}
```

## 3. 申请管理 API

### 3.1 获取用户申请列表

**GET** `/api/applications`

获取当前用户的申请列表。

**权限**: 所有认证用户

**查询参数**:

- `status` (可选): 申请状态筛选
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 10

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": "uuid",
        "activity_id": "activity-uuid",
        "user_id": "student-uuid",
        "status": "approved",
        "applied_credits": 2.0,
        "awarded_credits": 2.0,
        "submitted_at": "2024-01-01T20:00:00Z",
        "created_at": "2024-01-01T20:00:00Z",
        "updated_at": "2024-01-01T20:00:00Z",
        "activity": {
          "id": "activity-uuid",
          "title": "创新创业实践活动",
          "description": "参与创新创业项目，提升创新能力和实践技能",
          "category": "创新创业",
          "start_date": "2024-12-01T00:00:00Z",
          "end_date": "2024-12-31T00:00:00Z"
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

### 3.2 获取申请详情

**GET** `/api/applications/{id}`

获取申请详情。

**权限**: 所有认证用户（学生只能查看自己的申请）

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "activity_id": "activity-uuid",
    "user_id": "student-uuid",
    "status": "approved",
    "applied_credits": 2.0,
    "awarded_credits": 2.0,
    "submitted_at": "2024-01-01T20:00:00Z",
    "created_at": "2024-01-01T20:00:00Z",
    "updated_at": "2024-01-01T20:00:00Z",
    "activity": {
      "id": "activity-uuid",
      "title": "创新创业实践活动",
      "description": "参与创新创业项目，提升创新能力和实践技能",
      "category": "创新创业",
      "start_date": "2024-12-01T00:00:00Z",
      "end_date": "2024-12-31T00:00:00Z"
    }
  }
}
```

### 3.3 获取所有申请

**GET** `/api/applications/all`

获取所有申请列表（教师/管理员权限）。

**权限**: 教师、管理员

**查询参数**:

- `activity_id` (可选): 活动 ID 筛选
- `user_id` (可选): 用户 ID 筛选
- `page` (可选): 页码，默认 1
- `limit` (可选): 每页数量，默认 10

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": "uuid",
        "activity_id": "activity-uuid",
        "user_id": "student-uuid",
        "status": "approved",
        "applied_credits": 2.0,
        "awarded_credits": 2.0,
        "submitted_at": "2024-01-01T20:00:00Z",
        "created_at": "2024-01-01T20:00:00Z",
        "updated_at": "2024-01-01T20:00:00Z",
        "activity": {
          "id": "activity-uuid",
          "title": "创新创业实践活动",
          "description": "参与创新创业项目，提升创新能力和实践技能",
          "category": "创新创业",
          "start_date": "2024-12-01T00:00:00Z",
          "end_date": "2024-12-31T00:00:00Z"
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

### 3.4 导出申请数据

**GET** `/api/applications/export`

导出申请数据为 CSV 格式。

**权限**: 所有认证用户（学生只能导出自己的申请，教师/管理员可以导出所有申请）

**查询参数**:

- `format` (可选): 导出格式，支持 "csv", "excel"，默认 "csv"
- `activity_id` (可选): 活动 ID 筛选
- `status` (可选): 申请状态筛选
- `start_date` (可选): 开始日期筛选 (YYYY-MM-DD)
- `end_date` (可选): 结束日期筛选 (YYYY-MM-DD)

**响应**: 文件下载（CSV 或 Excel 格式）

### 3.5 获取申请统计

**GET** `/api/applications/stats`

获取申请统计信息。

**权限**: 所有认证用户

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_applications": 10,
    "total_credits": 20.0,
    "awarded_credits": 16.0
  }
}
```

## 4. 附件管理 API

### 4.1 获取附件列表

**GET** `/api/activities/{id}/attachments`

获取活动的附件列表。

**权限**: 所有认证用户

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "attachments": [
      {
        "id": "uuid",
        "activity_id": "activity-uuid",
        "file_name": "document.pdf",
        "original_name": "项目计划书.pdf",
        "file_size": 1024000,
        "file_type": ".pdf",
        "file_category": "document",
        "description": "项目计划书",
        "uploaded_by": "user-uuid",
        "uploaded_at": "2024-01-01T21:00:00Z",
        "download_count": 5,
        "download_url": "/api/activities/activity-uuid/attachments/uuid/download"
      }
    ],
    "total": 1
  }
}
```

### 4.2 上传单个附件

**POST** `/api/activities/{id}/attachments`

上传单个附件。

**权限**: 活动创建者、参与者、管理员

**请求体**: `multipart/form-data`

- `file`: 文件
- `description` (可选): 文件描述

**支持的文件类型**:

- 文档: .pdf, .doc, .docx, .txt, .rtf, .odt
- 图片: .jpg, .jpeg, .png, .gif, .bmp, .webp
- 视频: .mp4, .avi, .mov, .wmv, .flv
- 音频: .mp3, .wav, .ogg, .aac
- 压缩包: .zip, .rar, .7z, .tar, .gz
- 表格: .xls, .xlsx, .csv
- 演示文稿: .ppt, .pptx

**文件大小限制**: 20MB

**响应示例**:

```json
{
  "code": 0,
  "message": "附件上传成功",
  "data": {
    "id": "uuid",
    "activity_id": "activity-uuid",
    "file_name": "document.pdf",
    "original_name": "项目计划书.pdf",
    "file_size": 1024000,
    "file_type": ".pdf",
    "file_category": "document",
    "description": "项目计划书",
    "uploaded_by": "user-uuid",
    "uploaded_at": "2024-01-01T21:00:00Z",
    "download_count": 0,
    "download_url": "/api/activities/activity-uuid/attachments/uuid/download"
  }
}
```

### 4.3 批量上传附件

**POST** `/api/activities/{id}/attachments/batch`

批量上传多个附件（最多 10 个）。

**权限**: 活动创建者、参与者、管理员

**请求体**: `multipart/form-data`

- `files`: 文件数组

**响应示例**:

```json
{
  "code": 0,
  "message": "批量上传完成",
  "data": {
    "total_files": 3,
    "success_count": 2,
    "fail_count": 1,
    "results": [
      {
        "file_name": "document1.pdf",
        "status": "success",
        "file_id": "uuid1",
        "file_size": 1024000
      },
      {
        "file_name": "document2.pdf",
        "status": "success",
        "file_id": "uuid2",
        "file_size": 2048000
      },
      {
        "file_name": "invalid.txt",
        "status": "failed",
        "message": "不支持的文件类型"
      }
    ]
  }
}
```

### 4.4 下载附件

**GET** `/api/activities/{id}/attachments/{attachment_id}/download`

下载附件。

**权限**: 所有认证用户

**响应**: 文件下载

### 4.5 更新附件信息

**PUT** `/api/activities/{id}/attachments/{attachment_id}`

更新附件描述信息。

**权限**: 上传者、管理员

**请求体**:

```json
{
  "description": "更新后的文件描述"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "附件信息更新成功",
  "data": {
    "id": "uuid",
    "description": "更新后的文件描述",
    "updated_at": "2024-01-01T22:00:00Z"
  }
}
```

### 4.6 删除附件

**DELETE** `/api/activities/{id}/attachments/{attachment_id}`

删除附件。

**权限**: 上传者、管理员

**响应示例**:

```json
{
  "code": 0,
  "message": "附件删除成功",
  "data": {
    "id": "uuid",
    "deleted_at": "2024-01-01T23:00:00Z"
  }
}
```

## 5. 健康检查

### 5.1 健康检查

**GET** `/health`

检查服务健康状态。

**响应示例**:

```json
{
  "status": "ok",
  "service": "credit-activity-service"
}
```

## 6. 错误处理

### 6.1 常见错误码

| 错误码 | 说明           | 解决方案                      |
| ------ | -------------- | ----------------------------- |
| 400    | 请求参数错误   | 检查请求参数格式和必填字段    |
| 401    | 未认证         | 检查 Authorization 头是否正确 |
| 403    | 权限不足       | 检查用户角色和权限            |
| 404    | 资源不存在     | 检查资源 ID 是否正确          |
| 500    | 服务器内部错误 | 联系管理员                    |

### 6.2 错误响应示例

```json
{
  "code": 400,
  "message": "请求参数错误: 活动标题不能为空",
  "data": null
}
```

## 7. 状态说明

### 7.1 活动状态

- `draft`: 草稿状态，可以修改
- `pending_review`: 待审核状态，等待教师审核
- `approved`: 已通过，自动生成申请
- `rejected`: 已拒绝，可以修改后重新提交

### 7.2 申请状态

- `approved`: 已通过（固定状态，自动生成）

## 8. 权限说明

### 8.1 角色权限

| 功能           | 学生 | 教师 | 管理员 |
| -------------- | ---- | ---- | ------ |
| 创建活动       | ✓    | ✓    | ✓      |
| 编辑自己的活动 | ✓    | ✓    | ✓      |
| 删除自己的活动 | ✓    | ✓    | ✓      |
| 提交活动审核   | ✓    | ✓    | ✓      |
| 撤回活动       | ✓    | ✓    | ✓      |
| 审核活动       | ✗    | ✓    | ✓      |
| 添加参与者     | ✓    | ✓    | ✓      |
| 删除参与者     | ✓    | ✓    | ✓      |
| 设置学分       | ✓    | ✓    | ✓      |
| 退出活动       | ✓    | ✗    | ✗      |
| 查看自己的申请 | ✓    | ✓    | ✓      |
| 导出自己的申请 | ✓    | ✓    | ✓      |
| 查看所有申请   | ✗    | ✓    | ✓      |
| 导出所有申请   | ✗    | ✓    | ✓      |

### 8.2 特殊权限说明

1. **活动创建者权限**：

   - 可以编辑、删除、提交、撤回自己创建的活动
   - 可以管理自己活动的参与者
   - 不一定是活动的参与者

2. **教师权限**：

   - 可以创建和管理自己的活动
   - 可以审核所有活动
   - 可以查看和导出所有申请

3. **参与者限制**：
   - 只有学生可以参与活动
   - 教师和管理员不能参与活动

## 9. 使用示例

### 9.1 完整活动创建流程

1. **获取活动类别**

```bash
curl -X GET "http://localhost:8083/api/activities/categories"
```

2. **创建活动**

```bash
curl -X POST "http://localhost:8083/api/activities" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "创新创业实践活动",
    "description": "参与创新创业项目，提升创新能力和实践技能",
    "start_date": "2024-12-01",
    "end_date": "2024-12-31",
    "category": "创新创业",
    "requirements": "需要提交项目计划书和成果展示"
  }'
```

3. **添加参与者**

```bash
curl -X POST "http://localhost:8083/api/activities/activity-uuid/participants" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": ["student-uuid1", "student-uuid2"],
    "credits": 2.0
  }'
```

4. **提交审核**

```bash
curl -X POST "http://localhost:8083/api/activities/activity-uuid/submit" \
  -H "Authorization: Bearer your-token"
```

5. **审核活动（教师）**

```bash
curl -X POST "http://localhost:8083/api/activities/activity-uuid/review" \
  -H "Authorization: Bearer teacher-token" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "review_comments": "活动内容符合要求，同意通过"
  }'
```

### 9.2 申请查看和导出

1. **查看申请列表**

```bash
curl -X GET "http://localhost:8083/api/applications" \
  -H "Authorization: Bearer your-token"
```

2. **导出申请数据**

```bash
curl -X GET "http://localhost:8083/api/applications/export?format=csv" \
  -H "Authorization: Bearer your-token"
```

## 10. 统一搜索 API

### 10.1 统一活动搜索

**GET** `/api/search/activities`

提供统一的活动搜索功能，支持多条件筛选和分页。

**权限**: 所有认证用户（学生只能看到自己创建或参与的活动）

**查询参数**:

- `query` (可选): 搜索关键词，支持标题、描述、类别、要求的模糊搜索
- `category` (可选): 活动类别筛选
- `status` (可选): 活动状态筛选 (draft, pending_review, approved, rejected)
- `owner_id` (可选): 创建者 ID 筛选
- `start_date` (可选): 开始日期筛选 (YYYY-MM-DD)
- `end_date` (可选): 结束日期筛选 (YYYY-MM-DD)
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 10，最大 100
- `sort_by` (可选): 排序字段，默认 created_at
- `sort_order` (可选): 排序方向，默认 desc

**响应示例**:

```json
{
  "code": 0,
  "message": "搜索成功",
  "data": {
    "data": [
      {
        "id": "uuid",
        "title": "创新创业实践活动",
        "description": "参与创新创业项目，提升创新能力和实践技能",
        "start_date": "2024-12-01T00:00:00Z",
        "end_date": "2024-12-31T00:00:00Z",
        "status": "approved",
        "category": "创新创业",
        "requirements": "需要提交项目计划书和成果展示",
        "owner_id": "user-uuid",
        "reviewer_id": "teacher-uuid",
        "review_comments": "活动内容符合要求",
        "reviewed_at": "2024-01-01T10:00:00Z",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "participants": [],
        "applications": []
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1,
    "filters": {
      "query": "创新创业",
      "category": "创新创业",
      "status": "approved",
      "page": 1,
      "page_size": 10,
      "sort_by": "created_at",
      "sort_order": "desc"
    }
  }
}
```

### 10.2 统一申请搜索

**GET** `/api/search/applications`

提供统一的申请搜索功能，支持多条件筛选和分页。

**权限**: 所有认证用户（学生只能看到自己的申请）

**查询参数**:

- `query` (可选): 搜索关键词，通过活动信息搜索
- `activity_id` (可选): 活动 ID 筛选
- `user_id` (可选): 用户 ID 筛选
- `status` (可选): 申请状态筛选
- `start_date` (可选): 开始日期筛选 (YYYY-MM-DD)
- `end_date` (可选): 结束日期筛选 (YYYY-MM-DD)
- `min_credits` (可选): 最小学分筛选
- `max_credits` (可选): 最大学分筛选
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 10，最大 100
- `sort_by` (可选): 排序字段，默认 submitted_at
- `sort_order` (可选): 排序方向，默认 desc

**响应示例**:

```json
{
  "code": 0,
  "message": "搜索成功",
  "data": {
    "data": [
      {
        "id": "uuid",
        "activity_id": "activity-uuid",
        "user_id": "student-uuid",
        "status": "approved",
        "applied_credits": 2.0,
        "awarded_credits": 2.0,
        "submitted_at": "2024-01-01T10:00:00Z",
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z",
        "activity": {
          "id": "activity-uuid",
          "title": "创新创业实践活动",
          "description": "参与创新创业项目，提升创新能力和实践技能",
          "category": "创新创业",
          "start_date": "2024-12-01T00:00:00Z",
          "end_date": "2024-12-31T00:00:00Z"
        }
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1,
    "filters": {
      "query": "创新创业",
      "status": "approved",
      "page": 1,
      "page_size": 10,
      "sort_by": "submitted_at",
      "sort_order": "desc"
    }
  }
}
```

### 10.3 统一参与者搜索

**GET** `/api/search/participants`

提供统一的参与者搜索功能，支持多条件筛选和分页。

**权限**: 所有认证用户（学生只能看到自己参与的活动）

**查询参数**:

- `activity_id` (可选): 活动 ID 筛选
- `user_id` (可选): 用户 ID 筛选
- `min_credits` (可选): 最小学分筛选
- `max_credits` (可选): 最大学分筛选
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 10，最大 100
- `sort_by` (可选): 排序字段，默认 joined_at
- `sort_order` (可选): 排序方向，默认 desc

**响应示例**:

```json
{
  "code": 0,
  "message": "搜索成功",
  "data": {
    "data": [
      {
        "user_id": "student-uuid",
        "credits": 2.0,
        "joined_at": "2024-01-01T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1,
    "filters": {
      "activity_id": "activity-uuid",
      "page": 1,
      "page_size": 10,
      "sort_by": "joined_at",
      "sort_order": "desc"
    }
  }
}
```

### 10.4 统一附件搜索

**GET** `/api/search/attachments`

提供统一的附件搜索功能，支持多条件筛选和分页。

**权限**: 所有认证用户（学生只能看到自己创建或参与活动的附件）

**查询参数**:

- `query` (可选): 搜索关键词，支持文件名、原始文件名、描述的模糊搜索
- `activity_id` (可选): 活动 ID 筛选
- `uploader_id` (可选): 上传者 ID 筛选
- `file_type` (可选): 文件类型筛选
- `file_category` (可选): 文件分类筛选
- `min_size` (可选): 最小文件大小筛选（字节）
- `max_size` (可选): 最大文件大小筛选（字节）
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 10，最大 100
- `sort_by` (可选): 排序字段，默认 uploaded_at
- `sort_order` (可选): 排序方向，默认 desc

**响应示例**:

```json
{
  "code": 0,
  "message": "搜索成功",
  "data": {
    "data": [
      {
        "id": "uuid",
        "activity_id": "activity-uuid",
        "file_name": "project_plan.pdf",
        "original_name": "项目计划书.pdf",
        "file_size": 1024000,
        "file_type": "application/pdf",
        "file_category": "document",
        "description": "项目计划书",
        "uploaded_by": "student-uuid",
        "uploaded_at": "2024-01-01T10:00:00Z",
        "download_count": 5
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1,
    "filters": {
      "query": "项目计划书",
      "file_type": "application/pdf",
      "page": 1,
      "page_size": 10,
      "sort_by": "uploaded_at",
      "sort_order": "desc"
    }
  }
}
```

### 10.5 搜索功能说明

**搜索特性**:

1. **模糊搜索**: 支持关键词的模糊匹配
2. **多条件筛选**: 支持多个条件的组合筛选
3. **分页支持**: 支持分页查询，避免大量数据返回
4. **排序功能**: 支持多种字段的升序/降序排序
5. **权限控制**: 根据用户角色自动过滤数据

**性能优化**:

1. **索引优化**: 对常用搜索字段建立数据库索引
2. **查询优化**: 使用高效的 SQL 查询语句
3. **缓存支持**: 对搜索结果进行适当缓存
4. **分页限制**: 限制每页最大返回数量为 100

**使用建议**:

1. 合理使用搜索条件，避免过于复杂的查询
2. 使用分页功能处理大量数据
3. 根据实际需求选择合适的排序字段
4. 注意权限限制，确保只能访问有权限的数据

---

**版本**: 1.0.0  
**更新时间**: 2024-01-01  
**维护者**: 学分活动服务团队
