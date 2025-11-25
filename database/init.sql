-- 双创分申请平台数据库初始化脚本（优化版）
-- 整合所有约束定义和改进，确保数据库、后端、前端约束一致性

-- ========================================
-- 1. 扩展和设置
-- ========================================

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- ========================================
-- 2. 创建核心业务表（带优化约束）
-- ========================================

-- 枚举：用户身份
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type_enum') THEN
            CREATE TYPE user_type_enum AS ENUM ('student', 'teacher', 'admin');
        END IF;
    END
$$;
-- 枚举：账号状态
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status_enum') THEN
            CREATE TYPE user_status_enum AS ENUM ('active', 'inactive', 'suspended');
        END IF;
    END
$$;
-- 枚举：部门类型
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'dept_type_enum') THEN
            CREATE TYPE dept_type_enum AS ENUM ('school','faculty', 'college', 'major', 'class', 'office', 'others');
        END IF;
    END
$$;

-- 创建部门表
CREATE TABLE IF NOT EXISTS departments
(
    id         UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    name       VARCHAR(100)   NOT NULL,
    code       VARCHAR(20) UNIQUE,
    dept_type  dept_type_enum NOT NULL DEFAULT 'others',
    level      INT            NOT NULL DEFAULT 0,
    parent_id  UUID           REFERENCES departments (id) ON UPDATE CASCADE ON DELETE SET NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 创建用户表（统一用户、学生、教师信息）
CREATE TABLE IF NOT EXISTS users
(
    uuid         UUID PRIMARY KEY             DEFAULT gen_random_uuid(),
    student_id   VARCHAR(18) UNIQUE,
    teacher_id   VARCHAR(18) UNIQUE,
    username     VARCHAR(20) UNIQUE           NOT NULL,
    password     TEXT                         NOT NULL,
    email        VARCHAR(100) UNIQUE          NOT NULL CHECK (email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
    phone        VARCHAR(11) UNIQUE CHECK (phone IS NULL OR phone ~ '^1[3-9]\d{9}$'),
    real_name    VARCHAR(50)                  NOT NULL,
    user_type    user_type_enum               NOT NULL DEFAULT 'student',
    status       user_status_enum             NOT NULL DEFAULT 'active',
    avatar       TEXT,
    department_id UUID                        REFERENCES departments (id) ON UPDATE CASCADE ON DELETE SET NULL,
    last_login_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ                  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ                  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ,
    grade        VARCHAR(4),
    title        VARCHAR(50),
    CONSTRAINT ck_user_identity_consistency CHECK (
            (user_type = 'student' AND student_id IS NOT NULL AND teacher_id IS NULL)
        OR  (user_type = 'teacher' AND teacher_id IS NOT NULL AND student_id IS NULL)
        OR  (user_type = 'admin'   AND student_id IS NULL     AND teacher_id IS NULL)
    )
);

-- 创建学分活动表
CREATE TABLE IF NOT EXISTS credit_activities
(
    id              UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    title           VARCHAR(200) NOT NULL CHECK (LENGTH(TRIM(title)) > 0),
    description     TEXT,
    start_date      DATE         NOT NULL,
    end_date        DATE         NOT NULL CHECK (end_date >= start_date),
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_review', 'approved', 'rejected')),
    category        VARCHAR(100) NOT NULL CHECK (LENGTH(TRIM(category)) > 0),
    owner_id        UUID         NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    reviewer_id     UUID         REFERENCES users (uuid) ON DELETE SET NULL,
    review_comments TEXT,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ
);

-- 创建活动参与者表
CREATE TABLE IF NOT EXISTS activity_participants
(
    id          UUID PRIMARY KEY       DEFAULT gen_random_uuid(),
    activity_id UUID          NOT NULL REFERENCES credit_activities (id) ON DELETE CASCADE,
    user_id UUID          NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    credits     DECIMAL(5, 2) NOT NULL DEFAULT 0 CHECK (credits >= 0),
    joined_at   TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

-- 创建申请表
CREATE TABLE IF NOT EXISTS applications
(
    id              UUID PRIMARY KEY       DEFAULT gen_random_uuid(),
    activity_id     UUID          NOT NULL REFERENCES credit_activities (id) ON DELETE CASCADE,
    user_id    UUID          NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    status          VARCHAR(20)   NOT NULL DEFAULT 'approved' CHECK (status IN ('approved')),
    applied_credits DECIMAL(5, 2) NOT NULL CHECK (applied_credits >= 0),
    awarded_credits DECIMAL(5, 2) NOT NULL CHECK (awarded_credits >= 0),
    submitted_at    TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ,
    UNIQUE (activity_id, user_id)
);

-- 创建附件表
CREATE TABLE IF NOT EXISTS attachments
(
    id             UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    activity_id    UUID         NOT NULL REFERENCES credit_activities (id) ON DELETE CASCADE,
    file_name      VARCHAR(255) NOT NULL CHECK (LENGTH(TRIM(file_name)) > 0),
    original_name  VARCHAR(255) NOT NULL CHECK (LENGTH(TRIM(original_name)) > 0),
    file_size      BIGINT       NOT NULL CHECK (file_size > 0 AND file_size <= 20971520), -- 最大20MB
    file_type      VARCHAR(20)  NOT NULL CHECK (file_type IN
                                                ('.pdf', '.doc', '.docx', '.txt', '.rtf', '.odt', '.jpg', '.jpeg',
                                                 '.png', '.gif', '.bmp', '.webp', '.mp4', '.avi', '.mov', '.wmv',
                                                 '.flv', '.mp3', '.wav', '.ogg', '.aac', '.zip', '.rar', '.7z', '.tar',
                                                 '.gz', '.xls', '.xlsx', '.csv', '.ppt', '.pptx')),
    file_category  VARCHAR(50)  NOT NULL CHECK (file_category IN
                                                ('document', 'image', 'video', 'audio', 'archive', 'spreadsheet',
                                                 'presentation', 'other')),
    description    TEXT,
    uploaded_by    UUID         NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    uploaded_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    download_count INTEGER      NOT NULL DEFAULT 0 CHECK (download_count >= 0),
    md5_hash       VARCHAR(32),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMPTZ
);


-- ========================================
-- 3. 创建索引（优化版）
-- ========================================

-- =========================
-- 用户表索引（users）
-- =========================
-- 单列
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username); -- 按用户名登录/查询
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email); -- 邮箱登录
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users (user_type); -- 按身份过滤（student/teacher/admin）
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status); -- 按账号状态过滤（active/inactive/suspended）
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at); -- 软删除过滤
CREATE INDEX IF NOT EXISTS idx_users_type_status ON users (user_type, status);
-- 身份+状态复合过滤

