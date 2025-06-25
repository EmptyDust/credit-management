-- 数据库视图定义
-- 用于不同权限级别的数据访问控制

-- 学生基本信息视图（学生可查看其他学生的基本信息）
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

-- 教师基本信息视图（学生和教师可查看教师的基本信息）
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

-- 学生详细信息视图（教师可查看学生的详细信息）
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

-- 教师详细信息视图（管理员可查看教师的详细信息）
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

-- 用户统计视图（用于统计查询）
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