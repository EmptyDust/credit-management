# 事务管理服务 API 文档

## 概述
事务管理服务负责处理学分事务的创建、查询、更新、删除等操作，以及参与者管理和附件管理。

## 权限说明

### 用户角色
- **admin**: 管理员，可以访问所有功能，查看所有事务和申请
- **teacher**: 教师，可以创建、管理事务，查看所有事务和申请，审核事务
- **student**: 学生，可以创建事务，查看自己参与的事务，只能查看自己的申请

### 数据访问权限
- **管理员**: 可以查看所有事务和申请，可以创建、更新、删除事务，审核事务
- **教师**: 可以查看所有事务和申请，可以创建、更新事务，审核事务
- **学生**: 
  - 可以创建事务
  - 只能查看自己参与的事务
  - 只能查看自己的申请
  - 可以加入/退出事务
  - 可以修改自己参与的事务内容
  - 不能删除事务

### 事务状态
- `pending`: 待审核
- `approved`: 已通过
- `rejected`: 已驳回
- `active`: 进行中
- `completed`: 已完成
- `cancelled`: 已取消

## 认证
所有API都需要在请求头中包含有效的JWT token：
```
Authorization: Bearer <token>
```

## API 端点

### 1. 获取所有事务
**GET** `/api/affairs`

**权限要求**: 所有认证用户

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "uuid",
      "title": "学术讲座",
      "description": "计算机科学前沿讲座",
      "credit_value": 2.0,
      "max_participants": 100,
      "current_participants": 50,
      "status": "active",
      "owner_id": "user-uuid",
      "reviewer_id": "teacher-uuid",
      "review_comments": "审核通过"
    }
  ]
}
```

### 2. 获取指定事务
**GET** `/api/affairs/{id}`

**权限要求**: 所有认证用户（学生只能查看自己参与的事务）

**路径参数**:
- `id`: 事务UUID

### 3. 创建事务
**POST** `/api/affairs`

**权限要求**: 所有认证用户

**请求体**:
```json
{
  "title": "学术讲座",
  "description": "计算机科学前沿讲座",
  "credit_value": 2.0,
  "max_participants": 100,
  "start_date": "2024-01-01T09:00:00Z",
  "end_date": "2024-01-01T11:00:00Z",
  "category": "学术活动",
  "requirements": "计算机专业学生优先",
  "participants": ["user-uuid-1", "user-uuid-2"]
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "学术讲座",
    "status": "pending",
    "owner_id": "user-uuid"
  }
}
```

### 4. 更新事务
**PUT** `/api/affairs/{id}`

**权限要求**: 管理员、事务所有者、事务参与者

**路径参数**:
- `id`: 事务UUID

**说明**: 事务参与者可以修改事务内容，但只有管理员和教师可以修改状态

### 5. 删除事务
**DELETE** `/api/affairs/{id}`

**权限要求**: 仅管理员和事务所有者

**路径参数**:
- `id`: 事务UUID

### 6. 获取事务参与者
**GET** `/api/affairs/{id}/participants`

**权限要求**: 所有认证用户

### 7. 更新事务参与者
**PUT** `/api/affairs/{id}/participants`

**权限要求**: 仅教师和管理员

**请求体**:
```json
{
  "user_ids": ["user-uuid-1", "user-uuid-2", "user-uuid-3"]
}
```

### 8. 获取事务申请
**GET** `/api/affairs/{id}/applications`

**权限要求**: 所有认证用户（学生只能查看自己的申请）

### 9. 批量创建申请
**POST** `/api/affairs/{id}/applications`

**权限要求**: 仅教师和管理员

### 10. 同步申请信息
**POST** `/api/affairs/{id}/sync-applications`

**权限要求**: 仅教师和管理员

### 11. 加入事务
**POST** `/api/affairs/{id}/join`

**权限要求**: 仅学生

**说明**: 只能加入已审核通过的事务

### 12. 退出事务
**POST** `/api/affairs/{id}/leave`

**权限要求**: 仅学生

**说明**: 只能退出已参与的事务

### 13. 获取待审核事务
**GET** `/api/affairs/pending`

**权限要求**: 仅教师和管理员

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "uuid",
      "title": "待审核事务",
      "status": "pending",
      "owner_id": "user-uuid",
      "created_at": "2024-01-01T09:00:00Z"
    }
  ]
}
```

### 14. 审核事务
**POST** `/api/affairs/{id}/review`

**权限要求**: 仅教师和管理员

