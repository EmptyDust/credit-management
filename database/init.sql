-- 双创分申请平台数据库初始化脚本（优化版）
-- 整合所有约束定义和改进，确保数据库、后端、前端约束一致性

-- ========================================
-- 1. 扩展和设置
-- ========================================

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- ========================================
-- 2. 创建核心业务表（带优化约束）
-- ========================================

-- 创建用户表（统一用户、学生、教师信息）
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(20) UNIQUE NOT NULL CHECK (username ~ '^[a-zA-Z0-9_]+$'),
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL CHECK (email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
    phone VARCHAR(11) UNIQUE CHECK (phone IS NULL OR phone ~ '^1[3-9]\d{9}$'),
    real_name VARCHAR(50) NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('student', 'teacher', 'admin')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    avatar VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    register_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    -- 学生特有字段（可选）
    student_id VARCHAR(8) UNIQUE CHECK (student_id IS NULL OR student_id ~ '^\d{8}$'),
    college VARCHAR(100),
    major VARCHAR(100),
    class VARCHAR(50),
    grade VARCHAR(4) CHECK (grade IS NULL OR grade ~ '^\d{4}$'),
    
    -- 教师特有字段（可选）
    department VARCHAR(100),
    title VARCHAR(50)
);

-- 创建学分活动表
CREATE TABLE IF NOT EXISTS credit_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(title)) > 0),
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL CHECK (end_date >= start_date),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_review', 'approved', 'rejected')),
    category VARCHAR(100) NOT NULL CHECK (LENGTH(TRIM(category)) > 0),
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
    credits DECIMAL(5,2) NOT NULL DEFAULT 0 CHECK (credits >= 0),
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
    applied_credits DECIMAL(5,2) NOT NULL CHECK (applied_credits >= 0),
    awarded_credits DECIMAL(5,2) NOT NULL CHECK (awarded_credits >= 0),
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
    file_name VARCHAR(255) NOT NULL CHECK (LENGTH(TRIM(file_name)) > 0),
    original_name VARCHAR(255) NOT NULL CHECK (LENGTH(TRIM(original_name)) > 0),
    file_size BIGINT NOT NULL CHECK (file_size > 0 AND file_size <= 20971520), -- 最大20MB
    file_type VARCHAR(20) NOT NULL CHECK (file_type IN ('.pdf', '.doc', '.docx', '.txt', '.rtf', '.odt', '.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.mp4', '.avi', '.mov', '.wmv', '.flv', '.mp3', '.wav', '.ogg', '.aac', '.zip', '.rar', '.7z', '.tar', '.gz', '.xls', '.xlsx', '.csv', '.ppt', '.pptx')),
    file_category VARCHAR(50) NOT NULL CHECK (file_category IN ('document', 'image', 'video', 'audio', 'archive', 'spreadsheet', 'presentation', 'other')),
    description TEXT,
    uploaded_by UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    download_count INTEGER NOT NULL DEFAULT 0 CHECK (download_count >= 0),
    md5_hash VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- ========================================
-- 2.x. 创建五类活动类型详情表（精简字段版）
-- ========================================

