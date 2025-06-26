-- 数据库迁移脚本：升级到优化版本
-- 将现有数据库从旧版本升级到优化版本，包含约束修改、索引优化等

-- ========================================
-- 1. 备份提示
-- ========================================

DO $$
BEGIN
    RAISE NOTICE '开始数据库迁移到优化版本...';
    RAISE NOTICE '建议在执行此脚本前备份数据库！';
    RAISE NOTICE '迁移将修改表结构、添加约束、优化索引等';
END $$;

-- ========================================
-- 2. 用户表优化
-- ========================================

-- 2.1 修改用户名字段长度和格式约束
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(20);
ALTER TABLE users DROP CONSTRAINT IF EXISTS username_format;
ALTER TABLE users ADD CONSTRAINT username_format CHECK (username ~ '^[a-zA-Z0-9_]+$');

-- 2.2 修改手机号字段长度和格式约束
ALTER TABLE users ALTER COLUMN phone TYPE VARCHAR(11);
ALTER TABLE users DROP CONSTRAINT IF EXISTS phone_format;
ALTER TABLE users ADD CONSTRAINT phone_format CHECK (phone IS NULL OR phone ~ '^1[3-9]\d{9}$');

-- 2.3 修改学号字段长度和格式约束
ALTER TABLE users ALTER COLUMN student_id TYPE VARCHAR(8);
ALTER TABLE users DROP CONSTRAINT IF EXISTS student_id_format;
ALTER TABLE users ADD CONSTRAINT student_id_format CHECK (student_id IS NULL OR student_id ~ '^\d{8}$');

-- 2.4 修改年级字段长度和格式约束
ALTER TABLE users ALTER COLUMN grade TYPE VARCHAR(4);
ALTER TABLE users DROP CONSTRAINT IF EXISTS grade_format;
ALTER TABLE users ADD CONSTRAINT grade_format CHECK (grade IS NULL OR grade ~ '^\d{4}$');

-- 2.5 修改真实姓名字段长度约束
ALTER TABLE users ALTER COLUMN real_name TYPE VARCHAR(50);

-- 2.6 添加状态约束
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE users ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'inactive', 'suspended'));

-- ========================================
-- 3. 活动表优化
-- ========================================

-- 3.1 确保活动状态约束正确
ALTER TABLE credit_activities DROP CONSTRAINT IF EXISTS credit_activities_status_check;
ALTER TABLE credit_activities ADD CONSTRAINT credit_activities_status_check 
    CHECK (status IN ('draft', 'pending_review', 'approved', 'rejected'));

-- 3.2 添加标题和类别非空约束
ALTER TABLE credit_activities ADD CONSTRAINT credit_activities_title_check 
    CHECK (LENGTH(TRIM(title)) > 0);
ALTER TABLE credit_activities ADD CONSTRAINT credit_activities_category_check 
    CHECK (LENGTH(TRIM(category)) > 0);

-- 3.3 添加日期逻辑约束
ALTER TABLE credit_activities ADD CONSTRAINT credit_activities_date_check 
    CHECK (end_date >= start_date);

-- ========================================
-- 4. 参与者表优化
-- ========================================

-- 4.1 添加学分非负约束
ALTER TABLE activity_participants ADD CONSTRAINT activity_participants_credits_check 
    CHECK (credits >= 0);

-- ========================================
-- 5. 申请表优化
-- ========================================

-- 5.1 添加学分非负约束
ALTER TABLE applications ADD CONSTRAINT applications_applied_credits_check 
    CHECK (applied_credits >= 0);
ALTER TABLE applications ADD CONSTRAINT applications_awarded_credits_check 
    CHECK (awarded_credits >= 0);

-- ========================================
-- 6. 附件表优化
-- ========================================

-- 6.1 确保文件类型字段长度合适
ALTER TABLE attachments ALTER COLUMN file_type TYPE VARCHAR(20);

-- 6.2 确保文件类别字段长度合适
ALTER TABLE attachments ALTER COLUMN file_category TYPE VARCHAR(50);

-- 6.3 确保MD5哈希字段长度正确
ALTER TABLE attachments ALTER COLUMN md5_hash TYPE VARCHAR(32);

-- 6.4 添加文件大小约束
ALTER TABLE attachments ADD CONSTRAINT attachments_file_size_check 
    CHECK (file_size > 0 AND file_size <= 20971520);

