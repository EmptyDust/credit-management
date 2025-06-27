# 数据库触发器使用指南

## 概述

本文档详细说明了双创学分申请平台数据库中定义的各种触发器的功能、触发条件和使用方法。

## 触发器分类

### 1. 时间戳更新触发器

#### 功能描述

自动更新表中 `updated_at` 字段为当前时间戳，确保数据修改时间的准确性。

#### 涉及的触发器

```sql
-- 所有核心表的更新时间戳触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_credit_activities_updated_at BEFORE UPDATE ON credit_activities FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activity_participants_updated_at BEFORE UPDATE ON activity_participants FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON applications FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_attachments_updated_at BEFORE UPDATE ON attachments FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 活动类型详情表的更新时间戳触发器
CREATE TRIGGER update_innovation_activity_details_updated_at BEFORE UPDATE ON innovation_activity_details FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_competition_activity_details_updated_at BEFORE UPDATE ON competition_activity_details FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entrepreneurship_project_details_updated_at BEFORE UPDATE ON entrepreneurship_project_details FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entrepreneurship_practice_details_updated_at BEFORE UPDATE ON entrepreneurship_practice_details FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_paper_patent_details_updated_at BEFORE UPDATE ON paper_patent_details FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

#### 触发条件

- **触发时机**: BEFORE UPDATE
- **触发范围**: 每行数据更新时
- **执行函数**: `update_updated_at_column()`

#### 使用示例

```sql
-- 更新用户信息时，updated_at 会自动更新
UPDATE users SET real_name = '新姓名' WHERE user_id = 'some-uuid';

-- 更新活动状态时，updated_at 会自动更新
UPDATE credit_activities SET status = 'approved' WHERE id = 'some-uuid';
```

#### 实际效果

- 无需手动设置 `updated_at` 字段
- 确保数据修改时间的准确性
- 便于数据审计和追踪

### 2. 用户数据验证触发器

#### 功能描述

在插入或更新用户数据时，自动验证数据格式的合法性。

#### 触发器定义

```sql
CREATE TRIGGER trigger_validate_user_data
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION validate_user_data();
```

#### 验证规则

1. **邮箱格式验证**: 必须符合标准邮箱格式
2. **手机号格式验证**: 11 位数字，以 1 开头
3. **学号格式验证**: 8 位数字
4. **年级格式验证**: 4 位数字
5. **用户名格式验证**: 只能包含字母、数字和下划线

#### 触发条件

- **触发时机**: BEFORE INSERT OR UPDATE
- **触发范围**: 每行数据插入或更新时
- **执行函数**: `validate_user_data()`

#### 使用示例

```sql
-- 正确的用户数据插入
INSERT INTO users (username, password, email, real_name, user_type, student_id, grade)
VALUES ('student001', 'password123', 'student@example.com', '张三', 'student', '20210001', '2021');

-- 错误的用户数据插入（会触发异常）
INSERT INTO users (username, password, email, real_name, user_type, student_id, grade)
VALUES ('student@001', 'password123', 'invalid-email', '张三', 'student', '2021001', '202');
-- 错误: 用户名只能包含字母、数字和下划线
```

#### 实际效果

- 防止无效数据进入数据库
- 确保数据格式的一致性
- 减少应用层的验证负担

### 3. 活动申请自动生成触发器

#### 功能描述

当活动状态从非"approved"变为"approved"时，自动为所有参与者生成申请记录。

#### 触发器定义

```sql
CREATE TRIGGER trigger_generate_applications
    AFTER UPDATE ON credit_activities
    FOR EACH ROW
    EXECUTE FUNCTION generate_applications_on_activity_approval();
```

#### 业务逻辑

1. 检查活动状态是否从非"approved"变为"approved"
2. 查找该活动的所有参与者
3. 为每个参与者生成申请记录（如果不存在）
4. 申请状态默认为"approved"
5. 申请学分等于参与学分

#### 触发条件

- **触发时机**: AFTER UPDATE
- **触发范围**: 活动状态更新时
- **执行函数**: `generate_applications_on_activity_approval()`

#### 使用示例

```sql
-- 1. 首先创建活动和参与者
INSERT INTO credit_activities (title, description, start_date, end_date, status, category, owner_id)
VALUES ('创新创业大赛', '比赛描述', '2024-01-01', '2024-01-31', 'draft', 'competition', 'teacher-uuid');

INSERT INTO activity_participants (activity_id, user_id, credits)
VALUES ('activity-uuid', 'student1-uuid', 2.0),
       ('activity-uuid', 'student2-uuid', 2.0);

-- 2. 当活动被批准时，自动生成申请
UPDATE credit_activities SET status = 'approved' WHERE id = 'activity-uuid';

