# 用户管理微服务

## 概述

用户管理微服务是创新创业学分管理平台的核心服务之一，负责用户认证、权限管理、文件上传、通知系统等功能。

## 新增功能

### 1. 文件上传功能
- **文件上传**: 支持多种文件格式的上传，包括图片、文档、视频、音频等
- **文件预览**: 支持图片直接预览，文档生成HTML预览
- **缩略图生成**: 自动为图片生成缩略图
- **文件分类**: 支持按类别管理文件（头像、文档、证书等）
- **权限控制**: 支持公开/私有文件设置

### 2. 权限管理系统
- **基于角色的访问控制(RBAC)**: 支持角色和权限的灵活配置
- **用户权限**: 支持直接为用户分配权限
- **角色权限**: 支持为角色分配权限，用户继承角色权限
- **权限中间件**: 提供权限检查中间件，方便API权限控制
- **默认角色**: 系统预设admin、teacher、student、user等角色

### 3. 通知系统
- **用户通知**: 支持用户个人通知管理
- **系统通知**: 支持管理员发送系统通知
- **通知模板**: 提供预定义的通知模板
- **通知统计**: 提供通知数量统计功能
- **批量通知**: 支持批量发送通知

### 4. 报表统计
- **用户统计**: 用户数量、类型、状态等统计
- **文件统计**: 文件数量、大小、类型等统计
- **通知统计**: 通知数量、类型、时间等统计
- **实时数据**: 支持今日、本周、本月等时间维度统计

### 5. 移动端适配
- **响应式设计**: API支持移动端访问
- **文件上传优化**: 支持大文件分片上传
- **图片压缩**: 自动压缩上传的图片
- **缓存策略**: 支持文件缓存和CDN加速

### 6. 性能优化
- **数据库优化**: 使用索引优化查询性能
- **文件存储优化**: 支持本地存储和云存储
- **缓存机制**: 支持Redis缓存热点数据
- **并发处理**: 支持高并发用户访问

## API接口

### 用户管理
```
POST   /api/users/register          # 用户注册
POST   /api/users/login             # 用户登录
GET    /api/users/:username         # 获取用户信息
PUT    /api/users/:username         # 更新用户信息
DELETE /api/users/:username         # 删除用户
GET    /api/users                   # 获取所有用户
GET    /api/users/type/:userType    # 根据用户类型获取用户
POST   /api/users/validate-token    # 验证JWT token
GET    /api/users/stats             # 获取用户统计信息
POST   /api/users/avatar            # 上传头像
```

### 文件管理
```
POST   /api/files/upload            # 上传文件
GET    /api/files/download/:id      # 下载文件
GET    /api/files/:id               # 获取文件信息
GET    /api/files                   # 获取用户文件列表
PUT    /api/files/:id               # 更新文件信息
DELETE /api/files/:id               # 删除文件
GET    /api/files/public            # 获取公开文件列表
GET    /api/files/stats             # 获取文件统计信息
```

### 权限管理
```
POST   /api/permissions/roles       # 创建角色
GET    /api/permissions/roles       # 获取角色列表
GET    /api/permissions/roles/:id   # 获取角色详情
PUT    /api/permissions/roles/:id   # 更新角色
DELETE /api/permissions/roles/:id   # 删除角色
POST   /api/permissions             # 创建权限
GET    /api/permissions             # 获取权限列表
GET    /api/permissions/:id         # 获取权限详情
DELETE /api/permissions/:id         # 删除权限
POST   /api/permissions/assign-role # 分配角色给用户
DELETE /api/permissions/users/:userID/roles/:roleID # 移除用户角色
GET    /api/permissions/users/:userID/roles # 获取用户角色
POST   /api/permissions/assign-permission # 分配权限给用户
DELETE /api/permissions/users/:userID/permissions/:permissionID # 移除用户权限
GET    /api/permissions/users/:userID/permissions # 获取用户权限
POST   /api/permissions/roles/:roleID/permissions/:permissionID # 给角色分配权限
DELETE /api/permissions/roles/:roleID/permissions/:permissionID # 移除角色权限
POST   /api/permissions/initialize  # 初始化默认权限和角色
```

### 通知管理
```
GET    /api/notifications           # 获取用户通知列表
GET    /api/notifications/:id       # 获取通知详情
PUT    /api/notifications/:id/read  # 标记通知为已读
PUT    /api/notifications/read-all # 标记所有通知为已读
DELETE /api/notifications/:id       # 删除通知
GET    /api/notifications/unread-count # 获取未读通知数量
GET    /api/notifications/stats     # 获取通知统计信息

# 管理员功能
POST   /api/notifications/admin     # 创建通知
POST   /api/notifications/admin/system # 发送系统通知
POST   /api/notifications/admin/batch # 批量发送通知
GET    /api/notifications/admin/all # 获取所有通知
DELETE /api/notifications/admin/:id # 管理员删除通知
GET    /api/notifications/admin/stats # 获取系统通知统计
```

