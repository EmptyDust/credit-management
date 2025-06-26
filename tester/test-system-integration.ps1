#!/usr/bin/env pwsh

# 双创分申请平台综合测试脚本
# 测试所有服务和前端功能

Write-Host "=== 双创分申请平台综合测试 ===" -ForegroundColor Green
Write-Host "开始时间: $(Get-Date)" -ForegroundColor Yellow

# 配置
$API_BASE = "http://localhost:8080"
$FRONTEND_URL = "http://localhost:3000"
$ADMIN_USERNAME = "admin"
$ADMIN_PASSWORD = "admin123"
$STUDENT_USERNAME = "student001"
$STUDENT_PASSWORD = "password123"

# 测试结果统计
$totalTests = 0
$passedTests = 0
$failedTests = 0

function Test-API {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Endpoint,
        [string]$Body = "",
        [hashtable]$Headers = @{},
        [int]$ExpectedStatus = 200
    )
    
    $global:totalTests++
    Write-Host "测试: $Name" -ForegroundColor Cyan
    
    try {
        $uri = "$API_BASE$Endpoint"
        $headers["Content-Type"] = "application/json"
        
        $params = @{
            Uri     = $uri
            Method  = $Method
            Headers = $headers
        }
        
        if ($Body -and $Method -ne "GET") {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params -ErrorAction Stop
        $statusCode = $response.StatusCode
        
        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "  ✓ 通过" -ForegroundColor Green
            $global:passedTests++
            return $response
        }
        else {
            Write-Host "  ✗ 失败 - 状态码: $statusCode, 期望: $ExpectedStatus" -ForegroundColor Red
            $global:failedTests++
            return $null
        }
    }
    catch {
        Write-Host "  ✗ 失败 - $($_.Exception.Message)" -ForegroundColor Red
        $global:failedTests++
        return $null
    }
}

function Test-Login {
    param([string]$Username, [SecureString]$Password)
    
    $body = @{
        username = $Username
        password = $Password
    } | ConvertTo-Json
    
    $response = Test-API -Name "用户登录 ($Username)" -Method "POST" -Endpoint "/api/auth/login" -Body $body
    if ($response) {
        return $response.data.token
    }
    return $null
}

# 1. 测试服务健康状态
Write-Host "`n=== 1. 服务健康检查 ===" -ForegroundColor Magenta

Test-API -Name "API网关健康检查" -Method "GET" -Endpoint "/health"
Test-API -Name "认证服务健康检查" -Method "GET" -Endpoint "/api/auth/health"
Test-API -Name "用户服务健康检查" -Method "GET" -Endpoint "/api/users/health"
Test-API -Name "学分活动服务健康检查" -Method "GET" -Endpoint "/api/activities/health"

# 2. 测试认证功能
Write-Host "`n=== 2. 认证功能测试 ===" -ForegroundColor Magenta

$adminToken = Test-Login -Username $ADMIN_USERNAME -Password $ADMIN_PASSWORD
$studentToken = Test-Login -Username $STUDENT_USERNAME -Password $STUDENT_PASSWORD

if ($adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $adminToken" }
    Test-API -Name "验证管理员Token" -Method "POST" -Endpoint "/api/auth/validate-token" -Body '{"token":"'$adminToken'"}' -Headers $adminHeaders
}

if ($studentToken) {
    $studentHeaders = @{ "Authorization" = "Bearer $studentToken" }
    Test-API -Name "验证学生Token" -Method "POST" -Endpoint "/api/auth/validate-token" -Body '{"token":"'$studentToken'"}' -Headers $studentHeaders
}

# 3. 测试用户管理功能
Write-Host "`n=== 3. 用户管理功能测试 ===" -ForegroundColor Magenta

if ($adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $adminToken" }
    
    # 获取用户统计
    Test-API -Name "获取用户统计" -Method "GET" -Endpoint "/api/users/stats" -Headers $adminHeaders
    
    # 获取学生列表
    Test-API -Name "获取学生列表" -Method "GET" -Endpoint "/api/students" -Headers $adminHeaders
    
    # 获取教师列表
    Test-API -Name "获取教师列表" -Method "GET" -Endpoint "/api/teachers" -Headers $adminHeaders
    
    # 搜索用户
    Test-API -Name "搜索用户" -Method "GET" -Endpoint "/api/search/users?query=admin" -Headers $adminHeaders
}

# 4. 测试活动管理功能
Write-Host "`n=== 4. 活动管理功能测试 ===" -ForegroundColor Magenta

