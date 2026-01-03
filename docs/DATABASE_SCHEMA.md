# 双创学分申请平台数据库关系模式

## 概述

本文档描述了双创学分申请平台的完整数据库关系模式，包括表结构、字段定义、约束关系、索引设计和业务逻辑。

## 数据库技术栈

- **数据库类型**: PostgreSQL 15+
- **字符编码**: UTF-8
- **时区设置**: Asia/Shanghai
- **扩展**: uuid-ossp (UUID 生成)

## 核心表结构

### 1. 用户表 (users)

**表描述**: 统一管理所有用户信息，包括学生、教师和管理员

**字段定义**:

```sql
users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(20) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(11) UNIQUE,
    real_name VARCHAR(50) NOT NULL,
    user_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    avatar VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    register_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    -- 学生特有字段
    student_id VARCHAR(8) UNIQUE,
    college VARCHAR(100),
    major VARCHAR(100),
    class VARCHAR(50),
    grade VARCHAR(4),

    -- 教师特有字段
    department VARCHAR(100),
    title VARCHAR(50)
)
```

**约束条件**:

- `username`: 只能包含字母、数字和下划线
- `email`: 必须符合邮箱格式
- `phone`: 11 位数字，以 1 开头
- `user_type`: 只能是 'student', 'teacher', 'admin'
- `status`: 只能是 'active', 'inactive', 'suspended'
- `student_id`: 8 位数字
- `grade`: 4 位数字

### 2. 学分活动表 (credit_activities)

**表描述**: 存储所有学分活动的基本信息

**字段定义**:

```sql
credit_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    category VARCHAR(100) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 扩展字段，存储活动类型特定信息
    owner_id UUID NOT NULL,
    reviewer_id UUID,
    review_comments TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

**约束条件**:

- `title`: 不能为空字符串
- `end_date`: 必须大于等于 `start_date`（由应用层验证）
- `status`: 只能是 'draft', 'pending_review', 'approved', 'rejected'
- `category`: 不能为空字符串，必须在预定义的类别列表中
- `details`: JSONB 字段，存储不同活动类型的扩展信息，由应用层根据 category 决定结构
- 外键约束: `owner_id` 和 `reviewer_id` 引用 `users(id)`

### 3. 活动参与者表 (activity_participants)

**表描述**: 记录活动与参与者的多对多关系

**字段定义**:

```sql
activity_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    user_id UUID NOT NULL,  -- 注意：字段名是user_id，存储用户UUID
    credits DECIMAL(5,2) NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

**约束条件**:

- `credits`: 必须大于等于 0 且小于等于 10
- 外键约束: `activity_id` 引用 `credit_activities(id)`，级联删除
- 唯一约束: `(activity_id, user_id)` 在未删除状态下唯一
- 参与者限制: 只有学生用户可以作为参与者（由应用层验证）

### 4. 申请表 (applications)

**表描述**: 记录学生申请学分的情况（自动生成，无需手动创建）

**字段定义**:

```sql
applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    user_id UUID NOT NULL,  -- 注意：字段名是user_id，存储用户UUID
    status VARCHAR(20) NOT NULL DEFAULT 'approved',
    applied_credits DECIMAL(5,2) NOT NULL,
    awarded_credits DECIMAL(5,2) NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

**约束条件**:

- `status`: 固定为 'approved'（自动生成时设置）
- `applied_credits`, `awarded_credits`: 必须大于等于 0，通常相等
- 外键约束: `activity_id` 引用 `credit_activities(id)`，级联删除
- 唯一约束: `(activity_id, user_id)` 在未删除状态下唯一
- **自动生成**: 申请记录由系统在活动审核通过时自动创建，无需手动提交

### 5. 附件表 (attachments)

**表描述**: 存储活动相关的文件附件

**字段定义**:

```sql
attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    file_category VARCHAR(50) NOT NULL,
    description TEXT,
    uploaded_by UUID NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    download_count INTEGER NOT NULL DEFAULT 0,
    md5_hash VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

**约束条件**:

- `file_name`, `original_name`: 不能为空字符串
- `file_size`: 大于 0 且小于等于 20MB
- `file_type`: 必须在支持的文件类型白名单中
- `file_category`: 只能是预定义的类型
- `download_count`: 必须大于等于 0

## 活动详情存储策略

为了简化 schema 并降低迁移成本，历史上的五类详情表（创新创业、学科竞赛、大学生创业、创业实践、论文专利）已经移除。现在所有扩展字段统一写入 `credit_activities.details` JSONB 字段或 `description` 中，由后端根据 `category` 决定序列化结构并在应用层做校验。

## 关系图

```
users (1) ←→ (N) credit_activities
users (1) ←→ (N) activity_participants
users (1) ←→ (N) applications
users (1) ←→ (N) attachments

credit_activities (1) ←→ (N) activity_participants
credit_activities (1) ←→ (N) applications
credit_activities (1) ←→ (N) attachments
```

## 索引设计

### 用户表索引

- `idx_users_username`: 用户名索引
- `idx_users_email`: 邮箱索引
- `idx_users_user_type`: 用户类型索引
- `idx_users_status`: 状态索引
- `idx_users_student_id`: 学号索引
- `idx_users_teacher_id`: 工号索引
- `idx_users_deleted_at`: 删除时间索引
- `idx_users_type_status`: 用户类型+状态复合索引
- `idx_users_college_major`: 学部+专业复合索引
- `idx_users_department_title`: 部门+职称复合索引

