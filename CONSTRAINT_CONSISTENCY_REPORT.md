# 约束条件一致性检查报告

## 概述
本报告分析了学分管理系统中数据库、后端API、前端表单之间的约束条件一致性，发现了多个不一致的地方需要修复。

## 1. 用户模型约束条件对比

### 1.1 数据库约束 (init.sql)
```sql
-- users表约束
username VARCHAR(50) UNIQUE NOT NULL
password VARCHAR(255) NOT NULL
email VARCHAR(100) UNIQUE NOT NULL
phone VARCHAR(20) UNIQUE  -- 可为空
real_name VARCHAR(100) NOT NULL
user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('student', 'teacher', 'admin'))
status VARCHAR(20) NOT NULL DEFAULT 'active'
student_id VARCHAR(20) UNIQUE  -- 可为空
college VARCHAR(100)  -- 可为空
major VARCHAR(100)  -- 可为空
class VARCHAR(50)  -- 可为空
grade VARCHAR(10)  -- 可为空
department VARCHAR(100)  -- 可为空
title VARCHAR(50)  -- 可为空
specialty VARCHAR(200)  -- 可为空
```

### 1.2 后端API约束 (user-service/models/user.go)
```go
// UserRequest 约束
Username  string `json:"username" binding:"required,min=3,max=20,alphanum"`
Password  string `json:"password" binding:"required,min=8"`
Email     string `json:"email" binding:"required,email"`
Phone     string `json:"phone" binding:"omitempty,len=11"`
RealName  string `json:"real_name" binding:"required,min=2,max=50"`
UserType  string `json:"user_type" binding:"required,oneof=student teacher"`
StudentID string `json:"student_id" binding:"omitempty,len=8"`
College   string `json:"college" binding:"omitempty,max=100"`
Major     string `json:"major" binding:"omitempty,max=100"`
Class     string `json:"class" binding:"omitempty,max=50"`
Grade     string `json:"grade" binding:"omitempty,len=4"`
Department string `json:"department" binding:"omitempty,max=100"`
Title     string `json:"title" binding:"omitempty,max=50"`
Specialty string `json:"specialty" binding:"omitempty,max=200"`
```

### 1.3 前端表单约束 (Students.tsx)
```typescript
// 前端验证规则
username: z.string().min(1, "用户名不能为空").max(20, "用户名最多20个字符")
password: z.string().min(8, "密码至少8个字符")
  .regex(/[A-Z]/, "密码必须包含至少一个大写字母")
  .regex(/[a-z]/, "密码必须包含至少一个小写字母")
  .regex(/[0-9]/, "密码必须包含至少一个数字")
student_id: z.string().optional()
real_name: z.string().min(1, "姓名不能为空").max(50, "姓名最多50个字符")
college: z.string().min(1, "学院不能为空")
major: z.string().min(1, "专业不能为空")
class: z.string().min(1, "班级不能为空")
phone: z.string().regex(/^1[3-9]\d{9}$/, "请输入有效的11位手机号").optional()
email: z.string().email({ message: "请输入有效的邮箱地址" }).optional()
grade: z.string().min(1, "年级不能为空")
```

### 1.4 注册表单约束 (Register.tsx)
```typescript
// 学生注册验证规则
username: z.string()
  .min(3, "用户名至少3个字符")
  .max(20, "用户名最多20个字符")
  .regex(/^[a-zA-Z0-9_]+$/, "用户名只能包含字母、数字和下划线")
password: z.string()
  .min(8, "密码至少8个字符")
  .regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, "密码必须包含大小写字母和数字")
email: z.string().email("请输入有效的邮箱地址")
phone: z.string()
  .min(11, "手机号必须是11位数字")
  .max(11, "手机号必须是11位数字")
  .regex(/^1[3-9]\d{9}$/, "请输入有效的手机号")
real_name: z.string().min(2, "真实姓名至少2个字符").max(50, "真实姓名最多50个字符")
student_id: z.string()
  .min(8, "学号必须是8位数字")
  .max(8, "学号必须是8位数字")
  .regex(/^\d{8}$/, "学号必须是8位数字")
college: z.string().min(1, "请选择学院")
major: z.string().min(1, "请选择专业")
class: z.string().min(1, "请选择班级")
grade: z.string().length(4, "年级必须是4位数字").regex(/^\d{4}$/, "年级必须是4位数字")
```

## 2. 发现的不一致问题

### 2.1 用户名约束不一致
- **数据库**: VARCHAR(50) - 最大50字符
- **后端API**: min=3, max=20, alphanum - 3-20字符，只允许字母数字
- **前端Students**: min=1, max=20 - 1-20字符
- **前端Register**: min=3, max=20, 只允许字母数字下划线

**问题**: 数据库允许50字符，但API和前端都限制为20字符，且字符集规则不一致。

### 2.2 密码约束不一致
- **数据库**: VARCHAR(255) - 无特殊要求
- **后端API**: min=8 - 最少8字符
- **前端Students**: min=8 + 必须包含大小写字母和数字
- **前端Register**: min=8 + 必须包含大小写字母和数字

