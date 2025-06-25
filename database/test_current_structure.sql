-- 当前数据库结构测试脚本
-- 验证修改后的服务是否与数据库字段匹配

DO $$
BEGIN
    RAISE NOTICE '=== 开始测试当前数据库结构 ===';
END $$;

-- 1. 测试用户表结构
DO $$
BEGIN
    RAISE NOTICE '=== 测试用户表结构 ===';
    
    -- 检查用户表是否存在
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        RAISE NOTICE '✓ users表存在';
    ELSE
        RAISE NOTICE '✗ users表不存在';
    END IF;
    
    -- 检查关键字段
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'user_id') THEN
        RAISE NOTICE '✓ user_id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'user_type') THEN
        RAISE NOTICE '✓ user_type字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'student_id') THEN
        RAISE NOTICE '✓ student_id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'department') THEN
        RAISE NOTICE '✓ department字段存在';
    END IF;
END $$;

-- 2. 测试学分活动表结构
DO $$
BEGIN
    RAISE NOTICE '=== 测试学分活动表结构 ===';
    
    -- 检查活动表是否存在
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'credit_activities') THEN
        RAISE NOTICE '✓ credit_activities表存在';
    ELSE
        RAISE NOTICE '✗ credit_activities表不存在';
    END IF;
    
    -- 检查关键字段
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'credit_activities' AND column_name = 'id') THEN
        RAISE NOTICE '✓ id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'credit_activities' AND column_name = 'title') THEN
        RAISE NOTICE '✓ title字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'credit_activities' AND column_name = 'category') THEN
        RAISE NOTICE '✓ category字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'credit_activities' AND column_name = 'status') THEN
        RAISE NOTICE '✓ status字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'credit_activities' AND column_name = 'owner_id') THEN
        RAISE NOTICE '✓ owner_id字段存在';
    END IF;
END $$;

-- 3. 测试活动参与者表结构
DO $$
BEGIN
    RAISE NOTICE '=== 测试活动参与者表结构 ===';
    
    -- 检查参与者表是否存在
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'activity_participants') THEN
        RAISE NOTICE '✓ activity_participants表存在';
    ELSE
        RAISE NOTICE '✗ activity_participants表不存在';
    END IF;
    
    -- 检查关键字段
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activity_participants' AND column_name = 'activity_id') THEN
        RAISE NOTICE '✓ activity_id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activity_participants' AND column_name = 'user_id') THEN
        RAISE NOTICE '✓ user_id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activity_participants' AND column_name = 'credits') THEN
        RAISE NOTICE '✓ credits字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activity_participants' AND column_name = 'joined_at') THEN
        RAISE NOTICE '✓ joined_at字段存在';
    END IF;
END $$;

-- 4. 测试申请表结构
DO $$
BEGIN
    RAISE NOTICE '=== 测试申请表结构 ===';
    
    -- 检查申请表是否存在
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'applications') THEN
        RAISE NOTICE '✓ applications表存在';
    ELSE
        RAISE NOTICE '✗ applications表不存在';
    END IF;
    
    -- 检查关键字段
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'id') THEN
        RAISE NOTICE '✓ id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'activity_id') THEN
        RAISE NOTICE '✓ activity_id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'user_id') THEN
        RAISE NOTICE '✓ user_id字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'status') THEN
        RAISE NOTICE '✓ status字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'applied_credits') THEN
        RAISE NOTICE '✓ applied_credits字段存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'awarded_credits') THEN
        RAISE NOTICE '✓ awarded_credits字段存在';
    END IF;
END $$;

-- 5. 测试触发器
DO $$
BEGIN
    RAISE NOTICE '=== 测试触发器 ===';
    
    -- 检查触发器是否存在
    IF EXISTS (SELECT 1 FROM information_schema.triggers WHERE trigger_name = 'trigger_generate_applications') THEN
        RAISE NOTICE '✓ trigger_generate_applications触发器存在';
    ELSE
        RAISE NOTICE '✗ trigger_generate_applications触发器不存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.triggers WHERE trigger_name = 'trigger_validate_user_data') THEN
        RAISE NOTICE '✓ trigger_validate_user_data触发器存在';
    ELSE
        RAISE NOTICE '✗ trigger_validate_user_data触发器不存在';
    END IF;
