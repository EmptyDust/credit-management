# 简单的约束条件一致性测试脚本

$baseUrl = "http://localhost:8080/api"

Write-Host "=== Constraint Consistency Test ===" -ForegroundColor Green

# 1. 测试有效注册
Write-Host "`n1. Testing valid registration..." -ForegroundColor Yellow

$validData = @{
    username   = "testuser$(Get-Random)"
    password   = "TestPass123"
    email      = "test$(Get-Random)@example.com"
    phone      = "13800138000"
    real_name  = "Test User"
    student_id = "12345678"
    college    = "Computer Science"
    major      = "Software Engineering"
    class      = "SE2021-1"
    grade      = "2021"
    user_type  = "student"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($validData | ConvertTo-Json) -ContentType "application/json"
    if ($response.code -eq 0) {
        Write-Host "✅ Valid data registration successful" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Valid data registration failed: $($response.message)" -ForegroundColor Red
    }
}
catch {
    $errorData = $null
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
        try {
            $errorData = $_.ErrorDetails.Message | ConvertFrom-Json
        }
        catch {}
    }
    elseif ($_.Exception.Response -and ($_.Exception.Response -is [System.Net.HttpWebResponse] -or $_.Exception.Response -is [System.Net.WebResponse])) {
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $responseBody = $reader.ReadToEnd()
            $errorData = $responseBody | ConvertFrom-Json
        }
        catch {}
    }
    if ($errorData -and $errorData.message) {
        Write-Host "❌ Valid data registration failed: $($errorData.message)" -ForegroundColor Red
    }
    else {
        Write-Host "❌ Valid data registration failed: $($_ | Out-String)" -ForegroundColor Red
    }
}

# 2. 测试用户名约束
Write-Host "`n2. Testing username constraints..." -ForegroundColor Yellow

$invalidUsernameData = @{
    username   = "ab"  # 太短
    password   = "TestPass123"
    email      = "test2@example.com"
    phone      = "13800138001"
    real_name  = "Test User"
    student_id = "12345679"
    college    = "Computer Science"
    major      = "Software Engineering"
    class      = "SE2021-1"
    grade      = "2021"
    user_type  = "student"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($invalidUsernameData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "❌ Short username should fail but passed validation" -ForegroundColor Red
}
catch {
    $errorData = $null
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
        try {
            $errorData = $_.ErrorDetails.Message | ConvertFrom-Json
        }
        catch {}
    }
    elseif ($_.Exception.Response -and ($_.Exception.Response -is [System.Net.HttpWebResponse] -or $_.Exception.Response -is [System.Net.WebResponse])) {
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $responseBody = $reader.ReadToEnd()
            $errorData = $responseBody | ConvertFrom-Json
        }
        catch {}
    }
    if ($errorData -and $errorData.message) {
        if ($errorData.message -like "*username*" -or $errorData.message -like "*用户名*") {
            Write-Host "✅ Short username correctly rejected: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ Short username error message mismatch: $($errorData.message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "❌ Short username error: $($_ | Out-String)" -ForegroundColor Red
    }
}

# 3. 测试密码约束
Write-Host "`n3. Testing password constraints..." -ForegroundColor Yellow

$invalidPasswordData = @{
    username   = "testuser$(Get-Random)"
    password   = "short"  # 太短且不符合复杂度要求
    email      = "test3@example.com"
    phone      = "13800138002"
    real_name  = "Test User"
    student_id = "12345680"
    college    = "Computer Science"
    major      = "Software Engineering"
    class      = "SE2021-1"
    grade      = "2021"
    user_type  = "student"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($invalidPasswordData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "❌ Weak password should fail but passed validation" -ForegroundColor Red
}
catch {
    $errorData = $null
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
        try {
            $errorData = $_.ErrorDetails.Message | ConvertFrom-Json
        }
        catch {}
    }
    elseif ($_.Exception.Response -and ($_.Exception.Response -is [System.Net.HttpWebResponse] -or $_.Exception.Response -is [System.Net.WebResponse])) {
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $responseBody = $reader.ReadToEnd()
            $errorData = $responseBody | ConvertFrom-Json
        }
        catch {}
    }
    if ($errorData -and $errorData.message) {
        if ($errorData.message -like "*password*" -or $errorData.message -like "*密码*") {
            Write-Host "✅ Weak password correctly rejected: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ Weak password error message mismatch: $($errorData.message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "❌ Weak password error: $($_ | Out-String)" -ForegroundColor Red
    }
}

