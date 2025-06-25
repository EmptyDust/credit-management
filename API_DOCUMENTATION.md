# 学分管理系统 API 文档

## 概述

学分管理系统是一个基于微服务架构的学分申请和管理平台，包含多个独立的服务模块。

## 系统架构

系统由以下微服务组成：

- **affair-management-service** (端口: 8082) - 事务管理服务
- **application-management-service** (端口: 8083) - 申请管理服务
- **student-info-service** (端口: 8085) - 学生信息服务
- **user-management-service** (端口: 8084) - 用户管理服务
- **auth-service** (端口: 8081) - 认证服务
- **teacher-info-service** (端口: 8086) - 教师信息服务
- **api-gateway** (端口: 8080) - API 网关

## 统一返回格式

所有服务的 API 接口都使用统一的返回格式：

### 成功响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体的数据内容
  }
}
```

### 错误响应格式
```json
{
  "code": 400,  // 错误码
  "message": "错误描述",
  "data": null
}
```

### 错误码说明
- `400` - 请求参数错误
- `401` - 未认证
- `403` - 权限不足
- `404` - 资源不存在
- `500` - 服务器内部错误

## 服务文档

### 1. 事务管理服务
- **服务名称**: affair-management-service
- **端口**: 8082
- **基础路径**: `/api/affairs`
- **功能**: 管理学分申请相关的事务，包括事务的创建、查询、更新、删除等操作
- **详细文档**: [affair-management-service-API.md](API_DOCS/affair-management-service-API.md)

### 2. 申请管理服务
- **服务名称**: application-management-service
- **端口**: 8083
- **基础路径**: `/api/applications`
- **功能**: 处理学分申请的管理，包括申请的创建、查询、更新、删除等操作
- **详细文档**: [application-management-service-API.md](API_DOCS/application-management-service-API.md)

### 3. 学生信息服务
- **服务名称**: student-info-service
- **端口**: 8085
- **基础路径**: `/api/students`
- **功能**: 管理学生基本信息，包括学生信息的创建、查询、更新、删除等操作
- **详细文档**: [student-info-service-API.md](API_DOCS/student-info-service-API.md)

### 4. 用户管理服务
- **服务名称**: user-management-service
- **端口**: 8084
- **基础路径**: `/api/users`
- **功能**: 处理用户的基本信息管理，包括用户的创建、查询、更新、删除等操作
- **详细文档**: [user-management-service-API.md](API_DOCS/user-management-service-API.md)

### 5. 认证服务
- **服务名称**: auth-service
- **端口**: 8081
- **基础路径**: `/api/auth`
- **功能**: 处理用户认证和授权，包括用户登录、令牌管理、权限验证等操作
- **详细文档**: [auth-service-API.md](API_DOCS/auth-service-API.md)

### 6. 教师信息服务
- **服务名称**: teacher-info-service
- **端口**: 8086
- **基础路径**: `/api/teachers`
- **功能**: 管理教师基本信息，包括教师信息的创建、查询、更新、删除等操作
- **详细文档**: [teacher-info-service-API.md](API_DOCS/teacher-info-service-API.md)

## 数据模型关系

### 核心实体关系

1. **User** (用户)
   - 基础用户信息
   - 关联 Student 或 Teacher

2. **Student** (学生)
   - 学生详细信息
   - 关联 User (user_id)

3. **Teacher** (教师)
   - 教师详细信息
   - 关联 User (user_id)

4. **Affair** (事务)
   - 学分申请事务
   - 关联多个 Student (通过 AffairStudent)
   - 关联多个 Application

5. **Application** (申请)
   - 学分申请
   - 关联 Affair (affair_id)
   - 关联 Student (student_id)
   - 关联 User (user_id)

6. **Permission** (权限)
   - 系统权限定义
   - 关联 User (通过 UserPermission)

### 关系图

```
User (1) -----> (1) Student
   |
   +----> (1) Teacher

Affair (1) -----> (N) AffairStudent (N) -----> (1) Student
   |
   +----> (N) Application (N) -----> (1) Student
                              |
                              +----> (1) User

User (N) -----> (N) Permission (通过 UserPermission)
```

## 认证和授权

### 认证流程

1. 用户通过 `/api/auth/login` 登录
2. 获取访问令牌 (JWT)
3. 后续请求在 Authorization 头部携带令牌

### 权限控制

- 基于角色的访问控制 (RBAC)
- 细粒度权限控制
- 支持资源级别的权限管理

### 令牌管理

- 访问令牌有效期：24小时
- 刷新令牌有效期：7天
- 支持令牌自动刷新
- 支持令牌撤销

## 错误处理

所有服务都遵循统一的错误处理机制：

- 统一的错误码定义
- 详细的错误描述
- 一致的错误响应格式

## 开发指南

### 环境要求

- Go 1.19+
- PostgreSQL 13+
- Redis (可选，用于缓存)

### 启动服务

1. 启动数据库
2. 运行数据库迁移
3. 启动各个微服务
4. 启动 API 网关

### 测试

每个服务都提供了完整的测试脚本，位于 `tester/` 目录下。

## 部署

### Docker 部署

所有服务都提供了 Dockerfile，支持容器化部署。

### Kubernetes 部署

提供了 Kubernetes 部署配置文件，位于 `k8s/` 目录下。

## 监控和日志

- 统一的日志格式
- 结构化日志输出
- 支持日志聚合和分析

## 安全考虑

- JWT 令牌认证
- 密码加密存储
- 输入验证和清理
- CORS 配置
- 请求频率限制

## 版本控制

- API 版本通过 URL 路径控制
- 向后兼容性保证
- 版本迁移指南

## 联系信息

如有问题或建议，请联系开发团队。