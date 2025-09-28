# 数据库管理文档

## 概述

本项目使用PostgreSQL作为数据库，通过Docker容器进行部署和管理。

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

## 存储过程和触发器

### 活动管理存储过程

1. **update_activity_and_sync_applications(activity_id, title, description, credit_points)**
   - 更新活动信息并同步相关申请
   - 确保活动和申请数据一致性
   - 原子性操作，避免数据不一致

### 活动管理触发器

系统实现了完整的活动-申请同步机制：

1. **trigger_activity_update_sync**
   - 监听活动表更新事件
   - 自动同步更新相关申请表的标题、描述、学分
   - 确保数据一致性

2. **trigger_activity_delete_sync**
   - 监听活动表删除事件
   - 自动软删除相关申请
   - 保持数据完整性

3. **trigger_participant_change_sync**
   - 监听参与者表变更事件
   - 当参与者状态变为approved时自动创建申请
   - 当学分更新时同步更新申请学分

4. **trigger_validate_user_data**
   - 用户数据验证触发器
   - 确保用户名、邮箱唯一性
   - 验证必填字段

### 触发器功能说明

- **自动同步**：活动信息变更时，相关申请会自动更新
- **数据一致性**：确保活动和申请数据始终保持同步
- **软删除**：活动删除时，相关申请会被软删除而非物理删除
- **状态管理**：参与者状态变更时自动处理申请流程
- **数据验证**：用户数据插入和更新时自动验证唯一性和完整性

## 部署

### 使用Docker Compose

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
2. 创建索引
3. 创建触发器函数和触发器
4. 创建存储过程
5. 创建数据库视图
6. 插入初始数据

### 手动初始化

如果需要手动执行初始化脚本：

```bash
# 连接到数据库
docker exec -it credit_management_postgres psql -U postgres -d credit_management

# 执行初始化脚本
\i /docker-entrypoint-initdb.d/init.sql
```

## 使用示例

### 调用活动更新存储过程

```sql
-- 更新活动信息并同步申请
SELECT update_activity_and_sync_applications(1, '新活动标题', '新活动描述', 5);
```

### 触发器测试

```sql
-- 更新活动信息，观察申请表自动更新
UPDATE credit_activities SET title = '新标题' WHERE id = 1;

-- 更新参与者状态，观察申请自动创建
UPDATE activity_participants SET status = 'approved' WHERE activity_id = 1 AND id = 'user-id-1';

-- 测试用户数据验证
INSERT INTO users (username, password, email, real_name, user_type)
VALUES ('test_user', 'password', 'test@example.com', '测试用户', 'student');
```

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