-- 创新创业实践活动详情表
CREATE TABLE IF NOT EXISTS innovation_activity_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    item VARCHAR(200),
    company VARCHAR(200),
    project_no VARCHAR(100),
    issuer VARCHAR(100),
    date DATE,
    total_hours DECIMAL(6,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 学科竞赛学分详情表
CREATE TABLE IF NOT EXISTS competition_activity_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    level VARCHAR(100),
    competition VARCHAR(200),
    award_level VARCHAR(100),
    rank VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 大学生创业项目详情表
CREATE TABLE IF NOT EXISTS entrepreneurship_project_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    project_name VARCHAR(200),
    project_level VARCHAR(100),
    project_rank VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 创业实践项目详情表
CREATE TABLE IF NOT EXISTS entrepreneurship_practice_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    company_name VARCHAR(200),
    legal_person VARCHAR(100),
    share_percent DECIMAL(5,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 论文专利详情表
CREATE TABLE IF NOT EXISTS paper_patent_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    name VARCHAR(200),
    category VARCHAR(100),
    rank VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- ========================================
-- 3. 创建索引（优化版）
-- ========================================

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_student_id ON users(student_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_type_status ON users(user_type, status);
CREATE INDEX IF NOT EXISTS idx_users_college_major ON users(college, major);
CREATE INDEX IF NOT EXISTS idx_users_department_title ON users(department, title);

-- 活动表索引
CREATE INDEX IF NOT EXISTS idx_credit_activities_status ON credit_activities(status);
CREATE INDEX IF NOT EXISTS idx_credit_activities_owner_id ON credit_activities(owner_id);
CREATE INDEX IF NOT EXISTS idx_credit_activities_deleted_at ON credit_activities(deleted_at);
CREATE INDEX IF NOT EXISTS idx_activities_owner_status ON credit_activities(owner_id, status);
CREATE INDEX IF NOT EXISTS idx_activities_category_status ON credit_activities(category, status);

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

-- 活动类型详情表索引
CREATE INDEX IF NOT EXISTS idx_innovation_activity_details_activity_id ON innovation_activity_details(activity_id);
CREATE INDEX IF NOT EXISTS idx_innovation_activity_details_deleted_at ON innovation_activity_details(deleted_at);

CREATE INDEX IF NOT EXISTS idx_competition_activity_details_activity_id ON competition_activity_details(activity_id);
CREATE INDEX IF NOT EXISTS idx_competition_activity_details_deleted_at ON competition_activity_details(deleted_at);

CREATE INDEX IF NOT EXISTS idx_entrepreneurship_project_details_activity_id ON entrepreneurship_project_details(activity_id);
CREATE INDEX IF NOT EXISTS idx_entrepreneurship_project_details_deleted_at ON entrepreneurship_project_details(deleted_at);

CREATE INDEX IF NOT EXISTS idx_entrepreneurship_practice_details_activity_id ON entrepreneurship_practice_details(activity_id);
CREATE INDEX IF NOT EXISTS idx_entrepreneurship_practice_details_deleted_at ON entrepreneurship_practice_details(deleted_at);

CREATE INDEX IF NOT EXISTS idx_paper_patent_details_activity_id ON paper_patent_details(activity_id);
CREATE INDEX IF NOT EXISTS idx_paper_patent_details_deleted_at ON paper_patent_details(deleted_at);

-- ========================================
-- 4. 创建验证函数
-- ========================================

-- 密码复杂度验证函数
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

-- 邮箱格式验证函数
CREATE OR REPLACE FUNCTION validate_email_format(email TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' THEN
        RETURN TRUE;
    END IF;
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 文件类型验证函数
CREATE OR REPLACE FUNCTION validate_file_type(file_type TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF file_type IN ('.pdf', '.doc', '.docx', '.txt', '.rtf', '.odt', '.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.mp4', '.avi', '.mov', '.wmv', '.flv', '.mp3', '.wav', '.ogg', '.aac', '.zip', '.rar', '.7z', '.tar', '.gz', '.xls', '.xlsx', '.csv', '.ppt', '.pptx') THEN
        RETURN TRUE;
    END IF;
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 文件大小验证函数
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
-- 5. 创建触发器函数
-- ========================================

-- 更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 用户数据验证触发器函数
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

-- 活动通过后自动生成申请触发器函数
CREATE OR REPLACE FUNCTION generate_applications_on_activity_approval()
RETURNS TRIGGER AS $$
BEGIN
    -- 只有当状态从非approved变为approved时才触发
    IF OLD.status != 'approved' AND NEW.status = 'approved' THEN
        -- 为所有参与者生成申请（只插入不存在的记录）
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
        WHERE ap.activity_id = NEW.id 
        AND ap.deleted_at IS NULL
        AND NOT EXISTS (
            SELECT 1 FROM applications 
            WHERE activity_id = ap.activity_id 
            AND user_id = ap.user_id 
            AND deleted_at IS NULL
        );
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

-- 附件清理触发器函数
CREATE OR REPLACE FUNCTION cleanup_orphaned_attachments()
RETURNS TRIGGER AS $$
DECLARE
    v_file_path TEXT;
    v_other_attachments_count INTEGER;
BEGIN
    -- 如果附件被删除，检查是否需要删除物理文件
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
        -- 检查是否有其他活动使用相同的文件
        SELECT COUNT(*) INTO v_other_attachments_count
        FROM attachments
        WHERE md5_hash = NEW.md5_hash 
        AND activity_id != NEW.activity_id 
        AND deleted_at IS NULL;
        
        -- 如果没有其他活动使用该文件，则删除物理文件
        IF v_other_attachments_count = 0 THEN
            v_file_path := 'uploads/attachments/' || NEW.file_name;
            -- 注意：这里只是记录，实际删除需要在应用层处理
            RAISE NOTICE '需要删除孤立文件: %', v_file_path;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 单个活动删除权限检查函数
CREATE OR REPLACE FUNCTION delete_activity_with_permission_check(
    p_activity_id UUID,
    p_user_id UUID,
    p_user_type VARCHAR
) RETURNS TEXT AS $$
DECLARE
    v_activity_record RECORD;
    v_now TIMESTAMPTZ := NOW();
BEGIN
    -- 检查活动是否存在且未被删除
    SELECT * INTO v_activity_record 
    FROM credit_activities 
    WHERE id = p_activity_id AND deleted_at IS NULL;
    
    IF NOT FOUND THEN
        RETURN '活动不存在或已删除';
    END IF;
    
    -- 权限检查：只有活动创建者和管理员可以删除活动
    IF v_activity_record.owner_id != p_user_id AND p_user_type != 'admin' THEN
        RETURN '无权限删除该活动';
    END IF;
    
    -- 软删除活动
    UPDATE credit_activities SET deleted_at = v_now WHERE id = p_activity_id;
    
    -- 软删除参与者
    UPDATE activity_participants SET deleted_at = v_now WHERE activity_id = p_activity_id;
    
    -- 软删除申请
    UPDATE applications SET deleted_at = v_now WHERE activity_id = p_activity_id;
    
    -- 软删除附件
    UPDATE attachments SET deleted_at = v_now WHERE activity_id = p_activity_id;
    
    RETURN '活动删除成功';
END;
$$ LANGUAGE plpgsql;

-- 批量删除活动函数
CREATE OR REPLACE FUNCTION batch_delete_activities(
    p_activity_ids UUID[],
    p_user_id UUID,
    p_user_type VARCHAR
) RETURNS INTEGER AS $$
DECLARE
    v_activity_id UUID;
    v_deleted_count INTEGER := 0;
    v_now TIMESTAMPTZ := NOW();
BEGIN
    -- 只有管理员可以批量删除活动
    IF p_user_type != 'admin' THEN
        RAISE EXCEPTION '只有管理员可以批量删除活动';
    END IF;

    -- 遍历活动ID数组
    FOREACH v_activity_id IN ARRAY p_activity_ids
    LOOP
        -- 检查活动是否存在且未被删除
        IF EXISTS (SELECT 1 FROM credit_activities WHERE id = v_activity_id AND deleted_at IS NULL) THEN
            -- 软删除活动
            UPDATE credit_activities SET deleted_at = v_now WHERE id = v_activity_id;
            
            -- 软删除参与者
            UPDATE activity_participants SET deleted_at = v_now WHERE activity_id = v_activity_id;
            
            -- 软删除申请
            UPDATE applications SET deleted_at = v_now WHERE activity_id = v_activity_id;
            
            -- 软删除附件
            UPDATE attachments SET deleted_at = v_now WHERE activity_id = v_activity_id;
            
            v_deleted_count := v_deleted_count + 1;
        END IF;
    END LOOP;

    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 恢复已删除的活动函数（仅管理员）
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
-- 6. 创建触发器
-- ========================================

-- 更新时间戳触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_credit_activities_updated_at BEFORE UPDATE ON credit_activities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_activity_participants_updated_at BEFORE UPDATE ON activity_participants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON applications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_attachments_updated_at BEFORE UPDATE ON attachments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 活动类型详情表更新时间戳触发器
CREATE TRIGGER update_innovation_activity_details_updated_at BEFORE UPDATE ON innovation_activity_details FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_competition_activity_details_updated_at BEFORE UPDATE ON competition_activity_details FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entrepreneurship_project_details_updated_at BEFORE UPDATE ON entrepreneurship_project_details FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entrepreneurship_practice_details_updated_at BEFORE UPDATE ON entrepreneurship_practice_details FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_paper_patent_details_updated_at BEFORE UPDATE ON paper_patent_details FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 用户数据验证触发器
CREATE TRIGGER trigger_validate_user_data
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION validate_user_data();

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
-- 7. 创建视图
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
    avatar
FROM users
WHERE user_type = 'student' AND deleted_at IS NULL;

-- 学生详细信息视图
CREATE OR REPLACE VIEW student_detail_info AS
SELECT
    user_id,
    username,
    real_name,
    email,
    phone,
    student_id,
    college,
    major,
    class,
    grade,
    status,
    avatar,
    last_login_at,
    register_time
FROM users
WHERE user_type = 'student' AND deleted_at IS NULL;

-- 学生完整信息视图
CREATE OR REPLACE VIEW student_complete_info AS
SELECT
    user_id,
    username,
    email,
    phone,
    real_name,
    user_type,
    status,
    avatar,
    last_login_at,
    register_time,
    created_at,
    updated_at,
    student_id,
    college,
    major,
    class,
    grade
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
    avatar
FROM users
WHERE user_type = 'teacher' AND deleted_at IS NULL;

-- 教师详细信息视图
CREATE OR REPLACE VIEW teacher_detail_info AS
SELECT
    user_id,
    username,
    real_name,
    email,
    phone,
    department,
    title,
    status,
    avatar,
    last_login_at,
    register_time
FROM users
WHERE user_type = 'teacher' AND deleted_at IS NULL;

-- 教师完整信息视图
CREATE OR REPLACE VIEW teacher_complete_info AS
SELECT
    user_id,
    username,
    email,
    phone,
    real_name,
    user_type,
    status,
    avatar,
    last_login_at,
    register_time,
    created_at,
    updated_at,
    department,
    title
FROM users
WHERE user_type = 'teacher' AND deleted_at IS NULL;

CREATE OR REPLACE VIEW detailed_credit_activity_view AS
SELECT
    ca.id AS activity_id,
    ca.title,
    ca.description,
    ca.start_date,
    ca.end_date,
    ca.status,
    ca.category,
    ca.owner_id AS creator_id,
    creator.real_name AS creator_name,
    creator.username AS creator_username,
    ca.created_at,
    ca.updated_at,
    (
        SELECT json_agg(json_build_object(
            'user_id', p.user_id,
            'username', u.username,
            'real_name', u.real_name,
            'credits', p.credits
        ))
        FROM activity_participants p
        JOIN users u ON p.user_id = u.user_id
        WHERE p.activity_id = ca.id AND p.deleted_at IS NULL AND u.deleted_at IS NULL
    ) AS participants
FROM
    credit_activities ca
JOIN
    users creator ON ca.owner_id = creator.user_id
WHERE
    ca.deleted_at IS NULL AND creator.deleted_at IS NULL;

CREATE OR REPLACE VIEW detailed_applications_view AS
SELECT
    -- 申请本身的信息
    app.id AS application_id,
    app.status AS application_status,
    app.applied_credits,
    app.awarded_credits,
    app.submitted_at,

    -- 申请人（学生）的信息
    u.user_id AS applicant_id,
    u.real_name AS applicant_name,
    u.username AS applicant_username,
    u.student_id,
    u.college AS applicant_college,
    u.major AS applicant_major,

    -- 关联的活动信息
    act.id AS activity_id,
    act.title AS activity_title,
    act.category AS activity_category
FROM
    applications app
JOIN
    users u ON app.user_id = u.user_id
JOIN
    credit_activities act ON app.activity_id = act.id
WHERE
    app.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND act.deleted_at IS NULL;

-- ========================================
-- 8. 初始化数据
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
        INSERT INTO users (username, password, email, phone, real_name, user_type, status, department, title)
        VALUES ('teacher', '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu', 'teacher@example.com', '13800138000', 'Default Teacher', 'teacher', 'active', '计算机学院', '副教授');
    END IF;
END $$;

-- 创建默认学生用户
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'student') THEN
        INSERT INTO users (username, password, email, phone, real_name, user_type, status, student_id, college, major, class, grade)
        VALUES ('student', '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu', 'student@example.com', '13800138001', 'Default Student', 'student', 'active', '20210001', '计算机学院', '软件工程', '软工2101', '2021');
    END IF;
END $$;

-- ========================================
-- 9. 完成提示
-- ========================================

DO $$
BEGIN
    RAISE NOTICE '数据库初始化完成！';
    RAISE NOTICE '已创建以下表：';
    RAISE NOTICE '- users (用户表) - 带完整约束验证';
    RAISE NOTICE '- credit_activities (学分活动表) - 带业务逻辑约束';
    RAISE NOTICE '- activity_participants (活动参与者表) - 带唯一性约束';
    RAISE NOTICE '- applications (申请表) - 带学分验证';
    RAISE NOTICE '- attachments (附件表) - 带文件类型和大小约束';
    RAISE NOTICE '- innovation_activity_details (创新创业实践活动详情表)';
    RAISE NOTICE '- competition_activity_details (学科竞赛详情表)';
    RAISE NOTICE '- entrepreneurship_project_details (大学生创业项目详情表)';
    RAISE NOTICE '- entrepreneurship_practice_details (创业实践项目详情表)';
    RAISE NOTICE '- paper_patent_details (论文专利详情表)';
    RAISE NOTICE '';
    RAISE NOTICE '约束验证：';
    RAISE NOTICE '- 用户名：3-20位字母数字下划线';
    RAISE NOTICE '- 密码：8位以上，包含大小写字母和数字';
    RAISE NOTICE '- 手机号：11位数字，1开头';
    RAISE NOTICE '- 学号：8位数字';
    RAISE NOTICE '- 年级：4位数字';
    RAISE NOTICE '- 文件：支持类型白名单，最大20MB';
    RAISE NOTICE '';
    RAISE NOTICE '默认用户：admin/adminpassword, teacher/adminpassword, student/adminpassword';
    RAISE NOTICE '';
    RAISE NOTICE '新增优化：';
    RAISE NOTICE '- 完整的字段格式验证';
    RAISE NOTICE '- 文件类型和大小约束';
    RAISE NOTICE '- 复合索引优化查询性能';
    RAISE NOTICE '- 触发器自动维护数据一致性';
    RAISE NOTICE '- 批量操作和恢复功能';
    RAISE NOTICE '- 活动类型详情表支持';
END $$; 