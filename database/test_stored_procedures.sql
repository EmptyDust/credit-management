-- 存储过程和触发器测试脚本
-- 用于验证数据库功能是否正常工作

-- ========================================
-- 1. 测试触发器功能
-- ========================================

-- 创建测试活动
SELECT '=== 创建测试活动 ===' as test_name;
INSERT INTO credit_activities (
    title, description, activity_type, credit_points, 
    start_time, end_time, organizer_id, status
) VALUES (
    '测试活动', '这是一个测试活动', 'workshop', 5,
    CURRENT_TIMESTAMP + INTERVAL '1 day', CURRENT_TIMESTAMP + INTERVAL '2 days',
    (SELECT user_id FROM users WHERE username = 'admin' LIMIT 1),
    'published'
) RETURNING id;

-- 获取刚创建的活动ID
DO $$
DECLARE
    activity_id INTEGER;
    admin_user_id UUID;
    student_user_id UUID;
BEGIN
    -- 获取活动ID
    SELECT id INTO activity_id FROM credit_activities WHERE title = '测试活动' LIMIT 1;
    
    -- 获取管理员用户ID
    SELECT user_id INTO admin_user_id FROM users WHERE username = 'admin' LIMIT 1;
    
    -- 获取学生用户ID
    SELECT user_id INTO student_user_id FROM users WHERE username = 'student1' LIMIT 1;
    
    -- 添加参与者
    INSERT INTO activity_participants (
        activity_id, user_id, status, credit_points_earned
    ) VALUES (
        activity_id, student_user_id, 'approved', 5
    );
    
    RAISE NOTICE '测试活动ID: %, 参与者已添加', activity_id;
END $$;

-- 检查是否自动创建了申请
SELECT '=== 检查自动创建的申请 ===' as test_name;
SELECT 
    a.id as application_id,
    a.user_id,
    a.activity_id,
    a.title,
    a.status,
    a.credit_points_requested,
    a.credit_points_approved
FROM applications a
JOIN credit_activities ca ON a.activity_id = ca.id
WHERE ca.title = '测试活动';

-- 测试活动更新触发器
SELECT '=== 测试活动更新触发器 ===' as test_name;
UPDATE credit_activities 
SET title = '更新后的测试活动', credit_points = 8
WHERE title = '测试活动';

-- 检查申请是否同步更新
SELECT '=== 检查申请同步更新 ===' as test_name;
SELECT 
    a.id as application_id,
    a.title,
    a.credit_points_requested,
    a.updated_at
FROM applications a
JOIN credit_activities ca ON a.activity_id = ca.id
WHERE ca.title = '更新后的测试活动';

-- 测试参与者学分更新触发器
SELECT '=== 测试参与者学分更新触发器 ===' as test_name;
UPDATE activity_participants 
SET credit_points_earned = 8
WHERE activity_id = (SELECT id FROM credit_activities WHERE title = '更新后的测试活动' LIMIT 1);

-- 检查申请学分是否同步更新
SELECT '=== 检查申请学分同步更新 ===' as test_name;
SELECT 
    a.id as application_id,
    a.credit_points_approved,
    a.updated_at
FROM applications a
JOIN credit_activities ca ON a.activity_id = ca.id
WHERE ca.title = '更新后的测试活动';

-- ========================================
-- 2. 测试活动更新同步存储过程
-- ========================================

-- 测试活动更新同步存储过程
SELECT '=== 测试活动更新同步存储过程 ===' as test_name;
SELECT update_activity_and_sync_applications(
    (SELECT id FROM credit_activities WHERE title = '更新后的测试活动' LIMIT 1),
    '存储过程更新活动',
    '通过存储过程更新的活动描述',
    10
);

-- 检查存储过程更新结果
SELECT '=== 检查存储过程更新结果 ===' as test_name;
SELECT 
    ca.id as activity_id,
    ca.title as activity_title,
    ca.description as activity_description,
    ca.credit_points as activity_credit_points,
    a.title as application_title,
    a.description as application_description,
    a.credit_points_requested as application_credit_points
