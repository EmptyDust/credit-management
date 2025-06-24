# 用户注册修复测试文档

## 问题描述

在原始实现中，用户注册时只创建用户记录，没有同时创建对应的学生或教师记录，导致数据不一致。

## 修复方案

### 1. 后端修复（用户管理服务）

在用户注册成功后，根据用户类型自动创建对应的学生或教师记录：

- 修改 `user-management-service/handlers/user.go`
- 添加 `createUserProfile` 方法
- 添加 `createStudentProfile` 和 `createTeacherProfile` 方法
- 添加 `callExternalService` 方法调用外部服务

### 2. 前端修复（备用方案）

在前端注册逻辑中添加创建用户档案的调用：

- 修改 `frontend/src/pages/Register.tsx`
- 在用户注册成功后，根据用户类型调用对应的API创建档案

### 3. 配置修复

- 在 `docker-compose.yml` 中添加服务URL环境变量
- 在 `user-management-service/main.go` 中配置服务URL

## 测试步骤

### 1. 启动服务

```bash
docker-compose up -d
```

### 2. 测试学生注册

1. 访问前端注册页面：http://localhost:3000/register
2. 填写学生注册信息：
   - 用户名：test_student
   - 密码：Test123456
   - 邮箱：student@test.com
   - 真实姓名：测试学生
   - 用户类型：学生
3. 点击注册
4. 验证结果：
   - 检查用户表中是否有新用户记录
   - 检查学生表中是否有对应的学生记录
   - 验证用户ID和学生ID是否一致

### 3. 测试教师注册

1. 填写教师注册信息：
   - 用户名：test_teacher
   - 密码：Test123456
   - 邮箱：teacher@test.com
   - 真实姓名：测试教师
   - 用户类型：教师
2. 点击注册
3. 验证结果：
   - 检查用户表中是否有新用户记录
   - 检查教师表中是否有对应的教师记录
   - 验证用户ID和教师用户名是否一致

### 4. 数据库验证

```sql
-- 检查用户表
SELECT id, username, user_type, real_name FROM users WHERE username IN ('test_student', 'test_teacher');

-- 检查学生表
SELECT id, username, name FROM students WHERE username IN ('test_student', 'test_teacher');

-- 检查教师表
SELECT id, username, name FROM teachers WHERE username IN ('test_student', 'test_teacher');
```

## 预期结果

1. **用户注册成功**：用户记录创建成功
2. **档案创建成功**：对应的学生或教师记录创建成功
3. **数据一致性**：用户ID和档案记录关联正确
4. **错误处理**：如果档案创建失败，用户注册也会失败（后端方案）或继续但记录警告（前端方案）

## 回滚方案

如果修复出现问题，可以：

1. **回滚后端代码**：恢复 `user-management-service/handlers/user.go` 的原始版本
2. **回滚前端代码**：恢复 `frontend/src/pages/Register.tsx` 的原始版本
3. **回滚配置**：恢复 `docker-compose.yml` 的原始版本

## 注意事项

1. **服务依赖**：确保学生和教师服务在用户管理服务之前启动
2. **网络连接**：确保服务间网络连接正常
3. **错误处理**：添加适当的错误处理和日志记录
4. **性能考虑**：外部服务调用可能增加注册时间，需要设置合理的超时时间 