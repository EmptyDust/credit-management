# 数据库存储过程使用指南

## 概述

本文档详细说明了双创学分申请平台数据库中定义的各种存储过程的功能、参数、返回值和使用方法。

## 存储过程分类

### 1. 数据验证函数

#### 1.1 密码复杂度验证函数

**函数名**: `validate_password_complexity(password TEXT)`

**功能描述**: 验证密码是否符合复杂度要求

**参数**:

- `password TEXT`: 待验证的密码

**返回值**: `BOOLEAN`

- `TRUE`: 密码符合复杂度要求
- `FALSE`: 密码不符合复杂度要求

**验证规则**:

1. 密码长度至少 8 位
2. 必须包含大写字母
3. 必须包含小写字母
4. 必须包含数字

**使用示例**:

```sql
-- 验证密码复杂度
SELECT validate_password_complexity('Password123');  -- 返回 TRUE
SELECT validate_password_complexity('password');     -- 返回 FALSE (缺少大写字母和数字)
SELECT validate_password_complexity('PASS123');      -- 返回 FALSE (缺少小写字母)
SELECT validate_password_complexity('Pass');         -- 返回 FALSE (长度不足)

-- 在应用中使用
DO $$
BEGIN
    IF NOT validate_password_complexity('user_password') THEN
        RAISE EXCEPTION '密码不符合复杂度要求';
    END IF;
END $$;
```

#### 1.2 邮箱格式验证函数

**函数名**: `validate_email_format(email TEXT)`

**功能描述**: 验证邮箱格式是否符合标准

**参数**:

- `email TEXT`: 待验证的邮箱地址

**返回值**: `BOOLEAN`

- `TRUE`: 邮箱格式正确
- `FALSE`: 邮箱格式错误

**使用示例**:

```sql
-- 验证邮箱格式
SELECT validate_email_format('user@example.com');     -- 返回 TRUE
SELECT validate_email_format('user.name@domain.co.uk'); -- 返回 TRUE
SELECT validate_email_format('invalid-email');        -- 返回 FALSE
SELECT validate_email_format('user@');                -- 返回 FALSE

-- 在用户注册时使用
DO $$
BEGIN
    IF NOT validate_email_format('new_user@example.com') THEN
        RAISE EXCEPTION '邮箱格式不正确';
    END IF;
END $$;
```

#### 1.3 文件类型验证函数

**函数名**: `validate_file_type(file_type TEXT)`

**功能描述**: 验证文件类型是否在允许的白名单中

**参数**:

- `file_type TEXT`: 文件扩展名（包含点号）

**返回值**: `BOOLEAN`

- `TRUE`: 文件类型允许
- `FALSE`: 文件类型不允许

**支持的文件类型**:

- 文档: `.pdf`, `.doc`, `.docx`, `.txt`, `.rtf`, `.odt`
- 图片: `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`
- 视频: `.mp4`, `.avi`, `.mov`, `.wmv`, `.flv`
- 音频: `.mp3`, `.wav`, `.ogg`, `.aac`
- 压缩: `.zip`, `.rar`, `.7z`, `.tar`, `.gz`
- 表格: `.xls`, `.xlsx`, `.csv`
- 演示: `.ppt`, `.pptx`

**使用示例**:

```sql
-- 验证文件类型
SELECT validate_file_type('.pdf');    -- 返回 TRUE
SELECT validate_file_type('.docx');   -- 返回 TRUE
SELECT validate_file_type('.exe');    -- 返回 FALSE
SELECT validate_file_type('.php');    -- 返回 FALSE

-- 在文件上传时使用
DO $$
BEGIN
    IF NOT validate_file_type('.uploaded_file_extension') THEN
        RAISE EXCEPTION '不支持的文件类型';
    END IF;
END $$;
```

#### 1.4 文件大小验证函数

**函数名**: `validate_file_size(file_size BIGINT)`

**功能描述**: 验证文件大小是否在允许范围内

**参数**:

- `file_size BIGINT`: 文件大小（字节）

**返回值**: `BOOLEAN`

- `TRUE`: 文件大小在允许范围内
- `FALSE`: 文件大小超出限制

**限制条件**:

- 文件大小必须大于 0
- 文件大小不能超过 20MB (20,971,520 字节)

**使用示例**:

```sql
-- 验证文件大小
SELECT validate_file_size(1048576);      -- 1MB，返回 TRUE
SELECT validate_file_size(20971520);     -- 20MB，返回 TRUE
SELECT validate_file_size(20971521);     -- 20MB+1字节，返回 FALSE
SELECT validate_file_size(0);            -- 0字节，返回 FALSE
SELECT validate_file_size(-1);           -- 负数，返回 FALSE

-- 在文件上传时使用
DO $$
BEGIN
    IF NOT validate_file_size(1048576) THEN
        RAISE EXCEPTION '文件大小超出限制';
    END IF;
END $$;
```

### 2. 业务操作函数

#### 2.1 权限检查的活动删除函数

**函数名**: `delete_activity_with_permission_check(p_activity_id UUID, p_id UUID, p_user_type VARCHAR)`

