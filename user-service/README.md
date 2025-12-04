# 统一用户服务 (User Service)

这是合并后的统一用户服务，整合了原有的用户管理、学生信息和教师信息三个服务的功能。

## 功能特性

- 用户注册和管理
- 学生信息管理
- 教师信息管理
- 用户搜索和统计
- 权限控制

## 快速开始

### 环境要求

- Go 1.24+
- PostgreSQL
- Docker (可选)

### 本地运行

```bash
# 安装依赖
go mod download

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=credit_management
export JWT_SECRET=your-secret-key

# 运行服务
go run main.go
```

### Docker 运行

```bash
# 构建镜像
docker build -t user-service .

# 运行容器
docker run -d \
  --name user-service \
  -p 8084:8084 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  -e JWT_SECRET=your-jwt-secret \
  user-service
```

## API 文档

### 用户管理 API

#### 基础操作（所有认证用户）

```http
GET    /api/users/profile          # 获取当前用户信息
PUT    /api/users/profile          # 更新当前用户信息
GET    /api/users/{id}             # 获取指定用户详情
GET    /api/users/stats            # 获取用户统计信息
POST   /api/users/change_password  # 用户自助修改密码
GET    /api/users/activity         # 获取当前用户活动记录（预留）
GET    /api/users/{id}/activity    # 获取指定用户活动记录（预留）
```

#### 管理员专用操作

```http
POST   /api/users/teachers                    # 创建教师账户
POST   /api/users/students                    # 创建学生账户
PUT    /api/users/{id}                        # 更新用户信息
DELETE /api/users/{id}                        # 删除用户（软删除）
POST   /api/users/batch_delete                # 批量删除用户
POST   /api/users/batch_status                # 批量更新用户状态
POST   /api/users/reset_password              # 重置用户密码
GET    /api/users/export                      # 导出用户数据（支持JSON/CSV）
POST   /api/users/import-csv                  # CSV批量导入用户
GET    /api/users/csv-template                # 下载CSV导入模板
POST   /api/users/import                      # Excel批量导入用户（预留）
GET    /api/users/excel-template              # 下载Excel导入模板（预留）
```

#### 教师/管理员专用

```http
GET    /api/users/stats/students              # 获取学生统计信息
GET    /api/users/stats/teachers              # 获取教师统计信息
```

### 学生管理 API

```http
POST   /api/students/register                 # 学生自助注册（无需认证）
POST   /api/students                          # 管理员创建学生账户
PUT    /api/students/{id}                     # 更新学生信息
DELETE /api/students/{id}                     # 删除学生账户
GET    /api/students                          # 获取学生列表（教师/管理员）
GET    /api/students/{id}                     # 获取学生详情
GET    /api/students/stats                    # 获取学生统计
```

### 教师管理 API

```http
POST   /api/teachers                          # 创建教师账户（仅管理员）
PUT    /api/teachers/{id}                     # 更新教师信息（仅管理员）
DELETE /api/teachers/{id}                     # 删除教师账户（仅管理员）
GET    /api/teachers                          # 获取教师列表（仅管理员）
GET    /api/teachers/{id}                     # 获取教师详情
GET    /api/teachers/stats                    # 获取教师统计
```

### 搜索功能

```http
GET    /api/search/users                      # 搜索用户（支持姓名、用户名、学号等）
```

### 配置选项

```http
GET    /api/config/options                    # 获取系统配置选项（无需认证）
```

## 环境变量

| 变量名        | 默认值              | 说明            |
| ------------- | ------------------- | --------------- |
| `DB_HOST`     | `localhost`         | 数据库主机      |
| `DB_PORT`     | `5432`              | 数据库端口      |
| `DB_USER`     | `postgres`          | 数据库用户名    |
| `DB_PASSWORD` | `password`          | 数据库密码      |
| `DB_NAME`     | `credit_management` | 数据库名称      |
| `DB_SSLMODE`  | `disable`           | 数据库 SSL 模式 |
| `JWT_SECRET`  | `your-secret-key`   | JWT 密钥        |
| `PORT`        | `8084`              | 服务端口        |

## 健康检查

```
GET /health
```

## 新增用户管理 API（后端已实现，前端尚未实现）

### 1. 批量删除用户

- `POST /api/users/batch_delete`
- 权限：仅管理员
- 请求体：`{"ids": ["uuid1", "uuid2", ...]}`
- 返回：删除数量
- 说明：批量软删除用户。前端尚未实现。

### 2. 批量更新用户状态

- `POST /api/users/batch_status`
- 权限：仅管理员
- 请求体：`{"ids": ["uuid1", ...], "status": "active|inactive|suspended"}`
- 返回：更新数量及状态
- 说明：批量修改用户状态。前端尚未实现。

### 3. 重置用户密码

- `POST /api/users/reset_password`
- 权限：仅管理员
- 请求体：`{"id": "uuid", "new_password": "xxxx"}`
- 返回：成功消息
- 说明：管理员可重置任意用户密码。前端尚未实现。

### 4. 用户自助修改密码

- `POST /api/users/change_password`
- 权限：所有认证用户
- 请求体：`{"old_password": "xxx", "new_password": "yyy"}`
- 返回：成功消息
- 说明：用户可自助修改自己的密码。前端尚未实现。

### 5. 获取用户活动记录（预留）

- `GET /api/users/activity` 获取当前用户活动
- `GET /api/users/:id/activity` 获取指定用户活动
- 权限：所有认证用户（指定 id 需有权限）
- 返回：活动列表（目前为空）
- 说明：后端接口已预留，前端尚未实现。

### 6. 导出用户数据

- `GET /api/users/export?format=json&user_type=student&status=active`
- 权限：仅管理员
- 返回：用户数据（json，csv 待实现）
- 说明：支持按类型/状态导出。前端尚未实现。

## CSV 批量导入功能（后端已实现，前端尚未实现）

### 1. 用户 CSV 导入

- `POST /api/users/import-csv`
- 权限：仅管理员
- 请求：multipart/form-data，包含 file 字段（CSV 文件）和 user_type 字段（student/teacher）
- 功能：从 CSV 文件批量导入用户（学生或教师）
- 说明：支持最多 1000 行数据，文件大小限制 5MB。前端尚未实现。

### 2. 用户 CSV 模板下载

- `GET /api/users/csv-template?user_type=student`
- `GET /api/users/csv-template?user_type=teacher`
- 权限：仅管理员
- 功能：下载用户 CSV 导入模板
- 说明：提供学生和教师两种格式的 CSV 模板。前端尚未实现。

### CSV 格式要求

#### 学生 CSV 格式

必须包含以下列：

- username: 用户名（必需）
- password: 密码（必需，至少 8 位）
- email: 邮箱（必需）
- phone: 手机号（必需，11 位）
- real_name: 真实姓名（必需）
- student_id: 学号（必需，8 位数字）
- college: 学部（必需）
- major: 专业（必需）
- class: 班级（必需）
- grade: 年级（必需，4 位数字）

#### 教师 CSV 格式

必须包含以下列：

- username: 用户名（必需）
- password: 密码（必需，至少 8 位）
- email: 邮箱（必需）
- phone: 手机号（必需，11 位）
- real_name: 真实姓名（必需）
- department: 部门（必需）
- title: 职称（必需）

> 以上 CSV 导入功能均已在后端实现，前端页面和交互尚未开发。

> 以上 API 均已在后端实现，前端页面和交互尚未开发。