**请求体**:
```json
{
  "status": "approved",
  "review_comments": "审核通过，活动内容充实"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "affair_id": "uuid",
    "status": "approved",
    "reviewer_id": "teacher-uuid",
    "review_comments": "审核通过，活动内容充实",
    "reviewed_at": "2024-01-01T10:00:00Z"
  }
}
```

### 15. 获取附件列表
**GET** `/api/affairs/{id}/attachments`

**权限要求**: 所有认证用户

**查询参数**:
- `category`: 文件类别（document, image, video, audio, archive, spreadsheet, presentation）
- `file_type`: 文件类型（如 .pdf, .doc）
- `uploaded_by`: 上传者用户ID

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "attachments": [
      {
        "id": "uuid",
        "affair_id": "affair-uuid",
        "file_name": "文档.pdf",
        "file_size": 1024000,
        "file_type": ".pdf",
        "file_category": "document",
        "description": "重要文档",
        "uploaded_by": "user-uuid",
        "uploaded_at": "2024-01-01T09:00:00Z",
        "download_count": 5
      }
    ],
    "stats": {
      "total_count": 10,
      "total_size": 10485760,
      "category_count": {
        "document": 5,
        "image": 3,
        "video": 2
      }
    }
  }
}
```

### 16. 上传附件
**POST** `/api/affairs/{id}/attachments`

**权限要求**: 事务参与同学和管理员

**请求**: `multipart/form-data`
- `file`: 文件内容
- `description`: 文件描述（可选）

**支持的文件类型**:
- 文档：.pdf, .doc, .docx, .txt, .rtf, .odt
- 图片：.jpg, .jpeg, .png, .gif, .bmp, .webp
- 视频：.mp4, .avi, .mov, .wmv, .flv
- 音频：.mp3, .wav, .ogg, .aac
- 压缩包：.zip, .rar, .7z, .tar, .gz
- 表格：.xls, .xlsx, .csv
- 演示文稿：.ppt, .pptx

**文件大小限制**: 20MB

### 17. 批量上传附件
**POST** `/api/affairs/{id}/attachments/batch`

**权限要求**: 事务参与同学和管理员

**请求**: `multipart/form-data`
- `files`: 多个文件内容

**限制**: 一次最多上传10个文件

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_files": 5,
    "success_count": 4,
    "fail_count": 1,
    "results": [
      {
        "file_name": "文档1.pdf",
        "status": "success",
        "file_id": "uuid",
        "file_size": 1024000
      },
      {
        "file_name": "文档2.doc",
        "status": "failed",
        "message": "不支持的文件类型"
      }
    ]
  }
}
```

### 18. 更新附件信息
**PUT** `/api/affairs/{id}/attachments/{fileId}`

**权限要求**: 上传者或管理员

**请求**: `application/x-www-form-urlencoded`
- `description`: 文件描述

### 19. 下载附件
**GET** `/api/affairs/{id}/attachments/{fileId}`

**权限要求**: 所有认证用户

**响应**: 文件内容流

### 20. 删除附件
**DELETE** `/api/affairs/{id}/attachments/{fileId}`

**权限要求**: 上传者或管理员

## 错误码
- `0`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 冲突（如已参与事务）
- `500`: 服务器内部错误

## 使用说明
1. **事务创建**: 所有认证用户都可以创建事务，创建后状态为"待审核"
2. **事务审核**: 只有教师和管理员可以审核事务，审核通过后学生才能加入
3. **事务参与**: 学生可以加入已审核通过的事务，也可以退出已参与的事务
4. **事务修改**: 事务参与者可以修改事务内容，但只有管理员和教师可以修改状态
5. **事务删除**: 只有管理员和事务所有者可以删除事务
6. **附件管理**: 只有事务参与同学和管理员可以上传附件
7. **权限控制**: 学生只能查看自己参与的事务和相关的申请
8. **文件去重**: 通过MD5哈希检测重复文件
9. **文件分类**: 自动识别文件类型并分类
10. **批量操作**: 支持批量上传多个文件
11. **统计信息**: 提供附件数量和大小统计
12. **下载统计**: 记录文件下载次数
13. **文件描述**: 支持为附件添加描述信息
14. **多种格式**: 支持文档、图片、视频、音频等多种文件类型

## 事务流程
1. 用户创建事务 → 状态：pending（待审核）
2. 教师/管理员审核 → 状态：approved（已通过）或 rejected（已驳回）
3. 学生加入事务 → 状态：active（进行中）
4. 事务完成 → 状态：completed（已完成）
5. 事务取消 → 状态：cancelled（已取消） 