-- 数据库完整初始化脚本
-- 用于创建数据库、用户和所有相关结构

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- ========================================
-- 1. 创建基础表结构
-- ========================================

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20),
    real_name VARCHAR(100) NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('student', 'teacher', 'admin')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    avatar VARCHAR(255),
    last_login_at TIMESTAMP,
    register_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- 学生特有字段
    student_id VARCHAR(20) UNIQUE,
    college VARCHAR(100),
    major VARCHAR(100),
    class VARCHAR(50),
    grade VARCHAR(20),
    
    -- 教师特有字段
    department VARCHAR(100),
    title VARCHAR(50),
    specialty VARCHAR(100)
);

-- 创建学分活动表
CREATE TABLE IF NOT EXISTS credit_activities (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    activity_type VARCHAR(50) NOT NULL,
    credit_points INTEGER NOT NULL DEFAULT 0,
    max_participants INTEGER,
    current_participants INTEGER DEFAULT 0,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    registration_deadline TIMESTAMP,
    location VARCHAR(200),
    organizer_id UUID REFERENCES users(user_id),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'ongoing', 'completed', 'cancelled')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建活动参与者表
CREATE TABLE IF NOT EXISTS activity_participants (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'registered' CHECK (status IN ('registered', 'approved', 'rejected', 'completed', 'cancelled')),
    credit_points_earned INTEGER DEFAULT 0,
    attendance_status VARCHAR(20) DEFAULT 'pending' CHECK (attendance_status IN ('pending', 'present', 'absent', 'late')),
    feedback TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(activity_id, user_id)
);

-- 创建申请表
CREATE TABLE IF NOT EXISTS applications (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    activity_id INTEGER NOT NULL REFERENCES credit_activities(id) ON DELETE CASCADE,
    application_type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    credit_points_requested INTEGER DEFAULT 0,
    credit_points_approved INTEGER DEFAULT 0,
    reviewer_id UUID REFERENCES users(user_id),
    review_comment TEXT,
    review_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(user_id, activity_id)
);

-- 创建权限组表
CREATE TABLE IF NOT EXISTS permission_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- 创建用户权限关联表
CREATE TABLE IF NOT EXISTS user_permissions (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, permission_id)
);

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- 创建权限组权限关联表
CREATE TABLE IF NOT EXISTS permission_group_permissions (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, permission_id)
);

