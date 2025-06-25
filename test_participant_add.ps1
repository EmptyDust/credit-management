# 参与者添加测试脚本
Write-Host "=== 参与者添加测试 ===" -ForegroundColor Green

# 配置
$API_GATEWAY = "http://localhost:8080"
$CREDIT_SERVICE = "http://localhost:8083"

# 登录获取token
Write-Host "`n1. 登录获取token..." -ForegroundColor Yellow
$loginData = @{
    username = "student"
    password = "adminpassword"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_GATEWAY/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    if ($response.code -eq 0) {
        $studentToken = $response.data.token
        Write-Host "✓ 学生登录成功" -ForegroundColor Green
    }
    else {
        Write-Host "✗ 学生登录失败: $($response.message)" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "✗ 登录请求失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 创建活动
Write-Host "`n2. 创建测试活动..." -ForegroundColor Yellow
$activityData = @{
    title        = "参与者测试活动"
    description  = "这是一个用于测试参与者添加功能的活动"
    start_date   = "2024-12-01"
    end_date     = "2024-12-31"
    category     = "学科竞赛"
    requirements = "需要提交竞赛证书"
} | ConvertTo-Json

try {
    $headers = @{
        "Authorization" = "Bearer $studentToken"
        "Content-Type"  = "application/json"
    }
    
    $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities" -Method POST -Body $activityData -Headers $headers
    if ($response.code -eq 0) {
        $activityId = $response.data.id
        Write-Host "✓ 活动创建成功: $($response.data.title), ID: $activityId" -ForegroundColor Green
    }
    else {
        Write-Host "✗ 活动创建失败: $($response.message)" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "✗ 活动创建请求失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 获取学生用户ID
Write-Host "`n3. 获取学生用户ID..." -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer $studentToken"
    }
    
    $response = Invoke-RestMethod -Uri "$API_GATEWAY/api/search/users?query=student&user_type=student&limit=1" -Method GET -Headers $headers
    if ($response.code -eq 0 -and $response.data.users.Count -gt 0) {
        $studentUserId = $response.data.users[0].user_id
        Write-Host "✓ 获取学生用户ID成功: $studentUserId" -ForegroundColor Green
        Write-Host "  用户信息: $($response.data.users[0] | ConvertTo-Json)" -ForegroundColor Cyan
    }
    else {
        Write-Host "✗ 获取学生用户失败: $($response.message)" -ForegroundColor Red
        Write-Host "  响应详情: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Yellow
        exit 1
    }
}
catch {
    Write-Host "✗ 获取学生用户请求失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 添加参与者
Write-Host "`n4. 添加参与者..." -ForegroundColor Yellow
$participantData = @{
    user_ids = @($studentUserId)
    credits  = 2.0
} | ConvertTo-Json

try {
    $headers = @{
        "Authorization" = "Bearer $studentToken"
        "Content-Type"  = "application/json"
    }
    
    $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/participants" -Method POST -Body $participantData -Headers $headers
    if ($response.code -eq 0) {
        Write-Host "✓ 参与者添加成功: $($response.data.added_count) 人" -ForegroundColor Green
        Write-Host "  参与者详情: $($response.data.participants | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    }
    else {
        Write-Host "✗ 参与者添加失败: $($response.message)" -ForegroundColor Red
    }
}
catch {
    Write-Host "✗ 参与者添加请求失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green 