**功能描述**: 检查用户权限并删除活动及其相关数据

**参数**:

- `p_activity_id UUID`: 要删除的活动 ID
- `p_id UUID`: 执行删除操作的用户 ID
- `p_user_type VARCHAR`: 执行删除操作的用户类型

**返回值**: `TEXT`

- `'活动删除成功'`: 删除成功
- `'活动不存在或已删除'`: 活动不存在
- `'无权限删除该活动'`: 权限不足

**权限规则**:

- 只有活动创建者可以删除自己的活动
- 管理员可以删除任何活动

**操作内容**:

1. 软删除活动记录
2. 软删除相关参与者记录
3. 软删除相关申请记录
4. 软删除相关附件记录

**使用示例**:

```sql
-- 活动创建者删除自己的活动
SELECT delete_activity_with_permission_check(
    'activity-uuid',
    'creator-uuid',
    'teacher'
);

-- 管理员删除活动
SELECT delete_activity_with_permission_check(
    'activity-uuid',
    'admin-uuid',
    'admin'
);

-- 无权限用户尝试删除活动
SELECT delete_activity_with_permission_check(
    'activity-uuid',
    'other-user-uuid',
    'student'
);
-- 返回: '无权限删除该活动'

-- 在应用中使用
DO $$
DECLARE
    v_result TEXT;
BEGIN
    v_result := delete_activity_with_permission_check(
        'activity-uuid',
        'current-user-uuid',
        'teacher'
    );

    IF v_result != '活动删除成功' THEN
        RAISE EXCEPTION '删除失败: %', v_result;
    END IF;
END $$;
```

#### 2.2 批量删除活动函数

**函数名**: `batch_delete_activities(p_activity_ids UUID[], p_id UUID, p_user_type VARCHAR)`

**功能描述**: 批量删除多个活动（仅管理员可用）

**参数**:

- `p_activity_ids UUID[]`: 要删除的活动 ID 数组
- `p_id UUID`: 执行删除操作的用户 ID
- `p_user_type VARCHAR`: 执行删除操作的用户类型

**返回值**: `INTEGER`

- 返回成功删除的活动数量

**权限规则**:

- 只有管理员可以执行批量删除操作

**操作内容**:

1. 验证用户权限
2. 遍历活动 ID 数组
3. 对每个活动执行软删除操作
4. 统计删除成功的活动数量

**使用示例**:

```sql
-- 管理员批量删除活动
SELECT batch_delete_activities(
    ARRAY['activity1-uuid', 'activity2-uuid', 'activity3-uuid'],
    'admin-uuid',
    'admin'
);
-- 返回删除成功的活动数量

-- 非管理员尝试批量删除（会抛出异常）
SELECT batch_delete_activities(
    ARRAY['activity1-uuid', 'activity2-uuid'],
    'teacher-uuid',
    'teacher'
);
-- 抛出异常: '只有管理员可以批量删除活动'

-- 在应用中使用
DO $$
DECLARE
    v_deleted_count INTEGER;
    v_activity_ids UUID[] := ARRAY['activity1-uuid', 'activity2-uuid'];
BEGIN
    v_deleted_count := batch_delete_activities(
        v_activity_ids,
        'admin-uuid',
        'admin'
    );

    RAISE NOTICE '成功删除 % 个活动', v_deleted_count;
END $$;
```

#### 2.3 恢复已删除活动函数

**函数名**: `restore_deleted_activity(p_activity_id UUID, p_user_type VARCHAR)`

**功能描述**: 恢复已删除的活动及其相关数据（仅管理员可用）

**参数**:

- `p_activity_id UUID`: 要恢复的活动 ID
- `p_user_type VARCHAR`: 执行恢复操作的用户类型

**返回值**: `TEXT`

- `'活动恢复成功'`: 恢复成功
- `'活动不存在'`: 活动不存在
- `'只有管理员可以恢复活动'`: 权限不足

**权限规则**:

- 只有管理员可以恢复已删除的活动

**操作内容**:

1. 恢复活动记录
2. 恢复相关参与者记录
3. 恢复相关申请记录
4. 恢复相关附件记录

**使用示例**:

```sql
-- 管理员恢复活动
SELECT restore_deleted_activity('activity-uuid', 'admin');
-- 返回: '活动恢复成功'

-- 非管理员尝试恢复活动
SELECT restore_deleted_activity('activity-uuid', 'teacher');
-- 返回: '只有管理员可以恢复活动'

-- 恢复不存在的活动
SELECT restore_deleted_activity('non-existent-uuid', 'admin');
-- 返回: '活动不存在'

-- 在应用中使用
DO $$
DECLARE
    v_result TEXT;
BEGIN
    v_result := restore_deleted_activity('activity-uuid', 'admin');

    IF v_result != '活动恢复成功' THEN
        RAISE EXCEPTION '恢复失败: %', v_result;
    END IF;
END $$;
```

## 存储过程管理

### 查看现有存储过程

