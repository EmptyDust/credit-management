# 综合测试脚本
# 测试所有服务的集成功能

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

function Test-AllServices {
    Write-Host "=== 综合服务测试 ===" -ForegroundColor Yellow
    
    $global:testResults = @()
    $global:authToken = $null
    $global:testUserId = $null
    $global:testStudentId = $null
    $global:testTeacherId = $null
    $global:testActivityId = $null
    $global:testApplicationId = $null
    
    # 生成唯一测试用户名
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $testUsername = "testall_$timestamp"
    $testPassword = "password123"
    $testEmail = "$testUsername@example.com"
    
    Info "测试用户名: $testUsername"
    
    # 1. 测试用户注册
    Info "1. 测试用户注册"
    $registerBody = @{
        username = $testUsername
        password = $testPassword
        email = $testEmail
        real_name = "综合测试用户"
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
    
    # 3. 测试创建学生信息
    Info "3. 测试创建学生信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $studentBody = @{
            user_id = $global:testUserId
            student_id = "2024$timestamp"
            name = "综合测试学生"
            college = "计算机学院"
            major = "软件工程"
            class = "软工2024-1班"
            grade = "2024"
            status = "active"
            phone = "13800138000"
            email = $testEmail
        } | ConvertTo-Json
        
        try {
            $createStudentResp = Invoke-RestMethod -Uri "$BaseUrl/api/students" -Method POST -Body $studentBody -Headers $headers -ContentType "application/json"
            if ($createStudentResp.code -eq 0) {
                $global:testStudentId = $createStudentResp.data.id
                $global:testResults += Pass "创建学生信息成功，学生ID: $($createStudentResp.data.id)"
            } else {
                $global:testResults += Fail "创建学生信息失败: $($createStudentResp.message)"
            }
        } catch {
            $global:testResults += Fail "创建学生信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过创建学生信息测试 - 无有效令牌"
    }
    
    # 4. 测试创建教师信息
    Info "4. 测试创建教师信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $teacherBody = @{
            user_id = $global:testUserId
            teacher_id = "T2024$timestamp"
            name = "综合测试教师"
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
    
    # 5. 测试创建学分活动
    Info "5. 测试创建学分活动"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $activityBody = @{
            title = "综合测试学分活动 $timestamp"
            description = "这是一个综合测试学分活动"
            category = "academic"
            status = "draft"
            start_date = (Get-Date).ToString("yyyy-MM-dd")
            end_date = (Get-Date).AddDays(30).ToString("yyyy-MM-dd")
            max_participants = 10
            credit_value = 2.0
            participants = @($global:testUserId)
        } | ConvertTo-Json
        
        try {
            $createActivityResp = Invoke-RestMethod -Uri "$BaseUrl/api/activities" -Method POST -Body $activityBody -Headers $headers -ContentType "application/json"
            if ($createActivityResp.code -eq 0) {
                $global:testActivityId = $createActivityResp.data.id
                $global:testResults += Pass "创建学分活动成功，活动ID: $($createActivityResp.data.id)"
            } else {
                $global:testResults += Fail "创建学分活动失败: $($createActivityResp.message)"
            }
        } catch {
            $global:testResults += Fail "创建学分活动异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过创建学分活动测试 - 无有效令牌"
    }
    
    # 6. 测试创建申请
    Info "6. 测试创建申请"
    if ($global:authToken -and $global:testActivityId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $applicationBody = @{
            activity_id = $global:testActivityId
            user_id = $global:testUserId
            applied_credits = 2.0
        } | ConvertTo-Json
        
        try {
            $createApplicationResp = Invoke-RestMethod -Uri "$BaseUrl/api/applications" -Method POST -Body $applicationBody -Headers $headers -ContentType "application/json"
            if ($createApplicationResp.code -eq 0) {
                $global:testApplicationId = $createApplicationResp.data.id
                $global:testResults += Pass "创建申请成功，申请ID: $($createApplicationResp.data.id)"
            } else {
                $global:testResults += Fail "创建申请失败: $($createApplicationResp.message)"
            }
        } catch {
            $global:testResults += Fail "创建申请异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过创建申请测试 - 无有效令牌或无活动ID"
    }
    
    # 7. 测试获取活动参与者列表
    Info "7. 测试获取活动参与者列表"
    if ($global:authToken -and $global:testActivityId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $participantsResp = Invoke-RestMethod -Uri "$BaseUrl/api/activities/$global:testActivityId/participants" -Method GET -Headers $headers
            if ($participantsResp.code -eq 0) {
                $global:testResults += Pass "获取活动参与者列表成功"
            } else {
                $global:testResults += Fail "获取活动参与者列表失败: $($participantsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取活动参与者列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取活动参与者列表测试 - 无有效令牌或无活动ID"
    }
    
    # 8. 测试获取活动申请列表
    Info "8. 测试获取活动申请列表"
    if ($global:authToken -and $global:testActivityId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $applicationsResp = Invoke-RestMethod -Uri "$BaseUrl/api/applications" -Method GET -Headers $headers
            if ($applicationsResp.code -eq 0) {
                $global:testResults += Pass "获取活动申请列表成功"
            } else {
                $global:testResults += Fail "获取活动申请列表失败: $($applicationsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取活动申请列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取活动申请列表测试 - 无有效令牌或无活动ID"
    }
    
    # 9. 测试获取申请统计信息
    Info "9. 测试获取申请统计信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $statsResp = Invoke-RestMethod -Uri "$BaseUrl/api/applications/stats" -Method GET -Headers $headers
            if ($statsResp.code -eq 0) {
                $global:testResults += Pass "获取申请统计信息成功"
            } else {
                $global:testResults += Fail "获取申请统计信息失败: $($statsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取申请统计信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取申请统计信息测试 - 无有效令牌"
    }
    
    # 10. 测试获取用户统计信息
    Info "10. 测试获取用户统计信息"
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
    
    # 11. 测试搜索学生
    Info "11. 测试搜索学生"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $searchStudentResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/search?q=综合测试学生" -Method GET -Headers $headers
            if ($searchStudentResp.code -eq 0) {
                $global:testResults += Pass "搜索学生成功"
            } else {
                $global:testResults += Fail "搜索学生失败: $($searchStudentResp.message)"
            }
        } catch {
            $global:testResults += Fail "搜索学生异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过搜索学生测试 - 无有效令牌"
    }
    
    # 12. 测试搜索教师
    Info "12. 测试搜索教师"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $searchTeacherResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/search?q=综合测试教师" -Method GET -Headers $headers
            if ($searchTeacherResp.code -eq 0) {
                $global:testResults += Pass "搜索教师成功"
            } else {
                $global:testResults += Fail "搜索教师失败: $($searchTeacherResp.message)"
            }
        } catch {
            $global:testResults += Fail "搜索教师异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过搜索教师测试 - 无有效令牌"
    }
    
    # 13. 测试令牌验证
    Info "13. 测试令牌验证"
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
    
    # 14. 测试清理资源
    Info "14. 测试清理资源"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        # 删除活动
        if ($global:testActivityId) {
            try {
                $deleteActivityResp = Invoke-RestMethod -Uri "$BaseUrl/api/activities/$global:testActivityId" -Method DELETE -Headers $headers
                if ($deleteActivityResp.code -eq 0) {
                    $global:testResults += Pass "删除活动成功"
                } else {
                    $global:testResults += Fail "删除活动失败: $($deleteActivityResp.message)"
                }
            } catch {
                $global:testResults += Fail "删除活动异常: $($_.Exception.Message)"
            }
        }
        
        # 删除学生信息
        if ($global:testStudentId) {
            try {
                $deleteStudentResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/$global:testStudentId" -Method DELETE -Headers $headers
                if ($deleteStudentResp.code -eq 0) {
                    $global:testResults += Pass "删除学生信息成功"
                } else {
                    $global:testResults += Fail "删除学生信息失败: $($deleteStudentResp.message)"
                }
            } catch {
                $global:testResults += Fail "删除学生信息异常: $($_.Exception.Message)"
            }
        }
        
        # 删除教师信息
        if ($global:testTeacherId) {
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
        }
        
        # 删除用户
        if ($global:testUserId) {
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
        }
    } else {
        $global:testResults += Fail "跳过清理资源测试 - 无有效令牌"
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
Test-AllServices 