FROM credit_activities ca
LEFT JOIN applications a ON ca.id = a.activity_id
WHERE ca.title = '存储过程更新活动';

-- ========================================
-- 3. 测试活动删除触发器
-- ========================================

-- 创建另一个测试活动用于删除测试
SELECT '=== 创建删除测试活动 ===' as test_name;
INSERT INTO credit_activities (
    title, description, activity_type, credit_points, 
    start_time, end_time, organizer_id, status
) VALUES (
    '删除测试活动', '这个活动将被删除', 'lecture', 3,
    CURRENT_TIMESTAMP + INTERVAL '1 day', CURRENT_TIMESTAMP + INTERVAL '2 days',
    (SELECT user_id FROM users WHERE username = 'admin' LIMIT 1),
    'published'
) RETURNING id;

-- 为删除测试活动添加申请
DO $$
DECLARE
    activity_id INTEGER;
    student_user_id UUID;
BEGIN
    -- 获取活动ID
    SELECT id INTO activity_id FROM credit_activities WHERE title = '删除测试活动' LIMIT 1;
    
    -- 获取学生用户ID
    SELECT user_id INTO student_user_id FROM users WHERE username = 'student1' LIMIT 1;
    
    -- 添加参与者
    INSERT INTO activity_participants (
        activity_id, user_id, status, credit_points_earned
    ) VALUES (
        activity_id, student_user_id, 'approved', 3
    );
    
    RAISE NOTICE '删除测试活动ID: %, 参与者已添加', activity_id;
END $$;

-- 检查删除前的申请数量
SELECT '=== 删除前的申请数量 ===' as test_name;
SELECT COUNT(*) as applications_before_delete
FROM applications a
JOIN credit_activities ca ON a.activity_id = ca.id
WHERE ca.title = '删除测试活动';

-- 删除活动
SELECT '=== 删除活动 ===' as test_name;
DELETE FROM credit_activities WHERE title = '删除测试活动';

-- 检查删除后的申请状态
SELECT '=== 删除后的申请状态 ===' as test_name;
SELECT 
    a.id as application_id,
    a.title,
    a.deleted_at,
    a.status
FROM applications a
WHERE a.title = '删除测试活动';

-- ========================================
-- 4. 测试用户数据验证触发器
-- ========================================

-- 测试用户名唯一性验证
SELECT '=== 测试用户名唯一性验证 ===' as test_name;
DO $$
BEGIN
    BEGIN
        INSERT INTO users (username, password, email, real_name, user_type)
        VALUES ('admin', 'password', 'duplicate@test.com', '重复用户名', 'student');
        RAISE NOTICE '用户名唯一性验证失败';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE '用户名唯一性验证成功: %', SQLERRM;
    END;
END $$;

-- 测试邮箱唯一性验证
SELECT '=== 测试邮箱唯一性验证 ===' as test_name;
DO $$
BEGIN
    BEGIN
        INSERT INTO users (username, password, email, real_name, user_type)
        VALUES ('duplicate_email', 'password', 'admin@credit-management.com', '重复邮箱', 'student');
        RAISE NOTICE '邮箱唯一性验证失败';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE '邮箱唯一性验证成功: %', SQLERRM;
    END;
END $$;

-- ========================================
-- 5. 清理测试数据
-- ========================================

-- 清理测试数据
SELECT '=== 清理测试数据 ===' as test_name;
DELETE FROM activity_participants 
WHERE activity_id IN (
    SELECT id FROM credit_activities 
    WHERE title IN ('存储过程更新活动', '更新后的测试活动')
);

DELETE FROM applications 
WHERE title IN ('存储过程更新活动', '更新后的测试活动');

DELETE FROM credit_activities 
WHERE title IN ('存储过程更新活动', '更新后的测试活动');

SELECT '测试完成！' as completion_message; 