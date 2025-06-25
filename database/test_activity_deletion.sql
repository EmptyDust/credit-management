-- 测试删除活动存储过程的脚本
-- 使用方法：在PostgreSQL中执行此脚本来测试删除活动功能

-- 1. 创建测试用户
INSERT INTO users (user_id, username, password, email, real_name, user_type, status)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'test_student', 'password123', 'student@test.com', '测试学生', 'student', 'active'),
    ('550e8400-e29b-41d4-a716-446655440002', 'test_teacher', 'password123', 'teacher@test.com', '测试教师', 'teacher', 'active'),
    ('550e8400-e29b-41d4-a716-446655440003', 'test_admin', 'password123', 'admin@test.com', '测试管理员', 'admin', 'active')
ON CONFLICT (user_id) DO NOTHING;

-- 2. 创建测试活动
INSERT INTO credit_activities (id, title, description, start_date, end_date, status, category, requirements, owner_id)
VALUES 
    ('660e8400-e29b-41d4-a716-446655440001', '测试活动1', '这是测试活动1的描述', '2024-01-01', '2024-01-31', 'draft', '创新创业', '无特殊要求', '550e8400-e29b-41d4-a716-446655440002'),
    ('660e8400-e29b-41d4-a716-446655440002', '测试活动2', '这是测试活动2的描述', '2024-02-01', '2024-02-28', 'pending_review', '学科竞赛', '需要提交作品', '550e8400-e29b-41d4-a716-446655440002'),
    ('660e8400-e29b-41d4-a716-446655440003', '测试活动3', '这是测试活动3的描述', '2024-03-01', '2024-03-31', 'approved', '志愿服务', '需要参加培训', '550e8400-e29b-41d4-a716-446655440002')
ON CONFLICT (id) DO NOTHING;

-- 3. 添加参与者
INSERT INTO activity_participants (activity_id, user_id, credits)
VALUES 
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 2.0),
    ('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 3.0),
    ('660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 1.5)
ON CONFLICT (activity_id, user_id) DO NOTHING;

-- 4. 测试删除活动功能

-- 测试1：教师删除自己的活动（应该成功）
SELECT '测试1：教师删除自己的活动' as test_case;
SELECT delete_activity_with_permission_check(
    '660e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440002',
    'teacher'
);

-- 验证删除结果
SELECT '验证活动1是否被删除：' as verification;
SELECT id, title, deleted_at FROM credit_activities WHERE id = '660e8400-e29b-41d4-a716-446655440001';

SELECT '验证相关参与者是否被删除：' as verification;
SELECT activity_id, user_id, deleted_at FROM activity_participants WHERE activity_id = '660e8400-e29b-41d4-a716-446655440001';

-- 测试2：管理员删除任何活动（应该成功）
SELECT '测试2：管理员删除活动' as test_case;
SELECT delete_activity_with_permission_check(
    '660e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440003',
    'admin'
);

-- 测试3：学生尝试删除教师的活动（应该失败）
SELECT '测试3：学生尝试删除教师的活动' as test_case;
SELECT delete_activity_with_permission_check(
    '660e8400-e29b-41d4-a716-446655440003',
    '550e8400-e29b-41d4-a716-446655440001',
    'student'
);

-- 测试4：批量删除活动
SELECT '测试4：批量删除活动' as test_case;
SELECT batch_delete_activities(
    ARRAY['660e8400-e29b-41d4-a716-446655440003'],
    '550e8400-e29b-41d4-a716-446655440003',
    'admin'
);

-- 测试5：获取用户可删除的活动列表
SELECT '测试5：获取教师可删除的活动列表' as test_case;
SELECT * FROM get_user_deletable_activities(
    '550e8400-e29b-41d4-a716-446655440002',
    'teacher'
);

SELECT '测试6：获取管理员可删除的活动列表' as test_case;
SELECT * FROM get_user_deletable_activities(
    '550e8400-e29b-41d4-a716-446655440003',
    'admin'
);

-- 测试7：恢复已删除的活动（仅管理员）
SELECT '测试7：管理员恢复已删除的活动' as test_case;
SELECT restore_deleted_activity(
    '660e8400-e29b-41d4-a716-446655440001',
    'admin'
);

-- 验证恢复结果
SELECT '验证活动1是否被恢复：' as verification;
SELECT id, title, deleted_at FROM credit_activities WHERE id = '660e8400-e29b-41d4-a716-446655440001';

-- 清理测试数据（可选）
-- DELETE FROM activity_participants WHERE activity_id IN ('660e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440003');
-- DELETE FROM credit_activities WHERE id IN ('660e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440003');
-- DELETE FROM users WHERE user_id IN ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003'); 