# 学分管理系统 API 文档

## 概述
学分管理系统是一个基于微服务架构的完整系统，包含用户管理、认证授权、学生信息、教师信息、事务管理和申请管理等核心功能。

## 系统架构

### 微服务组成
1. **认证服务 (auth-service)** - 用户认证、授权和权限管理
2. **用户管理服务 (user-management-service)** - 用户基础信息管理
3. **学生信息服务 (student-info-service)** - 学生档案信息管理
4. **教师信息服务 (teacher-info-service)** - 教师档案信息管理
5. **事务管理服务 (affair-management-service)** - 学分事务管理
6. **申请管理服务 (application-management-service)** - 学分申请管理
7. **API网关 (api-gateway)** - 统一入口和路由管理

### 技术栈
- **后端**: Go (Gin框架)
- **数据库**: PostgreSQL
- **认证**: JWT
- **容器化**: Docker
- **前端**: React + TypeScript

## 权限模型

### 用户角色
- **admin**: 管理员，拥有所有权限
- **teacher**: 教师，可以审核事务和申请
- **student**: 学生，可以参与事务和申请

### 权限控制
- 基于角色的访问控制 (RBAC)
- 细粒度的资源权限管理
- 数据过滤和访问控制

## API 文档

### 核心服务文档
1. **[认证服务 API](./auth-service-API.md)** - 用户认证、授权和权限验证
2. **[用户管理服务 API](./user-management-service-API.md)** - 用户创建、管理和删除
3. **[学生信息服务 API](./student-info-service-API.md)** - 学生信息管理
4. **[教师信息服务 API](./teacher-info-service-API.md)** - 教师信息管理
5. **[事务管理服务 API](./affair-management-service-API.md)** - 事务创建、管理和审核
6. **[申请管理服务 API](./application-management-service-API.md)** - 申请创建、管理和审核

## 主要功能

### 用户管理
- 用户注册（仅限学生）
- 管理员创建学生和教师
- 用户信息更新和删除
- 密码重置和管理

### 认证授权
- JWT token认证
- 权限验证和角色控制
- Token刷新和登出
- 细粒度权限管理

### 学生管理
- 学生信息创建和更新
- 基于角色的数据访问控制
- 学生信息搜索和统计
- 批量导入功能

### 教师管理
- 教师信息创建和更新
- 基于角色的数据访问控制
- 教师信息搜索和统计
- 批量导入功能

### 事务管理
- 事务创建和审核
- 参与者管理
- 附件上传和管理
- 事务状态流转

### 申请管理
- 申请创建和审核
- 批量申请处理
- 申请状态管理
- 与事务管理集成

## 数据流程

### 用户注册流程
1. 学生注册 → 创建用户账号
2. 管理员创建学生/教师 → 创建用户和档案信息
3. 用户登录 → 获取JWT token
4. 权限验证 → 基于角色控制访问

### 事务管理流程
1. 用户创建事务 → 状态：pending（待审核）
2. 教师/管理员审核 → 状态：approved/rejected
3. 学生加入事务 → 状态：active（进行中）
4. 事务完成 → 状态：completed（已完成）

### 申请管理流程
1. 用户创建申请 → 状态：pending（待审核）
2. 教师/管理员审核 → 状态：approved/rejected
3. 申请通过 → 状态：completed（已完成）
4. 用户获得学分

## 安全特性

### 认证安全
- JWT token认证
- Token过期机制
- 密码加密存储
- 防止暴力破解

### 授权安全
- 基于角色的权限控制
- 资源级别的访问控制
- 数据过滤和脱敏
- 操作日志记录

### 数据安全
- 数据库连接加密
- 敏感数据加密
- SQL注入防护
- XSS攻击防护

## 部署说明

### 环境要求
- Docker & Docker Compose
- PostgreSQL 数据库
- Go 1.21+

### 快速启动
```bash
# 克隆项目
git clone <repository-url>

# 进入项目目录
cd credit-management

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 服务端口
- API网关: 8080
- 认证服务: 8081
- 事务管理服务: 8082
- 申请管理服务: 8083
- 用户管理服务: 8084
- 学生信息服务: 8085
- 教师信息服务: 8086

## 开发指南

### API调用示例
```bash
# 用户登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 获取学生列表
curl -X GET http://localhost:8080/api/students \
  -H "Authorization: Bearer <token>"
```

### 错误处理
所有API都使用统一的错误响应格式：
```json
{
  "code": 400,
  "message": "错误描述",
  "data": null
}
```

### 成功响应
所有API都使用统一的成功响应格式：
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 测试

### 自动化测试
项目包含完整的自动化测试脚本：
- 认证服务测试
- 用户管理测试
- 学生信息测试
- 教师信息测试
- 事务管理测试
- 申请管理测试

### 测试脚本位置
```
tester/
├── test-auth-service.ps1
├── test-user-management-service.ps1
├── test-student-info-service.ps1
├── test-teacher-info-service.ps1
├── test-affair-management-service.ps1
└── test-application-management-service.ps1
```

## 监控和日志

### 日志管理
- 结构化日志输出
- 错误日志记录
- 操作审计日志
- 性能监控日志

### 健康检查
每个服务都提供健康检查端点：
```
GET /health
```

## 常见问题

### Q: 如何重置管理员密码？
A: 使用用户管理服务的密码重置功能，或直接操作数据库。

### Q: 如何添加新的用户角色？
A: 修改认证服务的权限模型和相关的权限验证逻辑。

### Q: 如何扩展事务类型？
A: 在事务管理服务中添加新的事务类型和相应的处理逻辑。

### Q: 如何备份数据？
A: 使用PostgreSQL的备份工具或Docker卷备份。

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 完整的微服务架构
- 基于角色的权限控制
- 事务和申请管理功能
- 文件上传和管理
- 自动化测试脚本

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

如有问题或建议，请通过以下方式联系：
- 项目Issues: [GitHub Issues](https://github.com/your-repo/issues)
- 邮箱: your-email@example.com 