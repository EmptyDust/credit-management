# 附件功能自动化测试脚本
Write-Host "=== 附件功能自动化测试 ===" -ForegroundColor Green

$API_GATEWAY = "http://localhost:8080"
$CREDIT_SERVICE = "http://localhost:8083"

# 登录获取token
function Get-Token {
    param($username, $password)
    $loginData = @{ username = $username; password = $password } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "$API_GATEWAY/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        if ($response.code -eq 0) { return $response.data.token }
        else { Write-Host "登录失败: $($response.message)" -ForegroundColor Red; return $null }
    }
    catch { Write-Host "登录请求失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

# 创建活动
function Create-Activity {
    param($token)
    $activityData = @{ title = "附件测试活动"; description = "测试附件功能"; start_date = "2024-12-01"; end_date = "2024-12-31"; category = "学科竞赛"; requirements = "无" } | ConvertTo-Json
    try {
        $headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities" -Method POST -Body $activityData -Headers $headers
        if ($response.code -eq 0) { Write-Host "✓ 活动创建成功: $($response.data.title)" -ForegroundColor Green; return $response.data.id }
        else { Write-Host "✗ 活动创建失败: $($response.message)" -ForegroundColor Red; return $null }
    }
    catch { Write-Host "✗ 活动创建请求失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

# 上传附件
function Upload-Attachment {
    param($token, $activityId, $filePath)
    try {
        $headers = @{ "Authorization" = "Bearer $token" }
        $form = @{ file = Get-Item $filePath; description = "测试文件描述" }
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/attachments" -Method Post -Form $form -Headers $headers
        if ($response.code -eq 0) { Write-Host "✓ 附件上传成功: $($response.data.original_name)" -ForegroundColor Green; return $response.data.id }
        else { Write-Host "✗ 附件上传失败: $($response.message)" -ForegroundColor Red; return $null }
    }
    catch { Write-Host "✗ 附件上传请求失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

# 获取附件列表
function Get-Attachments {
    param($token, $activityId)
    try {
        $headers = @{ "Authorization" = "Bearer $token" }
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/attachments" -Method Get -Headers $headers
        if ($response.code -eq 0) {
            Write-Host "✓ 获取附件列表成功, 共 $($response.data.attachments.Count) 个" -ForegroundColor Green
            return $response.data.attachments
        }
        else { Write-Host "✗ 获取附件列表失败: $($response.message)" -ForegroundColor Red; return $null }
    }
    catch { Write-Host "✗ 获取附件列表请求失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

# 下载附件
function Download-Attachment {
    param($token, $activityId, $attachmentId, $savePath)
    try {
        $headers = @{ "Authorization" = "Bearer $token" }
        $url = "$CREDIT_SERVICE/api/activities/$activityId/attachments/$attachmentId/download"
        Invoke-WebRequest -Uri $url -Headers $headers -OutFile $savePath
        Write-Host "✓ 附件下载成功: $savePath" -ForegroundColor Green
    }
    catch { Write-Host "✗ 附件下载请求失败: $($_.Exception.Message)" -ForegroundColor Red }
}

# 更新附件描述
function Update-Attachment {
    param($token, $activityId, $attachmentId, $desc)
    try {
        $headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
        $body = @{ description = $desc } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/attachments/$attachmentId" -Method Put -Body $body -Headers $headers
        if ($response.code -eq 0) { Write-Host "✓ 附件描述更新成功" -ForegroundColor Green }
        else { Write-Host "✗ 附件描述更新失败: $($response.message)" -ForegroundColor Red }
    }
    catch { Write-Host "✗ 附件描述更新请求失败: $($_.Exception.Message)" -ForegroundColor Red }
}

# 删除附件
function Delete-Attachment {
    param($token, $activityId, $attachmentId)
    try {
        $headers = @{ "Authorization" = "Bearer $token" }
        $response = Invoke-RestMethod -Uri "$CREDIT_SERVICE/api/activities/$activityId/attachments/$attachmentId" -Method Delete -Headers $headers
        if ($response.code -eq 0) { Write-Host "✓ 附件删除成功" -ForegroundColor Green }
        else { Write-Host "✗ 附件删除失败: $($response.message)" -ForegroundColor Red }
    }
    catch { Write-Host "✗ 附件删除请求失败: $($_.Exception.Message)" -ForegroundColor Red }
}

# 主流程
Write-Host "1. 登录获取token..." -ForegroundColor Yellow
$studentToken = Get-Token "student" "adminpassword"
if (-not $studentToken) { exit 1 }

Write-Host "2. 创建测试活动..." -ForegroundColor Yellow
$activityId = Create-Activity $studentToken
if (-not $activityId) { exit 1 }

Write-Host "3. 上传附件..." -ForegroundColor Yellow
# 创建一个临时测试文件
$testFile = "test_upload.txt"
"Hello Attachment Test!" | Out-File $testFile -Encoding utf8
$attachmentId = Upload-Attachment $studentToken $activityId $testFile
if (-not $attachmentId) { Remove-Item $testFile -ErrorAction SilentlyContinue; exit 1 }

Write-Host "4. 获取附件列表..." -ForegroundColor Yellow
$attachments = Get-Attachments $studentToken $activityId
if (-not $attachments) { Remove-Item $testFile -ErrorAction SilentlyContinue; exit 1 }

Write-Host "5. 下载附件..." -ForegroundColor Yellow
$downloadPath = "downloaded_test_upload.txt"
Download-Attachment $studentToken $activityId $attachmentId $downloadPath

Write-Host "6. 更新附件描述..." -ForegroundColor Yellow
Update-Attachment $studentToken $activityId $attachmentId "新描述内容"

Write-Host "7. 删除附件..." -ForegroundColor Yellow
Delete-Attachment $studentToken $activityId $attachmentId

# 清理临时文件
Remove-Item $testFile -ErrorAction SilentlyContinue
Remove-Item $downloadPath -ErrorAction SilentlyContinue

Write-Host "=== 附件功能测试完成 ===" -ForegroundColor Green 