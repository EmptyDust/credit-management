# 统一用户服务测试脚本
# 测试合并后的用户服务功能

param(
    [string]$BaseUrl = "http://localhost:8084",
    [string]$ApiUrl = "http://localhost:8084/api",
    [string]$AuthUrl = "http://localhost:8081/api"
)

Write-Host "=== 统一用户服务测试 ===" -ForegroundColor Green
Write-Host "基础URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "API URL: $ApiUrl" -ForegroundColor Yellow
Write-Host "认证服务URL: $AuthUrl" -ForegroundColor Yellow
Write-Host ""

# 测试计数器
$totalTests = 0
$passedTests = 0
$failedTests = 0

# 测试结果记录
$testResults = @()

# 全局变量
$global:adminToken = $null
$global:studentToken = $null
$global:teacherToken = $null

# 测试函数
function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [string]$Body = "",
        [hashtable]$Headers = @{},
        [int]$ExpectedStatus = 200,
        [string]$Description = ""
    )
    
    $totalTests++
    Write-Host "测试: $Name" -ForegroundColor Cyan
    if ($Description) {
        Write-Host "  描述: $Description" -ForegroundColor Gray
    }
    
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
        $statusCode = 200  # 如果成功，状态码是200
        
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
        $statusCode = $_.Exception.Response.StatusCode.value__
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
    
    Write-Host ""
}

