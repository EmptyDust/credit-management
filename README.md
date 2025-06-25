# 学分管理系统 - 微服务架构

## 项目概述

这是一个基于微服务架构的学分管理系统，采用高聚合低耦合的设计原则，将系统拆分为多个独立的微服务。

## 系统架构

### 微服务组件

1. **auth-service** (端口: 8081)
   - 认证管理：用户登录、token验证、token刷新
   - 权限管理：角色管理、权限分配、权限验证

2. **user-management-service** (端口: 8084)
   - 用户基础信息管理：用户注册、用户信息维护
   - 通知管理：系统通知、用户通知

3. **credit-activity-service** (端口: 8083)
   - 学分活动管理：活动创建、状态管理、参与者管理
   - 申请管理：自动申请生成、申请审核、学分分配
   - 数据统计：活动统计、申请统计、学分统计

4. **student-info-service** (端口: 8085)
   - 学生信息管理：学生基础信息维护

5. **teacher-info-service** (端口: 8086)
   - 教师信息管理：教师基础信息维护

6. **api-gateway** (端口: 8080)
   - API网关：统一入口、路由转发、负载均衡

7. **frontend** (端口: 3000)
   - 前端应用：React + TypeScript + Tailwind CSS

8. **postgres** (端口: 5432)
   - 数据库：PostgreSQL

## 技术栈

### 后端
- **语言**: Go 1.24
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
- API网关: http://localhost:8080
- 健康检查: http://localhost:8080/health

### 服务端口映射

| 服务 | 端口 | 说明 |
|------|------|------|
| API网关 | 8080 | 统一API入口 |
| 认证服务 | 8081 | 认证和权限 |
| 学分活动服务 | 8083 | 活动和申请管理 |
| 用户管理 | 8084 | 用户基础信息 |
| 学生信息 | 8085 | 学生信息 |
| 教师信息 | 8086 | 教师信息 |
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

### 学分活动管理
- `GET /api/activities/categories` - 获取活动类别
- `POST /api/activities` - 创建活动
- `GET /api/activities` - 获取活动列表
- `GET /api/activities/{id}` - 获取活动详情
- `PUT /api/activities/{id}` - 更新活动
- `DELETE /api/activities/{id}` - 删除活动
- `POST /api/activities/{id}/submit` - 提交活动审核
- `POST /api/activities/{id}/withdraw` - 撤回活动
- `POST /api/activities/{id}/review` - 审核活动

### 参与者管理
- `POST /api/activities/{id}/participants` - 添加参与者
- `GET /api/activities/{id}/participants` - 获取参与者列表
- `PUT /api/activities/{id}/participants/batch-credits` - 批量设置学分
- `PUT /api/activities/{id}/participants/{user_id}/credits` - 设置单个学分
- `DELETE /api/activities/{id}/participants/{user_id}` - 删除参与者
- `POST /api/activities/{id}/leave` - 退出活动

### 申请管理
- `GET /api/applications` - 获取申请列表
- `GET /api/applications/{id}` - 获取申请详情
- `GET /api/applications/stats` - 获取申请统计
- `GET /api/applications/export` - 导出申请数据

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

#### 学分活动相关
- `activities` - 学分活动
- `participants` - 活动参与者
- `applications` - 申请记录（自动生成）

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