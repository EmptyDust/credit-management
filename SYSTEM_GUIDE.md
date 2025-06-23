# 创新学分管理系统 - 完整使用指南

## 系统概述

创新学分管理系统是一个基于微服务架构的完整解决方案，用于管理学生的创新学分申请、审核和管理流程。

### 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   API网关       │    │   PostgreSQL    │
│   (React)       │◄──►│   (Gin)         │◄──►│   数据库        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                    ┌─────────┼─────────┐
                    │         │         │
            ┌───────▼──────┐  │  ┌──────▼──────┐
            │用户管理服务   │  │  │认证服务     │
            │(8080)        │  │  │(8081)       │
            └──────────────┘  │  └─────────────┘
                              │
                    ┌─────────┼─────────┐
                    │         │         │
            ┌───────▼──────┐  │  ┌──────▼──────┐
            │申请管理服务   │  │  │事务管理服务 │
            │(8082)        │  │  │(8083)       │
            └──────────────┘  │  └─────────────┘
                              │
                    ┌─────────┼─────────┐
                    │         │         │
            ┌───────▼──────┐  │  ┌──────▼──────┐
            │学生信息服务   │  │  │教师信息服务 │
            │(8084)        │  │  │(8085)       │
            └──────────────┘  │  └─────────────┘
```

## 快速开始

### 环境要求

- Docker Desktop
- Docker Compose
- 至少 4GB 可用内存
- 至少 10GB 可用磁盘空间

### 启动系统

#### Windows 用户

1. 双击运行 `start-system.bat`
2. 等待所有服务启动完成
3. 访问 http://localhost:3000

#### Linux/Mac 用户

1. 运行 `./start-system.sh`
2. 等待所有服务启动完成
3. 访问 http://localhost:3000

### 系统测试

#### Windows 用户
```bash
test-system.bat
```

#### Linux/Mac 用户
```bash
./test-system.sh
```

### 系统监控

#### Windows 用户
```bash
monitor-system.bat
```

## 服务详情

### 1. 前端应用 (Port: 3000)

**技术栈**: React + TypeScript + Tailwind CSS + Shadcn/ui

**主要功能**:
- 用户登录/注册
- 仪表板
- 申请管理
- 学生管理
- 教师管理
- 事务管理
- 个人资料管理

**访问地址**: http://localhost:3000

### 2. API网关 (Port: 8000)

**技术栈**: Go + Gin

**主要功能**:
- 请求路由
- 负载均衡
- 认证中间件
- 请求日志
- 错误处理

**健康检查**: http://localhost:8000/health

### 3. 用户管理服务 (Port: 8080)

**技术栈**: Go + Gin + GORM

**主要功能**:
- 用户注册/管理
- 头像上传
- 通知管理
- 用户统计

**API端点**:
- `POST /api/users/register` - 用户注册
- `GET /api/users/profile` - 获取用户信息
- `PUT /api/users/profile` - 更新用户信息
- `POST /api/users/avatar` - 上传头像
- `GET /api/notifications` - 获取通知

### 4. 认证服务 (Port: 8081)

**技术栈**: Go + Gin + GORM + JWT

**主要功能**:
- 用户登录
- JWT token管理
- 权限管理
- 角色管理

**API端点**:
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/validate-token` - 验证token
- `POST /api/auth/refresh-token` - 刷新token
- `POST /api/permissions/roles` - 创建角色
- `GET /api/permissions/roles` - 获取角色列表

### 5. 申请管理服务 (Port: 8082)

**技术栈**: Go + Gin + GORM

**主要功能**:
- 创新学分申请
- 申请类型管理
- 文件上传/下载
- 申请状态管理

**API端点**:
- `POST /api/applications` - 创建申请
- `GET /api/applications` - 获取申请列表
- `PUT /api/applications/:id/status` - 更新申请状态
- `POST /api/files/upload` - 上传文件
- `GET /api/files/download/:fileID` - 下载文件

### 6. 事务管理服务 (Port: 8083)

**技术栈**: Go + Gin + GORM

**主要功能**:
- 事务管理
- 事务类型管理
- 事务状态跟踪

**API端点**:
- `POST /api/affairs` - 创建事务
- `GET /api/affairs` - 获取事务列表
- `PUT /api/affairs/:id/status` - 更新事务状态

### 7. 学生信息服务 (Port: 8084)

**技术栈**: Go + Gin + GORM

**主要功能**:
- 学生信息管理
- 学生档案
- 学分统计

**API端点**:
- `GET /api/students` - 获取学生列表
- `GET /api/students/:id` - 获取学生详情
- `PUT /api/students/:id` - 更新学生信息

### 8. 教师信息服务 (Port: 8085)

