# 教师信息服务测试脚本
# 测试教师信息服务的所有API接口

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

function Test-TeacherInfoService {
    Write-Host "=== 教师信息服务测试 ===" -ForegroundColor Yellow
    
    $global:testResults = @()
    $global:authToken = $null
    $global:testTeacherId = $null
    
    # 生成唯一测试用户名
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $testUsername = "testteacher_$timestamp"
    $testPassword = "password123"
    $testEmail = "$testUsername@example.com"
    
    Info "测试教师用户名: $testUsername"
    
    # 1. 测试用户注册
    Info "1. 测试用户注册"
    $registerBody = @{
        username = $testUsername
        password = $testPassword
        email = $testEmail
        real_name = "测试教师"
        user_type = "teacher"
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
    
    # 3. 测试创建教师信息
    Info "3. 测试创建教师信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $teacherBody = @{
            user_id = $global:testUserId
            teacher_id = "T2024$timestamp"
            name = "测试教师"
            department = "计算机学院"
            title = "副教授"
            status = "active"
            phone = "13800138000"
            email = $testEmail
            research_area = "软件工程"
        } | ConvertTo-Json
        
        try {
            $createTeacherResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers" -Method POST -Body $teacherBody -Headers $headers -ContentType "application/json"
            if ($createTeacherResp.code -eq 0) {
                $global:testTeacherId = $createTeacherResp.data.id
                $global:testResults += Pass "创建教师信息成功，教师ID: $($createTeacherResp.data.id)"
            } else {
                $global:testResults += Fail "创建教师信息失败: $($createTeacherResp.message)"
            }
        } catch {
            $global:testResults += Fail "创建教师信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过创建教师信息测试 - 无有效令牌"
    }
    
    # 4. 测试获取所有教师列表
    Info "4. 测试获取所有教师列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $teachersResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers" -Method GET -Headers $headers
            if ($teachersResp.code -eq 0) {
                $global:testResults += Pass "获取所有教师列表成功"
            } else {
                $global:testResults += Fail "获取所有教师列表失败: $($teachersResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取所有教师列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取所有教师列表测试 - 无有效令牌"
    }
    
    # 5. 测试获取特定教师信息
    Info "5. 测试获取特定教师信息"
    if ($global:authToken -and $global:testTeacherId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $teacherResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/$global:testTeacherId" -Method GET -Headers $headers
            if ($teacherResp.code -eq 0) {
                $global:testResults += Pass "获取特定教师信息成功"
            } else {
                $global:testResults += Fail "获取特定教师信息失败: $($teacherResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取特定教师信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取特定教师信息测试 - 无有效令牌或教师ID"
    }
    
    # 6. 测试按部门获取教师列表
    Info "6. 测试按部门获取教师列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $departmentResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/department/计算机学院" -Method GET -Headers $headers
            if ($departmentResp.code -eq 0) {
                $global:testResults += Pass "按部门获取教师列表成功"
            } else {
                $global:testResults += Fail "按部门获取教师列表失败: $($departmentResp.message)"
            }
        } catch {
            $global:testResults += Fail "按部门获取教师列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按部门获取教师列表测试 - 无有效令牌"
    }
    
    # 7. 测试按职称获取教师列表
    Info "7. 测试按职称获取教师列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $titleResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/title/副教授" -Method GET -Headers $headers
            if ($titleResp.code -eq 0) {
                $global:testResults += Pass "按职称获取教师列表成功"
            } else {
                $global:testResults += Fail "按职称获取教师列表失败: $($titleResp.message)"
            }
        } catch {
            $global:testResults += Fail "按职称获取教师列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按职称获取教师列表测试 - 无有效令牌"
    }
    
    # 8. 测试按状态获取教师列表
    Info "8. 测试按状态获取教师列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $statusResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/status/active" -Method GET -Headers $headers
            if ($statusResp.code -eq 0) {
                $global:testResults += Pass "按状态获取教师列表成功"
            } else {
                $global:testResults += Fail "按状态获取教师列表失败: $($statusResp.message)"
            }
        } catch {
            $global:testResults += Fail "按状态获取教师列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按状态获取教师列表测试 - 无有效令牌"
    }
    
    # 9. 测试搜索教师
    Info "9. 测试搜索教师"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $searchResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/search?q=测试教师" -Method GET -Headers $headers
            if ($searchResp.code -eq 0) {
                $global:testResults += Pass "搜索教师成功"
            } else {
                $global:testResults += Fail "搜索教师失败: $($searchResp.message)"
            }
        } catch {
            $global:testResults += Fail "搜索教师异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过搜索教师测试 - 无有效令牌"
    }
    
    # 10. 测试按用户名搜索教师
    Info "10. 测试按用户名搜索教师"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $searchByUsernameResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/search/username?username=$testUsername" -Method GET -Headers $headers
            if ($searchByUsernameResp.code -eq 0) {
                $global:testResults += Pass "按用户名搜索教师成功"
            } else {
                $global:testResults += Fail "按用户名搜索教师失败: $($searchByUsernameResp.message)"
            }
        } catch {
            $global:testResults += Fail "按用户名搜索教师异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按用户名搜索教师测试 - 无有效令牌"
    }
    
    # 11. 测试获取活跃教师列表
    Info "11. 测试获取活跃教师列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $activeResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/active" -Method GET -Headers $headers
            if ($activeResp.code -eq 0) {
                $global:testResults += Pass "获取活跃教师列表成功"
            } else {
                $global:testResults += Fail "获取活跃教师列表失败: $($activeResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取活跃教师列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取活跃教师列表测试 - 无有效令牌"
    }
    
    # 12. 测试更新教师信息
    Info "12. 测试更新教师信息"
    if ($global:authToken -and $global:testTeacherId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $updateTeacherBody = @{
            name = "更新后的测试教师"
            phone = "13900139000"
            email = "updated_$testEmail"
            research_area = "人工智能"
        } | ConvertTo-Json
        
        try {
            $updateTeacherResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/$global:testTeacherId" -Method PUT -Body $updateTeacherBody -Headers $headers -ContentType "application/json"
            if ($updateTeacherResp.code -eq 0) {
                $global:testResults += Pass "更新教师信息成功"
            } else {
                $global:testResults += Fail "更新教师信息失败: $($updateTeacherResp.message)"
            }
        } catch {
            $global:testResults += Fail "更新教师信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过更新教师信息测试 - 无有效令牌或教师ID"
    }
    
    # 13. 测试删除教师信息
    Info "13. 测试删除教师信息"
    if ($global:authToken -and $global:testTeacherId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deleteTeacherResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/$global:testTeacherId" -Method DELETE -Headers $headers
            if ($deleteTeacherResp.code -eq 0) {
                $global:testResults += Pass "删除教师信息成功"
            } else {
                $global:testResults += Fail "删除教师信息失败: $($deleteTeacherResp.message)"
            }
        } catch {
            $global:testResults += Fail "删除教师信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过删除教师信息测试 - 无有效令牌或教师ID"
    }
    
    # 14. 测试删除后获取教师信息（应该失败）
    Info "14. 测试删除后获取教师信息（应该失败）"
    if ($global:authToken -and $global:testTeacherId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deletedTeacherResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/$global:testTeacherId" -Method GET -Headers $headers
            if ($deletedTeacherResp.code -ne 0) {
                $global:testResults += Pass "删除后获取教师信息被正确拒绝"
            } else {
                $global:testResults += Fail "删除后获取教师信息应该被拒绝"
            }
        } catch {
            $global:testResults += Pass "删除后获取教师信息被正确拒绝"
        }
    } else {
        $global:testResults += Fail "跳过删除后获取教师信息测试 - 无有效令牌或教师ID"
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
Test-TeacherInfoService 