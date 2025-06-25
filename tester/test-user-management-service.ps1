# 用户管理服务测试脚本
# 测试用户管理服务的所有API接口

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

function Test-UserManagementService {
    Write-Host "=== 用户管理服务测试 ===" -ForegroundColor Yellow
    
    $global:testResults = @()
    $global:authToken = $null
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
    
    # 2. 测试用户登录获取令牌
    Info "2. 测试用户登录获取令牌"
    $loginBody = @{
        username = $testUsername
        password = $testPassword
    } | ConvertTo-Json
    
    try {
        $loginResp = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        if ($loginResp.code -eq 0) {
            $global:authToken = $loginResp.data.token
            $global:testResults += Pass "用户登录成功，获取到访问令牌"
        } else {
            $global:testResults += Fail "用户登录失败: $($loginResp.message)"
        }
    } catch {
        $global:testResults += Fail "用户登录异常: $($_.Exception.Message)"
    }
    
    # 3. 测试获取用户统计信息
    Info "3. 测试获取用户统计信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $statsResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/stats" -Method GET -Headers $headers
            if ($statsResp.code -eq 0) {
                $global:testResults += Pass "获取用户统计信息成功"
            } else {
                $global:testResults += Fail "获取用户统计信息失败: $($statsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取用户统计信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取用户统计信息测试 - 无有效令牌"
    }
    
    # 4. 测试获取用户个人资料
    Info "4. 测试获取用户个人资料"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $profileResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/profile" -Method GET -Headers $headers
            if ($profileResp.code -eq 0) {
                $global:testResults += Pass "获取用户个人资料成功"
            } else {
                $global:testResults += Fail "获取用户个人资料失败: $($profileResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取用户个人资料异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取用户个人资料测试 - 无有效令牌"
    }
    
    # 5. 测试更新用户个人资料
    Info "5. 测试更新用户个人资料"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $updateProfileBody = @{
            real_name = "更新后的测试用户"
            email = "updated_$testEmail"
        } | ConvertTo-Json
        
        try {
            $updateProfileResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/profile" -Method PUT -Body $updateProfileBody -Headers $headers -ContentType "application/json"
            if ($updateProfileResp.code -eq 0) {
                $global:testResults += Pass "更新用户个人资料成功"
            } else {
                $global:testResults += Fail "更新用户个人资料失败: $($updateProfileResp.message)"
            }
        } catch {
            $global:testResults += Fail "更新用户个人资料异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过更新用户个人资料测试 - 无有效令牌"
    }
    
    # 6. 测试获取所有用户列表
    Info "6. 测试获取所有用户列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $usersResp = Invoke-RestMethod -Uri "$BaseUrl/api/users" -Method GET -Headers $headers
            if ($usersResp.code -eq 0) {
                $global:testResults += Pass "获取所有用户列表成功"
            } else {
                $global:testResults += Fail "获取所有用户列表失败: $($usersResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取所有用户列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取所有用户列表测试 - 无有效令牌"
    }
    
    # 7. 测试按用户类型获取用户列表
    Info "7. 测试按用户类型获取用户列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $studentUsersResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/type/student" -Method GET -Headers $headers
            if ($studentUsersResp.code -eq 0) {
                $global:testResults += Pass "按用户类型获取用户列表成功"
            } else {
                $global:testResults += Fail "按用户类型获取用户列表失败: $($studentUsersResp.message)"
            }
        } catch {
            $global:testResults += Fail "按用户类型获取用户列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按用户类型获取用户列表测试 - 无有效令牌"
    }
    
    # 8. 测试获取特定用户信息
    Info "8. 测试获取特定用户信息"
    if ($global:authToken -and $global:testUserId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $userResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/$global:testUserId" -Method GET -Headers $headers
            if ($userResp.code -eq 0) {
                $global:testResults += Pass "获取特定用户信息成功"
            } else {
                $global:testResults += Fail "获取特定用户信息失败: $($userResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取特定用户信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取特定用户信息测试 - 无有效令牌或用户ID"
    }
    
    # 9. 测试更新特定用户信息
    Info "9. 测试更新特定用户信息"
    if ($global:authToken -and $global:testUserId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $updateUserBody = @{
            real_name = "管理员更新的测试用户"
            email = "admin_updated_$testEmail"
        } | ConvertTo-Json
        
        try {
            $updateUserResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/$global:testUserId" -Method PUT -Body $updateUserBody -Headers $headers -ContentType "application/json"
            if ($updateUserResp.code -eq 0) {
                $global:testResults += Pass "更新特定用户信息成功"
            } else {
                $global:testResults += Fail "更新特定用户信息失败: $($updateUserResp.message)"
            }
        } catch {
            $global:testResults += Fail "更新特定用户信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过更新特定用户信息测试 - 无有效令牌或用户ID"
    }
    
    # 10. 测试删除用户
    Info "10. 测试删除用户"
    if ($global:authToken -and $global:testUserId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deleteUserResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/$global:testUserId" -Method DELETE -Headers $headers
            if ($deleteUserResp.code -eq 0) {
                $global:testResults += Pass "删除用户成功"
            } else {
                $global:testResults += Fail "删除用户失败: $($deleteUserResp.message)"
            }
        } catch {
            $global:testResults += Fail "删除用户异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过删除用户测试 - 无有效令牌或用户ID"
    }
    
    # 11. 测试删除后获取用户信息（应该失败）
    Info "11. 测试删除后获取用户信息（应该失败）"
    if ($global:authToken -and $global:testUserId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deletedUserResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/$global:testUserId" -Method GET -Headers $headers
            if ($deletedUserResp.code -ne 0) {
                $global:testResults += Pass "删除后获取用户信息被正确拒绝"
            } else {
                $global:testResults += Fail "删除后获取用户信息应该被拒绝"
            }
        } catch {
            $global:testResults += Pass "删除后获取用户信息被正确拒绝"
        }
    } else {
        $global:testResults += Fail "跳过删除后获取用户信息测试 - 无有效令牌或用户ID"
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
Test-UserManagementService 