# 学分活动服务 (Credit Activity Service)

学分活动服务是学分管理系统的核心服务，负责处理学分活动的创建、管理、审核和申请生成。该服务合并了原有的事务管理和申请管理功能，简化了业务流程。

## 功能特性

- **活动管理**: 创建、编辑、删除、审核学分活动
- **参与者管理**: 添加、删除参与者，设置学分分配
- **申请生成**: 活动通过审核后自动生成申请记录
- **权限控制**: 基于角色的权限管理
- **撤回机制**: 支持活动创建者撤回活动
- **导出功能**: 支持申请数据导出

## 技术栈

- **语言**: Go 1.24.4
- **框架**: Gin
- **数据库**: PostgreSQL
- **ORM**: GORM
- **容器化**: Docker

## 快速开始

### 环境要求

- Go 1.24.4+
- PostgreSQL 15+
- Docker & Docker Compose

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd credit-activity-service
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=credit_management
export PORT=8083
```

4. **运行服务**
```bash
go run main.go
```

### Docker 部署

#### 单服务部署

1. **构建镜像**
```bash
docker build -t credit-activity-service .
```

2. **运行容器**
```bash
docker run -d \
  --name credit-activity-service \
  -p 8083:8083 \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=password \
  -e DB_NAME=credit_management \
  credit-activity-service
```

#### 使用 Docker Compose

1. **启动完整系统**
```bash
cd ../
docker-compose up -d
```

2. **仅启动学分活动服务**
```bash
docker-compose up -d credit-activity-service
```

## API 文档

### 基础信息

- **服务端口**: 8083
- **基础路径**: `/api`
- **认证方式**: JWT Token
- **数据格式**: JSON

### 主要端点

#### 活动管理
- `GET /api/activities` - 获取活动列表
- `POST /api/activities` - 创建活动
- `GET /api/activities/:id` - 获取活动详情
- `PUT /api/activities/:id` - 更新活动
- `DELETE /api/activities/:id` - 删除活动
- `POST /api/activities/:id/submit` - 提交活动审核
- `POST /api/activities/:id/withdraw` - 撤回活动
- `POST /api/activities/:id/review` - 审核活动

#### 参与者管理
- `POST /api/activities/:id/participants` - 添加参与者
- `PUT /api/activities/:id/participants/batch-credits` - 批量设置学分
- `PUT /api/activities/:id/participants/:user_id/credits` - 设置单个学分
- `DELETE /api/activities/:id/participants/:user_id` - 删除参与者
- `GET /api/activities/:id/participants` - 获取参与者列表
- `POST /api/activities/:id/leave` - 退出活动

#### 申请管理
- `GET /api/applications` - 获取用户申请列表
- `GET /api/applications/:id` - 获取申请详情
- `GET /api/applications/all` - 获取所有申请
- `GET /api/applications/export` - 导出申请数据
- `GET /api/applications/stats` - 获取申请统计

## 数据库设计

### 核心表结构

#### credit_activities
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

#### activity_participants
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

#### applications
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

## 监控和日志

### 健康检查
```bash
curl http://localhost:8083/health
```

### 日志级别
- `INFO`: 一般操作日志
- `WARN`: 警告信息
- `ERROR`: 错误信息

## 测试

### 运行测试
```bash
go test ./...
```

### API 测试
```bash
cd ../tester
.\test-credit-activity-service.ps1
```

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否运行
   - 验证连接参数是否正确
   - 确认网络连接

2. **权限错误**
   - 检查JWT token是否有效
   - 确认用户角色权限
   - 验证API路径是否正确

3. **触发器不工作**
   - 检查PostgreSQL版本是否支持
   - 确认触发器函数是否正确创建
   - 查看数据库日志

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License 