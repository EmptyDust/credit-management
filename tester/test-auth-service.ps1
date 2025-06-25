# 认证服务测试脚本
# 测试认证服务的所有API接口

param(
    [string]$BaseUrl = "http://localhost:8080"
)

function Fail($msg) {
    Write-Host "[FAIL] $msg" -ForegroundColor Red
    return $false
}

function Pass($msg) {
    Write-Host "[PASS] $msg" -ForegroundColor Green
    return $true
}

function Info($msg) {
    Write-Host "[INFO] $msg" -ForegroundColor Cyan
}

function Test-AuthService {
    Write-Host "=== 认证服务测试 ===" -ForegroundColor Yellow
    
    $global:testResults = @()
    $global:authToken = $null
    $global:refreshToken = $null
    $global:testUserId = $null
    
    # 生成唯一测试用户名
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $testUsername = "testuser_$timestamp"
    $testPassword = "password123"
    $testEmail = "$testUsername@example.com"
    
    Info "测试用户名: $testUsername"
    
    # 1. 测试用户注册
    Info "1. 测试用户注册"
    $registerBody = @{
        username = $testUsername
        password = $testPassword
        email = $testEmail
        real_name = "测试用户"
        user_type = "student"
    } | ConvertTo-Json
    
    try {
        $regResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/register" -Method POST -Body $registerBody -ContentType "application/json"
        if ($regResp.code -eq 0) {
            $global:testUserId = $regResp.data.id
            $global:testResults += Pass "用户注册成功，用户ID: $($regResp.data.id)"
        } else {
            $global:testResults += Fail "用户注册失败: $($regResp.message)"
        }
    } catch {
        $global:testResults += Fail "用户注册异常: $($_.Exception.Message)"
    }
    
    # 2. 测试用户登录
    Info "2. 测试用户登录"
    $loginBody = @{
        username = $testUsername
        password = $testPassword
    } | ConvertTo-Json
    
    try {
        $loginResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        if ($loginResp.code -eq 0) {
            $global:authToken = $loginResp.data.token
            $global:refreshToken = $loginResp.data.refresh_token
            $global:testResults += Pass "用户登录成功，获取到访问令牌"
        } else {
            $global:testResults += Fail "用户登录失败: $($loginResp.message)"
        }
    } catch {
        $global:testResults += Fail "用户登录异常: $($_.Exception.Message)"
    }
    
    # 3. 测试令牌验证
    Info "3. 测试令牌验证"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $validateResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/validate" -Method POST -Headers $headers -ContentType "application/json"
            if ($validateResp.code -eq 0 -and $validateResp.data.valid) {
                $global:testResults += Pass "令牌验证成功"
            } else {
                $global:testResults += Fail "令牌验证失败: $($validateResp.message)"
            }
        } catch {
            $global:testResults += Fail "令牌验证异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过令牌验证测试 - 无有效令牌"
    }
    
    # 4. 测试刷新令牌
    Info "4. 测试刷新令牌"
    if ($global:refreshToken) {
        $refreshBody = @{
            refresh_token = $global:refreshToken
        } | ConvertTo-Json
        
        try {
            $refreshResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/refresh" -Method POST -Body $refreshBody -ContentType "application/json"
            if ($refreshResp.code -eq 0) {
                $global:authToken = $refreshResp.data.token
                $global:refreshToken = $refreshResp.data.refresh_token
                $global:testResults += Pass "令牌刷新成功"
            } else {
                $global:testResults += Fail "令牌刷新失败: $($refreshResp.message)"
            }
        } catch {
            $global:testResults += Fail "令牌刷新异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过令牌刷新测试 - 无有效刷新令牌"
    }
    
    # 5. 测试错误密码登录
    Info "5. 测试错误密码登录"
    $wrongPasswordBody = @{
        username = $testUsername
        password = "wrongpassword"
    } | ConvertTo-Json
    
    try {
        $wrongLoginResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Body $wrongPasswordBody -ContentType "application/json"
        if ($wrongLoginResp.code -ne 0) {
            $global:testResults += Pass "错误密码登录被正确拒绝"
        } else {
            $global:testResults += Fail "错误密码登录应该被拒绝"
        }
    } catch {
        $global:testResults += Pass "错误密码登录被正确拒绝"
    }
    
    # 6. 测试不存在的用户登录
    Info "6. 测试不存在的用户登录"
    $nonexistentUserBody = @{
        username = "nonexistentuser_$timestamp"
        password = "password123"
    } | ConvertTo-Json
    
    try {
        $nonexistentLoginResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Body $nonexistentUserBody -ContentType "application/json"
        if ($nonexistentLoginResp.code -ne 0) {
            $global:testResults += Pass "不存在用户登录被正确拒绝"
        } else {
            $global:testResults += Fail "不存在用户登录应该被拒绝"
        }
    } catch {
        $global:testResults += Pass "不存在用户登录被正确拒绝"
    }
    
    # 7. 测试用户登出
    Info "7. 测试用户登出"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $logoutResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/logout" -Method POST -Headers $headers -ContentType "application/json"
            if ($logoutResp.code -eq 0) {
                $global:testResults += Pass "用户登出成功"
            } else {
                $global:testResults += Fail "用户登出失败: $($logoutResp.message)"
            }
        } catch {
            $global:testResults += Fail "用户登出异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过用户登出测试 - 无有效令牌"
    }
    
    # 8. 测试登出后令牌验证
    Info "8. 测试登出后令牌验证"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $validateAfterLogoutResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/validate" -Method POST -Headers $headers -ContentType "application/json"
            if ($validateAfterLogoutResp.code -ne 0) {
                $global:testResults += Pass "登出后令牌验证被正确拒绝"
            } else {
                $global:testResults += Fail "登出后令牌验证应该被拒绝"
            }
        } catch {
            $global:testResults += Pass "登出后令牌验证被正确拒绝"
        }
    } else {
        $global:testResults += Fail "跳过登出后令牌验证测试 - 无有效令牌"
    }
    
    # 输出测试结果统计
    Write-Host "`n=== 测试结果统计 ===" -ForegroundColor Yellow
    $passCount = ($global:testResults | Where-Object { $_ -eq $true }).Count
    $failCount = ($global:testResults | Where-Object { $_ -eq $false }).Count
    $totalCount = $global:testResults.Count
    
    Write-Host "总测试数: $totalCount" -ForegroundColor White
    Write-Host "通过: $passCount" -ForegroundColor Green
    Write-Host "失败: $failCount" -ForegroundColor Red
    
    if ($failCount -eq 0) {
        Write-Host "`n🎉 所有测试通过！" -ForegroundColor Green
        return $true
    } else {
        Write-Host "`n❌ 有 $failCount 个测试失败" -ForegroundColor Red
        return $false
    }
}

# 执行测试
Test-AuthService 