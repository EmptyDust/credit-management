# 学分活动服务统一 API 文档

## 概述

本文档描述了学分活动服务的完整 API 结构，包括活动管理、参与者管理、申请处理和附件管理功能。所有 API 都已统一并确保与 API 网关、服务实现和文档的一致性。

**基础 URL**: `http://localhost:8080` (通过 API 网关)
**服务 URL**: `http://localhost:8083` (直接访问服务)
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

## 完整 API 路由列表

### 1. 活动管理基础路由

| 方法   | 路径                                | 描述            | 权限               |
| ------ | ----------------------------------- | --------------- | ------------------ |
| GET    | `/api/activities/categories`        | 获取活动类别    | 无需认证           |
| GET    | `/api/activities/templates`         | 获取活动模板    | 无需认证           |
| POST   | `/api/activities`                   | 创建单个活动    | 所有认证用户       |
| POST   | `/api/activities/batch`             | 批量创建活动    | 所有认证用户       |
| GET    | `/api/activities`                   | 获取活动列表    | 所有认证用户       |
| GET    | `/api/activities/stats`             | 获取活动统计    | 所有认证用户       |
| GET    | `/api/activities/:id`               | 获取活动详情    | 所有认证用户       |
| PUT    | `/api/activities/:id`               | 更新活动        | 活动创建者、管理员 |
| DELETE | `/api/activities/:id`               | 删除活动        | 活动创建者、管理员 |
| POST   | `/api/activities/:id/submit`        | 提交活动审核    | 活动创建者         |
| POST   | `/api/activities/:id/withdraw`      | 撤回活动        | 活动创建者         |
| POST   | `/api/activities/:id/review`        | 审核活动        | 教师、管理员       |
| GET    | `/api/activities/pending`           | 获取待审核活动  | 教师、管理员       |
| POST   | `/api/activities/:id/copy`          | 复制活动        | 所有认证用户       |
| POST   | `/api/activities/:id/save-template` | 保存为模板      | 所有认证用户       |
| GET    | `/api/activities/search`            | 搜索活动        | 所有认证用户       |
| GET    | `/api/activities/deletable`         | 获取可删除活动  | 所有认证用户       |
| POST   | `/api/activities/batch-delete`      | 批量删除活动    | 管理员             |
| POST   | `/api/activities/import-csv`        | 从 CSV 导入活动 | 所有认证用户       |
| GET    | `/api/activities/csv-template`      | 获取 CSV 模板   | 所有认证用户       |
| GET    | `/api/activities/export`            | 导出活动数据    | 教师、管理员       |
| GET    | `/api/activities/report`            | 获取活动报表    | 教师、管理员       |

### 2. 活动参与者管理路由

| 方法   | 路径                                                | 描述               | 权限               |
| ------ | --------------------------------------------------- | ------------------ | ------------------ |
| POST   | `/api/activities/:id/participants`                  | 添加参与者         | 活动创建者、管理员 |
| GET    | `/api/activities/:id/participants`                  | 获取参与者列表     | 活动创建者、管理员 |
| PUT    | `/api/activities/:id/participants/batch-credits`    | 批量设置学分       | 活动创建者、管理员 |
| PUT    | `/api/activities/:id/participants/:user_id/credits` | 设置单个学分       | 活动创建者、管理员 |
| DELETE | `/api/activities/:id/participants/:user_id`         | 删除参与者         | 活动创建者、管理员 |
| POST   | `/api/activities/:id/participants/batch-remove`     | 批量删除参与者     | 活动创建者、管理员 |
| GET    | `/api/activities/:id/participants/search`           | 搜索参与者         | 活动创建者、管理员 |
| GET    | `/api/activities/:id/participants/stats`            | 获取参与者统计     | 活动创建者、管理员 |
| GET    | `/api/activities/:id/participants/export`           | 导出参与者名单     | 活动创建者、管理员 |
| POST   | `/api/activities/:id/leave`                         | 退出活动           | 学生参与者         |
| GET    | `/api/activities/:id/my-activities`                 | 获取用户参与的活动 | 所有认证用户       |