**技术栈**: Go + Gin + GORM

**主要功能**:
- 教师信息管理
- 教师档案
- 指导记录

**API端点**:
- `GET /api/teachers` - 获取教师列表
- `GET /api/teachers/:id` - 获取教师详情
- `PUT /api/teachers/:id` - 更新教师信息

### 9. PostgreSQL数据库 (Port: 5432)

**配置**:
- 数据库名: credit_management
- 用户名: postgres
- 密码: password

## 用户角色和权限

### 学生用户
- 查看个人信息
- 提交创新学分申请
- 查看申请状态
- 上传申请材料
- 查看通知

### 教师用户
- 查看个人信息
- 审核学生申请
- 管理学生信息
- 查看统计报告
- 发送通知

### 管理员用户
- 用户管理
- 权限管理
- 系统配置
- 数据统计
- 系统监控

## 数据库设计

### 核心表结构

1. **users** - 用户表
2. **roles** - 角色表
3. **permissions** - 权限表
4. **applications** - 申请表
5. **application_files** - 申请文件表
6. **affairs** - 事务表
7. **students** - 学生信息表
8. **teachers** - 教师信息表
9. **notifications** - 通知表

## 部署选项

### 1. 本地开发环境
使用 Docker Compose 进行本地开发和测试。

### 2. 生产环境
使用 Kubernetes 进行生产环境部署。

**部署步骤**:
```bash
# 1. 应用PostgreSQL配置
kubectl apply -f k8s/postgres-deployment.yaml

# 2. 应用所有服务配置
kubectl apply -f k8s/credit-service-deployment.yaml

# 3. 检查部署状态
kubectl get pods
kubectl get services
```

## 故障排除

### 常见问题

1. **Docker Desktop 未启动**
   - 启动 Docker Desktop
   - 等待服务完全启动

2. **端口冲突**
   - 检查端口占用: `netstat -ano | findstr :3000`
   - 停止占用端口的进程

3. **数据库连接失败**
   - 检查 PostgreSQL 容器状态
   - 查看数据库日志: `docker-compose logs postgres`

4. **服务启动失败**
   - 查看服务日志: `docker-compose logs [服务名]`
   - 检查环境变量配置

### 日志查看

```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs api-gateway
docker-compose logs user-management-service

# 实时查看日志
docker-compose logs -f [服务名]
```

### 系统维护

```bash
# 停止所有服务
docker-compose down

# 清理系统
docker system prune -a

# 重新构建
docker-compose up --build -d
```

## 开发指南

### 添加新功能

1. 在相应的微服务中添加新的API端点
2. 更新API网关的路由配置
3. 在前端添加相应的UI组件
4. 更新数据库模型（如需要）

### 代码结构

```
project/
├── frontend/                 # 前端应用
├── api-gateway/             # API网关
├── user-management-service/  # 用户管理服务
├── auth-service/            # 认证服务
├── application-management/  # 申请管理服务
├── affair-management-service/ # 事务管理服务
├── student-info-service/    # 学生信息服务
├── teacher-info-service/    # 教师信息服务
├── k8s/                    # Kubernetes配置
├── tester/                 # 测试工具
└── docs/                   # 文档
```

## 性能优化

### 数据库优化
- 添加适当的索引
- 优化查询语句
- 使用连接池

### 缓存策略
- Redis缓存热点数据
- 前端缓存静态资源
- API响应缓存

### 负载均衡
- 使用Nginx进行负载均衡
- 服务实例水平扩展
- 数据库读写分离

## 安全考虑

### 认证和授权
- JWT token认证
- 基于角色的权限控制
- API访问控制

### 数据安全
- 敏感数据加密
- SQL注入防护
- XSS攻击防护

### 网络安全
- HTTPS传输
- CORS配置
- 防火墙设置

## 监控和告警

### 系统监控
- 服务健康检查
- 性能指标监控
- 错误日志监控

### 告警机制
- 服务异常告警
- 性能阈值告警
- 磁盘空间告警

## 备份和恢复

### 数据备份
```bash
# 数据库备份
docker-compose exec postgres pg_dump -U postgres credit_management > backup.sql

# 文件备份
tar -czf uploads_backup.tar.gz application-management/uploads/
```

### 数据恢复
```bash
# 数据库恢复
docker-compose exec -T postgres psql -U postgres credit_management < backup.sql

# 文件恢复
tar -xzf uploads_backup.tar.gz
```

## 联系和支持

如有问题或建议，请通过以下方式联系：

- 项目仓库: [GitHub Repository]
- 问题反馈: [Issues]
- 文档更新: [Documentation]

---

**注意**: 本系统仅供学习和演示使用，生产环境部署前请进行充分的安全测试和性能优化。 