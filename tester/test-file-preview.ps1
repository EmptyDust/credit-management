# 文件预览功能测试脚本
# 使用方法: .\test-file-preview.ps1

Write-Host "=== 文件预览功能测试 ===" -ForegroundColor Green

# 配置
$API_BASE_URL = "http://localhost:8080"
$TEST_ACTIVITY_ID = "your-activity-id"  # 请替换为实际的活动ID
$TEST_ATTACHMENT_ID = "your-attachment-id"  # 请替换为实际的附件ID

# 测试1: 检查API网关健康状态
Write-Host "`n1. 测试API网关健康状态..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/health" -Method GET
    Write-Host "✓ API网关正常: $($response.status)" -ForegroundColor Green
}
catch {
    Write-Host "✗ API网关异常: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 测试2: 检查活动附件列表
Write-Host "`n2. 测试获取活动附件列表..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/activities/$TEST_ACTIVITY_ID/attachments" -Method GET
    Write-Host "✓ 附件列表获取成功，共 $($response.data.attachments.Count) 个附件" -ForegroundColor Green
    
    # 显示附件信息
    foreach ($attachment in $response.data.attachments) {
        Write-Host "  - $($attachment.original_name) ($($attachment.file_category), $($attachment.file_type))" -ForegroundColor Cyan
    }
}
catch {
    Write-Host "✗ 获取附件列表失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试3: 测试文件下载功能
Write-Host "`n3. 测试文件下载功能..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$API_BASE_URL/api/activities/$TEST_ACTIVITY_ID/attachments/$TEST_ATTACHMENT_ID/download" -Method GET
    Write-Host "✓ 文件下载成功，状态码: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "  内容类型: $($response.Headers.'Content-Type')" -ForegroundColor Cyan
    Write-Host "  文件大小: $($response.Content.Length) 字节" -ForegroundColor Cyan
}
catch {
    Write-Host "✗ 文件下载失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试4: 测试文件预览功能
Write-Host "`n4. 测试文件预览功能..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$API_BASE_URL/api/activities/$TEST_ACTIVITY_ID/attachments/$TEST_ATTACHMENT_ID/preview" -Method GET
    Write-Host "✓ 文件预览成功，状态码: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "  内容类型: $($response.Headers.'Content-Type')" -ForegroundColor Cyan
    Write-Host "  文件大小: $($response.Content.Length) 字节" -ForegroundColor Cyan
    
    # 检查响应头
    Write-Host "  响应头信息:" -ForegroundColor Cyan
    foreach ($header in $response.Headers.Keys) {
        Write-Host "    $header`: $($response.Headers[$header])" -ForegroundColor Gray
    }
}
catch {
    Write-Host "✗ 文件预览失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  错误详情: $($_.Exception.Response.StatusCode) - $($_.Exception.Response.StatusDescription)" -ForegroundColor Red
}

# 测试5: 检查Docker容器状态
Write-Host "`n5. 检查Docker容器状态..." -ForegroundColor Yellow
try {
    $containers = docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    Write-Host "当前运行的容器:" -ForegroundColor Cyan
    Write-Host $containers -ForegroundColor Gray
    
    # 检查credit-activity-service容器
    $creditService = docker ps --filter "name=credit_management_credit_activity" --format "{{.Names}}"
    if ($creditService) {
        Write-Host "✓ credit-activity-service 容器正在运行" -ForegroundColor Green
    }
    else {
        Write-Host "✗ credit-activity-service 容器未运行" -ForegroundColor Red
    }
}
catch {
    Write-Host "✗ 检查Docker容器失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试6: 检查文件存储目录
Write-Host "`n6. 检查文件存储目录..." -ForegroundColor Yellow
try {
    # 进入credit-activity-service容器检查文件
    $containerFiles = docker exec credit_management_credit_activity ls -la /app/uploads/attachments 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 文件存储目录存在" -ForegroundColor Green
        Write-Host "目录内容:" -ForegroundColor Cyan
        Write-Host $containerFiles -ForegroundColor Gray
    }
    else {
        Write-Host "✗ 文件存储目录不存在或为空" -ForegroundColor Red
    }
}
catch {
    Write-Host "✗ 检查文件存储目录失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green
Write-Host "请根据测试结果分析问题所在" -ForegroundColor Yellow 