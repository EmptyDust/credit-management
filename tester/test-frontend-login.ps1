#!/usr/bin/env pwsh

# 前端登录功能测试脚本

Write-Host "=== 前端登录功能测试 ===" -ForegroundColor Magenta

# 测试数据
$loginData = @{
    username = "admin"
    password = "password"
} | ConvertTo-Json

Write-Host "测试数据:" -ForegroundColor Cyan
Write-Host $loginData -ForegroundColor Gray
Write-Host ""

# 测试登录接口
Write-Host "测试管理员登录..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    Write-Host "✓ 管理员登录成功!" -ForegroundColor Green
    Write-Host "响应状态: $($response.code)" -ForegroundColor Gray
    Write-Host "用户信息: $($response.data.user.username) - $($response.data.user.user_type)" -ForegroundColor Gray
    Write-Host "Token长度: $($response.data.token.Length)" -ForegroundColor Gray
}
catch {
    Write-Host "✗ 管理员登录失败: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "状态码: $statusCode" -ForegroundColor Red
        try {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "错误响应: $errorBody" -ForegroundColor Red
        }
        catch {
            Write-Host "无法读取错误响应" -ForegroundColor Red
        }
    }
}

Write-Host ""

# 测试学生登录
$studentLoginData = @{
    username = "testuser777"
    password = "Password123"
} | ConvertTo-Json

Write-Host "测试学生登录..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body $studentLoginData -ContentType "application/json"
    Write-Host "✓ 学生登录成功!" -ForegroundColor Green
    Write-Host "响应状态: $($response.code)" -ForegroundColor Gray
    Write-Host "用户信息: $($response.data.user.username) - $($response.data.user.user_type)" -ForegroundColor Gray
    Write-Host "Token长度: $($response.data.token.Length)" -ForegroundColor Gray
}
catch {
    Write-Host "✗ 学生登录失败: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "状态码: $statusCode" -ForegroundColor Red
        try {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "错误响应: $errorBody" -ForegroundColor Red
        }
        catch {
            Write-Host "无法读取错误响应" -ForegroundColor Red
        }
    }
}

Write-Host ""

# 测试错误密码
$wrongPasswordData = @{
    username = "admin"
    password = "wrongpassword"
} | ConvertTo-Json

Write-Host "测试错误密码..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body $wrongPasswordData -ContentType "application/json"
    Write-Host "✗ 错误密码测试失败 - 应该返回401错误" -ForegroundColor Red
}
catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 401) {
            Write-Host "✓ 错误密码测试通过 - 正确返回401错误" -ForegroundColor Green
        }
        else {
            Write-Host "✗ 错误密码测试失败 - 期望401，实际$statusCode" -ForegroundColor Red
        }
    }
    else {
        Write-Host "✗ 错误密码测试失败 - 无响应" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== 测试完成 ===" -ForegroundColor Magenta 