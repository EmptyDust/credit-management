# 约束条件一致性测试脚本
# 验证数据库、后端API、前端验证的一致性

$baseUrl = "http://localhost:8080"

Write-Host "=== 约束条件一致性测试 ===" -ForegroundColor Green

# 1. 测试用户名约束
Write-Host "`n1. 测试用户名约束..." -ForegroundColor Yellow

# 1.1 测试用户名长度限制
$testCases = @(
    @{username = "ab"; expected = "用户名至少3个字符" },
    @{username = "a" * 21; expected = "用户名最多20个字符" },
    @{username = "test@user"; expected = "用户名只能包含字母、数字和下划线" },
    @{username = "test-user"; expected = "用户名只能包含字母、数字和下划线" },
    @{username = "test user"; expected = "用户名只能包含字母、数字和下划线" }
)

foreach ($testCase in $testCases) {
    $testData = @{
        username   = $testCase.username
        password   = "TestPass123"
        email      = "test@example.com"
        phone      = "13800138000"
        real_name  = "测试用户"
        student_id = "12345678"
        college    = "计算机学院"
        major      = "软件工程"
        class      = "软件2021-1班"
        grade      = "2021"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($testData | ConvertTo-Json) -ContentType "application/json"
        Write-Host "❌ 用户名 '$($testCase.username)' 应该失败但通过了验证" -ForegroundColor Red
    }
    catch {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $responseBody = $reader.ReadToEnd()
        $errorData = $responseBody | ConvertFrom-Json
        
        if ($errorData.message -like "*$($testCase.expected)*") {
            Write-Host "✅ 用户名 '$($testCase.username)' 正确被拒绝: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ 用户名 '$($testCase.username)' 错误信息不匹配: $($errorData.message)" -ForegroundColor Red
        }
    }
}

# 2. 测试密码约束
Write-Host "`n2. 测试密码约束..." -ForegroundColor Yellow

$passwordTestCases = @(
    @{password = "short"; expected = "密码至少8个字符" },
    @{password = "nouppercase123"; expected = "密码必须包含大小写字母和数字" },
    @{password = "NOLOWERCASE123"; expected = "密码必须包含大小写字母和数字" },
    @{password = "NoNumbers"; expected = "密码必须包含大小写字母和数字" }
)

foreach ($testCase in $passwordTestCases) {
    $testData = @{
        username   = "testuser$(Get-Random)"
        password   = $testCase.password
        email      = "test$(Get-Random)@example.com"
        phone      = "13800138000"
        real_name  = "测试用户"
        student_id = "12345678"
        college    = "计算机学院"
        major      = "软件工程"
        class      = "软件2021-1班"
        grade      = "2021"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($testData | ConvertTo-Json) -ContentType "application/json"
        Write-Host "❌ 密码 '$($testCase.password)' 应该失败但通过了验证" -ForegroundColor Red
    }
    catch {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $responseBody = $reader.ReadToEnd()
        $errorData = $responseBody | ConvertFrom-Json
        
        if ($errorData.message -like "*$($testCase.expected)*") {
            Write-Host "✅ 密码 '$($testCase.password)' 正确被拒绝: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ 密码 '$($testCase.password)' 错误信息不匹配: $($errorData.message)" -ForegroundColor Red
        }
    }
}

# 3. 测试手机号约束
Write-Host "`n3. 测试手机号约束..." -ForegroundColor Yellow

$phoneTestCases = @(
    @{phone = "12345678901"; expected = "手机号格式不正确" },
    @{phone = "1380013800"; expected = "手机号格式不正确" },
    @{phone = "138001380000"; expected = "手机号格式不正确" },
    @{phone = "12800138000"; expected = "手机号格式不正确" },
    @{phone = "12000138000"; expected = "手机号格式不正确" }
)

foreach ($testCase in $phoneTestCases) {
    $testData = @{
        username   = "testuser$(Get-Random)"
        password   = "TestPass123"
        email      = "test$(Get-Random)@example.com"
        phone      = $testCase.phone
        real_name  = "测试用户"
        student_id = "12345678"
        college    = "计算机学院"
        major      = "软件工程"
        class      = "软件2021-1班"
        grade      = "2021"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($testData | ConvertTo-Json) -ContentType "application/json"
        Write-Host "❌ 手机号 '$($testCase.phone)' 应该失败但通过了验证" -ForegroundColor Red
    }
    catch {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $responseBody = $reader.ReadToEnd()
        $errorData = $responseBody | ConvertFrom-Json
        
        if ($errorData.message -like "*$($testCase.expected)*") {
            Write-Host "✅ 手机号 '$($testCase.phone)' 正确被拒绝: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ 手机号 '$($testCase.phone)' 错误信息不匹配: $($errorData.message)" -ForegroundColor Red
        }
    }
}

