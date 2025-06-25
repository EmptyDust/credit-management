-- 学分管理系统数据库初始化脚本
-- 简化版本：移除复杂的RBAC系统，只保留基于user_type的权限控制

-- ========================================
-- 1. 扩展和设置
-- ========================================

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- ========================================
-- 2. 创建核心业务表
-- ========================================

-- 创建用户表（统一用户、学生、教师信息）
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    real_name VARCHAR(100) NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('student', 'teacher', 'admin')),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    avatar VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    register_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    -- 学生特有字段（可选）
    student_id VARCHAR(20) UNIQUE,
    college VARCHAR(100),
    major VARCHAR(100),
    class VARCHAR(50),
    grade VARCHAR(10),
    
    -- 教师特有字段（可选）
    department VARCHAR(100),
    title VARCHAR(50),
    specialty VARCHAR(200)
);

-- 创建学分活动表
CREATE TABLE IF NOT EXISTS credit_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_review', 'approved', 'rejected')),
    category VARCHAR(100) NOT NULL,
    requirements TEXT,
    owner_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    reviewer_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
    review_comments TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 创建活动参与者表
CREATE TABLE IF NOT EXISTS activity_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    credits DECIMAL(5,2) NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 创建申请表
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'approved' CHECK (status IN ('approved')),
    applied_credits DECIMAL(5,2) NOT NULL,
    awarded_credits DECIMAL(5,2) NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE(activity_id, user_id)
);

-- 创建附件表
CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    file_category VARCHAR(50) NOT NULL,
    description TEXT,
    uploaded_by UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    download_count INTEGER NOT NULL DEFAULT 0,
    md5_hash VARCHAR(32) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- ========================================
