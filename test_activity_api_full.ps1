# 活动API全流程自动化测试脚本

$baseUrl = "http://localhost:8080"
$teacherUser = @{ username = "teacher"; password = "adminpassword" }
$adminUser = @{ username = "admin"; password = "adminpassword" }

function Get-AuthToken {
    param($user)
    $loginBody = @{ username = $user.username; password = $user.password } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        return $response.data.token
    }
    catch { Write-Host "登录失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

function Create-Activity {
    param($token, $title)
    $headers = @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" }
    $body = @{ title = $title; description = "desc $title"; start_date = "2024-01-01"; end_date = "2024-12-31"; category = "学科竞赛"; requirements = "无" } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities" -Method POST -Body $body -Headers $headers
        return $response.data.id
    }
    catch { Write-Host "创建活动失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

function Get-Activities {
    param($token)
    $headers = @{ Authorization = "Bearer $token" }
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities" -Headers $headers
        return $response.data.data
    }
    catch { Write-Host "获取活动列表失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

function Get-Activity {
    param($token, $id)
    $headers = @{ Authorization = "Bearer $token" }
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities/$id" -Headers $headers
        return $response.data
    }
    catch { Write-Host "获取活动详情失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

function Upload-Attachment {
    param($token, $activityId, $content)
    $headers = @{ Authorization = "Bearer $token" }
    $tempFile = [System.IO.Path]::GetTempFileName()
    [System.IO.File]::WriteAllText($tempFile, $content)
    try {
        $form = @{ file = Get-Item $tempFile; description = "测试附件" }
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities/$activityId/attachments" -Method POST -Form $form -Headers $headers
        return $response.data.id
    }
    catch { Write-Host "上传附件失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
    finally { if (Test-Path $tempFile) { Remove-Item $tempFile } }
}

function Delete-Activity {
    param($token, $id)
    $headers = @{ Authorization = "Bearer $token" }
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities/$id" -Method DELETE -Headers $headers
        return $response
    }
    catch { Write-Host "删除活动失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

function Restore-Activity {
    param($token, $id)
    $headers = @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" }
    $body = @{ activity_id = $id; user_type = "admin" } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities/$id/restore" -Method POST -Body $body -Headers $headers
        return $response
    }
    catch { Write-Host "恢复活动失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

function Batch-Delete {
    param($token, $ids)
    $headers = @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" }
    $body = @{ activity_ids = $ids } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities/batch-delete" -Method POST -Body $body -Headers $headers
        return $response
    }
    catch { Write-Host "批量删除失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

function Get-Deletable {
    param($token)
    $headers = @{ Authorization = "Bearer $token" }
    try {
        $response = Invoke-RestMethod -Uri "${baseUrl}/api/activities/deletable" -Headers $headers
        return $response.data.activities
    }
    catch { Write-Host "获取可删除活动失败: $($_.Exception.Message)" -ForegroundColor Red; return $null }
}

Write-Host "1. 教师登录..." -ForegroundColor Yellow
$teacherToken = Get-AuthToken $teacherUser
if (-not $teacherToken) { exit 1 }
Write-Host "2. 管理员登录..." -ForegroundColor Yellow
$adminToken = Get-AuthToken $adminUser
if (-not $adminToken) { exit 1 }

Write-Host "3. 创建活动A..." -ForegroundColor Yellow
$activityA = Create-Activity $teacherToken "活动A"
Write-Host "4. 创建活动B..." -ForegroundColor Yellow
$activityB = Create-Activity $teacherToken "活动B"

Write-Host "5. 查询活动列表..." -ForegroundColor Yellow
$activities = Get-Activities $teacherToken
$activities | ConvertTo-Json -Depth 3 | Write-Host

Write-Host "6. 查询活动A详情..." -ForegroundColor Yellow
$detailA = Get-Activity $teacherToken $activityA
$detailA | ConvertTo-Json -Depth 3 | Write-Host

Write-Host "7. 上传附件到活动A..." -ForegroundColor Yellow
$attA = Upload-Attachment $teacherToken $activityA "附件内容A"
Write-Host "8. 上传附件到活动B..." -ForegroundColor Yellow
$attB = Upload-Attachment $teacherToken $activityB "附件内容B"

Write-Host "9. 删除活动A..." -ForegroundColor Yellow
$delA = Delete-Activity $teacherToken $activityA
$delA | ConvertTo-Json -Depth 3 | Write-Host

Write-Host "10. 批量删除活动B..." -ForegroundColor Yellow
$batchDel = Batch-Delete $adminToken @($activityB)
$batchDel | ConvertTo-Json -Depth 3 | Write-Host

Write-Host "11. 查询可删除活动列表..." -ForegroundColor Yellow
$deletable = Get-Deletable $adminToken
$deletable | ConvertTo-Json -Depth 3 | Write-Host

Write-Host "12. 恢复活动A（如API支持）..." -ForegroundColor Yellow
# 仅演示，实际API如未实现可跳过
# $restoreA = Restore-Activity $adminToken $activityA
# $restoreA | ConvertTo-Json -Depth 3 | Write-Host

Write-Host "=== 活动API全流程测试完成 ===" -ForegroundColor Green 