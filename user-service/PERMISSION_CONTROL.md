# 用户服务权限控制系统

## 概述

本系统实现了基于用户类型（user_type）的访问控制，确保不同用户类型的用户只能访问其权限范围内的数据。

## 权限矩阵

| 用户类型 | 学生信息 | 教师信息 | 管理员信息 | 备注 |
|---------|---------|---------|-----------|------|
| 学生 | 基本信息 | 基本信息 | 无权限 | 只能查看基本信息 |
| 教师 | 详细信息 | 基本信息 | 无权限 | 可查看学生详细信息 |
| 管理员 | 所有信息 | 所有信息 | 所有信息 | 完全访问权限 |

## 实现方案

### 1. 数据库视图层

创建了多个数据库视图来预过滤数据：

- `student_basic_info`: 学生基本信息视图
- `teacher_basic_info`: 教师基本信息视图  
- `student_detail_info`: 学生详细信息视图
- `teacher_detail_info`: 教师详细信息视图
- `user_stats_view`: 用户统计视图
- `student_stats_view`: 学生统计视图
- `teacher_stats_view`: 教师统计视图

### 2. 应用层权限控制

#### 中间件权限检查

- `AuthRequired()`: 认证中间件
- `AdminOnly()`: 仅管理员访问
- `StudentOnly()`: 仅学生访问
- `TeacherOnly()`: 仅教师访问
- `StudentTeacherOrAdmin()`: 学生、教师、管理员访问
- `TeacherOrAdmin()`: 教师、管理员访问

#### 响应结构体

根据权限级别定义了不同的响应结构体：

- `StudentBasicResponse`: 学生基本信息响应
- `TeacherBasicResponse`: 教师基本信息响应
- `StudentDetailResponse`: 学生详细信息响应
- `TeacherDetailResponse`: 教师详细信息响应

### 3. 权限控制工具函数

#### 用户类型获取函数

```go
func getCurrentUserRole(c *gin.Context) string
func getCurrentUserID(c *gin.Context) string
```

#### 响应转换函数

```go
func (h *UserHandler) convertToStudentBasicResponse(user models.User) models.StudentBasicResponse
func (h *UserHandler) convertToTeacherBasicResponse(user models.User) models.TeacherBasicResponse
func (h *UserHandler) convertToStudentDetailResponse(user models.User) models.StudentDetailResponse
func (h *UserHandler) convertToTeacherDetailResponse(user models.User) models.TeacherDetailResponse
func (h *UserHandler) convertToRoleBasedResponse(user models.User, currentUserRole string) interface{}
```

#### 权限检查函数

```go
func canViewUserDetails(currentUserRole, targetUserType string) bool
```

## API 权限控制

### 用户搜索 API

**端点**: `GET /api/search/users`

**权限控制**:
- 学生: 只能搜索学生和教师的基本信息
- 教师: 可以搜索学生详细信息和其他教师基本信息
- 管理员: 可以搜索所有用户的所有信息

**查询参数**:
- `user_type`: 用户类型过滤
- `query`: 搜索关键词
- `college`, `major`, `class`, `grade`: 学生相关过滤
- `department`, `title`: 教师相关过滤
- `status`: 状态过滤
- `page`, `page_size`: 分页参数

### 用户列表 API

**端点**: `GET /api/users/type/:userType`

**权限控制**:
- 学生: 只能查看学生和教师的基本信息
- 教师: 可以查看学生详细信息和其他教师基本信息
- 管理员: 可以查看所有用户的所有信息

### 用户详情 API

**端点**: `GET /api/users/:id`

**权限控制**:
- 查看自己的信息: 所有认证用户都可以查看自己的完整信息
- 查看他人信息: 根据用户类型权限过滤显示内容

## 路由配置

```go
// 学生、教师和管理员可以访问的路由（基于用户类型的权限控制）
studentTeacherOrAdmin := auth.Group("")
studentTeacherOrAdmin.Use(permissionMiddleware.StudentTeacherOrAdmin())
{
    studentTeacherOrAdmin.GET("/:id", userHandler.GetUser)                   // 获取指定用户信息（基于用户类型过滤）
    studentTeacherOrAdmin.GET("/type/:userType", userHandler.GetUsersByType) // 根据用户类型获取用户（基于用户类型过滤）
}
```

## 测试

### 权限控制测试脚本

运行 `tester/test-permission-control.ps1` 来测试权限控制功能：

```powershell
.\tester\test-permission-control.ps1
```

### 测试覆盖范围

1. **学生权限测试**
   - 搜索学生信息（基本信息）
   - 搜索教师信息（基本信息）
   - 搜索管理员信息（被拒绝）

2. **教师权限测试**
   - 搜索学生信息（详细信息）
   - 搜索教师信息（基本信息）
   - 搜索管理员信息（被拒绝）

3. **管理员权限测试**
   - 搜索所有用户信息（所有信息）

4. **权限边界测试**
   - 无效token测试
   - 无token测试
   - 越权访问测试

## 安全考虑

1. **数据脱敏**: 敏感信息（如密码、身份证号等）不会在任何响应中返回
2. **权限验证**: 每个API调用都会验证用户权限
3. **SQL注入防护**: 使用参数化查询防止SQL注入
4. **JWT验证**: 严格的JWT token验证
5. **错误信息**: 不泄露敏感的系统信息

## 扩展性

系统设计支持以下扩展：

1. **新增用户类型**: 可以轻松添加新的用户类型
2. **权限细化**: 可以进一步细化权限控制
3. **动态权限**: 可以支持动态权限配置
4. **审计日志**: 可以添加权限访问审计日志

## 部署说明

1. 启动服务时会自动创建数据库视图
2. 确保数据库用户有创建视图的权限
3. 建议在生产环境中使用HTTPS
4. 定期更新JWT密钥 