-- 3. 触发器会自动执行，生成以下申请记录：
-- INSERT INTO applications (activity_id, user_id, status, applied_credits, awarded_credits)
-- VALUES ('activity-uuid', 'student1-uuid', 'approved', 2.0, 2.0),
--        ('activity-uuid', 'student2-uuid', 'approved', 2.0, 2.0);
```

#### 实际效果

- 自动化申请流程
- 减少手动操作错误
- 确保数据一致性

### 4. 活动撤回申请删除触发器

#### 功能描述

当活动状态变为"draft"时，自动删除所有相关的申请记录。

#### 触发器定义

```sql
CREATE TRIGGER trigger_delete_applications_on_withdraw
    AFTER UPDATE ON credit_activities
    FOR EACH ROW
    EXECUTE FUNCTION delete_applications_on_activity_withdraw();
```

#### 业务逻辑

1. 检查活动状态是否变为"draft"
2. 删除该活动的所有申请记录
3. 确保申请与活动状态的一致性

#### 触发条件

- **触发时机**: AFTER UPDATE
- **触发范围**: 活动状态更新时
- **执行函数**: `delete_applications_on_activity_withdraw()`

#### 使用示例

```sql
-- 当活动被撤回时，自动删除相关申请
UPDATE credit_activities SET status = 'draft' WHERE id = 'activity-uuid';

-- 触发器会自动执行：
-- DELETE FROM applications WHERE activity_id = 'activity-uuid';
```

#### 实际效果

- 维护数据一致性
- 防止无效申请记录
- 简化活动管理流程

### 5. 附件清理触发器

#### 功能描述

当附件被删除时，检查是否需要删除物理文件，避免存储空间浪费。

#### 触发器定义

```sql
CREATE TRIGGER trigger_cleanup_orphaned_attachments
    BEFORE UPDATE ON attachments
    FOR EACH ROW
    EXECUTE FUNCTION cleanup_orphaned_attachments();
```

#### 业务逻辑

1. 检查附件是否被软删除
2. 通过 MD5 哈希查找是否有其他活动使用相同文件
3. 如果没有其他活动使用，记录需要删除的物理文件路径
4. 注意：实际文件删除需要在应用层处理

#### 触发条件

- **触发时机**: BEFORE UPDATE
- **触发范围**: 附件记录更新时
- **执行函数**: `cleanup_orphaned_attachments()`

#### 使用示例

```sql
-- 软删除附件
UPDATE attachments SET deleted_at = NOW() WHERE id = 'attachment-uuid';

-- 触发器会检查并输出类似以下信息：
-- NOTICE: 需要删除孤立文件: uploads/attachments/filename.pdf
```

#### 实际效果

- 识别孤立文件
- 优化存储空间
- 提供清理建议

## 触发器管理

### 查看现有触发器

```sql
-- 查看所有触发器
SELECT
    trigger_name,
    event_manipulation,
    event_object_table,
    action_statement
FROM information_schema.triggers
WHERE trigger_schema = 'public'
ORDER BY event_object_table, trigger_name;
```

### 禁用触发器

```sql
-- 临时禁用触发器
ALTER TABLE users DISABLE TRIGGER trigger_validate_user_data;

-- 重新启用触发器
ALTER TABLE users ENABLE TRIGGER trigger_validate_user_data;
```

### 删除触发器

```sql
-- 删除触发器
DROP TRIGGER IF EXISTS trigger_name ON table_name;
```

## 最佳实践

### 1. 触发器设计原则

- **单一职责**: 每个触发器只负责一个特定功能
- **性能考虑**: 避免在触发器中执行复杂查询
- **错误处理**: 合理处理异常情况
- **日志记录**: 重要操作需要记录日志

### 2. 使用建议

- **数据验证**: 使用 BEFORE 触发器进行数据验证
- **业务逻辑**: 使用 AFTER 触发器执行业务逻辑
- **级联操作**: 使用触发器维护数据一致性
- **审计追踪**: 使用触发器记录数据变更

### 3. 注意事项

- **性能影响**: 触发器会影响 INSERT/UPDATE/DELETE 性能
- **调试困难**: 触发器错误可能难以调试
- **维护成本**: 需要定期检查和维护触发器
- **测试覆盖**: 确保触发器逻辑得到充分测试

## 故障排除

### 常见问题

1. **触发器不执行**

   - 检查触发器是否启用
   - 验证触发条件是否正确
   - 确认函数是否存在且正确

2. **触发器执行错误**

   - 查看数据库日志
   - 检查函数逻辑
   - 验证数据约束

3. **性能问题**
   - 优化触发器函数
   - 减少不必要的查询
   - 考虑使用索引

### 调试方法

```sql
-- 启用详细日志
SET log_statement = 'all';
SET log_min_messages = 'notice';

-- 查看触发器执行情况
SELECT * FROM pg_stat_activity WHERE query LIKE '%trigger%';
```

## 总结

触发器是数据库自动化的重要工具，通过合理使用触发器可以：

- 确保数据一致性和完整性
- 自动化业务流程
- 减少应用层代码复杂度
- 提高系统可靠性

在使用触发器时，需要平衡自动化程度和系统性能，确保触发器的逻辑清晰、可维护。
