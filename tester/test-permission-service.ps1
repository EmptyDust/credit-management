# 权限管理服务测试脚本
# 测试权限管理服务的所有API接口

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

function Test-PermissionService {
    Write-Host "=== 权限管理服务测试 ===" -ForegroundColor Yellow
    
    $global:testResults = @()
    $global:authToken = $null
    $global:testUserId = $null
    $global:testRoleId = $null
    $global:testPermissionId = $null
    
    # 生成唯一测试用户名
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $testUsername = "testpermission_$timestamp"
    $testPassword = "password123"
    $testEmail = "$testUsername@example.com"
    
    Info "测试用户名: $testUsername"
    
    # 1. 测试用户注册
    Info "1. 测试用户注册"
    $registerBody = @{
        username = $testUsername
        password = $testPassword
        email = $testEmail
        real_name = "测试权限用户"
        user_type = "admin"
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
    
    # 3. 测试初始化权限系统
    Info "3. 测试初始化权限系统"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $initResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/init" -Method POST -Headers $headers -ContentType "application/json"
            if ($initResp.code -eq 0) {
                $global:testResults += Pass "初始化权限系统成功"
            } else {
                $global:testResults += Fail "初始化权限系统失败: $($initResp.message)"
            }
        } catch {
            $global:testResults += Fail "初始化权限系统异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过初始化权限系统测试 - 无有效令牌"
    }
    
    # 4. 测试创建角色
    Info "4. 测试创建角色"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $roleBody = @{
            name = "测试角色 $timestamp"
            description = "这是一个测试角色"
            permissions = @("read:users", "write:users")
        } | ConvertTo-Json
        
        try {
            $createRoleResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/roles" -Method POST -Body $roleBody -Headers $headers -ContentType "application/json"
            if ($createRoleResp.code -eq 0) {
                $global:testRoleId = $createRoleResp.data.id
                $global:testResults += Pass "创建角色成功，角色ID: $($createRoleResp.data.id)"
            } else {
                $global:testResults += Fail "创建角色失败: $($createRoleResp.message)"
            }
        } catch {
            $global:testResults += Fail "创建角色异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过创建角色测试 - 无有效令牌"
    }
    
    # 5. 测试获取所有角色列表
    Info "5. 测试获取所有角色列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $rolesResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/roles" -Method GET -Headers $headers
            if ($rolesResp.code -eq 0) {
                $global:testResults += Pass "获取所有角色列表成功"
            } else {
                $global:testResults += Fail "获取所有角色列表失败: $($rolesResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取所有角色列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取所有角色列表测试 - 无有效令牌"
    }
    
    # 6. 测试获取特定角色信息
    Info "6. 测试获取特定角色信息"
    if ($global:authToken -and $global:testRoleId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $roleResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/roles/$global:testRoleId" -Method GET -Headers $headers
            if ($roleResp.code -eq 0) {
                $global:testResults += Pass "获取特定角色信息成功"
            } else {
                $global:testResults += Fail "获取特定角色信息失败: $($roleResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取特定角色信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取特定角色信息测试 - 无有效令牌或角色ID"
    }
    
    # 7. 测试更新角色信息
    Info "7. 测试更新角色信息"
    if ($global:authToken -and $global:testRoleId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $updateRoleBody = @{
            name = "更新后的测试角色 $timestamp"
            description = "这是更新后的测试角色描述"
        } | ConvertTo-Json
        
        try {
            $updateRoleResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/roles/$global:testRoleId" -Method PUT -Body $updateRoleBody -Headers $headers -ContentType "application/json"
            if ($updateRoleResp.code -eq 0) {
                $global:testResults += Pass "更新角色信息成功"
            } else {
                $global:testResults += Fail "更新角色信息失败: $($updateRoleResp.message)"
            }
        } catch {
            $global:testResults += Fail "更新角色信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过更新角色信息测试 - 无有效令牌或角色ID"
    }
    
    # 8. 测试创建权限
    Info "8. 测试创建权限"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $permissionBody = @{
            name = "test:custom:permission:$timestamp"
            description = "测试自定义权限"
            resource = "test"
            action = "custom"
        } | ConvertTo-Json
        
        try {
            $createPermissionResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions" -Method POST -Body $permissionBody -Headers $headers -ContentType "application/json"
            if ($createPermissionResp.code -eq 0) {
                $global:testPermissionId = $createPermissionResp.data.id
                $global:testResults += Pass "创建权限成功，权限ID: $($createPermissionResp.data.id)"
            } else {
                $global:testResults += Fail "创建权限失败: $($createPermissionResp.message)"
            }
        } catch {
            $global:testResults += Fail "创建权限异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过创建权限测试 - 无有效令牌"
    }
    
    # 9. 测试获取所有权限列表
    Info "9. 测试获取所有权限列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $permissionsResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions" -Method GET -Headers $headers
            if ($permissionsResp.code -eq 0) {
                $global:testResults += Pass "获取所有权限列表成功"
            } else {
                $global:testResults += Fail "获取所有权限列表失败: $($permissionsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取所有权限列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取所有权限列表测试 - 无有效令牌"
    }
    
    # 10. 测试获取特定权限信息
    Info "10. 测试获取特定权限信息"
    if ($global:authToken -and $global:testPermissionId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $permissionResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/$global:testPermissionId" -Method GET -Headers $headers
            if ($permissionResp.code -eq 0) {
                $global:testResults += Pass "获取特定权限信息成功"
            } else {
                $global:testResults += Fail "获取特定权限信息失败: $($permissionResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取特定权限信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取特定权限信息测试 - 无有效令牌或权限ID"
    }
    
    # 11. 测试为用户分配角色
    Info "11. 测试为用户分配角色"
    if ($global:authToken -and $global:testUserId -and $global:testRoleId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $assignRoleBody = @{
            role_ids = @($global:testRoleId)
        } | ConvertTo-Json
        
        try {
            $assignRoleResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/users/$global:testUserId/roles" -Method POST -Body $assignRoleBody -Headers $headers -ContentType "application/json"
            if ($assignRoleResp.code -eq 0) {
                $global:testResults += Pass "为用户分配角色成功"
            } else {
                $global:testResults += Fail "为用户分配角色失败: $($assignRoleResp.message)"
            }
        } catch {
            $global:testResults += Fail "为用户分配角色异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过为用户分配角色测试 - 无有效令牌、用户ID或角色ID"
    }
    
    # 12. 测试获取用户角色列表
    Info "12. 测试获取用户角色列表"
    if ($global:authToken -and $global:testUserId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $userRolesResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/users/$global:testUserId/roles" -Method GET -Headers $headers
            if ($userRolesResp.code -eq 0) {
                $global:testResults += Pass "获取用户角色列表成功"
            } else {
                $global:testResults += Fail "获取用户角色列表失败: $($userRolesResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取用户角色列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取用户角色列表测试 - 无有效令牌或用户ID"
    }
    
    # 13. 测试获取用户权限列表
    Info "13. 测试获取用户权限列表"
    if ($global:authToken -and $global:testUserId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $userPermissionsResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/users/$global:testUserId/permissions" -Method GET -Headers $headers
            if ($userPermissionsResp.code -eq 0) {
                $global:testResults += Pass "获取用户权限列表成功"
            } else {
                $global:testResults += Fail "获取用户权限列表失败: $($userPermissionsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取用户权限列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取用户权限列表测试 - 无有效令牌或用户ID"
    }
    
    # 14. 测试为角色分配权限
    Info "14. 测试为角色分配权限"
    if ($global:authToken -and $global:testRoleId -and $global:testPermissionId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $assignPermissionBody = @{
            permission_ids = @($global:testPermissionId)
        } | ConvertTo-Json
        
        try {
            $assignPermissionResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/roles/$global:testRoleId/permissions" -Method POST -Body $assignPermissionBody -Headers $headers -ContentType "application/json"
            if ($assignPermissionResp.code -eq 0) {
                $global:testResults += Pass "为角色分配权限成功"
            } else {
                $global:testResults += Fail "为角色分配权限失败: $($assignPermissionResp.message)"
            }
        } catch {
            $global:testResults += Fail "为角色分配权限异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过为角色分配权限测试 - 无有效令牌、角色ID或权限ID"
    }
    
    # 15. 测试删除用户角色
    Info "15. 测试删除用户角色"
    if ($global:authToken -and $global:testUserId -and $global:testRoleId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deleteUserRoleResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/users/$global:testUserId/roles/$global:testRoleId" -Method DELETE -Headers $headers
            if ($deleteUserRoleResp.code -eq 0) {
                $global:testResults += Pass "删除用户角色成功"
            } else {
                $global:testResults += Fail "删除用户角色失败: $($deleteUserRoleResp.message)"
            }
        } catch {
            $global:testResults += Fail "删除用户角色异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过删除用户角色测试 - 无有效令牌、用户ID或角色ID"
    }
    
    # 16. 测试删除角色权限
    Info "16. 测试删除角色权限"
    if ($global:authToken -and $global:testRoleId -and $global:testPermissionId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deleteRolePermissionResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/roles/$global:testRoleId/permissions/$global:testPermissionId" -Method DELETE -Headers $headers
            if ($deleteRolePermissionResp.code -eq 0) {
                $global:testResults += Pass "删除角色权限成功"
            } else {
                $global:testResults += Fail "删除角色权限失败: $($deleteRolePermissionResp.message)"
            }
        } catch {
            $global:testResults += Fail "删除角色权限异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过删除角色权限测试 - 无有效令牌、角色ID或权限ID"
    }
    
    # 17. 测试删除权限
    Info "17. 测试删除权限"
    if ($global:authToken -and $global:testPermissionId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deletePermissionResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/$global:testPermissionId" -Method DELETE -Headers $headers
            if ($deletePermissionResp.code -eq 0) {
                $global:testResults += Pass "删除权限成功"
            } else {
                $global:testResults += Fail "删除权限失败: $($deletePermissionResp.message)"
            }
        } catch {
            $global:testResults += Fail "删除权限异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过删除权限测试 - 无有效令牌或权限ID"
    }
    
    # 18. 测试删除角色
    Info "18. 测试删除角色"
    if ($global:authToken -and $global:testRoleId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deleteRoleResp = Invoke-RestMethod -Uri "$BaseUrl/api/permissions/roles/$global:testRoleId" -Method DELETE -Headers $headers
            if ($deleteRoleResp.code -eq 0) {
                $global:testResults += Pass "删除角色成功"
            } else {
                $global:testResults += Fail "删除角色失败: $($deleteRoleResp.message)"
            }
        } catch {
            $global:testResults += Fail "删除角色异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过删除角色测试 - 无有效令牌或角色ID"
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
Test-PermissionService 