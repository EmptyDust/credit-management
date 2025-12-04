# 学分活动服务 (Credit Activity Service)

这是学分管理系统的学分活动服务，负责学分活动、参与者、申请和附件的管理。

## 功能特性

- **活动管理** - 创建、编辑、删除、审核学分活动
- **参与者管理** - 添加、删除参与者，设置学分
- **申请管理** - 自动生成申请，查看和导出申请数据
- **附件管理** - 上传、下载、预览活动附件
- **搜索功能** - 高级搜索活动、申请、参与者和附件
- **批量操作** - 支持批量创建、更新、删除、导入导出

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Docker (可选，用于文件存储)

### 本地运行

```bash
# 安装依赖
go mod download

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=credit_management
export DB_SSLMODE=disable
export PORT=8083

# 创建必要的目录
mkdir -p uploads/attachments

# 运行服务
go run main.go
```

### Docker 运行

```bash
# 构建镜像
docker build -t credit-activity-service .

# 运行容器
docker run -d \
  --name credit-activity-service \
  -p 8083:8083 \
  -v $(pwd)/uploads:/app/uploads \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  credit-activity-service
```

## API 文档

详细的 API 文档请参考：[docs/credit-activity-service-design.md](../docs/credit-activity-service-design.md)

### 主要 API 端点

#### 活动管理

```http
GET    /api/activities                    # 获取活动列表
POST   /api/activities                    # 创建活动
GET    /api/activities/{id}               # 获取活动详情
PUT    /api/activities/{id}               # 更新活动
DELETE /api/activities/{id}               # 删除活动
POST   /api/activities/{id}/submit        # 提交审核
POST   /api/activities/{id}/withdraw      # 撤回活动
POST   /api/activities/{id}/review        # 审核活动
```

#### 参与者管理

```http
GET    /api/activities/{id}/participants              # 获取参与者列表
POST   /api/activities/{id}/participants              # 添加参与者
PUT    /api/activities/{id}/participants/{uuid}/credits # 设置学分
DELETE /api/activities/{id}/participants/{uuid}       # 删除参与者
POST   /api/activities/{id}/leave                     # 退出活动（学生）
```

#### 申请管理

```http
GET    /api/applications                  # 获取申请列表
GET    /api/applications/{id}             # 获取申请详情
GET    /api/applications/stats            # 获取申请统计
GET    /api/applications/export           # 导出申请数据
```

#### 附件管理

```http
GET    /api/activities/{id}/attachments                    # 获取附件列表
POST   /api/activities/{id}/attachments                    # 上传附件
GET    /api/activities/{id}/attachments/{id}/download      # 下载附件
GET    /api/activities/{id}/attachments/{id}/preview       # 预览附件
DELETE /api/activities/{id}/attachments/{id}               # 删除附件
```

## 环境变量

| 变量名        | 说明            | 默认值              |
| ------------- | --------------- | ------------------- |
| `DB_HOST`     | 数据库主机      | `localhost`         |
| `DB_PORT`     | 数据库端口      | `5432`              |
| `DB_USER`     | 数据库用户名    | `postgres`          |
| `DB_PASSWORD` | 数据库密码      | `password`          |
| `DB_NAME`     | 数据库名称      | `credit_management` |
| `DB_SSLMODE`  | 数据库 SSL 模式 | `disable`           |
| `PORT`        | 服务端口        | `8083`              |

## 核心功能说明

### 活动状态流转

```
draft (草稿)
  ↓ 提交审核
pending_review (待审核)
  ↓ 审核通过/拒绝
approved (已通过) / rejected (已拒绝)
  ↓ 撤回
draft (草稿)
```

### 申请自动生成

当活动审核通过时，系统会自动：

1. 查询活动的所有参与者
2. 为每个参与者创建申请记录
3. 申请的学分继承参与者的学分设置
4. 申请状态固定为 `approved`

### 申请自动删除

当活动从已通过状态变为其他状态时，系统会自动软删除该活动的所有相关申请。

### 文件管理

- 附件存储在 `uploads/attachments/` 目录
- 支持多种文件格式（PDF、Word、Excel、图片等）
- 使用 MD5 哈希避免重复存储
- 活动删除时自动清理孤立文件

## 权限控制

服务内部实现了精细的权限控制：

- **活动创建者** - 可以管理自己的活动
- **教师/管理员** - 可以审核和管理所有活动
- **学生** - 可以创建和管理自己的活动，可以退出活动

详细权限说明请参考：[docs/PERMISSION_CONTROL_DIAGRAM.md](../docs/PERMISSION_CONTROL_DIAGRAM.md)

## 数据库依赖

服务依赖以下数据库表：

- `credit_activities`: 学分活动表
- `activity_participants`: 活动参与者表
- `applications`: 申请表
- `attachments`: 附件表
- `users`: 用户表（通过 User Service 查询）

## 健康检查

```http
GET /health
```

返回示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "service": "credit-activity-service"
  }
}
```

## 文件存储

附件文件存储在本地文件系统中：

- 存储路径: `uploads/attachments/`
- 文件命名: 使用 UUID 作为文件名
- 重复检测: 基于 MD5 哈希值

## 故障排除

### 常见问题

1. **数据库连接失败**

   - 检查数据库服务是否运行
   - 确认数据库连接配置正确
   - 检查数据库用户权限

2. **文件上传失败**

   - 检查 `uploads/attachments/` 目录权限
   - 确认磁盘空间充足
   - 检查文件大小限制

3. **申请未自动生成**
   - 检查活动是否已审核通过
   - 确认活动是否有参与者
   - 查看服务日志了解详细错误

## 开发指南

### 添加新的活动类型

1. 在 `config/activity_options.json` 中添加新类型配置
2. 更新前端选项获取逻辑
3. 根据需要在 `details` JSONB 字段中存储扩展信息

### 扩展附件功能

1. 在 `handlers/attachment.go` 中添加新功能
2. 更新附件模型（如需要）
3. 添加相应的 API 端点
