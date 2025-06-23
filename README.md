# 创新创业学分管理系统

一个基于微服务架构的创新创业学分管理平台，支持学生申请、教师审核、管理员管理等功能。

## 🏗️ 系统架构

### 微服务架构
- **用户管理服务** (Port: 8081) - 用户注册、登录、认证
- **学生信息服务** (Port: 8082) - 学生信息管理
- **教师信息服务** (Port: 8083) - 教师信息管理
- **事项管理服务** (Port: 8087) - 创新创业事项管理
- **通用申请服务** (Port: 8086) - 学分申请处理
- **API网关** (Port: 8080) - 统一入口和路由转发
- **前端应用** (Port: 3000) - React + TypeScript用户界面

### 技术栈
- **后端**: Go + Gin + GORM
- **数据库**: PostgreSQL
- **前端**: React + TypeScript + Tailwind CSS
- **容器化**: Docker + Docker Compose
- **API网关**: 反向代理 + 负载均衡

## 🚀 快速开始

### 环境要求
- Docker & Docker Compose
- Go 1.24.4+
- Node.js 18+

### 启动系统
```bash
# 克隆项目
git clone <repository-url>
cd credit-management

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 访问系统
- **前端界面**: http://localhost:3000
- **API网关**: http://localhost:8080
- **用户服务**: http://localhost:8081
- **学生服务**: http://localhost:8082
- **教师服务**: http://localhost:8083
- **申请服务**: http://localhost:8086
- **事项服务**: http://localhost:8087

## 📊 数据库设计

### 核心表结构
- `users` - 用户基础信息
- `students` - 学生详细信息
- `teachers` - 教师详细信息
- `affairs` - 创新创业事项
- `affair_students` - 事项-学生关联
- `applications` - 学分申请
- `proof_materials` - 证明材料
- `innovation_credits` - 创新学分
- `competition_credits` - 竞赛学分
- `patent_credits` - 专利学分
- `paper_credits` - 论文学分
- `project_credits` - 项目学分

### 统一学分字段
所有学分子表都包含 `recognized_credits` 字段，用于统一管理已认定的学分。

## 🔧 API接口

### 用户管理
```
POST   /api/users/register     # 用户注册
POST   /api/users/login        # 用户登录
GET    /api/users/:username    # 获取用户信息
PUT    /api/users/:username    # 更新用户信息
DELETE /api/users/:username    # 删除用户
```

### 学生管理
```
POST   /api/students           # 创建学生
GET    /api/students           # 获取所有学生
GET    /api/students/:id       # 获取学生详情
PUT    /api/students/:id       # 更新学生信息
DELETE /api/students/:id       # 删除学生
GET    /api/students/search    # 搜索学生
```

### 教师管理
```
POST   /api/teachers           # 创建教师
GET    /api/teachers           # 获取所有教师
GET    /api/teachers/:username # 获取教师详情
PUT    /api/teachers/:username # 更新教师信息
DELETE /api/teachers/:username # 删除教师
GET    /api/teachers/department/:department # 按院系查询
GET    /api/teachers/title/:title           # 按职称查询
GET    /api/teachers/search                 # 搜索教师
GET    /api/teachers/active                 # 获取活跃教师
```

### 事项管理
```
POST   /api/affairs            # 创建事项
GET    /api/affairs            # 获取所有事项
GET    /api/affairs/:id        # 获取事项详情
PUT    /api/affairs/:id        # 更新事项
DELETE /api/affairs/:id        # 删除事项
POST   /api/affair-students    # 关联学生到事项
GET    /api/affair-students/:affairId # 获取事项的学生
```

### 申请管理
```
POST   /api/applications       # 创建申请
GET    /api/applications       # 获取所有申请
GET    /api/applications/:id   # 获取申请详情
PUT    /api/applications/:id   # 更新申请
DELETE /api/applications/:id   # 删除申请
POST   /api/applications/:id/review # 审核申请
GET    /api/applications/user/:userID # 获取用户的申请
GET    /api/applications/student/:studentID # 获取学生的申请
```

## 🧪 测试

### 运行测试
```bash
# 进入测试目录
cd tester

# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestUserService
go test -v -run TestStudentService
go test -v -run TestTeacherService
go test -v -run TestAffairService
go test -v -run TestApplicationService
```

### 测试覆盖
- 用户注册、登录、CRUD操作
- 学生信息管理
- 教师信息管理
- 事项创建和关联
- 申请创建和审核流程

## 🎨 前端功能

### 主要页面
- **登录/注册** - 用户认证
- **仪表板** - 系统概览和统计
- **申请管理** - 学分申请处理
- **学生管理** - 学生信息维护
- **教师管理** - 教师信息维护
- **事项管理** - 创新创业事项
- **个人资料** - 用户信息设置

### 技术特性
- 响应式设计
- 深色/浅色主题切换
- 实时通知
- 表单验证
- 数据可视化

## 🔒 安全特性

- JWT Token认证
- 密码加密存储
- CORS跨域配置
- 输入验证和过滤
- SQL注入防护

## 📈 监控和日志

- 服务健康检查
- 请求日志记录
- 错误追踪
- 性能监控

## 🚀 部署

### 生产环境
```bash
# 构建生产镜像
docker-compose -f docker-compose.prod.yml build

# 启动生产服务
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes部署
```bash
# 应用K8s配置
kubectl apply -f k8s/

# 查看部署状态
kubectl get pods
kubectl get services
```

## 🔧 开发指南

### 本地开发
```bash
# 启动数据库
docker-compose up postgres -d

# 运行后端服务
cd user-management-service && go run main.go
cd student-info-service && go run main.go
# ... 其他服务

# 运行前端
cd frontend && npm install && npm run dev
```

### 代码规范
- Go代码遵循gofmt规范
- TypeScript使用ESLint + Prettier
- 提交信息使用Conventional Commits

## 📝 更新日志

### v1.0.0 (2024-01-XX)
- ✅ 完成基础微服务架构
- ✅ 实现用户认证系统
- ✅ 完成学生信息管理
- ✅ 完成教师信息管理
- ✅ 实现事项管理功能
- ✅ 完成申请处理流程
- ✅ 统一学分字段设计
- ✅ 前端界面开发
- ✅ API网关实现
- ✅ 容器化部署

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 📄 许可证

MIT License

## 📞 联系方式

- 项目维护者: [Your Name]
- 邮箱: [your.email@example.com]
- 项目地址: [GitHub Repository URL]

---

**注意**: 这是一个开发中的项目，请在生产环境使用前进行充分测试。 