-- 复合/专用
CREATE INDEX IF NOT EXISTS idx_users_status_type_username ON users (status, user_type, username); -- 状态+身份+用户名（后台列表/搜索）
CREATE INDEX IF NOT EXISTS idx_users_student_id ON users (student_id); -- 学号精确查找
CREATE INDEX IF NOT EXISTS idx_users_teacher_id ON users (teacher_id); -- 工号精确查找
CREATE INDEX IF NOT EXISTS idx_users_phone ON users (phone); -- 手机号登录
CREATE INDEX IF NOT EXISTS idx_users_department_id ON users (department_id); -- 按学院/专业/班级查人
CREATE INDEX IF NOT EXISTS idx_users_grade ON users (grade) WHERE grade IS NOT NULL;-- 按年级查学生
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users (last_login_at DESC); -- 最近登录排序
CREATE INDEX IF NOT EXISTS idx_users_active ON users (deleted_at) WHERE deleted_at IS NULL;
-- 仅活跃用户

-- 部门表索引
-- 按 parent_id 找子部门（树形）：
CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments (parent_id);
-- 按 code 精确查（如学院代码）：
CREATE INDEX IF NOT EXISTS idx_departments_code ON departments (code);
-- 按 dept_type 过滤（查所有学院/所有班级）：
CREATE INDEX IF NOT EXISTS idx_departments_type ON departments (dept_type);
-- 按层级排序：
CREATE INDEX IF NOT EXISTS idx_departments_level ON departments (level);
-- 软删除过滤：
CREATE INDEX IF NOT EXISTS idx_departments_deleted_at ON departments (deleted_at)
    WHERE deleted_at IS NULL;

-- 活动表索引
CREATE INDEX IF NOT EXISTS idx_credit_activities_status ON credit_activities (status);
CREATE INDEX IF NOT EXISTS idx_credit_activities_owner_id ON credit_activities (owner_id);
CREATE INDEX IF NOT EXISTS idx_credit_activities_deleted_at ON credit_activities (deleted_at);
CREATE INDEX IF NOT EXISTS idx_activities_owner_status ON credit_activities (owner_id, status);
CREATE INDEX IF NOT EXISTS idx_activities_category_status ON credit_activities (category, status);