### 3. 活动附件管理路由

| 方法   | 路径                                                      | 描述         | 权限                       |
| ------ | --------------------------------------------------------- | ------------ | -------------------------- |
| GET    | `/api/activities/:id/attachments`                         | 获取附件列表 | 所有认证用户               |
| POST   | `/api/activities/:id/attachments`                         | 上传单个附件 | 活动创建者、参与者、管理员 |
| POST   | `/api/activities/:id/attachments/batch`                   | 批量上传附件 | 活动创建者、参与者、管理员 |
| GET    | `/api/activities/:id/attachments/:attachment_id/download` | 下载附件     | 所有认证用户               |
| PUT    | `/api/activities/:id/attachments/:attachment_id`          | 更新附件信息 | 上传者、管理员             |
| DELETE | `/api/activities/:id/attachments/:attachment_id`          | 删除附件     | 上传者、管理员             |

### 4. 申请管理路由

| 方法 | 路径                       | 描述             | 权限         |
| ---- | -------------------------- | ---------------- | ------------ |
| GET  | `/api/applications`        | 获取用户申请列表 | 所有认证用户 |
| GET  | `/api/applications/:id`    | 获取申请详情     | 所有认证用户 |
| GET  | `/api/applications/stats`  | 获取申请统计     | 所有认证用户 |
| GET  | `/api/applications/export` | 导出申请数据     | 所有认证用户 |
| GET  | `/api/applications/all`    | 获取所有申请     | 教师、管理员 |

### 5. 统一搜索路由

| 方法 | 路径                       | 描述           | 权限         |
| ---- | -------------------------- | -------------- | ------------ |
| GET  | `/api/search/activities`   | 统一活动搜索   | 所有认证用户 |
| GET  | `/api/search/applications` | 统一申请搜索   | 所有认证用户 |
| GET  | `/api/search/participants` | 统一参与者搜索 | 所有认证用户 |
| GET  | `/api/search/attachments`  | 统一附件搜索   | 所有认证用户 |

## 权限矩阵

### 角色权限对照表

| 功能           | 学生 | 教师 | 管理员 |
| -------------- | ---- | ---- | ------ |
| 创建活动       | ✓    | ✓    | ✓      |
| 批量创建活动   | ✓    | ✓    | ✓      |
| 编辑自己的活动 | ✓    | ✓    | ✓      |
| 批量更新活动   | ✓    | ✓    | ✓      |
| 删除自己的活动 | ✓    | ✓    | ✓      |
| 批量删除活动   | ✗    | ✗    | ✓      |
| 提交活动审核   | ✓    | ✓    | ✓      |
| 撤回活动       | ✓    | ✓    | ✓      |
| 审核活动       | ✗    | ✓    | ✓      |
| 添加参与者     | ✓    | ✓    | ✓      |
| 删除参与者     | ✓    | ✓    | ✓      |
| 批量删除参与者 | ✓    | ✓    | ✓      |
| 设置学分       | ✓    | ✓    | ✓      |
| 批量设置学分   | ✓    | ✓    | ✓      |
| 退出活动       | ✓    | ✗    | ✗      |
| 查看自己的申请 | ✓    | ✓    | ✓      |
| 导出自己的申请 | ✓    | ✓    | ✓      |
| 查看所有申请   | ✗    | ✓    | ✓      |
| 导出所有申请   | ✗    | ✓    | ✓      |
| 上传附件       | ✓    | ✓    | ✓      |
| 批量上传附件   | ✓    | ✓    | ✓      |
| 搜索活动       | ✓\*  | ✓    | ✓      |
| 搜索申请       | ✓\*  | ✓    | ✓      |
| 搜索参与者     | ✓\*  | ✓    | ✓      |
| 搜索附件       | ✓\*  | ✓    | ✓      |
| 导出活动数据   | ✗    | ✓    | ✓      |
| 获取活动报表   | ✗    | ✓    | ✓      |

