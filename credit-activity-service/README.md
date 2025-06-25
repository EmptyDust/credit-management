# 学分活动服务

学分活动服务是一个完整的学分活动管理系统，提供活动创建、参与者管理、申请处理和附件管理等功能。

## 功能特性

- ✅ **活动管理**: 创建、编辑、删除、审核活动
- ✅ **参与者管理**: 添加参与者、设置学分、批量操作
- ✅ **申请处理**: 自动生成申请、查看申请、导出数据
- ✅ **附件管理**: 上传、下载、管理活动附件
- ✅ **权限控制**: 基于角色的细粒度权限控制
- ✅ **搜索功能**: 多字段模糊搜索和筛选
- ✅ **统计功能**: 活动和申请的统计信息
- ✅ **导出功能**: 支持CSV和Excel格式导出

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Docker (可选)

### 安装和运行

1. **克隆项目**
```bash
git clone <repository-url>
cd credit-management/credit-activity-service
```

2. **配置环境变量**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=credit_management
export PORT=8083
```

3. **运行服务**
```bash
go run main.go
```

### Docker 运行

```bash
docker-compose up credit-activity-service
```

## API 文档

详细的API文档请参考：[API_DOCUMENTATION.md](./API_DOCUMENTATION.md)

### 主要接口

#### 活动管理
- `GET /api/activities/categories` - 获取活动类别
- `GET /api/activities/templates` - 获取活动模板
- `POST /api/activities` - 创建活动
- `GET /api/activities` - 获取活动列表
- `GET /api/activities/{id}` - 获取活动详情
- `PUT /api/activities/{id}` - 更新活动
- `DELETE /api/activities/{id}` - 删除活动
- `POST /api/activities/{id}/submit` - 提交活动审核
- `POST /api/activities/{id}/withdraw` - 撤回活动
- `POST /api/activities/{id}/review` - 审核活动

#### 参与者管理
- `POST /api/activities/{id}/participants` - 添加参与者
- `PUT /api/activities/{id}/participants/batch-credits` - 批量设置学分
- `PUT /api/activities/{id}/participants/{user_id}/credits` - 设置单个学分
- `DELETE /api/activities/{id}/participants/{user_id}` - 删除参与者
- `GET /api/activities/{id}/participants` - 获取参与者列表
- `POST /api/activities/{id}/leave` - 退出活动

#### 申请管理
- `GET /api/applications` - 获取用户申请列表
- `GET /api/applications/{id}` - 获取申请详情
- `GET /api/applications/all` - 获取所有申请
- `GET /api/applications/export` - 导出申请数据
- `GET /api/applications/stats` - 获取申请统计

#### 附件管理
- `GET /api/activities/{id}/attachments` - 获取附件列表
- `POST /api/activities/{id}/attachments` - 上传单个附件
- `POST /api/activities/{id}/attachments/batch` - 批量上传附件
- `GET /api/activities/{id}/attachments/{attachment_id}/download` - 下载附件
- `PUT /api/activities/{id}/attachments/{attachment_id}` - 更新附件信息
- `DELETE /api/activities/{id}/attachments/{attachment_id}` - 删除附件

## 数据库设计

### 主要表结构

#### credit_activities (活动表)
```sql
CREATE TABLE credit_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    status TEXT DEFAULT 'draft',
    category TEXT,
    requirements TEXT,
    owner_id UUID NOT NULL,
    reviewer_id UUID,
    review_comments TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

#### activity_participants (参与者表)
```sql
CREATE TABLE activity_participants (
    activity_id UUID NOT NULL,
    user_id UUID NOT NULL,
    credits DECIMAL(5,2) NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (activity_id, user_id),
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE
);
```

#### applications (申请表)
```sql
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT DEFAULT 'approved',
    applied_credits DECIMAL(5,2) NOT NULL,
    awarded_credits DECIMAL(5,2) NOT NULL,
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE
);
```

#### attachments (附件表)
```sql
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    file_name TEXT NOT NULL,
    original_name TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    file_type TEXT NOT NULL,
    file_category TEXT NOT NULL,
    file_path TEXT NOT NULL,
    md5_hash TEXT NOT NULL,
    description TEXT,
    uploaded_by UUID NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    download_count BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE
);
```

## 业务流程

### 活动创建流程
1. 学生/教师创建活动（草稿状态）
2. 设置参与者和学分分配
3. 提交活动审核
4. 教师审核活动
5. 审核通过后自动生成申请