```sql
-- 查看所有函数
SELECT
    routine_name,
    routine_type,
    data_type,
    parameter_name,
    parameter_mode,
    parameter_default
FROM information_schema.routines r
LEFT JOIN information_schema.parameters p ON r.routine_name = p.specific_name
WHERE r.routine_schema = 'public'
ORDER BY r.routine_name, p.ordinal_position;

-- 查看函数定义
SELECT
    routine_name,
    routine_definition
FROM information_schema.routines
WHERE routine_schema = 'public'
AND routine_type = 'FUNCTION';
```

### 删除存储过程

```sql
-- 删除函数
DROP FUNCTION IF EXISTS function_name(parameter_types);

-- 示例
DROP FUNCTION IF EXISTS validate_password_complexity(TEXT);
DROP FUNCTION IF EXISTS delete_activity_with_permission_check(UUID, UUID, VARCHAR);
```

### 修改存储过程

```sql
-- 使用 CREATE OR REPLACE FUNCTION 重新定义函数
CREATE OR REPLACE FUNCTION function_name(parameters)
RETURNS return_type AS $$
BEGIN
    -- 新的函数逻辑
END;
$$ LANGUAGE plpgsql;
```

## 最佳实践

### 1. 函数设计原则

- **单一职责**: 每个函数只负责一个特定功能
- **参数验证**: 在函数内部验证参数的有效性
- **错误处理**: 使用 RAISE EXCEPTION 处理错误情况
- **返回值明确**: 返回值的含义要清晰明确

### 2. 使用建议

- **数据验证**: 在数据插入/更新前使用验证函数
- **权限控制**: 使用权限检查函数确保操作安全
- **批量操作**: 使用批量操作函数提高效率
- **错误处理**: 在应用层处理函数返回的错误信息

### 3. 性能考虑

- **避免复杂查询**: 在函数中避免执行复杂的查询
- **合理使用索引**: 确保函数中使用的字段有适当的索引
- **减少网络往返**: 将多个操作封装在一个函数中

### 4. 安全考虑

- **参数化查询**: 避免在函数中使用字符串拼接
- **权限检查**: 在敏感操作前进行权限验证
- **输入验证**: 对所有输入参数进行验证

## 实际应用场景

### 场景 1: 用户注册流程

```sql
-- 1. 验证密码复杂度
IF NOT validate_password_complexity('user_password') THEN
    RAISE EXCEPTION '密码不符合复杂度要求';
END IF;

-- 2. 验证邮箱格式
IF NOT validate_email_format('user@example.com') THEN
    RAISE EXCEPTION '邮箱格式不正确';
END IF;

-- 3. 插入用户数据
INSERT INTO users (username, password, email, real_name, user_type)
VALUES ('newuser', 'hashed_password', 'user@example.com', '张三', 'student');
```

### 场景 2: 文件上传流程

```sql
-- 1. 验证文件类型
IF NOT validate_file_type('.pdf') THEN
    RAISE EXCEPTION '不支持的文件类型';
END IF;

-- 2. 验证文件大小
IF NOT validate_file_size(1048576) THEN
    RAISE EXCEPTION '文件大小超出限制';
END IF;

-- 3. 插入附件记录
INSERT INTO attachments (activity_id, file_name, original_name, file_size, file_type, uploaded_by)
VALUES ('activity-uuid', 'stored_filename.pdf', 'original_filename.pdf', 1048576, '.pdf', 'user-uuid');
```

### 场景 3: 活动管理流程

```sql
-- 1. 删除活动（带权限检查）
SELECT delete_activity_with_permission_check(
    'activity-uuid',
    'current-user-uuid',
    'teacher'
);

-- 2. 批量删除活动（管理员）
SELECT batch_delete_activities(
    ARRAY['activity1-uuid', 'activity2-uuid'],
    'admin-uuid',
    'admin'
);

-- 3. 恢复已删除活动（管理员）
SELECT restore_deleted_activity('activity-uuid', 'admin');
```

## 故障排除

### 常见问题

1. **函数不存在**

   ```sql
   -- 检查函数是否存在
   SELECT routine_name FROM information_schema.routines
   WHERE routine_name = 'function_name';
   ```

2. **参数类型不匹配**

   ```sql
   -- 检查函数参数类型
   SELECT parameter_name, data_type
   FROM information_schema.parameters
   WHERE specific_name = 'function_name';
   ```

3. **权限不足**
   ```sql
   -- 检查用户权限
   SELECT current_user, session_user;
   ```

### 调试方法

```sql
-- 启用详细日志
SET log_statement = 'all';
SET log_min_messages = 'notice';

-- 在函数中添加调试信息
RAISE NOTICE '参数值: %', parameter_value;
```

## 总结

存储过程是数据库功能扩展的重要工具，通过合理使用存储过程可以：

- 封装复杂的业务逻辑
- 提高数据操作的安全性
- 减少应用层代码复杂度
- 提高系统性能

在使用存储过程时，需要：

- 合理设计函数接口
- 做好错误处理和权限控制
- 注意性能影响
- 保持代码的可维护性
  </rewritten_file>