if ($adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $adminToken" }
    
    # 获取活动类别
    Test-API -Name "获取活动类别" -Method "GET" -Endpoint "/api/activities/categories" -Headers $adminHeaders
    
    # 获取活动列表
    Test-API -Name "获取活动列表" -Method "GET" -Endpoint "/api/activities" -Headers $adminHeaders
    
    # 获取活动统计
    Test-API -Name "获取活动统计" -Method "GET" -Endpoint "/api/activities/stats" -Headers $adminHeaders
    
    # 创建新活动
    $newActivity = @{
        title        = "测试活动"
        description  = "这是一个测试活动"
        start_date   = (Get-Date).AddDays(1).ToString("yyyy-MM-ddTHH:mm:ssZ")
        end_date     = (Get-Date).AddDays(2).ToString("yyyy-MM-ddTHH:mm:ssZ")
        category     = "学术活动"
        requirements = "无特殊要求"
    } | ConvertTo-Json
    
    $createResponse = Test-API -Name "创建活动" -Method "POST" -Endpoint "/api/activities" -Body $newActivity -Headers $adminHeaders
    
    if ($createResponse) {
        $activityId = $createResponse.data.id
        Test-API -Name "获取活动详情" -Method "GET" -Endpoint "/api/activities/$activityId" -Headers $adminHeaders
    }
}

# 5. 测试申请管理功能
Write-Host "`n=== 5. 申请管理功能测试 ===" -ForegroundColor Magenta

if ($studentToken) {
    $studentHeaders = @{ "Authorization" = "Bearer $studentToken" }
    
    # 获取学生申请列表
    Test-API -Name "获取学生申请列表" -Method "GET" -Endpoint "/api/applications" -Headers $studentHeaders
    
    # 获取申请统计
    Test-API -Name "获取申请统计" -Method "GET" -Endpoint "/api/applications/stats" -Headers $studentHeaders
}

if ($adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $adminToken" }
    
    # 获取所有申请列表
    Test-API -Name "获取所有申请列表" -Method "GET" -Endpoint "/api/applications/all" -Headers $adminHeaders
}

# 6. 测试权限管理功能
Write-Host "`n=== 6. 权限管理功能测试 ===" -ForegroundColor Magenta

if ($adminToken) {
    $adminHeaders = @{ "Authorization" = "Bearer $adminToken" }
    
    # 获取角色列表
    Test-API -Name "获取角色列表" -Method "GET" -Endpoint "/api/permissions/roles" -Headers $adminHeaders
    
    # 获取权限列表
    Test-API -Name "获取权限列表" -Method "GET" -Endpoint "/api/permissions" -Headers $adminHeaders
}

# 7. 测试前端访问
Write-Host "`n=== 7. 前端访问测试 ===" -ForegroundColor Magenta

try {
    $frontendResponse = Invoke-WebRequest -Uri $FRONTEND_URL -Method GET -TimeoutSec 10
    if ($frontendResponse.StatusCode -eq 200) {
        Write-Host "  ✓ 前端访问正常" -ForegroundColor Green
        $global:passedTests++
    }
    else {
        Write-Host "  ✗ 前端访问失败 - 状态码: $($frontendResponse.StatusCode)" -ForegroundColor Red
        $global:failedTests++
    }
}
catch {
    Write-Host "  ✗ 前端访问失败 - $($_.Exception.Message)" -ForegroundColor Red
    $global:failedTests++
}

# 8. 测试数据库连接
Write-Host "`n=== 8. 数据库连接测试 ===" -ForegroundColor Magenta

try {
    $dbResponse = Invoke-RestMethod -Uri "http://localhost:5432" -Method GET -TimeoutSec 5
    Write-Host "  ✓ 数据库连接正常" -ForegroundColor Green
    $global:passedTests++
}
catch {
    Write-Host "  ✗ 数据库连接失败 - $($_.Exception.Message)" -ForegroundColor Red
    $global:failedTests++   
}

# 测试结果汇总
Write-Host "`n=== 测试结果汇总 ===" -ForegroundColor Magenta
Write-Host "总测试数: $totalTests" -ForegroundColor White
Write-Host "通过测试: $passedTests" -ForegroundColor Green
Write-Host "失败测试: $failedTests" -ForegroundColor Red
Write-Host "成功率: $([math]::Round(($passedTests / $totalTests) * 100, 2))%" -ForegroundColor Yellow

if ($failedTests -eq 0) {
    Write-Host "`n🎉 所有测试通过！系统运行正常。" -ForegroundColor Green
}
else {
    Write-Host "`n⚠️  有 $failedTests 个测试失败，请检查相关服务。" -ForegroundColor Yellow
}

Write-Host "`n结束时间: $(Get-Date)" -ForegroundColor Yellow 