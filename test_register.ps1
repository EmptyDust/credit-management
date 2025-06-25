# 用户注册接口+登录+用户信息获取测试脚本

Write-Host "=== 用户注册接口+登录+用户信息获取测试 ===" -ForegroundColor Green

# 测试数据
$testData = @{
    username = "teststudent001"
    password = "Password123"
    email = "teststudent001@example.com"
    phone = "13800138001"
    real_name = "测试学生001"
    user_type = "student"
    student_id = "20230001"
    college = "计算机学院"
    major = "软件工程"
    class = "软件2301"
    grade = "2023"
}

# 转换为JSON
$jsonBody = $testData | ConvertTo-Json

Write-Host "请求体:" -ForegroundColor Yellow
Write-Host $jsonBody -ForegroundColor Gray
Write-Host ""

# 注册
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8084/api/users/register" -Method POST -Body $jsonBody -ContentType "application/json"
    Write-Host "✓ 注册成功!" -ForegroundColor Green
    Write-Host "响应:" -ForegroundColor Yellow
    $response | ConvertTo-Json -Depth 10
}
catch {
    Write-Host "✗ 注册失败! (如已注册可忽略)" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
}

# 登录
$loginBody = @{username = $testData.username; password = $testData.password} | ConvertTo-Json
Write-Host "\n尝试登录..." -ForegroundColor Yellow
try {
    $loginResp = Invoke-RestMethod -Uri "http://localhost:8081/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "✓ 登录成功!" -ForegroundColor Green
    $token = $loginResp.data.token
    Write-Host "Token: $token" -ForegroundColor Gray
} catch {
    Write-Host "✗ 登录失败!" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 获取当前用户信息
Write-Host "\n请求当前用户信息..." -ForegroundColor Yellow
try {
    $headers = @{ Authorization = "Bearer $token" }
    $profileResp = Invoke-RestMethod -Uri "http://localhost:8084/api/users/profile" -Headers $headers -Method GET
    Write-Host "✓ 获取用户信息成功!" -ForegroundColor Green
    $profileResp | ConvertTo-Json -Depth 10
} catch {
    Write-Host "✗ 获取用户信息失败!" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "\n=== 测试完成 ===" -ForegroundColor Green

# 自动化测试：学生A获取学生B基本信息

Write-Host "=== 学生A获取学生B基本信息自动化测试 ===" -ForegroundColor Green

# 学生A数据
$studentA = @{
    username = "teststudentA"
    password = "Password123"
    email = "teststudentA@example.com"
    phone = "13800138011"
    real_name = "学生A"
    user_type = "student"
    student_id = "20230011"
    college = "计算机学院"
    major = "软件工程"
    class = "软件2301"
    grade = "2023"
}
# 学生B数据
$studentB = @{
    username = "teststudentB"
    password = "Password123"
    email = "teststudentB@example.com"
    phone = "13800138012"
    real_name = "学生B"
    user_type = "student"
    student_id = "20230012"
    college = "计算机学院"
    major = "软件工程"
    class = "软件2301"
    grade = "2023"
}

function Register-Student($student) {
    $jsonBody = $student | ConvertTo-Json
    try {
        Invoke-RestMethod -Uri "http://localhost:8084/api/users/register" -Method POST -Body $jsonBody -ContentType "application/json" | Out-Null
        Write-Host "✓ 注册 $($student.username) 成功!" -ForegroundColor Green
    } catch {
        Write-Host "✗ 注册 $($student.username) 失败! (如已注册可忽略)" -ForegroundColor Yellow
    }
}

function Login-Student($student) {
    $loginBody = @{username = $student.username; password = $student.password} | ConvertTo-Json
    try {
        $loginResp = Invoke-RestMethod -Uri "http://localhost:8081/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        Write-Host "✓ 登录 $($student.username) 成功!" -ForegroundColor Green
        return $loginResp.data.token
    } catch {
        Write-Host "✗ 登录 $($student.username) 失败!" -ForegroundColor Red
        exit 1
    }
}

# 注册两个学生
Register-Student $studentA
Register-Student $studentB

# 学生A登录
$tokenA = Login-Student $studentA

# 用学生A的token搜索学生B，获取user_id
$headersA = @{ Authorization = "Bearer $tokenA" }
$searchUrl = "http://localhost:8084/api/search/users?query=$($studentB.username)&user_type=student"
try {
    $searchResp = Invoke-RestMethod -Uri $searchUrl -Headers $headersA -Method GET
    $userB = $searchResp.data.users | Where-Object { $_.username -eq $studentB.username }
    if (-not $userB) {
        Write-Host "✗ 未找到学生B的user_id，测试失败" -ForegroundColor Red
        exit 1
    }
    $userB_id = $userB.user_id
    Write-Host "✓ 获取学生B的user_id: $userB_id" -ForegroundColor Green
} catch {
    Write-Host "✗ 搜索学生B失败!" -ForegroundColor Red
    exit 1
}

# 用学生A的token访问学生B的用户信息
$userInfoUrl = "http://localhost:8084/api/users/$userB_id"
try {
    $resp = Invoke-RestMethod -Uri $userInfoUrl -Headers $headersA -Method GET
    Write-Host "✓ 学生A成功获取学生B的基本信息!" -ForegroundColor Green
    $resp | ConvertTo-Json -Depth 10
    # 检查返回内容是否为基本信息（无email/phone等字段）
    if ($resp.data.email -or $resp.data.phone) {
        Write-Host "✗ 返回了不应有的敏感字段，权限系统有问题!" -ForegroundColor Red
    } else {
        Write-Host "✓ 权限系统正确，未返回敏感字段。" -ForegroundColor Green
    }
} catch {
    Write-Host "✗ 学生A获取学生B信息失败!" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
}

# 学生A测试更新自己的信息
Write-Host "\n=== 学生A测试更新自己的信息 ===" -ForegroundColor Green
$updateBody = @{ email = "teststudentA_new@example.com"; phone = "13800138099"; real_name = "学生A新名" } | ConvertTo-Json
try {
    $headersA = @{ Authorization = "Bearer $tokenA" }
    $resp = Invoke-RestMethod -Uri "http://localhost:8084/api/users/profile" -Headers $headersA -Method PUT -Body $updateBody -ContentType "application/json"
    Write-Host "✓ 更新成功!" -ForegroundColor Green
    $resp | ConvertTo-Json -Depth 10
    if ($resp.data.email -eq "teststudentA_new@example.com" -and $resp.data.phone -eq "13800138099" -and $resp.data.real_name -eq "学生A新名") {
        Write-Host "✓ 信息已正确更新。" -ForegroundColor Green
    } else {
        Write-Host "✗ 信息未正确更新!" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 更新失败!" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "\n=== 测试完成 ===" -ForegroundColor Green 