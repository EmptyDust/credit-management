# 用户管理服务测试

这个目录包含了用户管理服务的测试代码。

## 测试内容

- 用户注册测试
- 用户登录测试（包含JWT token验证）
- 用户信息查询测试
- 集成测试（按顺序执行所有测试）

## 运行测试

### 方法1：使用 go test 命令
```bash
# 运行所有测试
go test -v .

# 运行特定测试
go test -v -run TestUserRegister
go test -v -run TestUserLogin
go test -v -run TestGetUser
go test -v -run TestUserManagementIntegration
```

### 方法2：使用测试运行器
```bash
go run main.go
```

## 测试前提条件

1. 确保用户管理服务正在运行（端口8001）
2. 确保PostgreSQL数据库已启动
3. 确保所有微服务都已通过 `docker compose up -d` 启动

## 测试端口说明

测试代码中使用的端口与docker-compose.yml中的端口映射一致：
- 用户管理服务：8001
- 认证服务：8003
- 事项管理服务：8004
- 通用申请服务：8005
- 学生信息服务：8006
- 教师信息服务：8007
- API网关：8000

## 测试结果

测试会输出详细的请求和响应信息，包括：
- HTTP状态码
- 响应体内容
- 错误信息（如果有）

如果测试失败，请检查：
1. 服务是否正常运行
2. 数据库连接是否正常
3. 端口是否正确
4. 网络连接是否正常 