**问题**: 后端API没有要求密码复杂度，前端有复杂度要求。

### 2.3 手机号约束不一致
- **数据库**: VARCHAR(20) - 最大20字符
- **后端API**: omitempty, len=11 - 可选，必须是11位
- **前端Students**: 可选，11位，1[3-9]开头
- **前端Register**: 必填，11位，1[3-9]开头

**问题**: 前端注册要求必填，但后端API和数据库都允许为空。

### 2.4 学号约束不一致
- **数据库**: VARCHAR(20) - 最大20字符
- **后端API**: omitempty, len=8 - 可选，必须是8位
- **前端Students**: 可选
- **前端Register**: 必填，8位数字

**问题**: 前端注册要求必填，但后端API和数据库都允许为空。

### 2.5 年级约束不一致
- **数据库**: VARCHAR(10) - 最大10字符
- **后端API**: omitempty, len=4 - 可选，必须是4位
- **前端Students**: 必填
- **前端Register**: 必填，4位数字

**问题**: 前端要求必填，但后端API和数据库都允许为空。

## 3. 活动模型约束条件对比

### 3.1 数据库约束
```sql
-- credit_activities表约束
title VARCHAR(200) NOT NULL
description TEXT
start_date DATE NOT NULL
end_date DATE NOT NULL
status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_review', 'approved', 'rejected'))
category VARCHAR(100) NOT NULL
requirements TEXT
```

### 3.2 后端API约束
```go
// ActivityRequest 约束
Title        string `json:"title" binding:"required"`
Description  string `json:"description"`
StartDate    string `json:"start_date"`
EndDate      string `json:"end_date"`
Category     string `json:"category"`
Requirements string `json:"requirements"`
```

### 3.3 活动状态约束
- **数据库**: CHECK (status IN ('draft', 'pending_review', 'approved', 'rejected'))
- **后端模型**: 常量定义一致
- **前端**: 需要检查状态值使用

## 4. 附件模型约束条件对比

### 4.1 数据库约束
```sql
-- attachments表约束
file_name VARCHAR(255) NOT NULL
original_name VARCHAR(255) NOT NULL
file_size BIGINT NOT NULL
file_type VARCHAR(20) NOT NULL
file_category VARCHAR(50) NOT NULL
description TEXT
md5_hash VARCHAR(32) UNIQUE
```

### 4.2 后端模型约束
```go
// 文件大小限制
const MaxFileSize = 20 * 1024 * 1024  // 20MB

// 批量上传文件数量限制
const MaxBatchUploadCount = 10

// 支持的文件类型映射
var SupportedFileTypes = map[string]string{
    ".pdf": CategoryDocument,
    ".doc": CategoryDocument,
    // ... 更多文件类型
}
```

## 5. 修复建议

### 5.1 统一用户名约束
```sql
-- 修改数据库约束
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(20);
ALTER TABLE users ADD CONSTRAINT username_format CHECK (username ~ '^[a-zA-Z0-9_]+$');
```

### 5.2 统一密码约束
```go
// 修改后端API约束
Password string `json:"password" binding:"required,min=8,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
```

### 5.3 统一手机号约束
```go
// 修改后端API约束，使手机号在注册时必填
Phone string `json:"phone" binding:"required,len=11,startswith=1"`
```

### 5.4 统一学号约束
```go
// 修改后端API约束，使学生ID在注册时必填
StudentID string `json:"student_id" binding:"required,len=8,numeric"`
```

### 5.5 统一年级约束
```go
// 修改后端API约束，使年级在注册时必填
Grade string `json:"grade" binding:"required,len=4,numeric"`
```

### 5.6 添加数据库约束检查
```sql
-- 添加手机号格式约束
ALTER TABLE users ADD CONSTRAINT phone_format CHECK (phone ~ '^1[3-9]\d{9}$');

-- 添加学号格式约束
ALTER TABLE users ADD CONSTRAINT student_id_format CHECK (student_id ~ '^\d{8}$');

-- 添加年级格式约束
ALTER TABLE users ADD CONSTRAINT grade_format CHECK (grade ~ '^\d{4}$');
```

## 6. 实施计划

1. **第一阶段**: 修改数据库约束，确保与API一致
2. **第二阶段**: 更新后端API验证规则，确保与前端一致
3. **第三阶段**: 验证前端表单，确保所有约束条件统一
4. **第四阶段**: 全面测试，确保数据一致性

## 7. 测试验证

需要创建测试用例验证：
- 用户名格式和长度限制
- 密码复杂度要求
- 手机号格式验证
- 学号格式验证
- 年级格式验证
- 活动状态值验证
- 文件上传限制

## 8. 总结

当前系统存在多处约束条件不一致的问题，主要集中在用户注册和更新功能上。建议按照上述修复建议逐步统一约束条件，确保数据库、后端API、前端表单的验证规则完全一致。 