-- ========================================
-- 2. 创建索引
-- ========================================

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_student_id ON users(student_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 活动表索引
CREATE INDEX IF NOT EXISTS idx_credit_activities_organizer ON credit_activities(organizer_id);
CREATE INDEX IF NOT EXISTS idx_credit_activities_status ON credit_activities(status);
CREATE INDEX IF NOT EXISTS idx_credit_activities_start_time ON credit_activities(start_time);

-- 参与者表索引
CREATE INDEX IF NOT EXISTS idx_activity_participants_activity_id ON activity_participants(activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_participants_user_id ON activity_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_participants_status ON activity_participants(status);

-- 申请表索引
CREATE INDEX IF NOT EXISTS idx_applications_user_id ON applications(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_activity_id ON applications(activity_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_reviewer_id ON applications(reviewer_id);

-- ========================================
-- 3. 创建触发器函数
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
CREATE OR REPLACE FUNCTION validate_user_data() RETURNS TRIGGER AS $$
BEGIN
    -- 检查用户名不能为空
    IF NEW.username IS NULL OR LENGTH(TRIM(NEW.username)) = 0 THEN
        RAISE EXCEPTION '用户名不能为空';
    END IF;
    
    -- 检查邮箱不能为空
    IF NEW.email IS NULL OR LENGTH(TRIM(NEW.email)) = 0 THEN
        RAISE EXCEPTION '邮箱不能为空';
    END IF;
    
    -- 检查用户名唯一性（排除当前用户）
    IF EXISTS (SELECT 1 FROM users WHERE username = NEW.username AND user_id != NEW.user_id) THEN
        RAISE EXCEPTION '用户名已存在';
    END IF;
    
    -- 检查邮箱唯一性（排除当前用户）
    IF EXISTS (SELECT 1 FROM users WHERE email = NEW.email AND user_id != NEW.user_id) THEN
        RAISE EXCEPTION '邮箱已存在';
    END IF;
    
    -- 检查手机号唯一性（如果提供了手机号）
    IF NEW.phone IS NOT NULL AND LENGTH(TRIM(NEW.phone)) > 0 THEN
        IF EXISTS (SELECT 1 FROM users WHERE phone = NEW.phone AND user_id != NEW.user_id) THEN
            RAISE EXCEPTION '手机号已存在';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 活动与申请同步触发器函数
CREATE OR REPLACE FUNCTION sync_application_on_activity_update()
RETURNS TRIGGER AS $$
BEGIN
    -- 当活动信息更新时，同步更新相关申请
    UPDATE applications
    SET title = NEW.title,
        description = NEW.description,
        credit_points_requested = NEW.credit_points,
        updated_at = CURRENT_TIMESTAMP
    WHERE activity_id = NEW.id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sync_application_on_activity_delete()
RETURNS TRIGGER AS $$
BEGIN
    -- 当活动被删除时，删除相关申请
    DELETE FROM applications WHERE activity_id = OLD.id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sync_application_on_participant_change()
RETURNS TRIGGER AS $$
BEGIN
    -- 当参与者状态改变时，更新相关申请状态
    IF TG_OP = 'INSERT' THEN
        -- 新参与者加入时，创建对应的申请记录
        INSERT INTO applications (user_id, activity_id, application_type, title, description, status, credit_points_requested)
        SELECT 
            NEW.user_id,
            NEW.activity_id,
            'activity_participation',
            ca.title,
            ca.description,
            CASE 
                WHEN NEW.status = 'approved' THEN 'approved'
                WHEN NEW.status = 'rejected' THEN 'rejected'
                ELSE 'pending'
            END,
            ca.credit_points
        FROM credit_activities ca
        WHERE ca.id = NEW.activity_id
        ON CONFLICT (user_id, activity_id) DO UPDATE SET
            status = CASE 
                WHEN NEW.status = 'approved' THEN 'approved'
                WHEN NEW.status = 'rejected' THEN 'rejected'
                ELSE 'pending'
            END,
            updated_at = CURRENT_TIMESTAMP;
    ELSIF TG_OP = 'UPDATE' THEN
        -- 参与者状态更新时，同步更新申请状态
        UPDATE applications
        SET status = CASE 
            WHEN NEW.status = 'approved' THEN 'approved'
            WHEN NEW.status = 'rejected' THEN 'rejected'
            WHEN NEW.status = 'completed' THEN 'approved'
            WHEN NEW.status = 'cancelled' THEN 'cancelled'
            ELSE 'pending'
        END,
        credit_points_approved = CASE 
            WHEN NEW.status = 'completed' THEN NEW.credit_points_earned
            ELSE credit_points_approved
        END,
        updated_at = CURRENT_TIMESTAMP
        WHERE user_id = NEW.user_id AND activity_id = NEW.activity_id;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 4. 创建触发器
-- ========================================

-- 更新时间戳触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_credit_activities_updated_at BEFORE UPDATE ON credit_activities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_activity_participants_updated_at BEFORE UPDATE ON activity_participants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON applications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 数据验证触发器
CREATE TRIGGER trigger_validate_user_data
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION validate_user_data();

-- 同步触发器
CREATE TRIGGER trigger_activity_update_sync
    AFTER UPDATE ON credit_activities
    FOR EACH ROW
    EXECUTE FUNCTION sync_application_on_activity_update();

CREATE TRIGGER trigger_activity_delete_sync
    AFTER DELETE ON credit_activities
    FOR EACH ROW
    EXECUTE FUNCTION sync_application_on_activity_delete();

CREATE TRIGGER trigger_participant_change_sync
    AFTER INSERT OR UPDATE ON activity_participants
    FOR EACH ROW
    EXECUTE FUNCTION sync_application_on_participant_change();

-- ========================================
-- 5. 创建存储过程
-- ========================================

-- 活动更新同步存储过程
CREATE OR REPLACE FUNCTION update_activity_and_sync_applications(
    p_activity_id INTEGER,
    p_title VARCHAR,
    p_description TEXT,
    p_credit_points INTEGER
) RETURNS VOID AS $$
BEGIN
    -- 更新活动表
    UPDATE credit_activities
    SET title = p_title,
        description = p_description,
        credit_points = p_credit_points,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_activity_id;

    -- 同步申请表
    UPDATE applications
    SET title = p_title,
        description = p_description,
        credit_points_requested = p_credit_points,
        updated_at = CURRENT_TIMESTAMP
    WHERE activity_id = p_activity_id;
END;
$$ LANGUAGE plpgsql;

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

-- 用户统计视图
CREATE OR REPLACE VIEW user_stats_view AS
SELECT 
    user_type,
    status,
    COUNT(*) as count,
    DATE(created_at) as created_date
FROM users
WHERE deleted_at IS NULL
GROUP BY user_type, status, DATE(created_at);

-- 学生统计视图
CREATE OR REPLACE VIEW student_stats_view AS
SELECT 
    college,
    major,
    grade,
    status,
    COUNT(*) as count
FROM users
WHERE user_type = 'student' AND deleted_at IS NULL
GROUP BY college, major, grade, status;

-- 教师统计视图
CREATE OR REPLACE VIEW teacher_stats_view AS
SELECT 
    department,
    title,
    status,
    COUNT(*) as count
FROM users
WHERE user_type = 'teacher' AND deleted_at IS NULL
GROUP BY department, title, status;

-- 活动统计视图
CREATE OR REPLACE VIEW activity_stats_view AS
SELECT 
    activity_type,
    status,
    COUNT(*) as count,
    SUM(credit_points) as total_credit_points
FROM credit_activities 
WHERE deleted_at IS NULL
GROUP BY activity_type, status;

-- 参与者统计视图
CREATE OR REPLACE VIEW participant_stats_view AS
SELECT 
    ap.status,
    ap.attendance_status,
    COUNT(*) as count,
    SUM(ap.credit_points_earned) as total_credit_points_earned
FROM activity_participants ap
JOIN credit_activities ca ON ap.activity_id = ca.id
WHERE ca.deleted_at IS NULL
GROUP BY ap.status, ap.attendance_status;

-- 申请统计视图
CREATE OR REPLACE VIEW application_stats_view AS
SELECT 
    application_type,
    status,
    COUNT(*) as count,
    SUM(credit_points_requested) as total_credit_points_requested,
    SUM(credit_points_approved) as total_credit_points_approved
FROM applications
WHERE deleted_at IS NULL
GROUP BY application_type, status;

-- ========================================
-- 7. 插入初始数据
-- ========================================

-- 插入管理员用户（密码: Admin123456）
INSERT INTO users (
    user_id, username, password, email, real_name, user_type, status
) VALUES (
    '67ae4b2c-246b-459d-911d-c3e532bfdf07',
    'admin',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'admin@credit-management.com',
    '系统管理员',
    'admin',
    'active'
) ON CONFLICT (username) DO NOTHING;

-- 插入学生测试数据
INSERT INTO users (
    user_id, username, password, email, real_name, user_type, status,
    student_id, college, major, class, grade
) VALUES 
(
    gen_random_uuid(),
    'student1',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'student1@test.com',
    '张三',
    'student',
    'active',
    '20240001',
    '计算机学院',
    '软件工程',
    '软工2401',
    '2024'
),
(
    gen_random_uuid(),
    'student2',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'student2@test.com',
    '李四',
    'student',
    'active',
    '20240002',
    '计算机学院',
    '软件工程',
    '软工2401',
    '2024'
),
(
    gen_random_uuid(),
    'student3',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'student3@test.com',
    '王五',
    'student',
    'active',
    '20240003',
    '信息工程学院',
    '通信工程',
    '通信2401',
    '2024'
) ON CONFLICT (username) DO NOTHING;

-- 插入教师测试数据
INSERT INTO users (
    user_id, username, password, email, real_name, user_type, status,
    department, title, specialty
) VALUES 
(
    gen_random_uuid(),
    'teacher1',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'teacher1@test.com',
    '陈教授',
    'teacher',
    'active',
    '计算机学院',
    '教授',
    '软件工程'
),
(
    gen_random_uuid(),
    'teacher2',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'teacher2@test.com',
    '李副教授',
    'teacher',
    'active',
    '信息工程学院',
    '副教授',
    '通信工程'
) ON CONFLICT (username) DO NOTHING;

-- 插入角色数据
INSERT INTO roles (name, description) VALUES 
('admin', '系统管理员'),
('teacher', '教师'),
('student', '学生')
ON CONFLICT (name) DO NOTHING;

-- 插入权限数据
INSERT INTO permissions (name, description, resource, action) VALUES 
('user_manage', '用户管理', 'user', 'manage'),
('user_view', '用户查看', 'user', 'view'),
('activity_manage', '活动管理', 'activity', 'manage'),
('activity_view', '活动查看', 'activity', 'view'),
('application_manage', '申请管理', 'application', 'manage'),
('application_view', '申请查看', 'application', 'view'),
('permission_manage', '权限管理', 'permission', 'manage'),
('statistics_view', '统计查看', 'statistics', 'view')
ON CONFLICT (name) DO NOTHING;

-- 插入权限组数据
INSERT INTO permission_groups (name, description) VALUES 
('admin_group', '管理员权限组'),
('teacher_group', '教师权限组'),
('student_group', '学生权限组')
ON CONFLICT (name) DO NOTHING;

-- 插入角色权限关联
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('user_manage', 'user_view', 'activity_manage', 'activity_view', 'application_manage', 'application_view', 'permission_manage', 'statistics_view')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'teacher' AND p.name IN ('user_view', 'activity_manage', 'activity_view', 'application_manage', 'application_view', 'statistics_view')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'student' AND p.name IN ('activity_view', 'application_view')
ON CONFLICT DO NOTHING;

-- 插入用户角色关联
INSERT INTO user_roles (user_id, role_id)
SELECT u.user_id, r.id
FROM users u, roles r
WHERE u.username = 'admin' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT u.user_id, r.id
FROM users u, roles r
WHERE u.username LIKE 'teacher%' AND r.name = 'teacher'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT u.user_id, r.id
FROM users u, roles r
WHERE u.username LIKE 'student%' AND r.name = 'student'
ON CONFLICT DO NOTHING;

-- 插入测试活动数据
INSERT INTO credit_activities (
    title, description, activity_type, credit_points, max_participants,
    start_time, end_time, registration_deadline, location, organizer_id, status
) VALUES 
(
    '软件工程实践项目',
    '参与企业级软件项目开发，提升实践能力',
    'project',
    5,
    20,
    CURRENT_TIMESTAMP + INTERVAL '7 days',
    CURRENT_TIMESTAMP + INTERVAL '30 days',
    CURRENT_TIMESTAMP + INTERVAL '5 days',
    '计算机学院实验室',
    (SELECT user_id FROM users WHERE username = 'teacher1' LIMIT 1),
    'published'
),
(
    '学术讲座：人工智能前沿',
    '邀请知名专家分享AI领域最新研究成果',
    'lecture',
    2,
    100,
    CURRENT_TIMESTAMP + INTERVAL '3 days',
    CURRENT_TIMESTAMP + INTERVAL '4 days',
    CURRENT_TIMESTAMP + INTERVAL '2 days',
    '学术报告厅',
    (SELECT user_id FROM users WHERE username = 'teacher2' LIMIT 1),
    'published'
) ON CONFLICT DO NOTHING;

-- 提交事务
COMMIT;

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE '数据库初始化完成！';
    RAISE NOTICE '已创建所有表、视图、存储过程、触发器和初始数据';
    RAISE NOTICE '默认管理员账号: admin / Admin123456';
END $$;