### 活动表索引

- `idx_credit_activities_status`: 状态索引
- `idx_credit_activities_owner_id`: 创建者索引
- `idx_credit_activities_deleted_at`: 删除时间索引
- `idx_activities_owner_status`: 创建者+状态复合索引
- `idx_activities_category_status`: 类别+状态复合索引

### 参与者表索引

- `idx_activity_participants_activity_id`: 活动 ID 索引
- `idx_activity_participants_id`: 用户 ID 索引
- `idx_activity_participants_deleted_at`: 删除时间索引
- `uniq_activity_participants_active`: 活动+用户唯一索引（未删除）

### 申请表索引

- `idx_applications_activity_id`: 活动 ID 索引
- `idx_applications_id`: 用户 ID 索引
- `idx_applications_status`: 状态索引
- `idx_applications_deleted_at`: 删除时间索引

### 附件表索引

- `idx_attachments_activity_id`: 活动 ID 索引
- `idx_attachments_uploaded_by`: 上传者索引
- `idx_attachments_file_category`: 文件类别索引
- `idx_attachments_file_type`: 文件类型索引
- `idx_attachments_md5_hash`: MD5 哈希索引
- `idx_attachments_deleted_at`: 删除时间索引

## 应用层自动化逻辑

**重要说明**：系统已移除数据库触发器和存储过程，所有业务逻辑在应用层实现。

### 申请自动生成

- **实现位置**: `credit-activity-service/handlers/activity_side_effects.go`
- **触发时机**: 活动状态从非 `approved` 变为 `approved` 时
- **逻辑说明**: 自动为活动的所有参与者创建申请记录，申请的学分继承参与者的学分设置

### 申请自动删除

- **实现位置**: `credit-activity-service/handlers/activity_side_effects.go`
- **触发时机**: 活动从 `approved` 状态变为其他状态时
- **逻辑说明**: 自动软删除该活动的所有相关申请，保持数据一致性

### 文件清理

- **实现位置**: `credit-activity-service/handlers/activity_side_effects.go`
- **触发时机**: 活动删除时
- **逻辑说明**: 检测并清理孤立的附件文件（基于 MD5 哈希），如果文件未被其他附件引用则删除物理文件

### 时间戳自动更新

- **实现方式**: 使用 GORM 的 `autoCreateTime` 和 `autoUpdateTime` 标签
- **说明**: 自动维护 `created_at` 和 `updated_at` 字段

数据库现在只保留必要的约束、索引和视图，所有业务规则、数据校验以及文件清理均由后端代码负责，便于统一控制、调试和扩展。

## 视图设计

### 1. 用户信息视图

- `student_basic_info`: 学生基本信息
- `student_detail_info`: 学生详细信息
- `student_complete_info`: 学生完整信息
- `teacher_basic_info`: 教师基本信息
- `teacher_detail_info`: 教师详细信息
- `teacher_complete_info`: 教师完整信息

### 2. 业务视图

- `detailed_credit_activity_view`: 详细活动信息（包含参与者）
- `detailed_applications_view`: 详细申请信息（包含申请人和活动）

## 数据完整性约束

### 1. 实体完整性

- 所有表都有主键约束
- 使用 UUID 作为主键，确保全局唯一性

### 2. 参照完整性

- 外键约束确保数据一致性
- 级联删除和设置 NULL 策略

### 3. 域完整性

- 字段类型和长度约束
- CHECK 约束确保数据有效性
- 正则表达式验证格式

### 4. 业务完整性

- 唯一性约束防止重复数据
- 触发器维护业务规则
- 软删除机制保护数据

## 性能优化

### 1. 索引策略

- 为常用查询字段创建索引
- 复合索引优化多条件查询
- 部分索引减少索引大小

### 2. 查询优化

- 视图简化复杂查询
- 触发器减少应用层逻辑
- 软删除避免数据丢失

### 3. 存储优化

- UUID 扩展提供高效 ID 生成
- 合理的字段类型选择
- 索引覆盖查询减少 IO

## 安全设计

### 1. 数据验证

- 输入格式严格验证
- 文件类型白名单
- 大小限制防止攻击

### 2. 权限控制

- 数据库级权限检查
- 操作审计日志
- 软删除保护数据

### 3. 数据保护

- 密码加密存储
- 敏感信息脱敏
- 备份恢复机制

## 扩展性设计

### 1. 表结构扩展

- 预留字段支持功能扩展
- 详情表支持不同类型活动
- 软删除支持数据恢复

### 2. 性能扩展

- 索引支持查询优化
- 分区表支持大数据量
- 读写分离支持高并发

### 3. 功能扩展

- 触发器支持业务规则变更
- 存储过程支持复杂逻辑
- 视图支持报表需求

## 默认数据

### 初始用户

- 管理员: `admin/adminpassword`
- 教师: `teacher/adminpassword`
- 学生: `student/adminpassword`

## 维护建议

### 1. 定期维护

- 索引重建优化性能
- 统计信息更新
- 日志清理

### 2. 监控指标

- 查询性能监控
- 存储空间监控
- 连接数监控

### 3. 备份策略

- 定期全量备份
- 增量备份
- 异地备份

---

_本文档描述了双创学分申请平台的完整数据库关系模式，为系统开发和维护提供参考。_