# 4. 测试学号约束
Write-Host "`n4. 测试学号约束..." -ForegroundColor Yellow

$studentIDTestCases = @(
    @{student_id = "1234567"; expected = "学号格式不正确" },
    @{student_id = "123456789"; expected = "学号格式不正确" },
    @{student_id = "1234567a"; expected = "学号格式不正确" },
    @{student_id = "abcdefgh"; expected = "学号格式不正确" }
)

foreach ($testCase in $studentIDTestCases) {
    $testData = @{
        username   = "testuser$(Get-Random)"
        password   = "TestPass123"
        email      = "test$(Get-Random)@example.com"
        phone      = "13800138000"
        real_name  = "测试用户"
        student_id = $testCase.student_id
        college    = "计算机学院"
        major      = "软件工程"
        class      = "软件2021-1班"
        grade      = "2021"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($testData | ConvertTo-Json) -ContentType "application/json"
        Write-Host "❌ 学号 '$($testCase.student_id)' 应该失败但通过了验证" -ForegroundColor Red
    }
    catch {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $responseBody = $reader.ReadToEnd()
        $errorData = $responseBody | ConvertFrom-Json
        
        if ($errorData.message -like "*$($testCase.expected)*") {
            Write-Host "✅ 学号 '$($testCase.student_id)' 正确被拒绝: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ 学号 '$($testCase.student_id)' 错误信息不匹配: $($errorData.message)" -ForegroundColor Red
        }
    }
}

# 5. 测试年级约束
Write-Host "`n5. 测试年级约束..." -ForegroundColor Yellow

$gradeTestCases = @(
    @{grade = "202"; expected = "年级格式不正确" },
    @{grade = "20212"; expected = "年级格式不正确" },
    @{grade = "202a"; expected = "年级格式不正确" },
    @{grade = "abcd"; expected = "年级格式不正确" }
)

foreach ($testCase in $gradeTestCases) {
    $testData = @{
        username   = "testuser$(Get-Random)"
        password   = "TestPass123"
        email      = "test$(Get-Random)@example.com"
        phone      = "13800138000"
        real_name  = "测试用户"
        student_id = "12345678"
        college    = "计算机学院"
        major      = "软件工程"
        class      = "软件2021-1班"
        grade      = $testCase.grade
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($testData | ConvertTo-Json) -ContentType "application/json"
        Write-Host "❌ 年级 '$($testCase.grade)' 应该失败但通过了验证" -ForegroundColor Red
    }
    catch {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $responseBody = $reader.ReadToEnd()
        $errorData = $responseBody | ConvertFrom-Json
        
        if ($errorData.message -like "*$($testCase.expected)*") {
            Write-Host "✅ 年级 '$($testCase.grade)' 正确被拒绝: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ 年级 '$($testCase.grade)' 错误信息不匹配: $($errorData.message)" -ForegroundColor Red
        }
    }
}

# 6. 测试成功注册
Write-Host "`n6. 测试成功注册..." -ForegroundColor Yellow

$validData = @{
    username   = "testuser$(Get-Random)"
    password   = "TestPass123"
    email      = "test$(Get-Random)@example.com"
    phone      = "13800138000"
    real_name  = "测试用户"
    student_id = "12345678"
    college    = "计算机学院"
    major      = "软件工程"
    class      = "软件2021-1班"
    grade      = "2021"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($validData | ConvertTo-Json) -ContentType "application/json"
    if ($response.code -eq 0) {
        Write-Host "✅ 有效数据注册成功" -ForegroundColor Green
    }
    else {
        Write-Host "❌ 有效数据注册失败: $($response.message)" -ForegroundColor Red
    }
}
catch {
    $errorResponse = $_.Exception.Response.GetResponseStream()
    $reader = New-Object System.IO.StreamReader($errorResponse)
    $responseBody = $reader.ReadToEnd()
    $errorData = $responseBody | ConvertFrom-Json
    Write-Host "❌ 有效数据注册失败: $($errorData.message)" -ForegroundColor Red
}

Write-Host "`n=== 约束条件一致性测试完成 ===" -ForegroundColor Green 