# 获取认证令牌
function Get-AuthToken {
    param(
        [string]$Username,
        [SecureString]$Password
    )
    
    $loginBody = @{
        username = $Username
        password = $Password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$AuthUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        if ($response.code -eq 0) {
            return $response.data.token
        }
    } catch {
        Write-Host "登录失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $null
}

# 生成严格合规的手机号、学号、用户名
function New-ValidPhone {
    $prefix = Get-Random -Minimum 130 -Maximum 199
    $suffix = Get-Random -Minimum 10000000 -Maximum 99999999
    return "$prefix$suffix"
}
function New-ValidStudentID {
    return (Get-Random -Minimum 20230000 -Maximum 20999999).ToString()
}
function New-ValidUsername {
    $base = "user" + (Get-Random -Minimum 1000 -Maximum 9999)
    $rand = -join ((48..57)+(97..122) | Get-Random -Count 4 | ForEach-Object {[char]$_})
    return "$base$rand"
}

# 1. 健康检查测试
Write-Host "1. 健康检查测试" -ForegroundColor Magenta
Test-Endpoint -Name "健康检查" -Method "GET" -Url "$BaseUrl/health"

# 2. 获取认证令牌
Write-Host "2. 获取认证令牌" -ForegroundColor Magenta
$global:adminToken = Get-AuthToken -Username "admin" -Password "password"
if ($global:adminToken) {
    Write-Host "  ✓ 管理员令牌获取成功" -ForegroundColor Green
} else {
    Write-Host "  ✗ 管理员令牌获取失败" -ForegroundColor Red
}

# 3. 用户注册测试
Write-Host "3. 用户注册测试" -ForegroundColor Magenta

# 生成随机数据避免冲突
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$randomSuffix = Get-Random -Minimum 1000 -Maximum 9999

$registerBody = @{
    username = (New-ValidUsername)
    password = "Password123"
    email = "teststudent$timestamp$randomSuffix@example.com"
    phone = (New-ValidPhone)
    real_name = "测试学生$timestamp$randomSuffix"
    user_type = "student"
    student_id = (New-ValidStudentID)
    college = "计算机学院"
    major = "软件工程"
    class = "软件2301"
    grade = "2023"
} | ConvertTo-Json

Test-Endpoint -Name "学生注册" -Method "POST" -Url "$ApiUrl/users/register" -Body $registerBody -Description "测试学生用户注册功能"

# 3.1 密码强度测试
Write-Host "3.1 密码强度测试" -ForegroundColor Magenta

# 测试弱密码（只有小写字母）
$weakPasswordBody = @{
    username = (New-ValidUsername)
    password = "password"
    email = "teststudent_weak$timestamp$randomSuffix@example.com"
    phone = (New-ValidPhone)
    real_name = "测试学生弱密码"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "弱密码测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $weakPasswordBody -ExpectedStatus 400 -Description "测试密码强度验证"

# 测试短密码
$shortPasswordBody = @{
    username = (New-ValidUsername)
    password = "Abc1"
    email = "teststudent_short$timestamp$randomSuffix@example.com"
    phone = (New-ValidPhone)
    real_name = "测试学生短密码"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "短密码测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $shortPasswordBody -ExpectedStatus 400 -Description "测试密码长度验证"

# 3.2 手机号格式测试
Write-Host "3.2 手机号格式测试" -ForegroundColor Magenta

# 测试无效手机号
$invalidPhoneBody = @{
    username = (New-ValidUsername)
    password = "Password123"
    email = "teststudent_invalid_phone$timestamp$randomSuffix@example.com"
    phone = "12345678901"
    real_name = "测试学生无效手机号"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "无效手机号测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $invalidPhoneBody -ExpectedStatus 400 -Description "测试手机号格式验证"

# 3.3 学号格式测试
Write-Host "3.3 学号格式测试" -ForegroundColor Magenta

# 测试无效学号
$invalidStudentIDBody = @{
    username = (New-ValidUsername)
    password = "Password123"
    email = "teststudent_invalid_id$timestamp$randomSuffix@example.com"
    phone = (New-ValidPhone)
    real_name = "测试学生无效学号"
    user_type = "student"
    student_id = "1234567"
} | ConvertTo-Json

Test-Endpoint -Name "无效学号测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $invalidStudentIDBody -ExpectedStatus 400 -Description "测试学号格式验证"

# 3.4 用户名格式测试
Write-Host "3.4 用户名格式测试" -ForegroundColor Magenta

# 测试特殊字符用户名
$specialCharUsernameBody = @{
    username = "test@#$%$timestamp$randomSuffix"
    password = "Password123"
    email = "teststudent_special$timestamp$randomSuffix@example.com"
    phone = (New-ValidPhone)
    real_name = "测试学生特殊字符用户名"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "特殊字符用户名测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $specialCharUsernameBody -ExpectedStatus 400 -Description "测试用户名格式验证"

# 测试短用户名
$shortUsernameBody = @{
    username = "ab$timestamp$randomSuffix"
    password = "Password123"
    email = "teststudent_shortname$timestamp$randomSuffix@example.com"
    phone = (New-ValidPhone)
    real_name = "测试学生短用户名"
    user_type = "student"
} | ConvertTo-Json

Test-Endpoint -Name "短用户名测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $shortUsernameBody -ExpectedStatus 400 -Description "测试用户名长度验证"

# 3.5 成功注册测试
Write-Host "3.5 成功注册测试" -ForegroundColor Magenta
$validRegisterBody = @{
    username = (New-ValidUsername)
    password = "Password123"
    email = "validstudent$timestamp$randomSuffix@example.com"
    phone = (New-ValidPhone)
    real_name = "有效学生$timestamp$randomSuffix"
    user_type = "student"
    student_id = (New-ValidStudentID)
    college = "信息学院"
    major = "计算机科学"
    class = "计科2301"
    grade = "2023"
} | ConvertTo-Json

Test-Endpoint -Name "有效注册测试" -Method "POST" -Url "$ApiUrl/users/register" -Body $validRegisterBody -Description "测试有效学生注册"

# 4. 用户管理测试（需要管理员权限）
Write-Host "4. 用户管理测试" -ForegroundColor Magenta

if ($global:adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $global:adminToken" }
    
    # 4.1 创建教师用户
    Write-Host "4.1 创建教师用户" -ForegroundColor Magenta
    $createTeacherBody = @{
        username = (New-ValidUsername)
        password = "Password123"
        email = "testteacher$timestamp$randomSuffix@example.com"
        phone = (New-ValidPhone)
        real_name = "测试教师$timestamp$randomSuffix"
        user_type = "teacher"
        department = "计算机系"
        title = "副教授"
        specialty = "人工智能"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "创建教师" -Method "POST" -Url "$ApiUrl/users/teachers" -Body $createTeacherBody -Headers $adminHeaders -Description "管理员创建教师用户"
    
    # 4.2 创建学生用户
    Write-Host "4.2 创建学生用户" -ForegroundColor Magenta
    $createStudentBody = @{
        username = (New-ValidUsername)
        password = "Password123"
        email = "adminstudent$timestamp$randomSuffix@example.com"
        phone = (New-ValidPhone)
        real_name = "管理员创建学生$timestamp$randomSuffix"
        user_type = "student"
        student_id = (New-ValidStudentID)
        college = "机械学院"
        major = "机械工程"
        class = "机械2301"
        grade = "2023"
    } | ConvertTo-Json
    
    Test-Endpoint -Name "管理员创建学生" -Method "POST" -Url "$ApiUrl/users/students" -Body $createStudentBody -Headers $adminHeaders -Description "管理员创建学生用户"
    
    # 4.3 获取所有用户
    Write-Host "4.3 获取所有用户" -ForegroundColor Magenta
    Test-Endpoint -Name "获取所有用户" -Method "GET" -Url "$ApiUrl/users" -Headers $adminHeaders -Description "管理员获取所有用户列表"
    
    # 4.4 根据用户类型获取用户
    Write-Host "4.4 根据用户类型获取用户" -ForegroundColor Magenta
    Test-Endpoint -Name "获取学生用户" -Method "GET" -Url "$ApiUrl/users/type/student" -Headers $adminHeaders -Description "根据用户类型获取学生用户"
    Test-Endpoint -Name "获取教师用户" -Method "GET" -Url "$ApiUrl/users/type/teacher" -Headers $adminHeaders -Description "根据用户类型获取教师用户"
    
} else {
    Write-Host "  ⚠ 跳过管理员测试 - 无管理员令牌" -ForegroundColor Yellow
}

# 5. 学生管理测试
Write-Host "5. 学生管理测试" -ForegroundColor Magenta

if ($global:adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $global:adminToken" }
    
    # 5.1 获取所有学生
    Test-Endpoint -Name "获取所有学生" -Method "GET" -Url "$ApiUrl/students" -Headers $adminHeaders -Description "获取所有学生列表"
    
    # 5.2 获取学生统计信息
    Test-Endpoint -Name "获取学生统计" -Method "GET" -Url "$ApiUrl/students/stats" -Headers $adminHeaders -Description "获取学生统计信息"
    
    # 5.3 带查询参数的学生查询
    Test-Endpoint -Name "分页查询学生" -Method "GET" -Url "$ApiUrl/students?page=1&page_size=5" -Headers $adminHeaders -Description "分页查询学生"
    
    Test-Endpoint -Name "按学院查询学生" -Method "GET" -Url "$ApiUrl/students?college=计算机学院" -Headers $adminHeaders -Description "按学院筛选学生"
    
} else {
    Write-Host "  ⚠ 跳过学生管理测试 - 无管理员令牌" -ForegroundColor Yellow
}

# 6. 教师管理测试
Write-Host "6. 教师管理测试" -ForegroundColor Magenta

if ($global:adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $global:adminToken" }
    
    # 6.1 获取所有教师
    Test-Endpoint -Name "获取所有教师" -Method "GET" -Url "$ApiUrl/teachers" -Headers $adminHeaders -Description "获取所有教师列表"
    
    # 6.2 获取教师统计信息
    Test-Endpoint -Name "获取教师统计" -Method "GET" -Url "$ApiUrl/teachers/stats" -Headers $adminHeaders -Description "获取教师统计信息"
    
    # 6.3 带查询参数的教师查询
    Test-Endpoint -Name "分页查询教师" -Method "GET" -Url "$ApiUrl/teachers?page=1&page_size=5" -Headers $adminHeaders -Description "分页查询教师"
    
    Test-Endpoint -Name "按部门查询教师" -Method "GET" -Url "$ApiUrl/teachers?department=计算机系" -Headers $adminHeaders -Description "按部门筛选教师"
    
} else {
    Write-Host "  ⚠ 跳过教师管理测试 - 无管理员令牌" -ForegroundColor Yellow
}

# 7. 用户搜索测试
Write-Host "7. 用户搜索测试" -ForegroundColor Magenta

if ($global:adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $global:adminToken" }
    
    # 7.1 通用用户搜索
    Test-Endpoint -Name "通用用户搜索" -Method "GET" -Url "$ApiUrl/search/users?query=测试" -Headers $adminHeaders -Description "通用用户搜索功能"
    
    # 7.2 按用户类型搜索
    Test-Endpoint -Name "按类型搜索用户" -Method "GET" -Url "$ApiUrl/search/users?user_type=student" -Headers $adminHeaders -Description "按用户类型搜索"
    
    # 7.3 按学院搜索
    Test-Endpoint -Name "按学院搜索用户" -Method "GET" -Url "$ApiUrl/search/users?college=计算机学院" -Headers $adminHeaders -Description "按学院搜索用户"
    
    # 7.4 分页搜索
    Test-Endpoint -Name "分页搜索用户" -Method "GET" -Url "$ApiUrl/search/users?query=测试&page=1&page_size=5" -Headers $adminHeaders -Description "分页搜索用户"
    
} else {
    Write-Host "  ⚠ 跳过用户搜索测试 - 无管理员令牌" -ForegroundColor Yellow
}

# 8. 用户统计测试
Write-Host "8. 用户统计测试" -ForegroundColor Magenta

if ($global:adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $global:adminToken" }
    
    # 8.1 获取用户统计信息
    Test-Endpoint -Name "获取用户统计" -Method "GET" -Url "$ApiUrl/users/stats" -Headers $adminHeaders -Description "获取用户统计信息"
    
} else {
    Write-Host "  ⚠ 跳过用户统计测试 - 无管理员令牌" -ForegroundColor Yellow
}

# 9. 权限测试
Write-Host "9. 权限测试" -ForegroundColor Magenta

# 9.1 无令牌访问测试
Write-Host "9.1 无令牌访问测试" -ForegroundColor Magenta
Test-Endpoint -Name "无令牌访问用户列表" -Method "GET" -Url "$ApiUrl/users" -ExpectedStatus 401 -Description "测试无令牌访问被拒绝"

Test-Endpoint -Name "无令牌访问学生列表" -Method "GET" -Url "$ApiUrl/students" -ExpectedStatus 401 -Description "测试无令牌访问被拒绝"

# 9.2 无效令牌访问测试
Write-Host "9.2 无效令牌访问测试" -ForegroundColor Magenta
$invalidHeaders = @{ "Authorization" = "Bearer invalid_token" }
Test-Endpoint -Name "无效令牌访问用户列表" -Method "GET" -Url "$ApiUrl/users" -Headers $invalidHeaders -ExpectedStatus 401 -Description "测试无效令牌访问被拒绝"

# 10. 输出测试结果统计
Write-Host "`n=== 测试结果统计 ===" -ForegroundColor Yellow
Write-Host "总测试数: $totalTests" -ForegroundColor White
Write-Host "通过: $passedTests" -ForegroundColor Green
Write-Host "失败: $failedTests" -ForegroundColor Red

if ($failedTests -eq 0) {
    Write-Host "`n🎉 所有测试通过！" -ForegroundColor Green
} else {
    Write-Host "`n❌ 有 $failedTests 个测试失败" -ForegroundColor Red
    Write-Host "`n失败的测试详情:" -ForegroundColor Red
    $testResults | Where-Object { $_.Status -ne "PASS" } | ForEach-Object {
        Write-Host "  - $($_.Name): $($_.Status)" -ForegroundColor Red
        if ($_.Error) {
            Write-Host "    错误: $($_.Error)" -ForegroundColor Red
        }
    }
} 