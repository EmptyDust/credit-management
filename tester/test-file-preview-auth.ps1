# 文件预览认证功能测试脚本
# 使用方法: .\test-file-preview-auth.ps1

Write-Host "=== 文件预览认证功能测试 ===" -ForegroundColor Green

# 配置
$API_BASE_URL = "http://localhost:8080"
$TEST_ACTIVITY_ID = "eb69645c-9c60-4148-892c-492d131a4112"  # 从错误日志中获取的活动ID
$TEST_ATTACHMENT_ID = "2ad58636-6150-4786-9628-eb44d27d63f1"  # 从错误日志中获取的附件ID

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

# 测试2: 测试无认证的预览请求（应该返回401）
Write-Host "`n2. 测试无认证的预览请求..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$API_BASE_URL/api/activities/$TEST_ACTIVITY_ID/attachments/$TEST_ATTACHMENT_ID/preview" -Method GET
    Write-Host "✗ 意外成功，应该返回401" -ForegroundColor Red
}
catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✓ 正确返回401未认证错误" -ForegroundColor Green
    }
    else {
        Write-Host "✗ 返回了意外的状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

# 测试3: 测试带token参数的预览请求
Write-Host "`n3. 测试带token参数的预览请求..." -ForegroundColor Yellow
Write-Host "请提供有效的JWT token:" -ForegroundColor Cyan
$token = Read-Host

if ($token) {
    try {
        $response = Invoke-WebRequest -Uri "$API_BASE_URL/api/activities/$TEST_ACTIVITY_ID/attachments/$TEST_ATTACHMENT_ID/preview?token=$token" -Method GET
        Write-Host "✓ 带token的预览请求成功，状态码: $($response.StatusCode)" -ForegroundColor Green
        Write-Host "  内容类型: $($response.Headers.'Content-Type')" -ForegroundColor Cyan
        Write-Host "  文件大小: $($response.Content.Length) 字节" -ForegroundColor Cyan
    }
    catch {
        Write-Host "✗ 带token的预览请求失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            Write-Host "  状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
            Write-Host "  错误信息: $($_.Exception.Response.StatusDescription)" -ForegroundColor Red
        }
    }
}
else {
    Write-Host "跳过token测试" -ForegroundColor Yellow
}

# 测试4: 测试带Authorization头的预览请求
Write-Host "`n4. 测试带Authorization头的预览请求..." -ForegroundColor Yellow
if ($token) {
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
        }
        $response = Invoke-WebRequest -Uri "$API_BASE_URL/api/activities/$TEST_ACTIVITY_ID/attachments/$TEST_ATTACHMENT_ID/preview" -Method GET -Headers $headers
        Write-Host "✓ 带Authorization头的预览请求成功，状态码: $($response.StatusCode)" -ForegroundColor Green
        Write-Host "  内容类型: $($response.Headers.'Content-Type')" -ForegroundColor Cyan
        Write-Host "  文件大小: $($response.Content.Length) 字节" -ForegroundColor Cyan
    }
    catch {
        Write-Host "✗ 带Authorization头的预览请求失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            Write-Host "  状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
            Write-Host "  错误信息: $($_.Exception.Response.StatusDescription)" -ForegroundColor Red
        }
    }
}
else {
    Write-Host "跳过Authorization头测试" -ForegroundColor Yellow
}

# 测试5: 检查Docker容器日志
Write-Host "`n5. 检查Docker容器日志..." -ForegroundColor Yellow
try {
    Write-Host "API网关日志（最近10行）:" -ForegroundColor Cyan
    docker logs credit_management_gateway --tail 10 2>$null
    
    Write-Host "`nCredit Activity Service日志（最近10行）:" -ForegroundColor Cyan
    docker logs credit_management_credit_activity --tail 10 2>$null
}
catch {
    Write-Host "✗ 检查容器日志失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试6: 检查文件存储
Write-Host "`n6. 检查文件存储..." -ForegroundColor Yellow
try {
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
    Write-Host "✗ 检查文件存储失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green
Write-Host "如果带token的请求仍然失败，请检查:" -ForegroundColor Yellow
Write-Host "1. Token是否有效且未过期" -ForegroundColor Yellow
Write-Host "2. 用户是否有权限访问该活动的附件" -ForegroundColor Yellow
Write-Host "3. 活动ID和附件ID是否正确" -ForegroundColor Yellow 