# 学分管理系统 API 文档

## 概述
学分管理系统是一个基于微服务架构的完整系统，包含用户管理、认证授权、学生信息、教师信息、事务管理和申请管理等核心功能。

## 系统架构

### 微服务组成
1. **认证服务 (auth-service)** - 用户认证、授权和权限管理
2. **统一用户服务 (user-service)** - 用户、学生、教师信息统一管理
3. **事务管理服务 (affair-management-service)** - 学分事务管理
4. **申请管理服务 (application-management-service)** - 学分申请管理
5. **API网关 (api-gateway)** - 统一入口和路由管理

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
2. **[统一用户服务 API](./user-service-API.md)** - 用户、学生、教师信息统一管理
3. **[事务管理服务 API](./affair-management-service-API.md)** - 事务创建、管理和审核
4. **[申请管理服务 API](./application-management-service-API.md)** - 申请创建、管理和审核

## 主要功能

### 用户管理
- 用户注册（仅限学生）
- 管理员创建学生和教师
- 用户信息更新和删除
- 密码重置和管理
- 统一的学生和教师信息管理

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
- Go 1.24+

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
- 统一用户服务: 8084

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
- 统一用户服务测试
- 事务管理测试
- 申请管理测试

### 测试脚本位置
- `tester/` 目录包含所有服务的测试脚本
- 支持PowerShell和Bash环境
- 包含完整的API测试用例

## 服务合并说明

### 统一用户服务
原有的三个服务（用户管理、学生信息、教师信息）已合并为统一的用户服务，提供：
- 统一的用户管理API
- 简化的服务架构
- 更好的数据一致性
- 减少的服务间通信

### 迁移指南
详细的迁移说明请参考：[USER_SERVICE_MERGE_SUMMARY.md](../USER_SERVICE_MERGE_SUMMARY.md)

## 更新日志

### v2.0.0 (最新)
- 合并用户管理、学生信息、教师信息服务为统一用户服务
- 简化系统架构
- 优化API接口
- 提升系统性能

### v1.0.0
- 初始版本发布
- 微服务架构设计
- 基础功能实现

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