## 数据库设计

### 用户表 (users)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20),
    real_name VARCHAR(100),
    user_type VARCHAR(20) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    avatar VARCHAR(255),
    last_login_at TIMESTAMP,
    register_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### 角色表 (roles)
```sql
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### 权限表 (permissions)
```sql
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### 用户文件表 (user_files)
```sql
CREATE TABLE user_files (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT,
    file_type VARCHAR(20),
    mime_type VARCHAR(100),
    category VARCHAR(50),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    download_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### 通知表 (notifications)
```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(20) NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## 配置说明

### 环境变量
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=credit_management
DB_SSLMODE=disable

# JWT配置
JWT_SECRET=your-secret-key

# 服务配置
PORT=8081

# 文件上传配置
UPLOAD_DIR=./uploads
MAX_FILE_SIZE=52428800  # 50MB
PUBLIC_URL=http://localhost:8081/files
```

### 文件上传配置
```go
type FileConfig struct {
    UploadDir       string   // 上传目录
    MaxFileSize     int64    // 最大文件大小
    AllowedTypes    []string // 允许的文件类型
    ImageTypes      []string // 图片类型
    DocumentTypes   []string // 文档类型
    VideoTypes      []string // 视频类型
    AudioTypes      []string // 音频类型
    ArchiveTypes    []string // 压缩包类型
    ThumbnailDir    string   // 缩略图目录
    PreviewDir      string   // 预览文件目录
    TempDir         string   // 临时文件目录
    PublicURL       string   // 公共访问URL
    EnablePreview   bool     // 是否启用预览
    EnableThumbnail bool     // 是否启用缩略图
}
```

## 部署说明

### Docker部署
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/uploads ./uploads
EXPOSE 8081
CMD ["./main"]
```

### K8s部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-management-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-management-service
  template:
    metadata:
      labels:
        app: user-management-service
    spec:
      containers:
      - name: user-management-service
        image: user-management-service:latest
        ports:
        - containerPort: 8081
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        volumeMounts:
        - name: uploads
          mountPath: /app/uploads
      volumes:
      - name: uploads
        persistentVolumeClaim:
          claimName: uploads-pvc
```

## 测试

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v -run TestUserRegister

# 运行性能测试
go test -bench=.

# 运行并发测试
go test -v -run TestConcurrentUserRegistration
```

### API测试
```bash
# 用户注册
curl -X POST http://localhost:8081/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "email": "test@example.com",
    "phone": "13800138000",
    "real_name": "测试用户",
    "user_type": "student"
  }'

# 用户登录
curl -X POST http://localhost:8081/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'

# 文件上传
curl -X POST http://localhost:8081/api/files/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test.txt" \
  -F "category=document" \
  -F "description=测试文件" \
  -F "is_public=false"
```

## 性能优化建议

1. **数据库优化**
   - 为常用查询字段添加索引
   - 使用连接池管理数据库连接
   - 定期清理过期数据

2. **文件存储优化**
   - 使用CDN加速文件访问
   - 实现文件分片上传
   - 定期清理临时文件

3. **缓存优化**
   - 使用Redis缓存热点数据
   - 实现文件元数据缓存
   - 缓存用户权限信息

4. **并发优化**
   - 使用goroutine处理并发请求
   - 实现请求限流和熔断
   - 优化数据库连接池配置

## 监控和日志

### 健康检查
```bash
curl http://localhost:8081/health
```

### 日志配置
```go
// 配置结构化日志
log.SetFormatter(&log.JSONFormatter{})
log.SetLevel(log.InfoLevel)
```

### 监控指标
- 请求响应时间
- 错误率
- 并发用户数
- 文件上传成功率
- 数据库连接数

## 安全考虑

1. **认证安全**
   - 使用JWT进行身份认证
   - 实现token刷新机制
   - 密码加密存储

2. **权限安全**
   - 基于角色的访问控制
   - 最小权限原则
   - 定期权限审计

3. **文件安全**
   - 文件类型验证
   - 文件大小限制
   - 病毒扫描

4. **API安全**
   - 输入验证和过滤
   - SQL注入防护
   - XSS攻击防护

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接参数
   - 检查网络连通性

2. **文件上传失败**
   - 检查磁盘空间
   - 验证文件权限
   - 检查文件类型限制

3. **权限验证失败**
   - 检查JWT token有效性
   - 验证用户权限配置
   - 检查中间件配置

### 日志分析
```bash
# 查看错误日志
grep "ERROR" logs/app.log

# 查看慢查询日志
grep "SLOW" logs/app.log

# 查看访问日志
tail -f logs/access.log
``` 