# 4. 测试手机号约束
Write-Host "`n4. Testing phone number constraints..." -ForegroundColor Yellow

$invalidPhoneData = @{
    username   = "testuser$(Get-Random)"
    password   = "TestPass123"
    email      = "test4@example.com"
    phone      = "12345678901"  # 无效格式
    real_name  = "Test User"
    student_id = "12345681"
    college    = "Computer Science"
    major      = "Software Engineering"
    class      = "SE2021-1"
    grade      = "2021"
    user_type  = "student"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($invalidPhoneData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "❌ Invalid phone number should fail but passed validation" -ForegroundColor Red
}
catch {
    $errorData = $null
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
        try {
            $errorData = $_.ErrorDetails.Message | ConvertFrom-Json
        }
        catch {}
    }
    elseif ($_.Exception.Response -and ($_.Exception.Response -is [System.Net.HttpWebResponse] -or $_.Exception.Response -is [System.Net.WebResponse])) {
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $responseBody = $reader.ReadToEnd()
            $errorData = $responseBody | ConvertFrom-Json
        }
        catch {}
    }
    if ($errorData -and $errorData.message) {
        if ($errorData.message -like "*phone*" -or $errorData.message -like "*手机号*") {
            Write-Host "✅ Invalid phone number correctly rejected: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ Invalid phone number error message mismatch: $($errorData.message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "❌ Invalid phone number error: $($_ | Out-String)" -ForegroundColor Red
    }
}

# 5. 测试学号约束
Write-Host "`n5. Testing student ID constraints..." -ForegroundColor Yellow

$invalidStudentIDData = @{
    username   = "testuser$(Get-Random)"
    password   = "TestPass123"
    email      = "test5@example.com"
    phone      = "13800138003"
    real_name  = "Test User"
    student_id = "1234567"  # 太短
    college    = "Computer Science"
    major      = "Software Engineering"
    class      = "SE2021-1"
    grade      = "2021"
    user_type  = "student"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($invalidStudentIDData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "❌ Invalid student ID should fail but passed validation" -ForegroundColor Red
}
catch {
    $errorData = $null
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
        try {
            $errorData = $_.ErrorDetails.Message | ConvertFrom-Json
        }
        catch {}
    }
    elseif ($_.Exception.Response -and ($_.Exception.Response -is [System.Net.HttpWebResponse] -or $_.Exception.Response -is [System.Net.WebResponse])) {
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $responseBody = $reader.ReadToEnd()
            $errorData = $responseBody | ConvertFrom-Json
        }
        catch {}
    }
    if ($errorData -and $errorData.message) {
        if ($errorData.message -like "*student*" -or $errorData.message -like "*学号*") {
            Write-Host "✅ Invalid student ID correctly rejected: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ Invalid student ID error message mismatch: $($errorData.message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "❌ Invalid student ID error: $($_ | Out-String)" -ForegroundColor Red
    }
}

# 6. 测试年级约束
Write-Host "`n6. Testing grade constraints..." -ForegroundColor Yellow

$invalidGradeData = @{
    username   = "testuser$(Get-Random)"
    password   = "TestPass123"
    email      = "test6@example.com"
    phone      = "13800138004"
    real_name  = "Test User"
    student_id = "12345682"
    college    = "Computer Science"
    major      = "Software Engineering"
    class      = "SE2021-1"
    grade      = "202"  # 太短
    user_type  = "student"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method POST -Body ($invalidGradeData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "❌ Invalid grade should fail but passed validation" -ForegroundColor Red
}
catch {
    $errorData = $null
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
        try {
            $errorData = $_.ErrorDetails.Message | ConvertFrom-Json
        }
        catch {}
    }
    elseif ($_.Exception.Response -and ($_.Exception.Response -is [System.Net.HttpWebResponse] -or $_.Exception.Response -is [System.Net.WebResponse])) {
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $responseBody = $reader.ReadToEnd()
            $errorData = $responseBody | ConvertFrom-Json
        }
        catch {}
    }
    if ($errorData -and $errorData.message) {
        if ($errorData.message -like "*grade*" -or $errorData.message -like "*年级*") {
            Write-Host "✅ Invalid grade correctly rejected: $($errorData.message)" -ForegroundColor Green
        }
        else {
            Write-Host "❌ Invalid grade error message mismatch: $($errorData.message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "❌ Invalid grade error: $($_ | Out-String)" -ForegroundColor Red
    }
}

Write-Host "`n=== Constraint Consistency Test Completed ===" -ForegroundColor Green 