-- 参与者表索引
CREATE INDEX IF NOT EXISTS idx_activity_participants_activity_id ON activity_participants (activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_participants_user_id ON activity_participants (user_id);
CREATE INDEX IF NOT EXISTS idx_activity_participants_deleted_at ON activity_participants (deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_activity_participants_active ON activity_participants (activity_id, user_id) WHERE deleted_at IS NULL;

-- 申请表索引
CREATE INDEX IF NOT EXISTS idx_applications_activity_id ON applications (activity_id);
CREATE INDEX IF NOT EXISTS idx_applications_user_id ON applications (user_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications (status);
CREATE INDEX IF NOT EXISTS idx_applications_deleted_at ON applications (deleted_at);

-- 附件表索引
CREATE INDEX IF NOT EXISTS idx_attachments_activity_id ON attachments (activity_id);
CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by ON attachments (uploaded_by);
CREATE INDEX IF NOT EXISTS idx_attachments_file_category ON attachments (file_category);
CREATE INDEX IF NOT EXISTS idx_attachments_file_type ON attachments (file_type);
CREATE INDEX IF NOT EXISTS idx_attachments_md5_hash ON attachments (md5_hash);
CREATE INDEX IF NOT EXISTS idx_attachments_deleted_at ON attachments (deleted_at);


-- ========================================
-- 7. 创建视图
-- ========================================

-- 学生基本信息视图
CREATE OR REPLACE VIEW student_basic_info AS
SELECT u.uuid,
       u.username,
       u.real_name,
       u.student_id     AS student_id,
       college.name      AS college, -- 学院
       major.name        AS major,   -- 专业
       class.name        AS class,   -- 班级
       u.grade,
       u.avatar
FROM users u
-- 1. 先拿到班级
         JOIN departments class
              ON class.id = u.department_id
                  AND class.dept_type = 'class'
-- 2. 再拿专业
         JOIN departments major
              ON major.id = class.parent_id
                  AND major.dept_type = 'major'
-- 3. 最后拿学院
         JOIN departments college
              ON college.id = major.parent_id
                  AND college.dept_type = 'college'
WHERE u.user_type = 'student'
  AND u.deleted_at IS NULL;

-- 学生详细信息视图
CREATE OR REPLACE VIEW student_detail_info AS
SELECT u.uuid,
       u.username,
       u.real_name,
       u.email,
       u.phone,
       u.student_id     AS student_id, -- 学号
       college.name      AS college,    -- 学院
       major.name        AS major,      -- 专业
       class.name        AS class,      -- 班级
       u.grade,
       u.status,
       u.avatar,
       u.last_login_at
FROM users AS u
-- 1. 班级
         JOIN departments AS class
              ON class.id = u.department_id
                  AND class.dept_type = 'class'
-- 2. 专业
         JOIN departments AS major
              ON major.id = class.parent_id
                  AND major.dept_type = 'major'
-- 3. 学院
         JOIN departments AS college
              ON college.id = major.parent_id
                  AND college.dept_type = 'college'
WHERE u.user_type = 'student'
  AND u.deleted_at IS NULL;

-- 学生完整信息视图
CREATE OR REPLACE VIEW student_complete_info AS
SELECT u.uuid,
       u.username,
       u.email,
       u.phone,
       u.real_name,
       u.user_type,
       u.status,
       u.avatar,
       u.last_login_at,
       u.created_at,
       u.updated_at,
       u.student_id     AS student_id, -- 学号
       college.name      AS college,    -- 学院
       major.name        AS major,      -- 专业
       class.name        AS class,      -- 班级
       u.grade
FROM users AS u
-- 1. 班级
         JOIN departments AS class
              ON class.id = u.department_id
                  AND class.dept_type = 'class'
-- 2. 专业
         JOIN departments AS major
              ON major.id = class.parent_id
                  AND major.dept_type = 'major'
-- 3. 学院
         JOIN departments AS college
              ON college.id = major.parent_id
                  AND college.dept_type = 'college'
WHERE u.user_type = 'student'
  AND u.deleted_at IS NULL;

-- 教师基本信息视图
-- 教师基本信息视图
CREATE OR REPLACE VIEW teacher_basic_info AS
SELECT u.uuid,
       u.teacher_id     AS teacher_id,
       u.username,
       u.real_name,
       d.name            AS department,
       u.title,
       u.avatar
FROM users u
         LEFT JOIN departments d ON d.id = u.department_id
WHERE u.user_type = 'teacher'
  AND u.deleted_at IS NULL;

-- 教师详细信息视图
CREATE OR REPLACE VIEW teacher_detail_info AS
SELECT u.uuid,
       u.teacher_id     AS teacher_id,
       u.username,
       u.real_name,
       u.email,
       u.phone,
       d.name            AS department,
       u.title,
       u.status,
       u.avatar,
       u.last_login_at
FROM users u
         LEFT JOIN departments d ON d.id = u.department_id
WHERE u.user_type = 'teacher'
  AND u.deleted_at IS NULL;

-- 教师完整信息视图
CREATE OR REPLACE VIEW teacher_complete_info AS
SELECT u.uuid,
       u.teacher_id     AS teacher_id,
       u.username,
       u.email,
       u.phone,
       u.real_name,
       u.user_type,
       u.status,
       u.avatar,
       u.last_login_at,
       u.created_at,
       u.updated_at,
       d.name            AS department,
       u.title
FROM users u
         LEFT JOIN departments d ON d.id = u.department_id
WHERE u.user_type = 'teacher'
  AND u.deleted_at IS NULL;
--
CREATE OR REPLACE VIEW detailed_credit_activity_view AS
SELECT ca.id                        AS activity_id,
       ca.title,
       ca.description,
       ca.start_date,
       ca.end_date,
       ca.status,
       ca.category,
       ca.owner_id                  AS creator_id,
       creator.real_name            AS creator_name,
       creator.username             AS creator_username,
       ca.created_at,
       ca.updated_at,
       (SELECT json_agg(json_build_object(
               'id', p.id,
               'username', u.username,
               'real_name', u.real_name,
               'credits', p.credits
                        ))
        FROM activity_participants p
                 JOIN users u ON p.user_id = u.uuid
        WHERE p.activity_id = ca.id
          AND p.deleted_at IS NULL
          AND u.deleted_at IS NULL) AS participants
FROM credit_activities ca
         JOIN
     users creator ON ca.owner_id = creator.uuid
WHERE ca.deleted_at IS NULL
  AND creator.deleted_at IS NULL;
--
CREATE OR REPLACE VIEW detailed_applications_view AS
SELECT
    -- 申请本身
    app.id            AS application_id,
    app.status        AS application_status,
    app.applied_credits,
    app.awarded_credits,
    app.submitted_at,

    -- 申请人（学生）
    u.uuid            AS applicant_id,
    u.real_name       AS applicant_name,
    u.username        AS applicant_username,
    u.student_id AS student_id,        -- 学号
    coll.name         AS applicant_college, -- 学院
    maj.name          AS applicant_major,   -- 专业

    -- 活动
    act.id            AS activity_id,
    act.title         AS activity_title,
    act.category      AS activity_category
FROM applications app
         JOIN users u ON u.uuid = app.user_id-- 注意：一般是 app.user_id
         JOIN credit_activities act ON act.id = app.activity_id
         LEFT JOIN departments coll ON coll.id = u.department_id -- 学院
         LEFT JOIN departments maj ON maj.id = coll.parent_id -- 如果专业存的是 parent_id，可再关联
WHERE app.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND act.deleted_at IS NULL;

-- ========================================
-- 8. 初始化数据
-- ========================================
-- 上海电力大学
INSERT INTO departments (id, name, code, dept_type, parent_id)
VALUES (gen_random_uuid(), '上海电力大学', 'SUEP', 'school', NULL)
ON CONFLICT DO NOTHING;

-- 计算机科学与技术学院（挂在 上海电力大学 下）
WITH parent AS (SELECT id
                FROM departments
                WHERE name = '上海电力大学'
                  AND dept_type = 'school')
INSERT
INTO departments (id, name, code, dept_type, parent_id)
SELECT gen_random_uuid(), '计算机科学与技术学院', 'CS', 'college', parent.id
FROM parent
ON CONFLICT DO NOTHING;

-- 软件工程专业（挂在 计算机科学与技术学院 下）
WITH parent AS (SELECT id
                FROM departments
                WHERE name = '计算机科学与技术学院'
                  AND dept_type = 'college')
INSERT
INTO departments (id, name, code, dept_type, parent_id)
SELECT gen_random_uuid(), '软件工程', 'SE', 'major', parent.id
FROM parent
ON CONFLICT DO NOTHING;

-- 2024222（挂在 软件工程 下）
WITH parent AS (SELECT id
                FROM departments
                WHERE name = '软件工程'
                  AND dept_type = 'major')
INSERT
INTO departments (id, name, code, dept_type, parent_id)
SELECT gen_random_uuid(), '2024222', '2024222', 'class', parent.id
FROM parent
ON CONFLICT DO NOTHING;
-- 创建默认管理员用户（上海电力大学）
DO
$$
    DECLARE
        dept_id UUID;
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin') THEN
            SELECT id
            INTO dept_id
            FROM departments
            WHERE name = '上海电力大学'
              AND dept_type = 'school'
            LIMIT 1;

            INSERT INTO users (student_id, teacher_id, username, password, email, phone,
                               real_name, user_type, status, department_id)
            VALUES (NULL, NULL, 'admin',
                    '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu',
                    'admin@example.com', '13800000000',
                    'Administrator', 'admin', 'active', dept_id);
        END IF;
    END
$$;

-- 创建默认教师用户（软件工程专业）
DO
$$
    DECLARE
        dept_id UUID;
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM users WHERE teacher_id = 'T0000001') THEN
            SELECT id
            INTO dept_id
            FROM departments
            WHERE name = '软件工程'
              AND dept_type = 'major'
            LIMIT 1;

            INSERT INTO users (student_id, teacher_id, username, password, email, phone,
                               real_name, user_type, status, department_id, title)
            VALUES (NULL, 'T0000001', 'teacher',
                    '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu',
                    'teacher@example.com', '13800000001',
                    'Default Teacher', 'teacher', 'active', dept_id, '副教授');
        END IF;
    END
$$;

-- 创建默认学生用户（2024222 班级）
DO
$$
    DECLARE
        dept_id UUID;
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM users WHERE student_id = '20240000') THEN
            SELECT id
            INTO dept_id
            FROM departments
            WHERE name = '2024222'
              AND dept_type = 'class'
            LIMIT 1;

            INSERT INTO users (student_id, teacher_id, username, password, email, phone,
                               real_name, user_type, status, department_id, grade)
            VALUES ('20240000', NULL, 'student',
                    '$2a$10$BBpxLJa6o15NvrxwZcuLxOVCxRHychGgBSkWpp/qNwjc6eyHNoqhu',
                    'student@example.com', '13800000002',
                    'Default Student', 'student', 'active', dept_id, '2024');
        END IF;
    END
$$;

-- ========================================
-- 9. 完成提示
-- ========================================

DO
$$
    BEGIN
        RAISE NOTICE '数据库初始化完成！';
        RAISE NOTICE '已创建以下表：';
        RAISE NOTICE '- users (统一用户表，支持 student_id / teacher_id)';
        RAISE NOTICE '- departments (组织结构树)';
        RAISE NOTICE '- credit_activities (学分活动表)';
        RAISE NOTICE '- activity_participants (参与者表)';
        RAISE NOTICE '- applications (申请表)';
        RAISE NOTICE '- attachments (附件表)';
        RAISE NOTICE '';
        RAISE NOTICE '提示：校验、更新时间戳、活动审批派生申请等逻辑现已移至后端服务实现。';
        RAISE NOTICE '';
        RAISE NOTICE '字段约束仍覆盖：';
        RAISE NOTICE '- 用户名：3-20位字母数字下划线（由应用层校验）';
        RAISE NOTICE '- 邮箱：正则校验';
        RAISE NOTICE '- 手机号：11位 1 开头';
        RAISE NOTICE '- student_id：8位数字（由应用层校验）';
        RAISE NOTICE '- teacher_id：由应用层校验';
        RAISE NOTICE '- 年级：4位数字（由应用层校验）';
        RAISE NOTICE '- 文件：类型白名单，最大20MB';
        RAISE NOTICE '';
        RAISE NOTICE '默认用户：admin/adminpassword, teacher/adminpassword, student/adminpassword';
    END
$$;