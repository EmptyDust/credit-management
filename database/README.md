# 数据库管理文档

## 概述

本项目使用 PostgreSQL 作为数据库，通过 Docker 容器进行部署和管理。

## 数据库结构

### 主要表

1. **users** - 用户表

   - 存储所有用户信息（学生、教师、管理员）
   - 支持软删除
   - 包含用户类型特定字段

2. **credit_activities** - 学分活动表

   - 存储学分活动信息
   - 包含活动类型、学分点数、时间等信息

3. **activity_participants** - 活动参与者表

   - 记录用户参与活动的情况
   - 包含参与状态、获得的学分等

4. **applications** - 申请表

   - 存储用户提交的申请
   - 包含申请类型、状态、审核信息等

5. **权限相关表**
   - `roles` - 角色表
   - `permissions` - 权限表
   - `user_roles` - 用户角色关联表
   - `user_permissions` - 用户权限关联表
   - `role_permissions` - 角色权限关联表
   - `permission_groups` - 权限组表
   - `permission_group_permissions` - 权限组权限关联表

### 数据库视图

为了支持权限控制，创建了以下视图：

- `student_basic_info` - 学生基本信息视图
- `teacher_basic_info` - 教师基本信息视图
- `student_detail_info` - 学生详细信息视图
- `teacher_detail_info` - 教师详细信息视图
- `user_stats_view` - 用户统计视图
- `student_stats_view` - 学生统计视图
- `teacher_stats_view` - 教师统计视图
- `activity_stats_view` - 活动统计视图
- `participant_stats_view` - 参与者统计视图
- `application_stats_view` - 申请统计视图

## 应用层业务逻辑

**重要说明**：系统已移除数据库触发器和存储过程，所有业务逻辑在应用层实现，便于统一控制、调试和扩展。

### 自动化流程实现位置

所有自动化逻辑都在对应的微服务中实现：

| 功能         | 实现位置                                                    | 说明                                      |
| ------------ | ----------------------------------------------------------- | ----------------------------------------- |
| 申请自动生成 | `credit-activity-service/handlers/activity_side_effects.go` | 活动审核通过时自动生成申请                |
| 申请自动删除 | `credit-activity-service/handlers/activity_side_effects.go` | 活动状态变更时自动删除申请                |
| 文件清理     | `credit-activity-service/handlers/activity_side_effects.go` | 活动删除时清理孤立文件                    |
| 用户数据验证 | `user-service/utils/validator.go`                           | 用户名、邮箱、学号等格式验证              |
| 时间戳更新   | GORM 自动维护                                               | 使用 `autoCreateTime` 和 `autoUpdateTime` |

### 数据库层职责

数据库现在只负责：

- 表结构定义
- 索引优化
- 外键约束
- 唯一性约束
- CHECK 约束
- 视图定义

所有业务规则、数据校验、自动化流程均由应用层代码实现。

## 部署

### 使用 Docker Compose

```bash
# 启动所有服务（包括数据库）
docker-compose up -d

# 仅启动数据库
docker-compose up -d postgres

# 查看数据库日志
docker-compose logs postgres
```

### 数据库连接信息

- **主机**: localhost
- **端口**: 5432
- **数据库名**: credit_management
- **用户名**: postgres
- **密码**: password

### 连接示例

```bash
# 使用psql连接
psql -h localhost -p 5432 -U postgres -d credit_management

# 使用Docker容器连接
docker exec -it credit_management_postgres psql -U postgres -d credit_management
```

## 初始化

### 自动初始化

数据库容器启动时会自动执行 `init.sql` 脚本，该脚本包含：

1. 创建所有表结构
2. 创建索引和约束
3. 创建数据库视图
4. 插入初始数据（默认用户等）

**注意**：系统已移除数据库触发器和存储过程，所有业务逻辑在应用层实现。

### 手动初始化

如果需要手动执行初始化脚本：

```bash
# 连接到数据库
docker exec -it credit_management_postgres psql -U postgres -d credit_management

# 执行初始化脚本
\i /docker-entrypoint-initdb.d/init.sql
```

## 使用示例

### 数据库操作示例

```sql
-- 查询活动列表
SELECT * FROM credit_activities WHERE deleted_at IS NULL;

-- 查询活动的参与者
SELECT ap.*, u.real_name, u.username
FROM activity_participants ap
JOIN users u ON ap.user_id = u.id
WHERE ap.activity_id = 'your-activity-id' AND ap.deleted_at IS NULL;

-- 查询用户的申请
SELECT a.*, ca.title as activity_title
FROM applications a
JOIN credit_activities ca ON a.activity_id = ca.id
WHERE a.user_id = 'your-user-id' AND a.deleted_at IS NULL;

-- 统计活动数量
SELECT status, COUNT(*) as count
FROM credit_activities
WHERE deleted_at IS NULL
GROUP BY status;
```

**注意**：业务逻辑操作（如创建申请、同步数据等）应通过应用层 API 进行，不建议直接在数据库层面操作。

## 测试

### 运行测试脚本

为了验证存储过程和触发器是否正常工作，提供了完整的测试脚本：

```bash
# 连接到数据库
docker exec -it credit_management_postgres psql -U postgres -d credit_management

# 运行测试脚本
\i /docker-entrypoint-initdb.d/test_stored_procedures.sql
```

### 测试内容

测试脚本包含以下测试项目：

1. **触发器功能测试**

   - 活动创建和参与者添加
   - 活动信息更新同步
   - 参与者学分更新同步

2. **存储过程测试**

   - 活动更新同步存储过程

3. **删除触发器测试**

   - 活动删除时申请同步删除

4. **用户数据验证测试**

   - 用户名唯一性验证
   - 邮箱唯一性验证

5. **数据清理**
   - 自动清理测试数据

## 文件结构

```
database/
├── init.sql                    # 完整初始化脚本
├── test_stored_procedures.sql  # 测试脚本
├── backup.sh                   # 数据库备份脚本
├── restore.sh                  # 数据库恢复脚本
├── Dockerfile                  # 数据库容器配置
└── README.md                   # 本文档
```

## 备份和恢复

### 备份数据库

```bash
# 执行备份脚本
./database/backup.sh

# 备份文件保存在 database/backups/ 目录
```

### 恢复数据库

```bash
# 执行恢复脚本
./database/restore.sh backup_file.sql
```

## 注意事项

1. **数据一致性**：系统通过触发器确保活动和申请数据的一致性
2. **软删除**：删除操作使用软删除，数据不会物理删除
3. **唯一性约束**：用户名、邮箱等字段有唯一性约束
4. **权限控制**：通过视图和权限表实现细粒度权限控制
5. **测试环境**：建议在测试环境中充分测试后再部署到生产环境
