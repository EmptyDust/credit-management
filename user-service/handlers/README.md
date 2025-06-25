# User Service Handlers

用户服务的处理器模块，按功能进行了合理的拆分。

## 文件结构

### `user.go`
- **职责**: 定义核心结构体和构造函数
- **内容**: 
  - `UserHandler` 结构体定义
  - `NewUserHandler` 构造函数

### `validation.go`
- **职责**: 输入验证相关功能
- **内容**:
  - `validatePassword()` - 密码强度验证
  - `validatePhone()` - 手机号格式验证
  - `validateStudentID()` - 学号格式验证
  - `validateUserRequest()` - 统一用户请求验证

### `user_creation.go`
- **职责**: 用户创建相关功能
- **内容**:
  - `Register()` - 学生用户注册
  - `CreateTeacher()` - 管理员创建教师
  - `CreateStudent()` - 管理员创建学生

### `user_management.go`
- **职责**: 用户管理相关功能
- **内容**:
  - `GetUser()` - 获取用户信息
  - `UpdateUser()` - 更新用户信息
  - `DeleteUser()` - 删除用户
  - `GetAllUsers()` - 获取所有用户
  - `GetUsersByType()` - 根据类型获取用户

### `user_search.go`
- **职责**: 用户搜索和统计功能
- **内容**:
  - `SearchUsers()` - 搜索用户
  - `GetUserStats()` - 获取用户统计
  - `GetStudentStats()` - 获取学生统计
  - `GetTeacherStats()` - 获取教师统计

### `user_utils.go`
- **职责**: 工具方法
- **内容**:
  - `convertToUserResponse()` - 模型转换工具

## 设计原则

1. **单一职责**: 每个文件只负责一个特定的功能领域
2. **高内聚低耦合**: 相关功能聚集在一起，文件间依赖最小化
3. **可维护性**: 代码结构清晰，易于理解和维护
4. **可扩展性**: 新功能可以轻松添加到相应的文件中

## 使用方式

所有文件都在同一个 `handlers` 包中，可以相互调用方法。在 `main.go` 中只需要导入 `handlers` 包即可使用所有功能。

```go
import "credit-management/user-service/handlers"

// 创建处理器
userHandler := handlers.NewUserHandler(db)

// 使用各种方法
userHandler.Register(c)
userHandler.SearchUsers(c)
// ...
```