# 学分活动服务测试脚本
# 测试环境：直接删除所有数据和表格，重建新的

$baseUrl = "http://localhost:8083"
$token = "Bearer test-token"

Write-Host "=== 学分活动服务测试 ===" -ForegroundColor Green
Write-Host "测试环境：直接删除所有数据和表格，重建新的" -ForegroundColor Yellow
Write-Host ""

# 测试1：健康检查
Write-Host "1. 测试健康检查..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "✓ 健康检查通过: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 健康检查失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试2：获取活动类别
Write-Host "2. 测试获取活动类别..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/categories" -Method GET
    Write-Host "✓ 获取活动类别成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取活动类别失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试3：创建活动
Write-Host "3. 测试创建活动..." -ForegroundColor Cyan
$createActivityBody = @{
    title = "测试学术讲座"
    description = "这是一个测试学术讲座"
    start_date = "2024-01-01T09:00:00Z"
    end_date = "2024-01-01T11:00:00Z"
    category = "学术研究"
    requirements = "计算机专业学生优先"
} | ConvertTo-Json

try {
    $headers = @{
        "Authorization" = $token
        "Content-Type" = "application/json"
    }
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities" -Method POST -Body $createActivityBody -Headers $headers
    $activityId = $response.data.id
    Write-Host "✓ 创建活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
    Write-Host "  活动ID: $activityId" -ForegroundColor Yellow
} catch {
    Write-Host "✗ 创建活动失败: $($_.Exception.Message)" -ForegroundColor Red
    $activityId = "test-activity-id"
}
Write-Host ""

# 测试4：获取活动列表
Write-Host "4. 测试获取活动列表..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取活动列表成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取活动列表失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试5：获取活动详情
Write-Host "5. 测试获取活动详情..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取活动详情成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取活动详情失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试6：添加参与者
Write-Host "6. 测试添加参与者..." -ForegroundColor Cyan
$addParticipantsBody = @{
    user_ids = @("user-uuid-1", "user-uuid-2", "user-uuid-3")
    credits = 2.0
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/participants" -Method POST -Body $addParticipantsBody -Headers $headers
    Write-Host "✓ 添加参与者成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 添加参与者失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试7：批量设置学分
Write-Host "7. 测试批量设置学分..." -ForegroundColor Cyan
$batchCreditsBody = @{
    user_credits = @(
        @{user_id = "user-uuid-1"; credits = 2.5},
        @{user_id = "user-uuid-2"; credits = 1.5},
        @{user_id = "user-uuid-3"; credits = 3.0}
    )
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/participants/batch-credits" -Method PUT -Body $batchCreditsBody -Headers $headers
    Write-Host "✓ 批量设置学分成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 批量设置学分失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试8：设置单个参与者学分
Write-Host "8. 测试设置单个参与者学分..." -ForegroundColor Cyan
$singleCreditsBody = @{
    credits = 2.8
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/participants/user-uuid-1/credits" -Method PUT -Body $singleCreditsBody -Headers $headers
    Write-Host "✓ 设置单个参与者学分成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 设置单个参与者学分失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试9：获取参与者列表
Write-Host "9. 测试获取参与者列表..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/participants" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取参与者列表成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取参与者列表失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试10：提交活动审核
Write-Host "10. 测试提交活动审核..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/submit" -Method POST -Headers @{"Authorization" = $token}
    Write-Host "✓ 提交活动审核成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 提交活动审核失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试11：获取待审核活动
Write-Host "11. 测试获取待审核活动..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/pending" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取待审核活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取待审核活动失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试12：审核活动（通过）
Write-Host "12. 测试审核活动（通过）..." -ForegroundColor Cyan
$reviewBody = @{
    status = "approved"
    review_comments = "审核通过，活动内容符合要求"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/review" -Method POST -Body $reviewBody -Headers $headers
    Write-Host "✓ 审核活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 审核活动失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 等待一下让触发器执行
Write-Host "等待触发器执行..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

# 测试13：获取申请列表
Write-Host "13. 测试获取申请列表..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/applications" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取申请列表成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取申请列表失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试14：获取申请统计
Write-Host "14. 测试获取申请统计..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/applications/stats" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取申请统计成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取申请统计失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试15：获取活动统计
Write-Host "15. 测试获取活动统计..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/stats" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取活动统计成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取活动统计失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试16：获取所有申请（教师/管理员）
Write-Host "16. 测试获取所有申请（教师/管理员）..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/applications/all" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 获取所有申请成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 获取所有申请失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试17：退出活动
Write-Host "17. 测试退出活动..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/leave" -Method POST -Headers @{"Authorization" = $token}
    Write-Host "✓ 退出活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 退出活动失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试18：删除参与者
Write-Host "18. 测试删除参与者..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/participants/user-uuid-2" -Method DELETE -Headers @{"Authorization" = $token}
    Write-Host "✓ 删除参与者成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 删除参与者失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试19：更新活动
Write-Host "19. 测试更新活动..." -ForegroundColor Cyan
$updateActivityBody = @{
    title = "更新后的测试学术讲座"
    description = "这是更新后的测试学术讲座描述"
    category = "学术研究"
    requirements = "更新后的要求"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId" -Method PUT -Body $updateActivityBody -Headers $headers
    Write-Host "✓ 更新活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 更新活动失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试20：创建第二个活动并拒绝
Write-Host "20. 测试创建第二个活动并拒绝..." -ForegroundColor Cyan
$createActivity2Body = @{
    title = "测试活动2"
    description = "这是第二个测试活动"
    start_date = "2024-01-02T09:00:00Z"
    end_date = "2024-01-02T11:00:00Z"
    category = "学科竞赛"
    requirements = "所有学生都可以参加"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities" -Method POST -Body $createActivity2Body -Headers $headers
    $activity2Id = $response.data.id
    Write-Host "✓ 创建第二个活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
    
    # 添加参与者
    $addParticipants2Body = @{
        user_ids = @("user-uuid-4", "user-uuid-5")
        credits = 1.5
    } | ConvertTo-Json
    Invoke-RestMethod -Uri "$baseUrl/api/activities/$activity2Id/participants" -Method POST -Body $addParticipants2Body -Headers $headers
    
    # 提交审核
    Invoke-RestMethod -Uri "$baseUrl/api/activities/$activity2Id/submit" -Method POST -Headers @{"Authorization" = $token}
    
    # 拒绝活动
    $rejectBody = @{
        status = "rejected"
        review_comments = "活动内容不符合要求，请修改后重新提交"
    } | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activity2Id/review" -Method POST -Body $rejectBody -Headers $headers
    Write-Host "✓ 拒绝活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 创建第二个活动并拒绝失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试21：测试撤回活动
Write-Host "21. 测试撤回活动..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/activities/$activityId/withdraw" -Method POST -Headers @{"Authorization" = $token}
    Write-Host "✓ 撤回活动成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 撤回活动失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 测试22：导出申请数据
Write-Host "22. 测试导出申请数据..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/applications/export?format=csv" -Method GET -Headers @{"Authorization" = $token}
    Write-Host "✓ 导出申请数据成功: $($response | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "✗ 导出申请数据失败: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "=== 测试完成 ===" -ForegroundColor Green
Write-Host "所有测试已执行完毕，请检查上述结果。" -ForegroundColor Yellow 