# 活动审批和撤回测试脚本
Write-Host "=== 活动审批和撤回测试 ===" -ForegroundColor Green

# 配置
$API_GATEWAY = "http://localhost:8080"
$CREDIT_SERVICE = "http://localhost:8083"

# 存储token的变量
$adminToken = ""
$teacherToken = ""
$studentToken = ""
$activityId = ""

# 函数：登录获取token
function Get-Token {
    param($username, $password)
    
    $loginData = @{
        username = $username
        password = $password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$API_GATEWAY/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        if ($response.code -eq 0) {
            return $response.data.token
        }
        else {
            Write-Host "登录失败: $($response.message)" -ForegroundColor Red
            return $null
        }
    }
    catch {
        Write-Host "登录请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 函数：创建活动
function Create-Activity {
    param($token)
    
    $activityData = @{
        title        = "测试审批活动"
        description  = "这是一个用于测试审批和撤回功能的活动"
        start_date   = "2024-12-01"
        end_date     = "2024-12-31"
        category     = "学科竞赛"
        requirements = "需要提交竞赛证书"
    } | ConvertTo-Json
    
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
            "Content-Type"  = "application/json"
        }
        
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities" -Method POST -Body $activityData -Headers $headers
        if ($response.code -eq 0) {
            Write-Host "✓ 活动创建成功: $($response.data.title)" -ForegroundColor Green
            return $response.data.id
        }
        else {
            Write-Host "✗ 活动创建失败: $($response.message)" -ForegroundColor Red
            return $null
        }
    }
    catch {
        Write-Host "✗ 活动创建请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 函数：获取学生用户ID（与test_participant_add.ps1一致）
function Get-StudentUserId {
    param($token)
    try {
        $headers = @{ "Authorization" = "Bearer $token" }
        $response = Invoke-RestMethod -Uri "$API_GATEWAY/api/search/users?query=student&user_type=student&limit=1" -Method GET -Headers $headers
        if ($response.code -eq 0 -and $response.data.users.Count -gt 0) {
            Write-Host "✓ 获取学生用户ID成功: $($response.data.users[0].user_id)" -ForegroundColor Green
            Write-Host "  用户信息: $($response.data.users[0] | ConvertTo-Json)" -ForegroundColor Cyan
            return $response.data.users[0].user_id
        }
        else {
            Write-Host "✗ 获取学生用户失败: $($response.message)" -ForegroundColor Red
            Write-Host "  响应详情: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Yellow
            return $null
        }
    }
    catch {
        Write-Host "✗ 获取学生用户请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 函数：添加参与者（与test_participant_add.ps1一致）
function Add-Participant {
    param($token, $activityId, $studentId)
    $participantData = @{
        user_ids = @($studentId)
        credits  = 2.0
    } | ConvertTo-Json
    try {
        $headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/participants" -Method POST -Body $participantData -Headers $headers
        if ($response.code -eq 0) {
            Write-Host "✓ 参与者添加成功: $($response.data.added_count) 人" -ForegroundColor Green
            Write-Host "  参与者详情: $($response.data.participants | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
            return $true
        }
        else {
            Write-Host "✗ 参与者添加失败: $($response.message)" -ForegroundColor Red
            return $false
        }
    }
    catch {
        Write-Host "✗ 参与者添加请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 函数：提交活动审核
function Submit-Activity {
    param($token, $activityId)
    
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/submit" -Method POST -Headers $headers
        if ($response.code -eq 0) {
            Write-Host "✓ 活动提交审核成功" -ForegroundColor Green
            return $true
        }
        else {
            Write-Host "✗ 活动提交审核失败: $($response.message)" -ForegroundColor Red
            return $false
        }
    }
    catch {
        Write-Host "✗ 活动提交审核请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 函数：审批活动
function Approve-Activity {
    param($token, $activityId)
    
    $reviewData = @{
        status          = "approved"
        review_comments = "活动内容符合要求，同意通过"
    } | ConvertTo-Json
    
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
            "Content-Type"  = "application/json"
        }
        
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/review" -Method POST -Body $reviewData -Headers $headers
        if ($response.code -eq 0) {
            Write-Host "✓ 活动审批成功" -ForegroundColor Green
            return $true
        }
        else {
            Write-Host "✗ 活动审批失败: $($response.message)" -ForegroundColor Red
            return $false
        }
    }
    catch {
        Write-Host "✗ 活动审批请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 函数：检查申请
function Check-Applications {
    param($token, $activityId)
    
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/applications" -Method GET -Headers $headers
        if ($response.code -eq 0) {
            $applications = $response.data.data | Where-Object { $_.activity_id -eq $activityId }
            if ($applications.Count -gt 0) {
                Write-Host "✓ 找到 $($applications.Count) 个申请" -ForegroundColor Green
                foreach ($app in $applications) {
                    Write-Host "  - 申请ID: $($app.id), 状态: $($app.status), 学分: $($app.awarded_credits)" -ForegroundColor Cyan
                }
                return $true
            }
            else {
                Write-Host "✗ 未找到相关申请" -ForegroundColor Red
                return $false
            }
        }
        else {
            Write-Host "✗ 获取申请列表失败: $($response.message)" -ForegroundColor Red
            return $false
        }
    }
    catch {
        Write-Host "✗ 获取申请列表请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 函数：撤回活动
function Withdraw-Activity {
    param($token, $activityId)
    
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/withdraw" -Method POST -Headers $headers
        if ($response.code -eq 0) {
            Write-Host "✓ 活动撤回成功" -ForegroundColor Green
            return $true
        }
        else {
            Write-Host "✗ 活动撤回失败: $($response.message)" -ForegroundColor Red
            return $false
        }
    }
    catch {
        Write-Host "✗ 活动撤回请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 主测试流程
Write-Host "`n1. 登录获取token..." -ForegroundColor Yellow
$adminToken = Get-Token "admin" "adminpassword"
$teacherToken = Get-Token "teacher" "adminpassword"
$studentToken = Get-Token "student" "adminpassword"

if (-not $adminToken -or -not $teacherToken -or -not $studentToken) {
    Write-Host "✗ 登录失败，无法继续测试" -ForegroundColor Red
    exit 1
}

Write-Host "✓ 所有用户登录成功" -ForegroundColor Green

Write-Host "`n2. 创建测试活动..." -ForegroundColor Yellow
$activityId = Create-Activity $studentToken
if (-not $activityId) {
    Write-Host "✗ 活动创建失败，无法继续测试" -ForegroundColor Red
    exit 1
}

Write-Host "`n3. 获取学生用户ID..." -ForegroundColor Yellow
$studentUserId = Get-StudentUserId $studentToken
if ($studentUserId) {
    Write-Host "`n4. 添加参与者..." -ForegroundColor Yellow
    Add-Participant $studentToken $activityId $studentUserId
}
else {
    Write-Host "⚠ 无法获取学生用户ID，跳过参与者添加" -ForegroundColor Yellow
}

Write-Host "`n5. 提交活动审核..." -ForegroundColor Yellow
if (Submit-Activity $studentToken $activityId) {
    Write-Host "`n6. 教师审批活动..." -ForegroundColor Yellow
    if (Approve-Activity $teacherToken $activityId) {
        Write-Host "`n7. 检查申请是否生成..." -ForegroundColor Yellow
        Start-Sleep -Seconds 2  # 等待触发器执行
        Check-Applications $studentToken $activityId
        
        Write-Host "`n8. 学生撤回活动..." -ForegroundColor Yellow
        if (Withdraw-Activity $studentToken $activityId) {
            Write-Host "`n9. 检查申请是否删除..." -ForegroundColor Yellow
            Start-Sleep -Seconds 2  # 等待触发器执行
            Check-Applications $studentToken $activityId
        }
    }
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green 