\*学生只能搜索到自己创建或参与的内容

## 数据模型

### 活动状态

- `draft`: 草稿状态，可以修改
- `pending_review`: 待审核状态，等待教师审核
- `approved`: 已通过，自动生成申请
- `rejected`: 已拒绝，可以修改后重新提交

### 申请状态

- `approved`: 已通过（固定状态，自动生成）

### 活动类别

- `创新创业`: 创新创业活动
- `学科竞赛`: 学科竞赛活动
- `志愿服务`: 志愿服务活动
- `学术研究`: 学术研究活动
- `文体活动`: 文体活动

## 文件上传限制

- **单个文件大小**: 最大 20MB
- **批量上传**: 最多 10 个文件
- **支持的文件类型**:
  - 文档: `.pdf`, `.doc`, `.docx`, `.txt`, `.rtf`, `.odt`
  - 图片: `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`
  - 视频: `.mp4`, `.avi`, `.mov`, `.wmv`, `.flv`
  - 音频: `.mp3`, `.wav`, `.ogg`, `.aac`
  - 压缩包: `.zip`, `.rar`, `.7z`, `.tar`, `.gz`
  - 表格: `.xls`, `.xlsx`, `.csv`
  - 演示文稿: `.ppt`, `.pptx`

## 分页和搜索

### 分页参数

- `page`: 页码，默认 1
- `limit`: 每页数量，默认 10，最大 100

### 搜索功能

- 支持标题、描述、类别、要求的模糊搜索
- 支持多条件筛选（状态、类别、创建者等）
- 支持日期范围筛选

## 错误处理

### 常见错误码

| 错误码 | 说明           | 解决方案                      |
| ------ | -------------- | ----------------------------- |
| 400    | 请求参数错误   | 检查请求参数格式和必填字段    |
| 401    | 未认证         | 检查 Authorization 头是否正确 |
| 403    | 权限不足       | 检查用户角色和权限            |
| 404    | 资源不存在     | 检查资源 ID 是否正确          |
| 500    | 服务器内部错误 | 联系管理员                    |

### 错误响应示例

```json
{
  "code": 400,
  "message": "请求参数错误: 活动标题不能为空",
  "data": null
}
```

## 使用示例

### 完整活动创建流程

1. **获取活动类别**

```bash
curl -X GET "http://localhost:8080/api/activities/categories"
```

2. **创建活动**

```bash
curl -X POST "http://localhost:8080/api/activities" \
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
curl -X POST "http://localhost:8080/api/activities/activity-uuid/participants" \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": ["student-uuid1", "student-uuid2"],
    "credits": 2.0
  }'
```

4. **提交审核**

```bash
curl -X POST "http://localhost:8080/api/activities/activity-uuid/submit" \
  -H "Authorization: Bearer your-token"
```

5. **审核活动（教师）**

```bash
curl -X POST "http://localhost:8080/api/activities/activity-uuid/review" \
  -H "Authorization: Bearer teacher-token" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "review_comments": "活动内容符合要求，同意通过"
  }'
```

## 注意事项

1. **日期格式**: 支持多种日期格式，建议使用 ISO 8601 格式
2. **文件上传**: 单个文件最大 20MB，批量上传最多 10 个文件
3. **权限控制**: 严格按照角色权限进行访问控制
4. **数据验证**: 所有输入数据都会进行严格验证
5. **错误处理**: 所有错误都有明确的错误码和错误信息
6. **分页查询**: 默认每页 10 条记录，最大 100 条
7. **搜索功能**: 支持模糊搜索和多条件筛选
8. **API 网关**: 所有请求都通过 API 网关（端口 8080）路由到相应服务
9. **服务直连**: 也可以直接访问服务（端口 8083）进行调试

## 版本信息

- **API 版本**: 2.0.0
- **最后更新**: 2024-12-19
- **状态**: 生产就绪
- **兼容性**: 向后兼容
