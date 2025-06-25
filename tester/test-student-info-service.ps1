# 学生信息服务测试脚本
# 测试学生信息服务的所有API接口

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

function Test-StudentInfoService {
    Write-Host "=== 学生信息服务测试 ===" -ForegroundColor Yellow
    
    $global:testResults = @()
    $global:authToken = $null
    $global:testStudentId = $null
    
    # 生成唯一测试用户名
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $testUsername = "teststudent_$timestamp"
    $testPassword = "password123"
    $testEmail = "$testUsername@example.com"
    
    Info "测试学生用户名: $testUsername"
    
    # 1. 测试用户注册
    Info "1. 测试用户注册"
    $registerBody = @{
        username = $testUsername
        password = $testPassword
        email = $testEmail
        real_name = "测试学生"
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
            name = "测试学生"
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
    
    # 4. 测试获取所有学生列表
    Info "4. 测试获取所有学生列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $studentsResp = Invoke-RestMethod -Uri "$BaseUrl/api/students" -Method GET -Headers $headers
            if ($studentsResp.code -eq 0) {
                $global:testResults += Pass "获取所有学生列表成功"
            } else {
                $global:testResults += Fail "获取所有学生列表失败: $($studentsResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取所有学生列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取所有学生列表测试 - 无有效令牌"
    }
    
    # 5. 测试搜索学生
    Info "5. 测试搜索学生"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $searchResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/search?q=测试学生" -Method GET -Headers $headers
            if ($searchResp.code -eq 0) {
                $global:testResults += Pass "搜索学生成功"
            } else {
                $global:testResults += Fail "搜索学生失败: $($searchResp.message)"
            }
        } catch {
            $global:testResults += Fail "搜索学生异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过搜索学生测试 - 无有效令牌"
    }
    
    # 6. 测试按用户名搜索学生
    Info "6. 测试按用户名搜索学生"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $searchByUsernameResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/search/username?username=$testUsername" -Method GET -Headers $headers
            if ($searchByUsernameResp.code -eq 0) {
                $global:testResults += Pass "按用户名搜索学生成功"
            } else {
                $global:testResults += Fail "按用户名搜索学生失败: $($searchByUsernameResp.message)"
            }
        } catch {
            $global:testResults += Fail "按用户名搜索学生异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按用户名搜索学生测试 - 无有效令牌"
    }
    
    # 7. 测试按学院获取学生列表
    Info "7. 测试按学院获取学生列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $collegeResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/college/计算机学院" -Method GET -Headers $headers
            if ($collegeResp.code -eq 0) {
                $global:testResults += Pass "按学院获取学生列表成功"
            } else {
                $global:testResults += Fail "按学院获取学生列表失败: $($collegeResp.message)"
            }
        } catch {
            $global:testResults += Fail "按学院获取学生列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按学院获取学生列表测试 - 无有效令牌"
    }
    
    # 8. 测试按专业获取学生列表
    Info "8. 测试按专业获取学生列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $majorResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/major/软件工程" -Method GET -Headers $headers
            if ($majorResp.code -eq 0) {
                $global:testResults += Pass "按专业获取学生列表成功"
            } else {
                $global:testResults += Fail "按专业获取学生列表失败: $($majorResp.message)"
            }
        } catch {
            $global:testResults += Fail "按专业获取学生列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按专业获取学生列表测试 - 无有效令牌"
    }
    
    # 9. 测试按班级获取学生列表
    Info "9. 测试按班级获取学生列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $classResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/class/软工2024-1班" -Method GET -Headers $headers
            if ($classResp.code -eq 0) {
                $global:testResults += Pass "按班级获取学生列表成功"
            } else {
                $global:testResults += Fail "按班级获取学生列表失败: $($classResp.message)"
            }
        } catch {
            $global:testResults += Fail "按班级获取学生列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按班级获取学生列表测试 - 无有效令牌"
    }
    
    # 10. 测试按状态获取学生列表
    Info "10. 测试按状态获取学生列表"
    if ($global:authToken) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $statusResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/status/active" -Method GET -Headers $headers
            if ($statusResp.code -eq 0) {
                $global:testResults += Pass "按状态获取学生列表成功"
            } else {
                $global:testResults += Fail "按状态获取学生列表失败: $($statusResp.message)"
            }
        } catch {
            $global:testResults += Fail "按状态获取学生列表异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过按状态获取学生列表测试 - 无有效令牌"
    }
    
    # 11. 测试获取特定学生信息
    Info "11. 测试获取特定学生信息"
    if ($global:authToken -and $global:testStudentId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $studentResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/$global:testStudentId" -Method GET -Headers $headers
            if ($studentResp.code -eq 0) {
                $global:testResults += Pass "获取特定学生信息成功"
            } else {
                $global:testResults += Fail "获取特定学生信息失败: $($studentResp.message)"
            }
        } catch {
            $global:testResults += Fail "获取特定学生信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过获取特定学生信息测试 - 无有效令牌或学生ID"
    }
    
    # 12. 测试更新学生信息
    Info "12. 测试更新学生信息"
    if ($global:authToken -and $global:testStudentId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        $updateStudentBody = @{
            name = "更新后的测试学生"
            phone = "13900139000"
            email = "updated_$testEmail"
        } | ConvertTo-Json
        
        try {
            $updateStudentResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/$global:testStudentId" -Method PUT -Body $updateStudentBody -Headers $headers -ContentType "application/json"
            if ($updateStudentResp.code -eq 0) {
                $global:testResults += Pass "更新学生信息成功"
            } else {
                $global:testResults += Fail "更新学生信息失败: $($updateStudentResp.message)"
            }
        } catch {
            $global:testResults += Fail "更新学生信息异常: $($_.Exception.Message)"
        }
    } else {
        $global:testResults += Fail "跳过更新学生信息测试 - 无有效令牌或学生ID"
    }
    
    # 13. 测试删除学生信息
    Info "13. 测试删除学生信息"
    if ($global:authToken -and $global:testStudentId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
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
    } else {
        $global:testResults += Fail "跳过删除学生信息测试 - 无有效令牌或学生ID"
    }
    
    # 14. 测试删除后获取学生信息（应该失败）
    Info "14. 测试删除后获取学生信息（应该失败）"
    if ($global:authToken -and $global:testStudentId) {
        $headers = @{ "Authorization" = "Bearer $global:authToken" }
        try {
            $deletedStudentResp = Invoke-RestMethod -Uri "$BaseUrl/api/students/$global:testStudentId" -Method GET -Headers $headers
            if ($deletedStudentResp.code -ne 0) {
                $global:testResults += Pass "删除后获取学生信息被正确拒绝"
            } else {
                $global:testResults += Fail "删除后获取学生信息应该被拒绝"
            }
        } catch {
            $global:testResults += Pass "删除后获取学生信息被正确拒绝"
        }
    } else {
        $global:testResults += Fail "跳过删除后获取学生信息测试 - 无有效令牌或学生ID"
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
Test-StudentInfoService 