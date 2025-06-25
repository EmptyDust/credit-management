# 申请管理服务 API 文档

## 概述

申请管理服务负责处理学生与事务之间的申请关系。申请本质上是一个关联关系，大部分信息（如事务标题、描述、附件等）都从事务服务继承。

### 核心概念

- **申请（Application）**: 学生和事务之间的关联关系
- **申请状态**: pending（待审核）、approved（已通过）、rejected（已拒绝）、completed（已完成）、cancelled（已取消）
- **申请学分**: 学生申请获得的学分，不能超过事务学分
- **审核流程**: 教师或管理员可以审核申请，设置状态和审核意见

### 权限控制

- **学生**: 只能查看和管理自己的申请
- **教师**: 可以查看所有申请，审核申请
- **管理员**: 可以查看所有申请，审核申请，批量创建申请

## API 端点

### 1. 创建申请

**POST** `/applications`

学生创建对某个事务的申请。

**请求体:**
```json
{
  "affair_id": "uuid",
  "applied_credits": 2.0
}
```

**响应:**
```json
{
  "id": "uuid",
  "affair_id": "uuid",
  "user_id": "uuid",
  "status": "pending",
  "applied_credits": 2.0,
  "submitted_at": "2024-01-01T00:00:00Z",
  "reviewed_at": null,
  "reviewer_id": null,
  "review_comments": "",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "affair_title": "事务标题",
  "affair_description": "事务描述",
  "affair_category": "学术活动",
  "affair_credit_value": 3.0,
  "user_name": "学生姓名",
  "student_id": "2021001"
}
```

**错误响应:**
- `400 Bad Request`: 参数错误、事务不存在、申请学分超过事务学分
- `409 Conflict`: 已经申请过该事务
- `401 Unauthorized`: 用户未认证

### 2. 获取用户申请列表

**GET** `/applications`

获取当前用户的申请列表。

**查询参数:**
- `status` (可选): 按状态筛选 (pending, approved, rejected, completed, cancelled)
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10

**响应:**
```json
{
  "applications": [
    {
      "id": "uuid",
      "affair_id": "uuid",
      "user_id": "uuid",
      "status": "pending",
      "applied_credits": 2.0,
      "submitted_at": "2024-01-01T00:00:00Z",
      "reviewed_at": null,
      "reviewer_id": null,
      "review_comments": "",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "affair_title": "事务标题",
      "affair_description": "事务描述",
      "affair_category": "学术活动",
      "affair_credit_value": 3.0,
      "user_name": "学生姓名",
      "student_id": "2021001"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1
}
```

### 3. 获取申请统计

**GET** `/applications/stats`

获取当前用户的申请统计信息。

**响应:**
```json
{
  "user_id": "uuid",
  "total_applications": 10,
  "pending_count": 3,
  "approved_count": 5,
  "rejected_count": 1,
  "completed_count": 1,
  "total_credits": 20.0,
  "awarded_credits": 15.0
}
```

### 4. 获取单个申请

**GET** `/applications/{id}`

获取指定申请的详细信息。

**响应:**
```json
{
  "id": "uuid",
  "affair_id": "uuid",
  "user_id": "uuid",
  "status": "pending",
  "applied_credits": 2.0,
  "submitted_at": "2024-01-01T00:00:00Z",
  "reviewed_at": null,
  "reviewer_id": null,
  "review_comments": "",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "affair_title": "事务标题",
  "affair_description": "事务描述",
  "affair_category": "学术活动",
  "affair_credit_value": 3.0,
  "user_name": "学生姓名",
  "student_id": "2021001"
}
```

**错误响应:**
- `404 Not Found`: 申请不存在
- `403 Forbidden`: 无权限查看此申请

### 5. 更新申请

**PUT** `/applications/{id}`

更新申请信息。学生只能更新申请学分，教师/管理员可以审核申请。

**请求体:**
```json
{
  "applied_credits": 2.5,
  "status": "approved",
  "review_comments": "审核通过"
}
```

**权限说明:**
- 学生: 只能更新 `applied_credits`，且只能更新待审核状态的申请
- 教师/管理员: 可以更新所有字段，包括审核状态

**响应:** 同获取单个申请

**错误响应:**
- `400 Bad Request`: 参数错误、申请学分超过事务学分
- `403 Forbidden`: 无权限更新此申请
- `404 Not Found`: 申请不存在

### 6. 删除申请

**DELETE** `/applications/{id}`

