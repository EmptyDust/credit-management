#!/usr/bin/env pwsh

# 前端注册功能测试脚本

Write-Host "=== 前端注册功能测试 ===" -ForegroundColor Magenta

# 生成随机数据
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$randomSuffix = Get-Random -Minimum 1000 -Maximum 9999
$prefix = Get-Random -Minimum 130 -Maximum 199
$suffix = Get-Random -Minimum 10000000 -Maximum 99999999
$phone = "$prefix$suffix"
$studentId = (Get-Random -Minimum 20230000 -Maximum 20999999).ToString()
$base = "user" + (Get-Random -Minimum 1000 -Maximum 9999)
$rand = -join ((48..57) + (97..122) | Get-Random -Count 4 | ForEach-Object { [char]$_ })
$username = "$base$rand"

$registerData = @{
    username   = $username
    password   = "Password123"
    email      = "teststudent$timestamp$randomSuffix@example.com"
    phone      = $phone
    real_name  = "测试学生$timestamp$randomSuffix"
    user_type  = "student"
    student_id = $studentId
    college    = "计算机学院"
    major      = "软件工程"
    class      = "软件2301"
    grade      = "2023"
} | ConvertTo-Json

Write-Host "测试数据:" -ForegroundColor Cyan
Write-Host $registerData -ForegroundColor Gray
Write-Host ""

# 测试注册接口
Write-Host "测试注册接口..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/users/register" -Method POST -Body $registerData -ContentType "application/json"
    Write-Host "✓ 注册成功!" -ForegroundColor Green
    Write-Host "响应: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Gray
}
catch {
    Write-Host "✗ 注册失败: $($_.Exception.Message)" -ForegroundColor Red
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
Write-Host "=== 测试完成 ===" -ForegroundColor Magenta 