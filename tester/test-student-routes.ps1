# 测试学生创建路由
Write-Host "=== 测试学生创建路由 ===" -ForegroundColor Green

# 1. 管理员登录获取token
Write-Host "`n1. 管理员登录获取token..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body @"
{
    "username": "admin",
    "password": "adminpassword"
}
"@

if ($loginResponse.code -eq 0) {
    Write-Host "管理员登录成功，获取到token" -ForegroundColor Green
    $token = $loginResponse.data.token
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    # 2. 测试不同的学生创建路由
    Write-Host "`n2. 测试不同的学生创建路由..." -ForegroundColor Yellow
    
    $timestamp = Get-Date -Format "MMddHHmm"
    $randomStudent1 = "s1$timestamp"
    $randomStudent2 = "s2$timestamp"
    $studentID1 = (Get-Random -Minimum 10000000 -Maximum 99999999).ToString()
    $studentID2 = (Get-Random -Minimum 10000000 -Maximum 99999999).ToString()
    
    # 测试 /api/users/students 路由
    Write-Host "  测试 /api/users/students 路由..." -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8084/api/users/students" -Method POST -Headers $headers -Body @"
{
    "username": "$randomStudent1",
    "password": "Password123",
    "email": "$randomStudent1@example.com",
    "phone": "13800138002",
    "real_name": "李同学1",
    "user_type": "student",
    "student_id": "$studentID1",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023"
}
"@
        Write-Host "  ✓ /api/users/students 成功: $($response.code)" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ /api/users/students 失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    # 测试 /api/students 路由
    Write-Host "  测试 /api/students 路由..." -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8084/api/students" -Method POST -Headers $headers -Body @"
{
    "username": "$randomStudent2",
    "password": "Password123",
    "email": "$randomStudent2@example.com",
    "phone": "13800138003",
    "real_name": "李同学2",
    "user_type": "student",
    "student_id": "$studentID2",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023"
}
"@
        Write-Host "  ✓ /api/students 成功: $($response.code)" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ /api/students 失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host "`n=== 测试完成 ===" -ForegroundColor Green

} else {
    Write-Host "管理员登录失败: $($loginResponse.message)" -ForegroundColor Red
}