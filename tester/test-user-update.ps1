# 测试用户更新API
Write-Host "=== 测试用户更新API ===" -ForegroundColor Green

# 1. 注册学生A
Write-Host "`n1. 注册学生A..." -ForegroundColor Yellow
$registerResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/register" -Method POST -ContentType "application/json" -Body @"
{
    "username": "studentupdate99",
    "password": "Password123",
    "email": "student_update99@example.com",
    "phone": "13800138099",
    "real_name": "测试学生",
    "user_type": "student",
    "student_id": "20230099",
    "college": "计算机学院",
    "major": "软件工程",
    "class": "软件2301",
    "grade": "2023"
}
"@

if ($registerResponse.code -eq 0) {
    Write-Host "学生A注册成功" -ForegroundColor Green
} else {
    Write-Host "学生A注册失败: $($registerResponse.message)" -ForegroundColor Red
    exit 1
}

# 2. 登录获取token
Write-Host "`n2. 登录获取token..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body @"
{
    "username": "studentupdate99",
    "password": "Password123"
}
"@

if ($loginResponse.code -eq 0) {
    $token = $loginResponse.data.token
    Write-Host "登录成功，获取到token" -ForegroundColor Green
} else {
    Write-Host "登录失败: $($loginResponse.message)" -ForegroundColor Red
    exit 1
}

# 3. 获取当前用户信息
Write-Host "`n3. 获取当前用户信息..." -ForegroundColor Yellow
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

$profileResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/profile" -Method GET -Headers $headers

if ($profileResponse.code -eq 0) {
    Write-Host "获取用户信息成功" -ForegroundColor Green
    Write-Host "当前邮箱: $($profileResponse.data.email)" -ForegroundColor Cyan
    Write-Host "当前手机: $($profileResponse.data.phone)" -ForegroundColor Cyan
} else {
    Write-Host "获取用户信息失败: $($profileResponse.message)" -ForegroundColor Red
    exit 1
}

# 4. 更新用户信息
Write-Host "`n4. 更新用户信息..." -ForegroundColor Yellow
$updateResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/profile" -Method PUT -Headers $headers -Body @"
{
    "email": "updated_email@example.com",
    "phone": "13900139000",
    "real_name": "更新后的姓名",
    "college": "信息学院",
    "major": "计算机科学",
    "class": "计科2301"
}
"@

if ($updateResponse.code -eq 0) {
    Write-Host "更新用户信息成功" -ForegroundColor Green
    Write-Host "更新后邮箱: $($updateResponse.data.email)" -ForegroundColor Cyan
    Write-Host "更新后手机: $($updateResponse.data.phone)" -ForegroundColor Cyan
    Write-Host "更新后姓名: $($updateResponse.data.real_name)" -ForegroundColor Cyan
} else {
    Write-Host "更新用户信息失败: $($updateResponse.message)" -ForegroundColor Red
    Write-Host "错误码: $($updateResponse.code)" -ForegroundColor Red
    exit 1
}

# 5. 验证更新结果
Write-Host "`n5. 验证更新结果..." -ForegroundColor Yellow
$verifyResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/profile" -Method GET -Headers $headers

if ($verifyResponse.code -eq 0) {
    Write-Host "验证成功" -ForegroundColor Green
    Write-Host "验证邮箱: $($verifyResponse.data.email)" -ForegroundColor Cyan
    Write-Host "验证手机: $($verifyResponse.data.phone)" -ForegroundColor Cyan
    Write-Host "验证姓名: $($verifyResponse.data.real_name)" -ForegroundColor Cyan
    
    if ($verifyResponse.data.email -eq "updated_email@example.com" -and 
        $verifyResponse.data.phone -eq "13900139000" -and 
        $verifyResponse.data.real_name -eq "更新后的姓名") {
        Write-Host "`n✅ 用户更新API测试通过！" -ForegroundColor Green
    } else {
        Write-Host "`n❌ 用户更新API测试失败：数据未正确更新" -ForegroundColor Red
    }
} else {
    Write-Host "验证失败: $($verifyResponse.message)" -ForegroundColor Red
    exit 1
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green 