### 申请处理流程
1. 活动审核通过后自动为参与者生成申请
2. 学生可以查看自己的申请
3. 教师/管理员可以查看所有申请
4. 支持申请数据导出

## 权限控制

### 角色权限矩阵

| 功能 | 学生 | 教师 | 管理员 |
|------|------|------|--------|
| 创建活动 | ✓ | ✓ | ✓ |
| 编辑自己的活动 | ✓ | ✓ | ✓ |
| 删除自己的活动 | ✓ | ✓ | ✓ |
| 提交活动审核 | ✓ | ✓ | ✓ |
| 撤回活动 | ✓ | ✓ | ✓ |
| 审核活动 | ✗ | ✓ | ✓ |
| 添加参与者 | ✓ | ✓ | ✓ |
| 删除参与者 | ✓ | ✓ | ✓ |
| 设置学分 | ✓ | ✓ | ✓ |
| 退出活动 | ✓ | ✗ | ✗ |
| 查看自己的申请 | ✓ | ✓ | ✓ |
| 导出自己的申请 | ✓ | ✓ | ✓ |
| 查看所有申请 | ✗ | ✓ | ✓ |
| 导出所有申请 | ✗ | ✓ | ✓ |

## 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| DB_HOST | localhost | 数据库主机 |
| DB_PORT | 5432 | 数据库端口 |
| DB_USER | postgres | 数据库用户名 |
| DB_PASSWORD | password | 数据库密码 |
| DB_NAME | credit_management | 数据库名称 |
| PORT | 8083 | 服务端口 |
| USER_SERVICE_URL | http://user-service:8084 | 用户服务URL |

### 文件上传配置

- 单个文件最大大小: 20MB
- 批量上传最大文件数: 10个
- 支持的文件类型: PDF, DOC, DOCX, TXT, JPG, PNG, MP4, MP3, ZIP等

## 开发指南

### 项目结构

```
credit-activity-service/
├── handlers/          # 处理器
│   ├── activity.go    # 活动管理
│   ├── participant.go # 参与者管理
│   ├── application.go # 申请管理
│   └── attachment.go  # 附件管理
├── models/            # 数据模型
│   ├── activity.go    # 活动相关模型
│   └── attachment.go  # 附件相关模型
├── utils/             # 工具函数
│   ├── middleware.go  # 中间件
│   └── common.go      # 通用工具
├── main.go            # 主程序
├── go.mod             # Go模块文件
├── go.sum             # Go依赖文件
├── Dockerfile         # Docker配置
└── README.md          # 说明文档
```

### 添加新功能

1. 在 `models/` 目录下定义数据模型
2. 在 `handlers/` 目录下实现处理器
3. 在 `main.go` 中添加路由
4. 更新数据库迁移脚本
5. 添加相应的测试

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./handlers -v

# 运行基准测试
go test -bench=.
```

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t credit-activity-service .

# 运行容器
docker run -d \
  --name credit-activity-service \
  -p 8083:8083 \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=password \
  -e DB_NAME=credit_management \
  credit-activity-service
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: credit-activity-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: credit-activity-service
  template:
    metadata:
      labels:
        app: credit-activity-service
    spec:
      containers:
      - name: credit-activity-service
        image: credit-activity-service:latest
        ports:
        - containerPort: 8083
        env:
        - name: DB_HOST
          value: postgres-service
        - name: DB_PORT
          value: "5432"
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        - name: DB_NAME
          value: credit_management
```

## 监控和日志

### 健康检查

```bash
curl http://localhost:8083/health
```

### 日志级别

- INFO: 一般信息
- WARN: 警告信息
- ERROR: 错误信息
- DEBUG: 调试信息（开发环境）

### 监控指标

- 活动创建数量
- 活动审核通过率
- 申请生成数量
- API响应时间
- 错误率

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否运行
   - 验证数据库连接参数
   - 确认网络连接

2. **文件上传失败**
   - 检查文件大小是否超限
   - 验证文件类型是否支持
   - 确认存储目录权限

3. **权限验证失败**
   - 检查Authorization头格式
   - 验证token是否有效
   - 确认用户角色权限

### 日志查看

```bash
# 查看服务日志
docker logs credit-activity-service

# 查看实时日志
docker logs -f credit-activity-service
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证。

## 联系方式

- 项目维护者: 学分活动服务团队
- 邮箱: support@example.com
- 问题反馈: [GitHub Issues](https://github.com/example/credit-management/issues)

---

**版本**: 1.0.0  
**更新时间**: 2024-01-01 