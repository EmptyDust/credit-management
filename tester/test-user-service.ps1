# 统一用户服务测试脚本
# 测试合并后的用户服务功能

param(
    [string]$BaseUrl = "http://localhost:8084",
    [string]$ApiUrl = "http://localhost:8084/api"
)

Write-Host "=== 统一用户服务测试 ===" -ForegroundColor Green
Write-Host "基础URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "API URL: $ApiUrl" -ForegroundColor Yellow
Write-Host ""

# 测试计数器
$totalTests = 0
$passedTests = 0
$failedTests = 0

# 测试结果记录
$testResults = @()

# 测试函数
function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [string]$Body = "",
        [hashtable]$Headers = @{},
        [int]$ExpectedStatus = 200
    )
    
    $totalTests++
    Write-Host "测试: $Name" -ForegroundColor Cyan
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            Headers = $Headers
            ContentType = "application/json"
        }
        
        if ($Body -and $Body -ne "") {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params -ErrorAction Stop
        $statusCode = $response.StatusCode
        
        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "  ✓ 通过 (状态码: $statusCode)" -ForegroundColor Green
            $passedTests++
            $testResults += @{Name = $Name; Status = "PASS"; StatusCode = $statusCode}
        } else {
            Write-Host "  ✗ 失败 (期望: $ExpectedStatus, 实际: $statusCode)" -ForegroundColor Red
            $failedTests++
            $testResults += @{Name = $Name; Status = "FAIL"; StatusCode = $statusCode; Expected = $ExpectedStatus}
        }
    }
    catch {
        Write-Host "  ✗ 错误: $($_.Exception.Message)" -ForegroundColor Red
        $failedTests++
        $testResults += @{Name = $Name; Status = "ERROR"; Error = $_.Exception.Message}
    }
    
    Write-Host ""
}

# 1. 健康检查测试
Write-Host "1. 健康检查测试" -ForegroundColor Magenta
Test-Endpoint -Name "健康检查" -Method "GET" -Url "$BaseUrl/health"

# 2. 用户注册测试
Write-Host "2. 用户注册测试" -ForegroundColor Magenta
$registerBody = @{
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
} | ConvertTo-Json

Test-Endpoint -Name "学生注册" -Method "POST" -Url "$ApiUrl/users/register" -Body $registerBody

# 2.1 密码强度测试
Write-Host "2.1 密码强度测试" -ForegroundColor Magenta

