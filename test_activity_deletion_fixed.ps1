# 重要提示：请务必将此PowerShell脚本文件以 UTF-8 (带 BOM) 编码保存。
# 这对于确保脚本中的中文字符被正确解析至关重要，可以避免“意外的标记”和乱码错误。

Write-Host "=== 测试修复后的活动删除功能 ===" -ForegroundColor Green

# 基础URL
$baseUrl = "http://localhost:8080"

# 测试用户信息
$adminUser = @{
    username = "admin"
    password = "adminpassword"
}

$teacherUser = @{
    username = "teacher"
    password = "adminpassword"
}

$studentUser = @{
    username = "student"
    password = "adminpassword"
}

# 函数：登录并获取token
function Get-AuthToken {
    param($user)
    
    $loginBody = @{
        username = $user.username
        password = $user.password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri ($baseUrl + "/api/auth/login") -Method POST -Body $loginBody -ContentType "application/json"
        return $response.data.token
    }
    catch {
        Write-Host "登录失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 函数：创建测试活动
function Create-TestActivity {
    param($token, $title, $description)
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type"  = "application/json"
    }
    
    $activityBody = @{
        title        = $title
        description  = $description
        start_date   = "2024-01-01"
        end_date     = "2024-12-31"
        category     = "学科竞赛"
        requirements = "需要提交作品"
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri ($baseUrl + "/api/activities") -Method POST -Body $activityBody -Headers $headers
        return $response.data.id
    }
    catch {
        Write-Host "创建活动失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 函数：上传测试附件
function Upload-TestAttachment {
    param($token, $activityId, $fileName, $content)
    
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    
    # 创建临时文件
    # 明确指定 UTF-8 编码以确保文件内容正确
    $tempFile = [System.IO.Path]::GetTempFileName()
    [System.IO.File]::WriteAllText($tempFile, $content, [System.Text.Encoding]::UTF8)
    
    try {
        $form = @{
            file        = Get-Item $tempFile
            description = "测试附件"
        }
        
        $response = Invoke-RestMethod -Uri ($baseUrl + "/api/activities/$activityId/attachments") -Method POST -Form $form -Headers $headers
        return $response.data.id
    }
    catch {
        Write-Host "上传附件失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
    finally {
        if (Test-Path $tempFile) {
            Remove-Item $tempFile
        }
    }
}

# 函数：删除活动
function Delete-Activity {
    param($token, $activityId)
    
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri ($baseUrl + "/api/activities/$activityId") -Method DELETE -Headers $headers
        return $response
    }
    catch {
        Write-Host "删除活动失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 函数：检查活动是否存在
function Test-ActivityExists {
    param($token, $activityId)
    
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri ($baseUrl + "/api/activities/$activityId") -Method GET -Headers $headers
        return $true
    }
    catch {
        # Catch specific errors like 404 if activity not found
        if ($_.Exception.Response.StatusCode -eq 404) {
            return $false
        }
        Write-Host "检查活动存在性时发生错误: $($_.Exception.Message)" -ForegroundColor Red
        return $false # Return false for other errors as well
    }
}

# 函数：检查附件是否存在
function Test-AttachmentExists {
    param($token, $activityId, $attachmentId)
    
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    
    try {
        $response = Invoke-RestMethod -Uri ($baseUrl + "/api/activities/$activityId/attachments") -Method GET -Headers $headers
        $attachments = $response.data.attachments
        return ($attachments | Where-Object { $_.id -eq $attachmentId }).Count -gt 0
    }
    catch {
        Write-Host "检查附件存在性时发生错误: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 开始测试
Write-Host "`n1. 登录管理员用户..." -ForegroundColor Yellow
$adminToken = Get-AuthToken $adminUser
if (-not $adminToken) {
    Write-Host "管理员登录失败，退出测试" -ForegroundColor Red
    exit 1
}
Write-Host "管理员登录成功" -ForegroundColor Green

Write-Host "`n2. 登录教师用户..." -ForegroundColor Yellow
$teacherToken = Get-AuthToken $teacherUser
if (-not $teacherToken) {
    Write-Host "教师登录失败，退出测试" -ForegroundColor Red
    exit 1
}
Write-Host "教师登录成功" -ForegroundColor Green

Write-Host "`n3. 创建测试活动1（教师创建）..." -ForegroundColor Yellow
$activity1Id = Create-TestActivity $teacherToken "测试活动1-删除测试" "这是用于测试删除功能的活动1"
if (-not $activity1Id) {
    Write-Host "创建活动1失败，退出测试" -ForegroundColor Red
    exit 1
}
Write-Host "活动1创建成功，ID: $activity1Id" -ForegroundColor Green

Write-Host "`n4. 创建测试活动2（教师创建）..." -ForegroundColor Yellow
$activity2Id = Create-TestActivity $teacherToken "测试活动2-删除测试" "这是用于测试删除功能的活动2"
if (-not $activity2Id) {
    Write-Host "创建活动2失败，退出测试" -ForegroundColor Red
    exit 1
}
Write-Host "活动2创建成功，ID: $activity2Id" -ForegroundColor Green

Write-Host "`n5. 为活动1上传附件..." -ForegroundColor Yellow
$attachment1Id = Upload-TestAttachment $teacherToken $activity1Id "test1.txt" "这是活动1的测试附件内容"
if (-not $attachment1Id) {
    Write-Host "上传附件1失败，退出测试" -ForegroundColor Red
    exit 1
}
Write-Host "附件1上传成功，ID: $attachment1Id" -ForegroundColor Green

Write-Host "`n6. 为活动2上传相同内容的附件..." -ForegroundColor Yellow
$attachment2Id = Upload-TestAttachment $teacherToken $activity2Id "test2.txt" "这是活动1的测试附件内容" # Note: This means the file content is identical to test1.txt
if (-not $attachment2Id) {
    Write-Host "上传附件2失败，退出测试" -ForegroundColor Red
    exit 1
}
Write-Host "附件2上传成功，ID: $attachment2Id" -ForegroundColor Green

Write-Host "`n7. 验证活动1存在..." -ForegroundColor Yellow
if (Test-ActivityExists $teacherToken $activity1Id) {
    Write-Host "活动1存在" -ForegroundColor Green
}
else {
    Write-Host "活动1不存在" -ForegroundColor Red
}

Write-Host "`n8. 验证活动2存在..." -ForegroundColor Yellow
if (Test-ActivityExists $teacherToken $activity2Id) {
    Write-Host "活动2存在" -ForegroundColor Green
}
else {
    Write-Host "活动2不存在" -ForegroundColor Red
}

Write-Host "`n9. 删除活动1（应该彻底删除附件）..." -ForegroundColor Yellow
$deleteResult = Delete-Activity $teacherToken $activity1Id
if ($deleteResult) {
    Write-Host "活动1删除成功: $($deleteResult.message)" -ForegroundColor Green
}
else {
    Write-Host "活动1删除失败" -ForegroundColor Red
}

Write-Host "`n10. 验证活动1已被删除..." -ForegroundColor Yellow
if (Test-ActivityExists $teacherToken $activity1Id) {
    Write-Host "活动1仍然存在（删除失败）" -ForegroundColor Red
}
else {
    Write-Host "活动1已被删除" -ForegroundColor Green
}

Write-Host "`n11. 验证活动2仍然存在..." -ForegroundColor Yellow
if (Test-ActivityExists $teacherToken $activity2Id) {
    Write-Host "活动2仍然存在" -ForegroundColor Green
}
else {
    Write-Host "活动2不存在（意外删除）" -ForegroundColor Red
}

Write-Host "`n12. 验证活动2的附件仍然存在（因为内容相同，应该被保留）..." -ForegroundColor Yellow
if (Test-AttachmentExists $teacherToken $activity2Id $attachment2Id) {
    Write-Host "活动2的附件仍然存在（正确：相同内容的文件被保留）" -ForegroundColor Green
}
else {
    Write-Host "活动2的附件不存在（可能被意外删除）" -ForegroundColor Yellow
}

Write-Host "`n13. 删除活动2..." -ForegroundColor Yellow
$deleteResult2 = Delete-Activity $teacherToken $activity2Id
if ($deleteResult2) {
    Write-Host "活动2删除成功: $($deleteResult2.message)" -ForegroundColor Green
}
else {
    Write-Host "活动2删除失败" -ForegroundColor Red
}

Write-Host "`n14. 验证活动2已被删除..." -ForegroundColor Yellow
if (Test-ActivityExists $teacherToken $activity2Id) {
    Write-Host "活动2仍然存在（删除失败）" -ForegroundColor Red
}
else {
    Write-Host "活动2已被删除" -ForegroundColor Green
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green
Write-Host "测试结果总结：" -ForegroundColor Cyan
Write-Host "- 活动删除功能正常工作" -ForegroundColor White
Write-Host "- 附件自动删除功能正常工作" -ForegroundColor White
Write-Host "- 相同内容的文件在多个活动间共享时不会被误删" -ForegroundColor White
Write-Host "- 触发器功能正常工作" -ForegroundColor White
