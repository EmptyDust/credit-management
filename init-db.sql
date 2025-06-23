-- 创建数据库
CREATE DATABASE IF NOT EXISTS credit_management;

-- 使用数据库
\c credit_management;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'student',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建学生信息表
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id),
    username VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    student_no VARCHAR(20) UNIQUE NOT NULL,
    college VARCHAR(100),
    major VARCHAR(100),
    class VARCHAR(50),
    contact VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建教师信息表
CREATE TABLE IF NOT EXISTS teachers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id),
    username VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    department VARCHAR(100),
    title VARCHAR(50),
    contact VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建事项表
CREATE TABLE IF NOT EXISTS affairs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建申请表
CREATE TABLE IF NOT EXISTS applications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    student_id INTEGER NOT NULL REFERENCES students(id),
    affair_id INTEGER NOT NULL REFERENCES affairs(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    credits DECIMAL(5,2) NOT NULL DEFAULT 0,
    attachment_url VARCHAR(500),
    reviewer_id INTEGER REFERENCES teachers(id),
    review_comment TEXT,
    review_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_students_user_id ON students(user_id);
CREATE INDEX IF NOT EXISTS idx_students_student_no ON students(student_no);
CREATE INDEX IF NOT EXISTS idx_teachers_user_id ON teachers(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_user_id ON applications(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_student_id ON applications(student_id);
CREATE INDEX IF NOT EXISTS idx_applications_affair_id ON applications(affair_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);

-- 插入示例数据
INSERT INTO users (username, password, email, role) VALUES
('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin@example.com', 'admin'),
('student1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'student1@example.com', 'student'),
('teacher1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'teacher1@example.com', 'teacher')
ON CONFLICT (username) DO NOTHING;

INSERT INTO students (user_id, username, name, student_no, college, major, class, contact) VALUES
(2, 'student1', '张三', '2021001', '计算机学院', '软件工程', '软工2101', '13800138001')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO teachers (user_id, username, name, department, title, contact) VALUES
(3, 'teacher1', '李老师', '计算机学院', '副教授', '13900139001')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO affairs (title, description, type, status, start_date, end_date) VALUES
('创新创业大赛', '年度创新创业项目大赛', 'competition', 'active', '2024-01-01', '2024-12-31'),
('科研项目申请', '学生科研项目申请', 'research', 'active', '2024-01-01', '2024-12-31'),
('实习实践', '企业实习实践项目', 'internship', 'active', '2024-01-01', '2024-12-31')
ON CONFLICT DO NOTHING; 