-- 3. 创建索引
-- ========================================

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_student_id ON users(student_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 活动表索引
CREATE INDEX IF NOT EXISTS idx_credit_activities_status ON credit_activities(status);
CREATE INDEX IF NOT EXISTS idx_credit_activities_owner_id ON credit_activities(owner_id);
CREATE INDEX IF NOT EXISTS idx_credit_activities_deleted_at ON credit_activities(deleted_at);

-- 参与者表索引
CREATE INDEX IF NOT EXISTS idx_activity_participants_activity_id ON activity_participants(activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_participants_user_id ON activity_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_participants_deleted_at ON activity_participants(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_activity_participants_active ON activity_participants(activity_id, user_id) WHERE deleted_at IS NULL;

-- 申请表索引
CREATE INDEX IF NOT EXISTS idx_applications_activity_id ON applications(activity_id);
CREATE INDEX IF NOT EXISTS idx_applications_user_id ON applications(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_deleted_at ON applications(deleted_at);

-- 附件表索引
CREATE INDEX IF NOT EXISTS idx_attachments_activity_id ON attachments(activity_id);
CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by ON attachments(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_attachments_file_category ON attachments(file_category);
CREATE INDEX IF NOT EXISTS idx_attachments_file_type ON attachments(file_type);
CREATE INDEX IF NOT EXISTS idx_attachments_md5_hash ON attachments(md5_hash);
CREATE INDEX IF NOT EXISTS idx_attachments_deleted_at ON attachments(deleted_at);

-- ========================================
-- 4. 创建触发器函数
-- ========================================

-- 更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 活动通过后自动生成申请触发器函数
CREATE OR REPLACE FUNCTION generate_applications_on_activity_approval()
RETURNS TRIGGER AS $$
BEGIN
    -- 只有当状态从非approved变为approved时才触发
    IF OLD.status != 'approved' AND NEW.status = 'approved' THEN
        -- 为所有参与者生成申请
        INSERT INTO applications (id, activity_id, user_id, status, applied_credits, awarded_credits, submitted_at, created_at, updated_at)
        SELECT 
            gen_random_uuid(),
            ap.activity_id,
            ap.user_id,
            'approved',
            ap.credits,
            ap.credits,
            NOW(),
            NOW(),
            NOW()
        FROM activity_participants ap
        WHERE ap.activity_id = NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 活动撤回时删除相关申请触发器函数
CREATE OR REPLACE FUNCTION delete_applications_on_activity_withdraw()
RETURNS TRIGGER AS $$
BEGIN
    -- 当活动状态变为draft时，删除所有相关申请
    IF OLD.status != 'draft' AND NEW.status = 'draft' THEN
        DELETE FROM applications WHERE activity_id = NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 附件清理触发器函数 - 当附件被软删除时检查是否还有其他活动使用该文件
CREATE OR REPLACE FUNCTION cleanup_orphaned_attachments()
RETURNS TRIGGER AS $$
DECLARE
    v_other_activities_count INTEGER;
    v_file_path TEXT;
BEGIN
    -- 如果附件被软删除，检查是否还有其他活动使用相同的文件
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
        -- 检查是否有其他活动使用相同的MD5哈希文件
        SELECT COUNT(*) INTO v_other_activities_count
        FROM attachments a
        WHERE a.md5_hash = NEW.md5_hash 
        AND a.deleted_at IS NULL 
        AND a.id != NEW.id;
        
        -- 如果没有其他活动使用该文件，则彻底删除附件记录和物理文件
        IF v_other_activities_count = 0 THEN
            -- 删除物理文件
            v_file_path := 'uploads/attachments/' || NEW.file_name;
            -- 注意：这里只是记录，实际文件删除需要在应用层处理
            RAISE NOTICE '彻底删除附件文件: %', v_file_path;
            
            -- 彻底删除附件记录（不是软删除）
            DELETE FROM attachments WHERE id = NEW.id;
            RETURN NULL; -- 阻止软删除操作
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 4.1 活动删除存储过程（权限校验+彻底删除附件）
-- ========================================

CREATE OR REPLACE FUNCTION delete_activity_with_permission_check(
    p_activity_id UUID,
    p_user_id UUID,
    p_user_type VARCHAR
) RETURNS TEXT AS $$
DECLARE
    v_owner_id UUID;
    v_status VARCHAR;
    v_now TIMESTAMPTZ := NOW();
    v_attachment RECORD;
    v_file_path TEXT;
    v_other_activities_count INTEGER;
BEGIN
    -- 查询活动拥有者和状态
    SELECT owner_id, status INTO v_owner_id, v_status FROM credit_activities WHERE id = p_activity_id AND deleted_at IS NULL;
    IF NOT FOUND THEN
        RETURN '活动不存在或已删除';
    END IF;

    -- 权限校验：管理员可删，教师/学生只能删自己创建的
    IF p_user_type = 'admin' THEN
        -- 管理员可删除任何活动
        NULL;
    ELSIF p_user_id = v_owner_id THEN
        -- 活动创建者可删除
        NULL;
    ELSE
        RETURN '无权限删除该活动';
    END IF;

    -- 处理附件：彻底删除不被其他活动使用的附件
    FOR v_attachment IN 
        SELECT id, file_name, md5_hash 
        FROM attachments 
        WHERE activity_id = p_activity_id AND deleted_at IS NULL
    LOOP
        -- 检查是否有其他活动使用相同的文件
        SELECT COUNT(*) INTO v_other_activities_count
        FROM attachments a
        WHERE a.md5_hash = v_attachment.md5_hash 
        AND a.deleted_at IS NULL 
        AND a.activity_id != p_activity_id;
        
        -- 如果没有其他活动使用该文件，则彻底删除
        IF v_other_activities_count = 0 THEN
            -- 记录需要删除的物理文件路径
            v_file_path := 'uploads/attachments/' || v_attachment.file_name;
            RAISE NOTICE '需要删除物理文件: %', v_file_path;
            
            -- 彻底删除附件记录
            DELETE FROM attachments WHERE id = v_attachment.id;
        ELSE
            -- 有其他活动使用，只软删除
            UPDATE attachments SET deleted_at = v_now WHERE id = v_attachment.id;
        END IF;
    END LOOP;

    -- 软删除活动
    UPDATE credit_activities SET deleted_at = v_now WHERE id = p_activity_id;
    -- 软删除参与者
    UPDATE activity_participants SET deleted_at = v_now WHERE activity_id = p_activity_id;
    -- 软删除申请
    UPDATE applications SET deleted_at = v_now WHERE activity_id = p_activity_id;

    RETURN '活动删除成功';
END;
$$ LANGUAGE plpgsql;

-- 批量删除活动存储过程
CREATE OR REPLACE FUNCTION batch_delete_activities(
    p_activity_ids UUID[],
    p_user_id UUID,
    p_user_type VARCHAR
) RETURNS INTEGER AS $$
DECLARE
    v_activity_id UUID;
    v_deleted_count INTEGER := 0;
    v_result TEXT;
BEGIN
    -- 遍历活动ID数组
    FOREACH v_activity_id IN ARRAY p_activity_ids
    LOOP
        -- 调用单个删除函数
        SELECT delete_activity_with_permission_check(v_activity_id, p_user_id, p_user_type) INTO v_result;
        
        -- 如果删除成功，增加计数
        IF v_result = '活动删除成功' THEN
            v_deleted_count := v_deleted_count + 1;
        END IF;
    END LOOP;
    
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 获取用户可删除的活动列表
CREATE OR REPLACE FUNCTION get_user_deletable_activities(
    p_user_id UUID,
    p_user_type VARCHAR
) RETURNS TABLE (
    activity_id UUID,
    title VARCHAR(200),
    description TEXT,
    status VARCHAR(20),
    category VARCHAR(100),
    owner_id UUID,
    created_at TIMESTAMPTZ,
    can_delete BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ca.id,
        ca.title,
        ca.description,
        ca.status,
        ca.category,
        ca.owner_id,
        ca.created_at,
        CASE 
            WHEN p_user_type = 'admin' THEN true
            WHEN ca.owner_id = p_user_id THEN true
            ELSE false
        END as can_delete
    FROM credit_activities ca
    WHERE ca.deleted_at IS NULL
    AND (p_user_type = 'admin' OR ca.owner_id = p_user_id)
    ORDER BY ca.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 恢复已删除的活动（仅管理员）
CREATE OR REPLACE FUNCTION restore_deleted_activity(
    p_activity_id UUID,
    p_user_type VARCHAR
) RETURNS TEXT AS $$
DECLARE
    v_now TIMESTAMPTZ := NOW();
BEGIN
    -- 只有管理员可以恢复活动
    IF p_user_type != 'admin' THEN
        RETURN '只有管理员可以恢复活动';
    END IF;

    -- 恢复活动
    UPDATE credit_activities SET deleted_at = NULL WHERE id = p_activity_id;
    IF NOT FOUND THEN
        RETURN '活动不存在';
    END IF;

    -- 恢复参与者
    UPDATE activity_participants SET deleted_at = NULL WHERE activity_id = p_activity_id;
    
    -- 恢复申请
    UPDATE applications SET deleted_at = NULL WHERE activity_id = p_activity_id;
    
    -- 恢复附件
    UPDATE attachments SET deleted_at = NULL WHERE activity_id = p_activity_id;

    RETURN '活动恢复成功';
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 5. 创建触发器
-- ========================================

-- 更新时间戳触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_credit_activities_updated_at BEFORE UPDATE ON credit_activities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_activity_participants_updated_at BEFORE UPDATE ON activity_participants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON applications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_attachments_updated_at BEFORE UPDATE ON attachments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 活动通过后自动生成申请触发器
CREATE TRIGGER trigger_generate_applications
    AFTER UPDATE ON credit_activities
    FOR EACH ROW
    EXECUTE FUNCTION generate_applications_on_activity_approval();

-- 活动撤回时删除相关申请触发器
CREATE TRIGGER trigger_delete_applications_on_withdraw
    AFTER UPDATE ON credit_activities
    FOR EACH ROW
    EXECUTE FUNCTION delete_applications_on_activity_withdraw();

-- 附件清理触发器
CREATE TRIGGER trigger_cleanup_orphaned_attachments
    BEFORE UPDATE ON attachments
    FOR EACH ROW
    EXECUTE FUNCTION cleanup_orphaned_attachments();

-- ========================================
-- 6. 创建视图
-- ========================================

-- 学生基本信息视图
CREATE OR REPLACE VIEW student_basic_info AS
SELECT 
    user_id,
    username,
    real_name,
    student_id,
    college,
    major,
    class,
    grade,
    status,
    avatar,
    register_time,
    created_at,
    updated_at
FROM users 
WHERE user_type = 'student' AND deleted_at IS NULL;

-- 教师基本信息视图
CREATE OR REPLACE VIEW teacher_basic_info AS
SELECT 
    user_id,
    username,
    real_name,
    department,
    title,
    status,
    avatar,
    register_time,
    created_at,
    updated_at
FROM users 
WHERE user_type = 'teacher' AND deleted_at IS NULL;

-- 学生详细信息视图
CREATE OR REPLACE VIEW student_detail_info AS
SELECT 
    user_id,
    username,
    email,
    phone,
    real_name,
    student_id,
    college,
    major,
    class,
    grade,
    status,
    avatar,
    last_login_at,
    register_time,
    created_at,
    updated_at
FROM users 
WHERE user_type = 'student' AND deleted_at IS NULL;

-- 教师详细信息视图
CREATE OR REPLACE VIEW teacher_detail_info AS
SELECT 
    user_id,
    username,
    email,
    phone,
    real_name,
    department,
    title,
    specialty,
    status,
    avatar,
    last_login_at,
    register_time,
    created_at,
    updated_at
FROM users 
WHERE user_type = 'teacher' AND deleted_at IS NULL;

-- ========================================
-- 7. 初始化数据
-- ========================================

-- 创建默认管理员用户
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin') THEN
        INSERT INTO users (username, password, email, real_name, user_type, status)
        VALUES ('admin', '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu', 'admin@example.com', 'Administrator', 'admin', 'active');
    END IF;
END $$;

-- 创建默认教师用户
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'teacher') THEN
        INSERT INTO users (username, password, email, real_name, user_type, status, department, title)
        VALUES ('teacher', '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu', 'teacher@example.com', 'Default Teacher', 'teacher', 'active', '计算机学院', '副教授');
    END IF;
END $$;

-- 创建默认学生用户
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'student') THEN
        INSERT INTO users (username, password, email, real_name, user_type, status, student_id, college, major, class, grade)
        VALUES ('student', '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu', 'student@example.com', 'Default Student', 'student', 'active', '20210001', '计算机学院', '软件工程', '软工2101', '2021');
    END IF;
END $$;

-- ========================================
-- 8. 完成提示
-- ========================================

DO $$
BEGIN
    RAISE NOTICE '数据库初始化完成！';
    RAISE NOTICE '已创建以下表：';
    RAISE NOTICE '- users (用户表)';
    RAISE NOTICE '- credit_activities (学分活动表)';
    RAISE NOTICE '- activity_participants (活动参与者表)';
    RAISE NOTICE '- applications (申请表)';
    RAISE NOTICE '- attachments (附件表)';
    RAISE NOTICE '';
    RAISE NOTICE '权限控制已简化为基于user_type的系统';
    RAISE NOTICE '默认用户：admin/adminpassword, teacher/adminpassword, student/adminpassword';
    RAISE NOTICE '';
    RAISE NOTICE '新增功能：';
    RAISE NOTICE '- 活动删除时自动彻底删除不被其他活动使用的附件';
    RAISE NOTICE '- 附件触发器自动清理孤立文件';
    RAISE NOTICE '- 批量删除活动功能';
    RAISE NOTICE '- 活动恢复功能（仅管理员）';
END $$; 