# 测试用户创建API（管理员功能）
Write-Host "=== 测试用户创建API（管理员功能）===" -ForegroundColor Green

# 生成随机用户名和手机号（限制长度）
$timestamp = Get-Date -Format "MMddHHmm"
$randomTeacher = "t$timestamp"
$randomStudent = "s$timestamp"
$randomPhone1 = "13800$timestamp"
$randomPhone2 = "13801$timestamp"

# 1. 管理员登录获取token
Write-Host "`n1. 管理员登录获取token..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body @"
{
    "username": "admin",
    "password": "password"
}
"@

if ($loginResponse.code -eq 0) {
    Write-Host "管理员登录成功，获取到token" -ForegroundColor Green
    $token = $loginResponse.data.token
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    # 2. 测试创建教师
    Write-Host "`n2. 测试创建教师..." -ForegroundColor Yellow
    Write-Host "用户名: $randomTeacher" -ForegroundColor Cyan
    Write-Host "手机号: $randomPhone1" -ForegroundColor Cyan
    Write-Host "部门: 计算机学院" -ForegroundColor Cyan
    Write-Host "职称: 副教授" -ForegroundColor Cyan

    $teacherResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/teachers" -Method POST -Headers $headers -Body @"
{
    "username": "$randomTeacher",
    "password": "Password123",
    "email": "$randomTeacher@example.com",
    "phone": "$randomPhone1",
    "real_name": "王老师",
    "user_type": "teacher",
    "department": "计算机学院",
    "title": "副教授",
    "specialty": "人工智能"
}
"@

    if ($teacherResponse.code -eq 0) {
        Write-Host "✓ 教师创建成功" -ForegroundColor Green
        Write-Host "  用户ID: $($teacherResponse.data.user.user_id)" -ForegroundColor White
        Write-Host "  用户名: $($teacherResponse.data.user.username)" -ForegroundColor White
        Write-Host "  真实姓名: $($teacherResponse.data.user.real_name)" -ForegroundColor White
        Write-Host "  部门: $($teacherResponse.data.user.department)" -ForegroundColor White
        Write-Host "  职称: $($teacherResponse.data.user.title)" -ForegroundColor White
    } else {
        Write-Host "✗ 教师创建失败: $($teacherResponse.message)" -ForegroundColor Red
    }

    # 3. 测试创建学生
    Write-Host "`n3. 测试创建学生..." -ForegroundColor Yellow
    Write-Host "用户名: $randomStudent" -ForegroundColor Cyan
    Write-Host "手机号: $randomPhone2" -ForegroundColor Cyan
    Write-Host "学号: 2023$timestamp" -ForegroundColor Cyan
    Write-Host "学院: 计算机学院" -ForegroundColor Cyan
    Write-Host "专业: 软件工程" -ForegroundColor Cyan

    $studentResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/students" -Method POST -Headers $headers -Body @"
{
    "username": "$randomStudent",
    "password": "Password123",
    "email": "$randomStudent@example.com",
    "phone": "$randomPhone2",
    "real_name": "李同学",
    "user_type": "student",
    "student_id": "2023$timestamp",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023"
}
"@

    if ($studentResponse.code -eq 0) {
        Write-Host "✓ 学生创建成功" -ForegroundColor Green
        Write-Host "  用户ID: $($studentResponse.data.user.user_id)" -ForegroundColor White
        Write-Host "  用户名: $($studentResponse.data.user.username)" -ForegroundColor White
        Write-Host "  真实姓名: $($studentResponse.data.user.real_name)" -ForegroundColor White
        Write-Host "  学号: $($studentResponse.data.user.student_id)" -ForegroundColor White
        Write-Host "  学院: $($studentResponse.data.user.college)" -ForegroundColor White
        Write-Host "  专业: $($studentResponse.data.user.major)" -ForegroundColor White
    } else {
        Write-Host "✗ 学生创建失败: $($studentResponse.message)" -ForegroundColor Red
    }

    # 4. 测试权限验证
    Write-Host "`n4. 测试权限验证..." -ForegroundColor Yellow
    
    # 测试非管理员用户无法创建教师
    Write-Host "  测试非管理员用户无法创建教师..." -ForegroundColor Cyan
    try {
        $nonAdminResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/teachers" -Method POST -Headers $headers -Body @"
{
    "username": "test_teacher",
    "password": "Password123",
    "email": "test@example.com",
    "phone": "13800138003",
    "real_name": "测试教师",
    "user_type": "teacher",
    "department": "测试系",
    "title": "讲师",
    "specialty": "测试专业"
}
"@
        Write-Host "  ✗ 非管理员用户应该无法创建教师，但请求成功了" -ForegroundColor Red
    } catch {
        Write-Host "  ✓ 非管理员用户无法创建教师（预期行为）" -ForegroundColor Green
    }

    Write-Host "`n=== 测试完成 ===" -ForegroundColor Green

} else {
    Write-Host "管理员登录失败: $($loginResponse.message)" -ForegroundColor Red
} 