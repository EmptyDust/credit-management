# 学分管理系统 - 微服务架构

## 项目概述

这是一个基于微服务架构的学分管理系统，采用高聚合低耦合的设计原则，将系统拆分为多个独立的微服务。

## 系统架构

### 微服务组件

1. **auth-service** (端口: 8081)
   - 认证管理：用户登录、token验证、token刷新
   - 权限管理：角色管理、权限分配、权限验证

2. **user-management-service** (端口: 8080)
   - 用户基础信息管理：用户注册、用户信息维护
   - 通知管理：系统通知、用户通知

3. **application-management-service** (端口: 8082)
   - 申请信息管理：五种申请类型的管理
   - 文件管理：申请附件及配套文件管理

4. **affair-management-service** (端口: 8083)
   - 事务信息管理：学分事务的管理

5. **student-info-service** (端口: 8084)
   - 学生信息管理：学生基础信息维护

6. **teacher-info-service** (端口: 8085)
   - 教师信息管理：教师基础信息维护

7. **api-gateway** (端口: 8000)
   - API网关：统一入口、路由转发、负载均衡

8. **frontend** (端口: 3000)
   - 前端应用：React + TypeScript + Tailwind CSS

9. **postgres** (端口: 5432)
   - 数据库：PostgreSQL

## 技术栈

### 后端
- **语言**: Go 1.21
- **框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **认证**: JWT
- **容器化**: Docker

### 前端
- **框架**: React 18
- **语言**: TypeScript
- **样式**: Tailwind CSS
- **构建工具**: Vite
- **UI组件**: shadcn/ui

## 快速开始

### 环境要求
- Docker
- Docker Compose

### 启动步骤

1. 克隆项目
```bash
git clone <repository-url>
cd credit-management
```

2. 启动所有服务
```bash
docker-compose up -d
```

3. 访问应用
- 前端: http://localhost:3000
- API网关: http://localhost:8000
- 健康检查: http://localhost:8000/health

### 服务端口映射

| 服务 | 端口 | 说明 |
|------|------|------|
| API网关 | 8000 | 统一API入口 |
| 用户管理 | 8080 | 用户基础信息 |
| 认证服务 | 8081 | 认证和权限 |
| 申请管理 | 8082 | 申请和文件 |
| 事务管理 | 8083 | 事务信息 |
| 学生信息 | 8084 | 学生信息 |
| 教师信息 | 8085 | 教师信息 |
| 数据库 | 5432 | PostgreSQL |
| 前端 | 3000 | React应用 |

## API接口

### 认证相关
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/validate-token` - 验证token
- `POST /api/auth/refresh-token` - 刷新token
- `POST /api/auth/logout` - 用户登出

### 权限管理
- `GET /api/permissions/roles` - 获取角色列表
- `POST /api/permissions/roles` - 创建角色
- `GET /api/permissions` - 获取权限列表
- `POST /api/permissions/users/{userID}/roles` - 分配角色

### 用户管理
- `POST /api/users/register` - 用户注册
- `GET /api/users/profile` - 获取用户信息
- `PUT /api/users/profile` - 更新用户信息
- `GET /api/users/stats` - 获取用户统计

### 申请管理
- `GET /api/application-types` - 获取申请类型
- `POST /api/applications` - 创建申请
- `GET /api/applications` - 获取申请列表
- `PUT /api/applications/{id}/status` - 更新申请状态

### 文件管理
- `POST /api/files/upload` - 上传文件
- `GET /api/files/download/{fileID}` - 下载文件
- `GET /api/files/application/{applicationID}` - 获取申请文件

### 通知管理
- `GET /api/notifications` - 获取用户通知
- `PUT /api/notifications/{id}/read` - 标记已读
- `GET /api/notifications/unread-count` - 获取未读数量

## 数据库设计

### 核心表结构

#### 用户相关
- `users` - 用户基础信息
- `roles` - 角色定义
- `permissions` - 权限定义
- `user_roles` - 用户角色关联
- `user_permissions` - 用户权限关联
- `role_permissions` - 角色权限关联

#### 申请相关
- `application_types` - 申请类型
- `applications` - 申请记录
- `application_files` - 申请文件

#### 通知相关
- `notifications` - 通知记录

## 开发指南

### 本地开发

1. 安装Go 1.21+
2. 安装PostgreSQL
3. 配置环境变量
4. 运行服务

```bash
# 启动数据库
docker-compose up postgres -d

# 运行服务
cd auth-service && go run main.go
cd user-management-service && go run main.go
# ... 其他服务
```

### 代码结构

每个微服务都遵循相同的目录结构：
```
service-name/
├── main.go          # 服务入口
├── go.mod           # Go模块
├── Dockerfile       # 容器配置
├── handlers/        # 处理器
├── models/          # 数据模型
└── utils/           # 工具函数
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户 | postgres |
| DB_PASSWORD | 数据库密码 | password |
| DB_NAME | 数据库名称 | credit_management |
| JWT_SECRET | JWT密钥 | your-secret-key |
| PORT | 服务端口 | 服务默认端口 |

## 部署

### 生产环境部署

1. 修改环境变量
2. 配置数据库连接
3. 设置JWT密钥
4. 启动服务

```bash
# 生产环境启动
docker-compose -f docker-compose.prod.yml up -d
```

### 监控和日志

- 健康检查: `/health`
- 日志收集: 使用Docker日志
- 监控: 可集成Prometheus + Grafana

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 许可证

MIT License

## 联系方式

如有问题，请提交Issue或联系开发团队。 