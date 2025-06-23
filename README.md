# 创新创业学分管理平台

一个基于微服务架构的创新创业学分管理系统，支持学生申请学分、教师审核、管理员管理等功能。

## 技术栈

### 后端
- **语言**: Go 1.24.4
- **框架**: Gin
- **数据库**: PostgreSQL
- **ORM**: GORM
- **架构**: 微服务

### 前端
- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI库**: Tailwind CSS + shadcn/ui
- **路由**: React Router
- **状态管理**: React Context

### 部署
- **容器化**: Docker + Docker Compose
- **网关**: 自定义API网关

## 项目结构

```
credit-management/
├── api-gateway/                 # API网关服务
├── user-management-service/     # 用户管理服务
├── student-info-service/        # 学生信息服务
├── teacher-info-service/        # 教师信息服务
├── affair-management-service/   # 事项管理服务
├── general-application-service/ # 通用申请服务
├── frontend/                   # 前端应用
├── tester/                     # 测试工具
├── docker-compose.yml          # Docker编排文件
├── init-db.sql                # 数据库初始化脚本
└── README.md                  # 项目说明
```

## 微服务说明

### 1. 用户管理服务 (Port: 8081)
- 用户注册、登录
- JWT认证
- 用户信息管理

### 2. 学生信息服务 (Port: 8084)
- 学生信息CRUD
- 学生档案管理

### 3. 教师信息服务 (Port: 8085)
- 教师信息CRUD
- 教师档案管理

### 4. 事项管理服务 (Port: 8083)
- 创新创业事项管理
- 事项类型和状态管理

### 5. 通用申请服务 (Port: 8086)
- 学分申请管理
- 申请审核流程
- 认定学分管理

### 6. API网关 (Port: 8080)
- 统一API入口
- 路由转发
- CORS处理

## 快速开始

### 环境要求
- Docker
- Docker Compose

### 域名配置（可选）

如果要使用域名访问，请将域名解析到服务器IP：

```bash
# 检查域名配置
./check-domain.sh
```

### 启动步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd credit-management
```

2. **启动所有服务**
```bash
./start.sh
```

3. **访问应用**
- 前端应用: http://lab.emptydust.com (域名访问) 或 http://localhost (本地访问)
- API网关: http://lab.emptydust.com/api (域名访问) 或 http://localhost:8080 (本地访问)
- 数据库: localhost:5432

### 管理命令

```bash
# 启动服务
./start.sh

# 停止服务
./stop.sh

# 查看状态
./status.sh

# 检查域名配置
./check-domain.sh

# 系统测试
./test-system.sh
```

### 默认账号

系统预置了以下测试账号：

| 用户名 | 密码 | 角色 | 说明 |
|--------|------|------|------|
| admin | password | 管理员 | 系统管理员 |
| student1 | password | 学生 | 测试学生账号 |
| teacher1 | password | 教师 | 测试教师账号 |

## API文档

### 用户管理API
- `POST /api/users/register` - 用户注册
- `POST /api/users/login` - 用户登录
- `GET /api/users/:id` - 获取用户信息
- `PUT /api/users/:id` - 更新用户信息

### 学生信息API
- `POST /api/students` - 创建学生信息
- `GET /api/students` - 获取学生列表
- `GET /api/students/:id` - 获取学生信息
- `PUT /api/students/:id` - 更新学生信息

### 教师信息API
- `POST /api/teachers` - 创建教师信息
- `GET /api/teachers` - 获取教师列表
- `GET /api/teachers/:id` - 获取教师信息
- `PUT /api/teachers/:id` - 更新教师信息

### 事项管理API
- `POST /api/affairs` - 创建事项
- `GET /api/affairs` - 获取事项列表
- `GET /api/affairs/:id` - 获取事项详情
- `PUT /api/affairs/:id` - 更新事项

### 申请管理API
- `POST /api/applications` - 创建申请
- `GET /api/applications` - 获取申请列表
- `GET /api/applications/:id` - 获取申请详情
- `PUT /api/applications/:id` - 更新申请
- `POST /api/applications/:id/review` - 审核申请

## 数据库设计

### 核心表结构

1. **users** - 用户表
2. **students** - 学生信息表
3. **teachers** - 教师信息表
4. **affairs** - 事项表
5. **applications** - 申请表

### 关键特性

- **认定学分统一管理**: 所有学分申请都通过统一的申请表管理，便于集中审核和统计
- **微服务架构**: 各功能模块独立部署，便于扩展和维护
- **JWT认证**: 安全的用户认证机制
- **响应式前端**: 现代化的用户界面

## 开发说明

### 本地开发

1. **启动数据库**
```bash
docker-compose up postgres -d
```

2. **启动单个服务**
```bash
cd user-management-service
go run main.go
```

3. **前端开发**
```bash
cd frontend
pnpm install
pnpm dev
```

### 测试

项目包含完整的测试用例：

```bash
cd tester
go test ./...
```

## 部署

### 生产环境部署

1. **构建镜像**
```bash
docker-compose build
```

2. **启动服务**
```bash
docker-compose up -d
```

3. **查看日志**
```bash
docker-compose logs -f
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

如有问题，请提交 Issue 或联系开发团队。 