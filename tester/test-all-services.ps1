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
    $global:testActivityId = $null
    $global:testApplicationId = $null
    
    # 生成唯一测试用户名
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $testUsername = "testall_$timestamp"
    $testPassword = "Password123"
    $testEmail = "$testUsername@example.com"
    
    Info "测试用户名: $testUsername"
    
    # 1. 测试用户注册
    Info "1. 测试用户注册"
    $registerBody = @{
        username = $testUsername
        password = $testPassword
        email = $testEmail
        phone = "13800138000"
        real_name = "综合测试学生"
        user_type = "student"
        student_id = "2024$timestamp"
        college = "计算机学院"
        major = "软件工程"
        class = "软工2024-1班"
        grade = "2024"
    } | ConvertTo-Json
    
    try {
        $regResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/register" -Method POST -Body $registerBody -ContentType "application/json"
        if ($regResp.code -eq 0) {
            $global:testUserId = $regResp.data.user.user_id
            $global:testResults += Pass "用户注册成功，用户ID: $($regResp.data.user.user_id)"
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
    
    # 3. 测试获取用户信息
    Info "3. 测试获取用户信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        try {
            $userResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/profile" -Method GET -Headers $headers -ContentType "application/json"
            if ($userResp.code -eq 0) {
                $global:testResults += Pass "获取用户信息成功"
            } else {
                $global:testResults += Fail "获取用户信息失败: $($userResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取用户信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取用户信息测试 - 无有效令牌"
    }
    
    # 4. 测试获取学生列表
    Info "4. 测试获取学生列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        try {
            $studentsResp = Invoke-RestMethod -Uri "$BaseUrl/api/students" -Method GET -Headers $headers -ContentType "application/json"
            if ($studentsResp.code -eq 0) {
                $global:testResults += Pass "获取学生列表成功"
            } else {
                $global:testResults += Fail "获取学生列表失败: $($studentsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取学生列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取学生列表测试 - 无有效令牌"
    }
    
    # 5. 测试获取教师列表
    Info "5. 测试获取教师列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        try {
            $teachersResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers" -Method GET -Headers $headers -ContentType "application/json"
            if ($teachersResp.code -eq 0) {
                $global:testResults += Pass "获取教师列表成功"
            } else {
                $global:testResults += Fail "获取教师列表失败: $($teachersResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取教师列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取教师列表测试 - 无有效令牌"
    }
    
    # 6. 测试创建学分活动
    Info "6. 测试创建学分活动"
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
    
    # 7. 测试创建申请
    Info "7. 测试创建申请"
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
    
    # 8. 测试搜索功能
    Info "8. 测试搜索功能"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        try {
            $searchResp = Invoke-RestMethod -Uri "$BaseUrl/api/search/users?query=testall" -Method GET -Headers $headers -ContentType "application/json"
            if ($searchResp.code -eq 0) {
                $global:testResults += Pass "搜索用户成功"
            } else {
                $global:testResults += Fail "搜索用户失败: $($searchResp.message)"
            }
        } catch {
            $global:testResults += Fail "搜索用户异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过搜索功能测试 - 无有效令牌"
    }
    
    # 9. 测试获取统计信息
    Info "9. 测试获取统计信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        try {
            $statsResp = Invoke-RestMethod -Uri "$BaseUrl/api/users/stats" -Method GET -Headers $headers -ContentType "application/json"
            if ($statsResp.code -eq 0) {
                $global:testResults += Pass "获取用户统计信息成功"
            } else {
                $global:testResults += Fail "获取用户统计信息失败: $($statsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取用户统计信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取统计信息测试 - 无有效令牌"
    }
    
    # 10. 测试获取学生统计信息
    Info "10. 测试获取学生统计信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        try {
            $studentStatsResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/stats" -Method GET -Headers $headers -ContentType "application/json"
            if ($studentStatsResp.code -eq 0) {
                $global:testResults += Pass "获取学生统计信息成功"
            } else {
                $global:testResults += Fail "获取学生统计信息失败: $($studentStatsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取学生统计信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取学生统计信息测试 - 无有效令牌"
    }
    
    # 11. 测试获取教师统计信息
    Info "11. 测试获取教师统计信息"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        try {
            $teacherStatsResp = Invoke-RestMethod -Uri "$BaseUrl/api/teachers/stats" -Method GET -Headers $headers -ContentType "application/json"
            if ($teacherStatsResp.code -eq 0) {
                $global:testResults += Pass "获取教师统计信息成功"
            } else {
                $global:testResults += Fail "获取教师统计信息失败: $($teacherStatsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取教师统计信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取教师统计信息测试 - 无有效令牌"
    }
    
    # 12. 测试资源清理
    Info "12. 测试资源清理"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        
        # 清理测试数据
        if ($global:testApplicationId) {
            try {
                Invoke-RestMethod -Uri "$BaseUrl/api/applications/$global:testApplicationId" -Method DELETE -Headers $headers -ContentType "application/json" | Out-Null
                $global:testResults += Pass "清理测试申请成功"
            } catch {
                $global:testResults += Fail "清理测试申请失败: $($_.Exception.Message)"
            }
        }
        
        if ($global:testActivityId) {
            try {
                Invoke-RestMethod -Uri "$BaseUrl/api/activities/$global:testActivityId" -Method DELETE -Headers $headers -ContentType "application/json" | Out-Null
                $global:testResults += Pass "清理测试活动成功"
            } catch {
                $global:testResults += Fail "清理测试活动失败: $($_.Exception.Message)"
            }
        }
        
        if ($global:testUserId) {
            try {
                Invoke-RestMethod -Uri "$BaseUrl/api/users/$global:testUserId" -Method DELETE -Headers $headers -ContentType "application/json" | Out-Null
                $global:testResults += Pass "清理测试用户成功"
            } catch {
                $global:testResults += Fail "清理测试用户失败: $($_.Exception.Message)"
            }
        }
    } else {
        $global:testResults += Fail "跳过资源清理测试 - 无有效令牌"
    }
    
    # 输出测试结果
    Write-Host ""
    Write-Host "=== 测试结果汇总 ===" -ForegroundColor Yellow
    $passCount = ($global:testResults | Where-Object { $_ -like "*[PASS]*" }).Count
    $failCount = ($global:testResults | Where-Object { $_ -like "*[FAIL]*" }).Count
    $totalCount = $global:testResults.Count
    
    Write-Host "总测试数: $totalCount" -ForegroundColor White
    Write-Host "通过: $passCount" -ForegroundColor Green
    Write-Host "失败: $failCount" -ForegroundColor Red
    
    if ($failCount -eq 0) {
        Write-Host "所有测试通过！" -ForegroundColor Green
        return $true
    } else {
        Write-Host "有 $failCount 个测试失败" -ForegroundColor Red
        return $false
    }
}

# 主程序
try {
    Write-Host "开始综合服务测试..." -ForegroundColor Green
    Write-Host "API网关地址: $BaseUrl" -ForegroundColor Yellow
    Write-Host ""
    
    $result = Test-AllServices
    
    if ($result) {
        Write-Host ""
        Write-Host "综合测试完成，所有服务运行正常！" -ForegroundColor Green
        exit 0
    } else {
        Write-Host ""
        Write-Host "综合测试完成，发现一些问题，请检查服务状态。" -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-Host "测试过程中发生错误: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
} 