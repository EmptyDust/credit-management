# API 统一重构总结

## 概述

本次重构将所有用户查询功能统一到 `/search/users` 端点，删除了冗余的代码，简化了 API 结构。

## 主要变更

### 后端变更

#### 删除的方法

- `GetAllUsers()` - 获取所有用户
- `GetUsersByType()` - 根据用户类型获取用户
- `GetStudents()` - 获取学生列表
- `GetTeachers()` - 获取教师列表

#### 保留的方法

- `SearchUsers()` - 统一用户搜索（支持所有查询功能）
- `GetUserStats()` - 用户统计信息
- `GetStudentStats()` - 学生统计信息
- `GetTeacherStats()` - 教师统计信息

#### 路由变更

- 删除了 `/users/type/:userType` 路由
- 删除了 `/students` 和 `/teachers` 的 GET 路由
- 保留了 `/search/users` 作为统一查询端点
- 保留了统计信息路由：`/users/stats/students` 和 `/users/stats/teachers`

### 前端变更

#### 更新的 API 调用

1. **学生页面 (Students.tsx)**

   - 查询：`/users/type/student` → `/search/users?user_type=student`
   - 创建：`/students` → `/users/students`
   - 更新：`/students/:id` → `/users/:id`
   - 导入：`/users/import/students` → `/users/import-csv` (添加 user_type 参数)
   - 导出：`/users/export/students` → `/users/export?user_type=student`

2. **教师页面 (Teachers.tsx)**

   - 查询：`/users/type/teacher` → `/search/users?user_type=teacher`
   - 创建：`/teachers` → `/users/teachers`
   - 更新：`/teachers/:id` → `/users/:id`
   - 导入：`/users/import/teachers` → `/users/import-csv` (添加 user_type 参数)
   - 导出：`/users/export/teachers` → `/users/export?user_type=teacher`

3. **用户服务 (userService.ts)**

   - `getStudents()`: 使用 `/search/users?user_type=student`
   - `getTeachers()`: 使用 `/search/users?user_type=teacher`
   - `searchUsers()`: 使用 `/search/users`
   - `searchTeachersByUsername()`: 使用 `/search/users?user_type=teacher&query=...`
   - 新增 `searchStudentsByUsername()`: 使用 `/search/users?user_type=student&query=...`

4. **Profile 页面 (Profile.tsx)**
   - 密码修改：`/users/change-password` → `/users/change_password`
   - 参数名：`current_password` → `old_password`

#### 保持不变的 API 调用

- `/users/profile` - 获取用户信息
- `/users/register` - 用户注册
- `/users/stats` - 用户统计
- `/search/users` - 活动参与者搜索

### API 网关变更

#### 删除的路由

- `GET /users/type/:userType`
- `GET /students`
- `GET /teachers`
- `GET /students/search`
- `GET /students/stats`
- `GET /teachers/search`
- `GET /teachers/stats`
- 各种特定的搜索路由

#### 保留的路由

- `GET /search/users` - 统一用户搜索
- `GET /users/stats/students` - 学生统计
- `GET /users/stats/teachers` - 教师统计
- 用户管理相关路由（创建、更新、删除）

## 新的统一 API 格式

### 搜索用户 API

```
GET /api/search/users
```

**查询参数：**

- `query` - 搜索关键词（用户名、邮箱、真实姓名、手机号）
- `user_type` - 用户类型过滤（student/teacher/admin）
- `college` - 学院过滤
- `major` - 专业过滤
- `class` - 班级过滤
- `grade` - 年级过滤
- `department` - 院系过滤
- `title` - 职称过滤
- `status` - 状态过滤
- `page` - 页码
- `page_size` - 每页数量

**响应格式：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [...],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10
  }
}
```

### 权限控制

- 学生：只能查看学生和教师的基本信息
- 教师：可以查看学生详细信息和其他教师基本信息
- 管理员：可以查看所有用户的所有信息

## 兼容性说明

由于不考虑向后兼容性，所有旧的 API 端点已被删除。前端代码已全部更新为使用新的统一 API 格式。

## 测试建议

1. 测试学生列表页面的查询、创建、更新、删除功能
2. 测试教师列表页面的查询、创建、更新、删除功能
3. 测试用户搜索功能（包括各种过滤条件）
4. 测试导入导出功能
5. 测试权限控制（不同用户类型的访问限制）
6. 测试统计信息功能

## 优势

1. **代码简化**：删除了大量重复的查询逻辑
2. **维护性提升**：统一的搜索逻辑，便于维护和扩展
3. **性能优化**：减少了重复的数据库查询
4. **API 一致性**：所有用户查询都使用相同的端点和格式
5. **权限统一**：统一的权限控制逻辑
