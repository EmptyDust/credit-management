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
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    category VARCHAR(100) NOT NULL,
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
- `end_date`: 必须大于等于 `start_date`
- `status`: 只能是 'draft', 'pending_review', 'approved', 'rejected'
- `category`: 不能为空字符串

### 3. 活动参与者表 (activity_participants)

**表描述**: 记录活动与参与者的多对多关系

**字段定义**:

```sql
activity_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    id UUID NOT NULL,
    credits DECIMAL(5,2) NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

**约束条件**:

- `credits`: 必须大于等于 0
- 唯一约束: `(activity_id, id)` 在未删除状态下唯一

### 4. 申请表 (applications)

**表描述**: 记录学生申请学分的情况

**字段定义**:

```sql
applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    id UUID NOT NULL,
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

- `status`: 只能是 'approved'
- `applied_credits`, `awarded_credits`: 必须大于等于 0
- 唯一约束: `(activity_id, id)`

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

## 活动类型详情表

### 6. 创新创业实践活动详情表 (innovation_activity_details)

**表描述**: 存储创新创业实践活动的具体信息

```sql
innovation_activity_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    item VARCHAR(200),
    company VARCHAR(200),
    project_no VARCHAR(100),
    issuer VARCHAR(100),
    date DATE,
    total_hours DECIMAL(6,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

### 7. 学科竞赛学分详情表 (competition_activity_details)

**表描述**: 存储学科竞赛活动的具体信息

```sql
competition_activity_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    level VARCHAR(100),
    competition VARCHAR(200),
    award_level VARCHAR(100),
    rank VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

### 8. 大学生创业项目详情表 (entrepreneurship_project_details)

**表描述**: 存储大学生创业项目的具体信息

```sql
entrepreneurship_project_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    project_name VARCHAR(200),
    project_level VARCHAR(100),
    project_rank VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

### 9. 创业实践项目详情表 (entrepreneurship_practice_details)

**表描述**: 存储创业实践项目的具体信息

```sql
entrepreneurship_practice_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    company_name VARCHAR(200),
    legal_person VARCHAR(100),
    share_percent DECIMAL(5,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

### 10. 论文专利详情表 (paper_patent_details)

**表描述**: 存储论文专利活动的具体信息

```sql
paper_patent_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    name VARCHAR(200),
    category VARCHAR(100),
    rank VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
)
```

## 关系图

```
users (1) ←→ (N) credit_activities
users (1) ←→ (N) activity_participants
users (1) ←→ (N) applications
users (1) ←→ (N) attachments

credit_activities (1) ←→ (N) activity_participants
credit_activities (1) ←→ (N) applications
credit_activities (1) ←→ (N) attachments
credit_activities (1) ←→ (1) innovation_activity_details
credit_activities (1) ←→ (1) competition_activity_details
credit_activities (1) ←→ (1) entrepreneurship_project_details
credit_activities (1) ←→ (1) entrepreneurship_practice_details
credit_activities (1) ←→ (1) paper_patent_details
```

## 索引设计

### 用户表索引

- `idx_users_username`: 用户名索引
- `idx_users_email`: 邮箱索引
- `idx_users_user_type`: 用户类型索引
- `idx_users_status`: 状态索引
- `idx_users_student_id`: 学号索引
- `idx_users_deleted_at`: 删除时间索引
- `idx_users_type_status`: 用户类型+状态复合索引
- `idx_users_college_major`: 学院+专业复合索引
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

## 触发器设计

### 1. 时间戳更新触发器

- 所有表都有 `updated_at` 字段的自动更新触发器

### 2. 数据验证触发器

- `trigger_validate_user_data`: 用户数据格式验证

### 3. 业务逻辑触发器

- `trigger_generate_applications`: 活动通过后自动生成申请
- `trigger_delete_applications_on_withdraw`: 活动撤回时删除申请
- `trigger_cleanup_orphaned_attachments`: 清理孤立附件

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

## 存储过程

### 1. 数据验证函数

- `validate_password_complexity()`: 密码复杂度验证
- `validate_email_format()`: 邮箱格式验证
- `validate_file_type()`: 文件类型验证
- `validate_file_size()`: 文件大小验证

### 2. 业务操作函数

- `delete_activity_with_permission_check()`: 权限检查的活动删除
- `batch_delete_activities()`: 批量删除活动
- `restore_deleted_activity()`: 恢复已删除活动

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
