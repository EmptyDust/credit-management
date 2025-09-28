# 用户服务 Utils 使用指南

## 概述

用户服务的utils包提供了统一的工具函数，用于减少代码冗余和提高代码质量。参考活动服务的utils结构，我们创建了以下工具类：

## 文件结构

```
user-service/utils/
├── common.go      # 通用工具函数
├── response.go    # 统一响应处理
├── validator.go   # 数据验证工具
├── auth.go        # 认证和权限工具
├── middleware.go  # 中间件（已存在）
├── pointer.go     # 指针工具（已存在）
└── README.md      # 使用说明
```

## 使用方法

### 1. 响应处理 (response.go)

**替换前：**
```go
c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": data})
c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "错误信息", "data": nil})
```

**替换后：**
```go
import "credit-management/user-service/utils"

// 成功响应
utils.SendSuccessResponse(c, data)

// 错误响应
utils.SendBadRequest(c, "错误信息")
utils.SendUnauthorized(c)
utils.SendForbidden(c, "权限不足")
utils.SendNotFound(c, "用户不存在")
utils.SendConflict(c, "用户已存在")
utils.SendInternalServerError(c, err)
```

### 2. 数据验证 (validator.go)

**替换前：**
```go
// 分散在各个handler中的验证逻辑
if len(password) < 8 {
    c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码长度至少8位", "data": nil})
    return
}
```

**替换后：**
```go
import "credit-management/user-service/utils"

validator := utils.NewValidator()

// 验证密码
if err := validator.ValidatePassword(password); err != nil {
    utils.SendBadRequest(c, err.Error())
    return
}

// 验证邮箱
if err := validator.ValidateEmail(email); err != nil {
    utils.SendBadRequest(c, err.Error())
    return
}

// 验证手机号
if err := validator.ValidatePhone(phone); err != nil {
    utils.SendBadRequest(c, err.Error())
    return
}
```

### 3. 认证和权限 (auth.go)

**替换前：**
```go
// 分散的权限检查逻辑
claims, exists := c.Get("claims")
if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
    return
}
claimsMap, ok := claims.(jwt.MapClaims)
if !ok || claimsMap["user_type"] != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
    return
}
```

**替换后：**
```go
import "credit-management/user-service/utils"

// 获取用户信息
userID := utils.GetCurrentUserID(c)
userRole := utils.GetCurrentUserRole(c)
username := utils.GetUsername(c)

// 权限检查
if !utils.IsAdmin(userRole) {
    utils.SendForbidden(c, "需要管理员权限")
    return
}

// 组合权限检查
if !utils.IsTeacherOrAdmin(userRole) {
    utils.SendForbidden(c, "需要教师或管理员权限")
    return
}

// 获取用户声明
claims, exists := utils.GetUserClaims(c)
if !exists {
    utils.SendUnauthorized(c)
    return
}
```

### 4. 通用工具 (common.go)

```go
import "credit-management/user-service/utils"

// 获取环境变量
dbURL := utils.GetDatabaseURL()
port := utils.GetServerPort()

// 时间格式化
formattedTime := utils.FormatTime(time.Now())

// 字符串处理
if utils.IsEmptyString(username) {
    // 处理空字符串
}

// 兼容现有验证函数
if !utils.ValidatePasswordComplexity(password) {
    // 处理密码复杂度
}
```

## 重构建议

### 1. 逐步替换响应处理

在handlers中逐步替换手动构造的响应为统一的响应函数：

```go
// 替换前
c.JSON(http.StatusOK, gin.H{
    "code":    0,
    "message": "success",
    "data":    response,
})

// 替换后
utils.SendSuccessResponse(c, response)
```

### 2. 统一验证逻辑

将分散的验证逻辑集中到validator中：

```go
// 在handler中
validator := utils.NewValidator()

// 验证请求参数
if err := validator.ValidateEmail(req.Email); err != nil {
    utils.SendBadRequest(c, err.Error())
    return
}

if err := validator.ValidatePhone(req.Phone); err != nil {
    utils.SendBadRequest(c, err.Error())
    return
}
```

### 3. 使用认证工具

替换分散的权限检查逻辑：

```go
// 替换前
claims, exists := c.Get("claims")
if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证", "data": nil})
    return
}
claimsMap, ok := claims.(jwt.MapClaims)
if !ok || claimsMap["user_type"] != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足", "data": nil})
    return
}

// 替换后
if !utils.IsAdmin(utils.GetCurrentUserRole(c)) {
    utils.SendForbidden(c, "需要管理员权限")
    return
}
```

### 4. 使用通用工具函数

替换硬编码的配置和重复的工具函数：

```go
// 替换前
maxFileSize := 10 * 1024 * 1024 // 10MB

// 替换后
maxFileSize := utils.GetMaxFileSize()
```

## 优势

1. **代码复用**：减少重复代码，提高维护性
2. **统一标准**：确保所有响应格式一致
3. **易于测试**：工具函数可以独立测试
4. **易于扩展**：新增功能只需在utils中添加
5. **权限管理**：统一的权限检查逻辑

## 完整示例

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 权限检查
    if !utils.IsAdmin(utils.GetCurrentUserRole(c)) {
        utils.SendForbidden(c, "需要管理员权限")
        return
    }

    // 绑定请求
    var req models.UserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.SendBadRequest(c, "请求参数错误: "+err.Error())
        return
    }

    // 数据验证
    validator := utils.NewValidator()
    if err := validator.ValidateEmail(req.Email); err != nil {
        utils.SendBadRequest(c, err.Error())
        return
    }
    if err := validator.ValidatePassword(req.Password); err != nil {
        utils.SendBadRequest(c, err.Error())
        return
    }

    // 业务逻辑
    user := models.User{...}
    if err := h.db.Create(&user).Error; err != nil {
        utils.SendInternalServerError(c, err)
        return
    }

    // 成功响应
    utils.SendCreatedResponse(c, "用户创建成功", user)
}
``` 