# 测试用户服务的JWT Token解析
Write-Host "=== 测试用户服务的JWT Token解析 ===" -ForegroundColor Green

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

    # 2. 测试用户服务的认证中间件
    Write-Host "`n2. 测试用户服务的认证中间件..." -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8084/api/users/profile" -Method GET -Headers $headers
        Write-Host "✓ 用户服务认证成功: $($response.code)" -ForegroundColor Green
        Write-Host "  用户类型: $($response.data.user_type)" -ForegroundColor White
        Write-Host "  用户名: $($response.data.username)" -ForegroundColor White
    } catch {
        Write-Host "✗ 用户服务认证失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    # 3. 测试管理员权限中间件
    Write-Host "`n3. 测试管理员权限中间件..." -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8084/api/users" -Method GET -Headers $headers
        Write-Host "✓ 管理员权限验证成功: $($response.code)" -ForegroundColor Green
    } catch {
        Write-Host "✗ 管理员权限验证失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    # 4. 测试创建学生（应该成功）
    Write-Host "`n4. 测试创建学生..." -ForegroundColor Yellow
    $timestamp = Get-Date -Format "MMddHHmm"
    $randomStudent = "s$timestamp"
    
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8084/api/users/students" -Method POST -Headers $headers -Body @"
{
    "username": "$randomStudent",
    "password": "Password123",
    "email": "$randomStudent@example.com",
    "phone": "13800138002",
    "real_name": "李同学",
    "user_type": "student",
    "student_id": "2023$timestamp",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023"
}
"@
        Write-Host "✓ 学生创建成功: $($response.code)" -ForegroundColor Green
    } catch {
        Write-Host "✗ 学生创建失败: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host "`n=== 测试完成 ===" -ForegroundColor Green

} else {
    Write-Host "管理员登录失败: $($loginResponse.message)" -ForegroundColor Red
} 