-- 6.5 添加文件类型白名单约束
ALTER TABLE attachments ADD CONSTRAINT attachments_file_type_check 
    CHECK (file_type IN ('.pdf', '.doc', '.docx', '.txt', '.rtf', '.odt', '.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.mp4', '.avi', '.mov', '.wmv', '.flv', '.mp3', '.wav', '.ogg', '.aac', '.zip', '.rar', '.7z', '.tar', '.gz', '.xls', '.xlsx', '.csv', '.ppt', '.pptx'));

-- 6.6 添加文件类别约束
ALTER TABLE attachments ADD CONSTRAINT attachments_file_category_check 
    CHECK (file_category IN ('document', 'image', 'video', 'audio', 'archive', 'spreadsheet', 'presentation', 'other'));

-- 6.7 添加文件名非空约束
ALTER TABLE attachments ADD CONSTRAINT attachments_file_name_check 
    CHECK (LENGTH(TRIM(file_name)) > 0);
ALTER TABLE attachments ADD CONSTRAINT attachments_original_name_check 
    CHECK (LENGTH(TRIM(original_name)) > 0);

-- 6.8 添加下载次数非负约束
ALTER TABLE attachments ADD CONSTRAINT attachments_download_count_check 
    CHECK (download_count >= 0);

-- ========================================
-- 7. 创建验证函数
-- ========================================

