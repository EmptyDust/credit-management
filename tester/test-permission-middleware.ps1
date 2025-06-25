# 测试权限中间件
Write-Host "=== 测试权限中间件 ===" -ForegroundColor Green

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

    # 2. 测试管理员权限路由
    Write-Host "`n2. 测试管理员权限路由..." -ForegroundColor Yellow
    
    # 测试获取所有用户（需要管理员权限）
    Write-Host "  测试获取所有用户..." -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8084/api/users" -Method GET -Headers $headers
        Write-Host "  ✓ 获取所有用户成功: $($response.code)" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ 获取所有用户失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    # 测试创建教师（需要管理员权限）
    Write-Host "  测试创建教师..." -ForegroundColor Cyan
    $timestamp = Get-Date -Format "MMddHHmm"
    $randomTeacher = "t$timestamp"
    
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8084/api/users/teachers" -Method POST -Headers $headers -Body @"
{
    "username": "$randomTeacher",
    "password": "Password123",
    "email": "$randomTeacher@example.com",
    "phone": "13800138001",
    "real_name": "王老师",
    "user_type": "teacher",
    "department": "计算机学院",
    "title": "副教授",
    "specialty": "人工智能"
}
"@
        Write-Host "  ✓ 创建教师成功: $($response.code)" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ 创建教师失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host "`n=== 测试完成 ===" -ForegroundColor Green

} else {
    Write-Host "管理员登录失败: $($loginResponse.message)" -ForegroundColor Red
} 