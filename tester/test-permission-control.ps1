# 权限控制测试脚本
# 测试基于角色的用户搜索和查看功能

param(
    [string]$BaseUrl = "http://localhost:8084"
)

Write-Host "=== 权限控制测试 ===" -ForegroundColor Green

# 测试数据
$testUsers = @{
    "student" = @{
        "username" = "test_student"
        "password" = "Test123456"
        "email" = "student@test.com"
        "real_name" = "测试学生"
        "student_id" = "20240001"
        "college" = "计算机学院"
        "major" = "软件工程"
        "class" = "软工2401"
        "grade" = "2024"
    }
    "teacher" = @{
        "username" = "test_teacher"
        "password" = "Test123456"
        "email" = "teacher@test.com"
        "real_name" = "测试教师"
        "department" = "计算机系"
        "title" = "副教授"
        "specialty" = "软件工程"
    }
    "admin" = @{
        "username" = "test_admin"
        "password" = "Test123456"
        "email" = "admin@test.com"
        "real_name" = "测试管理员"
    }
}

# 存储token
$tokens = @{}

# 1. 注册测试用户
Write-Host "`n1. 注册测试用户..." -ForegroundColor Yellow

# 注册学生
$studentData = $testUsers.student
$studentResponse = Invoke-RestMethod -Uri "$BaseUrl/api/users/register" -Method POST -ContentType "application/json" -Body ($studentData | ConvertTo-Json)
if ($studentResponse.code -eq 0) {
    Write-Host "✓ 学生注册成功: $($studentData.username)" -ForegroundColor Green
} else {
    Write-Host "✗ 学生注册失败: $($studentResponse.message)" -ForegroundColor Red
}

# 2. 登录获取token
Write-Host "`n2. 登录获取token..." -ForegroundColor Yellow

# 学生登录
$loginData = @{
    "username" = $testUsers.student.username
    "password" = $testUsers.student.password
}
$loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -ContentType "application/json" -Body ($loginData | ConvertTo-Json)
if ($loginResponse.code -eq 0) {
    $tokens.student = $loginResponse.data.token
    Write-Host "✓ 学生登录成功" -ForegroundColor Green
} else {
    Write-Host "✗ 学生登录失败: $($loginResponse.message)" -ForegroundColor Red
}

# 3. 测试学生权限
Write-Host "`n3. 测试学生权限..." -ForegroundColor Yellow

if ($tokens.student) {
    $headers = @{
        "Authorization" = "Bearer $($tokens.student)"
        "Content-Type" = "application/json"
    }

    # 测试学生搜索学生（应该看到基本信息）
    Write-Host "  测试学生搜索学生..." -ForegroundColor Cyan
    $searchResponse = Invoke-RestMethod -Uri "$BaseUrl/api/search/users?user_type=student" -Method GET -Headers $headers
    if ($searchResponse.code -eq 0) {
        Write-Host "  ✓ 学生可以搜索学生信息" -ForegroundColor Green
        Write-Host "  返回用户数量: $($searchResponse.data.total)" -ForegroundColor Gray
    } else {
        Write-Host "  ✗ 学生搜索学生失败: $($searchResponse.message)" -ForegroundColor Red
    }

    # 测试学生搜索教师（应该看到基本信息）
    Write-Host "  测试学生搜索教师..." -ForegroundColor Cyan
    $searchResponse = Invoke-RestMethod -Uri "$BaseUrl/api/search/users?user_type=teacher" -Method GET -Headers $headers
    if ($searchResponse.code -eq 0) {
        Write-Host "  ✓ 学生可以搜索教师信息" -ForegroundColor Green
    } else {
        Write-Host "  ✗ 学生搜索教师失败: $($searchResponse.message)" -ForegroundColor Red
    }

    # 测试学生搜索管理员（应该被拒绝）
    Write-Host "  测试学生搜索管理员..." -ForegroundColor Cyan
    try {
        $searchResponse = Invoke-RestMethod -Uri "$BaseUrl/api/search/users?user_type=admin" -Method GET -Headers $headers
        Write-Host "  ✗ 学生不应该能搜索管理员信息" -ForegroundColor Red
    } catch {
        Write-Host "  ✓ 学生搜索管理员被正确拒绝" -ForegroundColor Green
    }
}

# 4. 测试教师权限（需要管理员创建教师）
Write-Host "`n4. 测试教师权限..." -ForegroundColor Yellow

# 这里需要管理员权限来创建教师，暂时跳过
Write-Host "  需要管理员权限创建教师，跳过教师权限测试" -ForegroundColor Yellow

# 5. 测试管理员权限
Write-Host "`n5. 测试管理员权限..." -ForegroundColor Yellow

# 这里需要管理员token，暂时跳过
Write-Host "  需要管理员token，跳过管理员权限测试" -ForegroundColor Yellow

# 6. 测试权限边界
Write-Host "`n6. 测试权限边界..." -ForegroundColor Yellow

if ($tokens.student) {
    $headers = @{
        "Authorization" = "Bearer $($tokens.student)"
        "Content-Type" = "application/json"
    }

    # 测试无效token
    Write-Host "  测试无效token..." -ForegroundColor Cyan
    $invalidHeaders = @{
        "Authorization" = "Bearer invalid_token"
        "Content-Type" = "application/json"
    }
    try {
        $searchResponse = Invoke-RestMethod -Uri "$BaseUrl/api/search/users" -Method GET -Headers $invalidHeaders
        Write-Host "  ✗ 无效token应该被拒绝" -ForegroundColor Red
    } catch {
        Write-Host "  ✓ 无效token被正确拒绝" -ForegroundColor Green
    }

    # 测试无token
    Write-Host "  测试无token..." -ForegroundColor Cyan
    try {
        $searchResponse = Invoke-RestMethod -Uri "$BaseUrl/api/search/users" -Method GET -ContentType "application/json"
        Write-Host "  ✗ 无token应该被拒绝" -ForegroundColor Red
    } catch {
        Write-Host "  ✓ 无token被正确拒绝" -ForegroundColor Green
    }
}

Write-Host "`n=== 权限控制测试完成 ===" -ForegroundColor Green
Write-Host "注意：完整的权限测试需要管理员权限来创建教师和管理员用户" -ForegroundColor Yellow 