删除申请。只能删除自己的申请，且状态为待审核。

**响应:**
```json
{
  "message": "申请删除成功"
}
```

**错误响应:**
- `403 Forbidden`: 无权限删除此申请
- `400 Bad Request`: 只能删除待审核状态的申请
- `404 Not Found`: 申请不存在

### 7. 批量创建申请

**POST** `/applications/batch`

教师或管理员批量为学生创建申请。

**请求体:**
```json
{
  "affair_id": "uuid",
  "user_ids": ["uuid1", "uuid2", "uuid3"],
  "applied_credits": 2.0
}
```

**响应:**
```json
{
  "affair_id": "uuid",
  "total_users": 3,
  "created_count": 2,
  "failed_count": 1,
  "applications": [
    {
      "id": "uuid",
      "affair_id": "uuid",
      "user_id": "uuid1",
      "status": "pending",
      "applied_credits": 2.0,
      "submitted_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "errors": [
    {
      "user_id": "uuid2",
      "error": "已经申请过该事务"
    }
  ],
  "message": "成功创建 2 个申请，失败 1 个"
}
```

**错误响应:**
- `403 Forbidden`: 只有管理员和教师可以批量创建申请
- `400 Bad Request`: 参数错误、事务不存在、申请学分超过事务学分

### 8. 获取事务的申请列表

**GET** `/affairs/{affair_id}/applications`

获取指定事务的所有申请。管理员/教师可以查看所有，学生只能查看自己创建的或参与的事务。

**查询参数:**
- `status` (可选): 按状态筛选
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10

**响应:** 同获取用户申请列表

**错误响应:**
- `403 Forbidden`: 无权限查看此事务的申请

## 数据模型

### Application

```go
type Application struct {
    ID             string         `json:"id" gorm:"primaryKey;type:uuid"`
    AffairID       string         `json:"affair_id" gorm:"type:uuid;not null;index"`
    UserID         string         `json:"user_id" gorm:"type:uuid;not null;index"`
    Status         string         `json:"status" gorm:"default:'pending';index"`
    AppliedCredits float64        `json:"applied_credits" gorm:"not null"`
    SubmittedAt    time.Time      `json:"submitted_at" gorm:"default:CURRENT_TIMESTAMP"`
    ReviewedAt     *time.Time     `json:"reviewed_at"`
    ReviewerID     string         `json:"reviewer_id" gorm:"type:uuid"`
    ReviewComments string         `json:"review_comments"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
```

### ApplicationResponse

```go
type ApplicationResponse struct {
    ID                 string     `json:"id"`
    AffairID           string     `json:"affair_id"`
    UserID             string     `json:"user_id"`
    Status             string     `json:"status"`
    AppliedCredits     float64    `json:"applied_credits"`
    SubmittedAt        time.Time  `json:"submitted_at"`
    ReviewedAt         *time.Time `json:"reviewed_at"`
    ReviewerID         string     `json:"reviewer_id"`
    ReviewComments     string     `json:"review_comments"`
    CreatedAt          time.Time  `json:"created_at"`
    UpdatedAt          time.Time  `json:"updated_at"`
    
    // 从事务继承的信息
    AffairTitle        string  `json:"affair_title,omitempty"`
    AffairDescription  string  `json:"affair_description,omitempty"`
    AffairCategory     string  `json:"affair_category,omitempty"`
    AffairCreditValue  float64 `json:"affair_credit_value,omitempty"`
    UserName           string  `json:"user_name,omitempty"`
    StudentID          string  `json:"student_id,omitempty"`
}
```

## 状态码说明

- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 用户未认证
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突（如重复申请）
- `500 Internal Server Error`: 服务器内部错误

## 使用说明

### 申请流程

1. **学生申请**: 学生浏览事务列表，选择感兴趣的事务进行申请
2. **教师审核**: 教师或管理员查看待审核的申请，进行审核
3. **状态更新**: 申请状态根据审核结果更新
4. **学分记录**: 通过的申请计入学生的学分统计

### 权限控制

- 学生只能管理自己的申请
- 教师和管理员可以查看和审核所有申请
- 只有管理员和教师可以批量创建申请

### 数据一致性

- 申请学分不能超过事务学分
- 一个学生只能对同一个事务申请一次
- 申请状态变更会记录审核时间和审核人

### 附件处理

申请的所有附件都与关联的事务附件相同，通过事务服务获取附件信息。 