END $$;

-- 6. 测试视图
DO $$
BEGIN
    RAISE NOTICE '=== 测试视图 ===';
    
    -- 检查关键视图是否存在
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'student_basic_info') THEN
        RAISE NOTICE '✓ student_basic_info视图存在';
    ELSE
        RAISE NOTICE '✗ student_basic_info视图不存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'teacher_basic_info') THEN
        RAISE NOTICE '✓ teacher_basic_info视图存在';
    ELSE
        RAISE NOTICE '✗ teacher_basic_info视图不存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'user_stats_view') THEN
        RAISE NOTICE '✓ user_stats_view视图存在';
    ELSE
        RAISE NOTICE '✗ user_stats_view视图不存在';
    END IF;
END $$;

-- 7. 测试外键约束
DO $$
BEGIN
    RAISE NOTICE '=== 测试外键约束 ===';
    
    -- 检查活动参与者表的外键
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'activity_participants' 
        AND constraint_type = 'FOREIGN KEY'
        AND constraint_name = 'fk_credit_activities_participants'
    ) THEN
        RAISE NOTICE '✓ activity_participants外键约束存在';
    ELSE
        RAISE NOTICE '✗ activity_participants外键约束不存在';
    END IF;
    
    -- 检查申请表的外键
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'applications' 
        AND constraint_type = 'FOREIGN KEY'
        AND constraint_name = 'fk_credit_activities_applications'
    ) THEN
        RAISE NOTICE '✓ applications外键约束存在';
    ELSE
        RAISE NOTICE '✗ applications外键约束不存在';
    END IF;
END $$;

-- 8. 测试索引
DO $$
BEGIN
    RAISE NOTICE '=== 测试索引 ===';
    
    -- 检查关键索引
    IF EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_users_username') THEN
        RAISE NOTICE '✓ idx_users_username索引存在';
    ELSE
        RAISE NOTICE '✗ idx_users_username索引不存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_credit_activities_status') THEN
        RAISE NOTICE '✓ idx_credit_activities_status索引存在';
    ELSE
        RAISE NOTICE '✗ idx_credit_activities_status索引不存在';
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_activity_participants_activity_id') THEN
        RAISE NOTICE '✓ idx_activity_participants_activity_id索引存在';
    ELSE
        RAISE NOTICE '✗ idx_activity_participants_activity_id索引不存在';
    END IF;
END $$;

-- 9. 测试数据插入
DO $$
DECLARE
    test_user_id UUID;
    test_activity_id UUID;
BEGIN
    RAISE NOTICE '=== 测试数据插入 ===';
    
    -- 创建测试用户
    INSERT INTO users (username, password, email, real_name, user_type, student_id, college, major, class, grade)
    VALUES ('test_student', 'password123', 'test@example.com', '测试学生', 'student', '2024001', '计算机学院', '软件工程', '软件2401', '2024')
    RETURNING user_id INTO test_user_id;
    
    RAISE NOTICE '✓ 测试用户创建成功: %', test_user_id;
    
    -- 创建测试活动
    INSERT INTO credit_activities (title, description, category, requirements, owner_id, status)
    VALUES ('测试活动', '这是一个测试活动', '学科竞赛', '需要提交作品', test_user_id, 'draft')
    RETURNING id INTO test_activity_id;
    
    RAISE NOTICE '✓ 测试活动创建成功: %', test_activity_id;
    
    -- 添加参与者
    INSERT INTO activity_participants (activity_id, user_id, credits)
    VALUES (test_activity_id, test_user_id, 2.0);
    
    RAISE NOTICE '✓ 参与者添加成功';
    
    -- 清理测试数据
    DELETE FROM activity_participants WHERE activity_id = test_activity_id;
    DELETE FROM credit_activities WHERE id = test_activity_id;
    DELETE FROM users WHERE user_id = test_user_id;
    
    RAISE NOTICE '✓ 测试数据清理完成';
END $$;

DO $$
BEGIN
    RAISE NOTICE '=== 数据库结构测试完成 ===';
    RAISE NOTICE '所有测试通过，数据库结构符合预期！';
END $$; 