# 测试弱密码（只有小写字母）
$weakPasswordBody = @{
    username = "teststudent002"
    password = "password"
    email = "teststudent002@example.com"
    phone = "13800138002"
    real_name = "测试学生002"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "弱密码测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $weakPasswordBody -ExpectedStatus 400

# 测试短密码
$shortPasswordBody = @{
    username = "teststudent003"
    password = "Abc1"
    email = "teststudent003@example.com"
    phone = "13800138003"
    real_name = "测试学生003"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "短密码测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $shortPasswordBody -ExpectedStatus 400

# 2.2 手机号格式测试
Write-Host "2.2 手机号格式测试" -ForegroundColor Magenta

# 测试无效手机号
$invalidPhoneBody = @{
    username = "teststudent004"
    password = "Password123"
    email = "teststudent004@example.com"
    phone = "12345678901"
    real_name = "测试学生004"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "无效手机号测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $invalidPhoneBody -ExpectedStatus 400

# 2.3 学号格式测试
Write-Host "2.3 学号格式测试" -ForegroundColor Magenta

# 测试无效学号
$invalidStudentIDBody = @{
    username = "teststudent005"
    password = "Password123"
    email = "teststudent005@example.com"
    phone = "13800138005"
    real_name = "测试学生005"
    user_type = "student"
    student_id = "1234567"
} | ConvertTo-Json

Test-Endpoint -Name "无效学号测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $invalidStudentIDBody -ExpectedStatus 400

# 2.4 用户名格式测试
Write-Host "2.4 用户名格式测试" -ForegroundColor Magenta

# 测试特殊字符用户名
$specialCharUsernameBody = @{
    username = "test@#$%"
    password = "Password123"
    email = "teststudent006@example.com"
    phone = "13800138006"
    real_name = "测试学生006"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "特殊字符用户名测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $specialCharUsernameBody -ExpectedStatus 400

# 测试短用户名
$shortUsernameBody = @{
    username = "ab"
    password = "Password123"
    email = "teststudent007@example.com"
    phone = "13800138007"
    real_name = "测试学生007"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "短用户名测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $shortUsernameBody -ExpectedStatus 400

# 2.5 成功注册测试
Write-Host "2.5 成功注册测试" -ForegroundColor Magenta
$validRegisterBody = @{
    username = "validstudent001"
    password = "Password123"
    email = "validstudent001@example.com"
    phone = "13800138008"
    real_name = "有效学生001"
    user_type = "student"
    student_id = "20230002"
    college = "信息学院"
    major = "计算机科学"
    class = "计科2301"
    grade = "2023"
} | ConvertTo-Json

Test-Endpoint -Name "有效注册测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $validRegisterBody

# 3. 用户登录测试（需要认证服务）
Write-Host "3. 用户登录测试" -ForegroundColor Magenta
Write-Host "  注意: 需要认证服务运行在8081端口" -ForegroundColor Yellow

$loginBody = @{
    username = "validstudent001"
    password = "Password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/auth/login" -Method "POST" -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-Host "  ✓ 登录成功，获取到token" -ForegroundColor Green
    $passedTests++
} catch {
    Write-Host "  ✗ 登录失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  跳过需要认证的测试..." -ForegroundColor Yellow
    $token = ""
    $failedTests++
}

# 4. 需要认证的API测试
if ($token) {
    Write-Host "4. 需要认证的API测试" -ForegroundColor Magenta
    
    $authHeaders = @{
        "Authorization" = "Bearer $token"
    }
    
    # 获取当前用户信息
    Test-Endpoint -Name "获取当前用户信息" -Method "GET" -Url "$ApiUrl/users/profile" -Headers $authHeaders
    
    # 获取用户统计信息
    Test-Endpoint -Name "获取用户统计信息" -Method "GET" -Url "$ApiUrl/users/stats" -Headers $authHeaders
    
    # 获取学生列表
    Test-Endpoint -Name "获取学生列表" -Method "GET" -Url "$ApiUrl/students" -Headers $authHeaders
    
    # 获取教师列表
    Test-Endpoint -Name "获取教师列表" -Method "GET" -Url "$ApiUrl/teachers" -Headers $authHeaders
    
    # 搜索用户
    Test-Endpoint -Name "搜索用户" -Method "GET" -Url "$ApiUrl/search/users?query=test" -Headers $authHeaders
}

# 5. 管理员功能测试（需要管理员token）
Write-Host "5. 管理员功能测试" -ForegroundColor Magenta
Write-Host "  注意: 需要管理员权限" -ForegroundColor Yellow

# 尝试使用管理员账号登录
$adminLoginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $adminLoginResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/auth/login" -Method "POST" -Body $adminLoginBody -ContentType "application/json"
    $adminToken = $adminLoginResponse.data.token
    Write-Host "  ✓ 管理员登录成功" -ForegroundColor Green
    $passedTests++
    
    $adminHeaders = @{
        "Authorization" = "Bearer $adminToken"
    }
    
    # 创建教师
    $createTeacherBody = @{
        username = "testteacher001"
        password = "Password123"
        email = "testteacher001@example.com"
        phone = "13800138002"
        real_name = "测试教师001"
        user_type = "teacher"
        department = "计算机系"
        title = "副教授"
        specialty = "人工智能"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "创建教师" -Method "POST" -Url "$ApiUrl/users/teachers" -Body $createTeacherBody -Headers $adminHeaders
    
    # 创建学生
    $createStudentBody = @{
        username = "teststudent002"
        password = "Password123"
        email = "teststudent002@example.com"
        phone = "13800138003"
        real_name = "测试学生002"
        user_type = "student"
        student_id = "20230003"
        college = "信息学院"
        major = "计算机科学"
        class = "计科2301"
        grade = "2023"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "创建学生" -Method "POST" -Url "$ApiUrl/users/students" -Body $createStudentBody -Headers $adminHeaders
    
    # 获取所有用户
    Test-Endpoint -Name "获取所有用户" -Method "GET" -Url "$ApiUrl/users" -Headers $adminHeaders
    
    # 获取学生统计信息
    Test-Endpoint -Name "获取学生统计信息" -Method "GET" -Url "$ApiUrl/students/stats" -Headers $adminHeaders
    
    # 获取教师统计信息
    Test-Endpoint -Name "获取教师统计信息" -Method "GET" -Url "$ApiUrl/teachers/stats" -Headers $adminHeaders
    
} catch {
    Write-Host "  ✗ 管理员登录失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  跳过管理员功能测试..." -ForegroundColor Yellow
    $failedTests++
}

# 6. 错误处理测试
Write-Host "6. 错误处理测试" -ForegroundColor Magenta

# 测试重复注册
Test-Endpoint -Name "重复注册测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $registerBody -ExpectedStatus 409

# 测试无效的注册数据
$invalidBody = @{
    username = ""
    password = ""
    email = "invalid-email"
} | ConvertTo-Json

Test-Endpoint -Name "无效注册数据测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $invalidBody -ExpectedStatus 400

# 测试不存在的端点
Test-Endpoint -Name "不存在的端点测试" -Method "GET" -Url "$ApiUrl/nonexistent" -ExpectedStatus 404

# 测试结果汇总
Write-Host "=== 测试结果汇总 ===" -ForegroundColor Green
Write-Host "总测试数: $totalTests" -ForegroundColor White
Write-Host "通过: $passedTests" -ForegroundColor Green
Write-Host "失败: $failedTests" -ForegroundColor Red
Write-Host "成功率: $([math]::Round($passedTests / $totalTests * 100, 2))%" -ForegroundColor Yellow

if ($failedTests -gt 0) {
    Write-Host ""
    Write-Host "失败的测试:" -ForegroundColor Red
    $testResults | Where-Object { $_.Status -ne "PASS" } | ForEach-Object {
        Write-Host "  - $($_.Name): $($_.Status)" -ForegroundColor Red
        if ($_.Error) {
            Write-Host "    错误: $($_.Error)" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "测试完成!" -ForegroundColor Green 