-- 7.1 密码复杂度验证函数
CREATE OR REPLACE FUNCTION validate_password_complexity(password TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- 检查密码长度至少8位
    IF LENGTH(password) < 8 THEN
        RETURN FALSE;
    END IF;
    
    -- 检查是否包含大写字母
    IF password !~ '[A-Z]' THEN
        RETURN FALSE;
    END IF;
    
    -- 检查是否包含小写字母
    IF password !~ '[a-z]' THEN
        RETURN FALSE;
    END IF;
    
    -- 检查是否包含数字
    IF password !~ '[0-9]' THEN
        RETURN FALSE;
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- 7.2 邮箱格式验证函数
CREATE OR REPLACE FUNCTION validate_email_format(email TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN
        RETURN TRUE;
    END IF;
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 7.3 文件类型验证函数
CREATE OR REPLACE FUNCTION validate_file_type(file_type TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF file_type IN ('.pdf', '.doc', '.docx', '.txt', '.rtf', '.odt', '.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.mp4', '.avi', '.mov', '.wmv', '.flv', '.mp3', '.wav', '.ogg', '.aac', '.zip', '.rar', '.7z', '.tar', '.gz', '.xls', '.xlsx', '.csv', '.ppt', '.pptx') THEN
        RETURN TRUE;
    END IF;
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 7.4 文件大小验证函数
CREATE OR REPLACE FUNCTION validate_file_size(file_size BIGINT)
RETURNS BOOLEAN AS $$
BEGIN
    IF file_size > 0 AND file_size <= 20971520 THEN -- 20MB
        RETURN TRUE;
    END IF;
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 8. 创建用户数据验证触发器
-- ========================================

-- 8.1 用户数据验证触发器函数
CREATE OR REPLACE FUNCTION validate_user_data()
RETURNS TRIGGER AS $$
BEGIN
    -- 验证邮箱格式
    IF NOT validate_email_format(NEW.email) THEN
        RAISE EXCEPTION '邮箱格式不正确';
    END IF;
    
    -- 验证手机号格式（如果提供）
    IF NEW.phone IS NOT NULL AND NEW.phone !~ '^1[3-9]\d{9}$' THEN
        RAISE EXCEPTION '手机号格式不正确';
    END IF;
    
    -- 验证学号格式（如果提供）
    IF NEW.student_id IS NOT NULL AND NEW.student_id !~ '^\d{8}$' THEN
        RAISE EXCEPTION '学号格式不正确';
    END IF;
    
    -- 验证年级格式（如果提供）
    IF NEW.grade IS NOT NULL AND NEW.grade !~ '^\d{4}$' THEN
        RAISE EXCEPTION '年级格式不正确';
    END IF;
    
    -- 验证用户名格式
    IF NEW.username !~ '^[a-zA-Z0-9_]+$' THEN
        RAISE EXCEPTION '用户名只能包含字母、数字和下划线';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 8.2 创建触发器
DROP TRIGGER IF EXISTS trigger_validate_user_data ON users;
CREATE TRIGGER trigger_validate_user_data
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION validate_user_data();

-- ========================================
-- 9. 添加复合索引优化
-- ========================================

-- 9.1 用户表复合索引
CREATE INDEX IF NOT EXISTS idx_users_type_status ON users(user_type, status);
CREATE INDEX IF NOT EXISTS idx_users_college_major ON users(college, major);
CREATE INDEX IF NOT EXISTS idx_users_department_title ON users(department, title);

-- 9.2 活动表复合索引
CREATE INDEX IF NOT EXISTS idx_activities_owner_status ON credit_activities(owner_id, status);
CREATE INDEX IF NOT EXISTS idx_activities_category_status ON credit_activities(category, status);

-- ========================================
-- 10. 数据清理和修复
-- ========================================

-- 10.1 清理可能存在的无效数据
UPDATE users SET phone = NULL WHERE phone IS NOT NULL AND phone !~ '^1[3-9]\d{9}$';
UPDATE users SET student_id = NULL WHERE student_id IS NOT NULL AND student_id !~ '^\d{8}$';
UPDATE users SET grade = NULL WHERE grade IS NOT NULL AND grade !~ '^\d{4}$';

-- 10.2 修复用户名格式
UPDATE users SET username = REGEXP_REPLACE(username, '[^a-zA-Z0-9_]', '_', 'g') 
WHERE username !~ '^[a-zA-Z0-9_]+$';

-- 10.3 修复活动标题和类别
UPDATE credit_activities SET title = TRIM(title) WHERE title IS NOT NULL AND LENGTH(TRIM(title)) = 0;
UPDATE credit_activities SET category = TRIM(category) WHERE category IS NOT NULL AND LENGTH(TRIM(category)) = 0;

-- 10.4 修复附件文件名
UPDATE attachments SET file_name = TRIM(file_name) WHERE file_name IS NOT NULL AND LENGTH(TRIM(file_name)) = 0;
UPDATE attachments SET original_name = TRIM(original_name) WHERE original_name IS NOT NULL AND LENGTH(TRIM(original_name)) = 0;

-- ========================================
-- 11. 验证迁移结果
-- ========================================

-- 11.1 检查约束是否生效
DO $$
DECLARE
    constraint_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO constraint_count
    FROM information_schema.table_constraints 
    WHERE table_name = 'users' 
    AND constraint_type = 'CHECK';
    
    RAISE NOTICE '用户表CHECK约束数量: %', constraint_count;
END $$;

-- 11.2 检查字段长度
DO $$
DECLARE
    username_length INTEGER;
    phone_length INTEGER;
    student_id_length INTEGER;
    grade_length INTEGER;
BEGIN
    SELECT character_maximum_length INTO username_length
    FROM information_schema.columns 
    WHERE table_name = 'users' AND column_name = 'username';
    
    SELECT character_maximum_length INTO phone_length
    FROM information_schema.columns 
    WHERE table_name = 'users' AND column_name = 'phone';
    
    SELECT character_maximum_length INTO student_id_length
    FROM information_schema.columns 
    WHERE table_name = 'users' AND column_name = 'student_id';
    
    SELECT character_maximum_length INTO grade_length
    FROM information_schema.columns 
    WHERE table_name = 'users' AND column_name = 'grade';
    
    RAISE NOTICE '字段长度验证:';
    RAISE NOTICE '- username: %', username_length;
    RAISE NOTICE '- phone: %', phone_length;
    RAISE NOTICE '- student_id: %', student_id_length;
    RAISE NOTICE '- grade: %', grade_length;
END $$;

-- 11.3 验证触发器
DO $$
DECLARE
    trigger_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO trigger_count
    FROM information_schema.triggers 
    WHERE event_object_table = 'users';
    
    RAISE NOTICE '用户表触发器数量: %', trigger_count;
END $$;

-- ========================================
-- 12. 完成提示
-- ========================================

DO $$
BEGIN
    RAISE NOTICE '数据库迁移完成！';
    RAISE NOTICE '';
    RAISE NOTICE '迁移内容：';
    RAISE NOTICE '- 优化了用户表字段长度和格式约束';
    RAISE NOTICE '- 添加了活动表的业务逻辑约束';
    RAISE NOTICE '- 优化了附件表的文件类型和大小约束';
    RAISE NOTICE '- 添加了复合索引提升查询性能';
    RAISE NOTICE '- 创建了数据验证触发器和函数';
    RAISE NOTICE '- 清理了无效数据';
    RAISE NOTICE '';
    RAISE NOTICE '约束验证已与后端API和前端验证规则保持一致';
    RAISE NOTICE '建议运行测试脚本验证约束一致性';
END $$; 