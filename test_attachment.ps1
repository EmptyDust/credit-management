# 活动服务附件功能测试脚本
# 测试环境：通过API网关访问服务

$ErrorActionPreference = "Stop"

# 配置
$GATEWAY_URL = "http://localhost:8080"
$API_PREFIX = "/api"

# 测试用户信息
$TEST_USERS = @{
    "admin"   = @{
        username = "admin"
        password = "adminpassword"
    }
    "teacher" = @{
        username = "teacher"
        password = "adminpassword"
    }
    "student" = @{
        username = "student"
        password = "adminpassword"
    }
}

# 全局变量
$global:authTokens = @{}
$global:testActivityId = $null
$global:testAttachmentId = $null

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# HTTP请求函数
function Invoke-APIRequest {
    param(
        [string]$Method = "GET",
        [string]$Endpoint,
        [object]$Body = $null,
        [hashtable]$Headers = @{},
        [string]$ContentType = "application/json",
        [string]$User = "admin"
    )
    
    $url = "$GATEWAY_URL$API_PREFIX$Endpoint"
    $headers["Content-Type"] = $ContentType
    
    if ($global:authTokens.ContainsKey($User)) {
        $headers["Authorization"] = "Bearer $($global:authTokens[$User])"
    }
    
    $params = @{
        Method  = $Method
        Uri     = $url
        Headers = $headers
    }
    
    if ($Body) {
        if ($ContentType -eq "application/json") {
            $params.Body = $Body | ConvertTo-Json -Depth 10
        }
        else {
            $params.Body = $Body
        }
    }
    
    try {
        $response = Invoke-RestMethod @params
        return @{
            Success = $true
            Data    = $response
        }
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            $reader.Close()
            
            try {
                $errorData = $errorBody | ConvertFrom-Json
                return @{
                    Success    = $false
                    Error      = $errorData
                    StatusCode = $errorResponse.StatusCode
                }
            }
            catch {
                return @{
                    Success    = $false
                    Error      = @{ message = $errorBody }
                    StatusCode = $errorResponse.StatusCode
                }
            }
        }
        else {
            return @{
                Success    = $false
                Error      = @{ message = $_.Exception.Message }
                StatusCode = 0
            }
        }
    }
}

# 登录函数
function Test-Login {
    param([string]$UserType)
    
    Write-ColorOutput "=== 测试用户登录: $UserType ===" "Yellow"
    
    $user = $TEST_USERS[$UserType]
    $loginData = @{
        username = $user.username
        password = $user.password
    }
    
    $result = Invoke-APIRequest -Method "POST" -Endpoint "/auth/login" -Body $loginData -User $UserType
    if ($result.Success) {
        $global:authTokens[$UserType] = $result.Data.data.token
        Write-ColorOutput "✓ $UserType 登录成功" "Green"
        return $true
    }
    else {
        Write-ColorOutput "✗ $UserType 登录失败: $($result.Error.message)" "Red"
        return $false
    }
}

# 创建测试活动
function Test-CreateActivity {
    param([string]$UserType)
    
    Write-ColorOutput "=== 创建测试活动 ===" "Yellow"
    
    $activityData = @{
        title        = "附件功能测试活动"
        description  = "这是一个用于测试附件功能的测试活动"
        start_date   = "2024-12-01"
        end_date     = "2024-12-31"
        category     = "创新创业"
        requirements = "需要上传相关文档和图片"
    }
    
    $result = Invoke-APIRequest -Method "POST" -Endpoint "/activities" -Body $activityData -User $UserType
    if ($result.Success) {
        $global:testActivityId = $result.Data.data.id
        Write-ColorOutput "✓ 测试活动创建成功，ID: $global:testActivityId" "Green"
        return $true
    }
    else {
        Write-ColorOutput "✗ 测试活动创建失败: $($result.Error.message)" "Red"
        return $false
    }
}

# 创建测试文件
function Create-TestFile {
    param(
        [string]$FileName,
        [string]$Content = "This is a test file content."
    )
    
    $tempDir = "temp_test_files"
    if (-not (Test-Path $tempDir)) {
        New-Item -ItemType Directory -Path $tempDir | Out-Null
    }
    
    $filePath = Join-Path $tempDir $FileName
    $Content | Out-File -FilePath $filePath -Encoding UTF8
    return $filePath
}

# 测试上传单个附件
function Test-UploadSingleAttachment {
    param([string]$UserType)
    
    Write-ColorOutput "=== 测试上传单个附件 ===" "Yellow"
    
    # 创建测试文件
    $testFile = Create-TestFile "test_document.txt" "这是一个测试文档文件，用于测试附件上传功能。"
    
    try {
        # 构建multipart表单数据
        $boundary = [System.Guid]::NewGuid().ToString()
        $LF = "`r`n"
        
        $bodyLines = @(
            "--$boundary",
            "Content-Disposition: form-data; name=`"file`"; filename=`"test_document.txt`"",
            "Content-Type: text/plain",
            "",
            [System.IO.File]::ReadAllText($testFile),
            "--$boundary",
            "Content-Disposition: form-data; name=`"description`"",
            "",
            "测试文档描述",
            "--$boundary--"
        )
        
        $body = $bodyLines -join $LF
        
        $headers = @{
            "Authorization" = "Bearer $($global:authTokens[$UserType])"
            "Content-Type"  = "multipart/form-data; boundary=$boundary"
        }
        
        $url = "$GATEWAY_URL$API_PREFIX/activities/$global:testActivityId/attachments"
        
        $result = Invoke-RestMethod -Method "POST" -Uri $url -Headers $headers -Body $body
        
        if ($result.code -eq 0) {
            $global:testAttachmentId = $result.data.id
            Write-ColorOutput "✓ 单个附件上传成功，ID: $global:testAttachmentId" "Green"
            Write-ColorOutput "  文件名: $($result.data.original_name)" "Cyan"
            Write-ColorOutput "  文件大小: $($result.data.file_size) bytes" "Cyan"
            return $true
        }
        else {
            Write-ColorOutput "✗ 单个附件上传失败: $($result.message)" "Red"
            return $false
        }
    }
    catch {
        Write-ColorOutput "✗ 单个附件上传异常: $($_.Exception.Message)" "Red"
        return $false
    }
    finally {
        # 清理测试文件
        if (Test-Path $testFile) {
            Remove-Item $testFile -Force
        }
    }
}

# 测试获取附件列表
function Test-GetAttachments {
    param([string]$UserType)
    
    Write-ColorOutput "=== 测试获取附件列表 ===" "Yellow"
    
    $result = Invoke-APIRequest -Method "GET" -Endpoint "/activities/$global:testActivityId/attachments" -User $UserType
    if ($result.Success) {
        $attachments = $result.Data.data.attachments
        $stats = $result.Data.data.stats
        
        Write-ColorOutput "✓ 获取附件列表成功" "Green"
        Write-ColorOutput "  附件总数: $($stats.total_count)" "Cyan"
        Write-ColorOutput "  总大小: $($stats.total_size) bytes" "Cyan"
        
        foreach ($attachment in $attachments) {
            Write-ColorOutput "  - $($attachment.original_name) ($($attachment.file_size) bytes, $($attachment.file_type))" "Cyan"
        }
        return $true
    }
    else {
        Write-ColorOutput "✗ 获取附件列表失败: $($result.Error.message)" "Red"
        return $false
    }
}

# 测试下载附件
function Test-DownloadAttachment {
    param([string]$UserType)
    
    Write-ColorOutput "=== 测试下载附件 ===" "Yellow"
    
    if (-not $global:testAttachmentId) {
        Write-ColorOutput "✗ 没有可用的测试附件ID" "Red"
        return $false
    }
    
    $url = "$GATEWAY_URL$API_PREFIX/activities/$global:testActivityId/attachments/$global:testAttachmentId/download"
    $headers = @{
        "Authorization" = "Bearer $($global:authTokens[$UserType])"
    }
    
    try {
        $response = Invoke-RestMethod -Method "GET" -Uri $url -Headers $headers -OutFile "downloaded_file.tmp"
        Write-ColorOutput "✓ 附件下载成功" "Green"
        
        # 检查下载的文件
        if (Test-Path "downloaded_file.tmp") {
            $fileSize = (Get-Item "downloaded_file.tmp").Length
            Write-ColorOutput "  下载文件大小: $fileSize bytes" "Cyan"
            Remove-Item "downloaded_file.tmp" -Force
        }
        return $true
    }
    catch {
        Write-ColorOutput "✗ 附件下载失败: $($_.Exception.Message)" "Red"
        return $false
    }
}

# 测试更新附件信息
function Test-UpdateAttachment {
    param([string]$UserType)
    
    Write-ColorOutput "=== 测试更新附件信息 ===" "Yellow"
    
    if (-not $global:testAttachmentId) {
        Write-ColorOutput "✗ 没有可用的测试附件ID" "Red"
        return $false
    }
    
    $updateData = @{
        description = "更新后的附件描述 - 测试更新功能"
    }
    
    $result = Invoke-APIRequest -Method "PUT" -Endpoint "/activities/$global:testActivityId/attachments/$global:testAttachmentId" -Body $updateData -User $UserType
    if ($result.Success) {
        Write-ColorOutput "✓ 附件信息更新成功" "Green"
        Write-ColorOutput "  新描述: $($result.Data.data.description)" "Cyan"
        return $true
    }
    else {
        Write-ColorOutput "✗ 附件信息更新失败: $($result.Error.message)" "Red"
        return $false
    }
}

# 测试删除附件
function Test-DeleteAttachment {
    param([string]$UserType)
    
    Write-ColorOutput "=== 测试删除附件 ===" "Yellow"
    
    if (-not $global:testAttachmentId) {
        Write-ColorOutput "✗ 没有可用的测试附件ID" "Red"
        return $false
    }
    
    $result = Invoke-APIRequest -Method "DELETE" -Endpoint "/activities/$global:testActivityId/attachments/$global:testAttachmentId" -User $UserType
    if ($result.Success) {
        Write-ColorOutput "✓ 附件删除成功" "Green"
        $global:testAttachmentId = $null
        return $true
    }
    else {
        Write-ColorOutput "✗ 附件删除失败: $($result.Error.message)" "Red"
        return $false
    }
}

# 主测试函数
function Test-AttachmentFeatures {
    Write-ColorOutput "=========================================" "Magenta"
    Write-ColorOutput "开始测试活动服务附件功能" "Magenta"
    Write-ColorOutput "=========================================" "Magenta"
    
    # 检查服务是否可用
    Write-ColorOutput "检查服务可用性..." "Yellow"
    try {
        $healthCheck = Invoke-RestMethod -Uri "$GATEWAY_URL/health" -Method GET
        Write-ColorOutput "✓ 服务健康检查通过" "Green"
    }
    catch {
        Write-ColorOutput "✗ 服务不可用，请确保服务已启动" "Red"
        return
    }
    
    # 登录管理员用户
    if (-not (Test-Login "admin")) {
        Write-ColorOutput "管理员登录失败，无法继续测试" "Red"
        return
    }
    
    # 创建测试活动
    if (-not (Test-CreateActivity "admin")) {
        Write-ColorOutput "测试活动创建失败，无法继续测试" "Red"
        return
    }
    
    # 测试附件功能
    $testResults = @()
    
    $testResults += @{ Name = "上传单个附件"; Result = Test-UploadSingleAttachment "admin" }
    $testResults += @{ Name = "获取附件列表"; Result = Test-GetAttachments "admin" }
    $testResults += @{ Name = "下载附件"; Result = Test-DownloadAttachment "admin" }
    $testResults += @{ Name = "更新附件信息"; Result = Test-UpdateAttachment "admin" }
    $testResults += @{ Name = "删除附件"; Result = Test-DeleteAttachment "admin" }
    
    # 输出测试结果摘要
    Write-ColorOutput "=========================================" "Magenta"
    Write-ColorOutput "测试结果摘要" "Magenta"
    Write-ColorOutput "=========================================" "Magenta"
    
    $passed = 0
    $failed = 0
    
    foreach ($test in $testResults) {
        if ($test.Result) {
            Write-ColorOutput "✓ $($test.Name)" "Green"
            $passed++
        }
        else {
            Write-ColorOutput "✗ $($test.Name)" "Red"
            $failed++
        }
    }
    
    Write-ColorOutput "-----------------------------------------" "Magenta"
    Write-ColorOutput "总计: $($testResults.Count) 个测试" "White"
    Write-ColorOutput "通过: $passed 个" "Green"
    Write-ColorOutput "失败: $failed 个" "Red"
    
    if ($failed -eq 0) {
        Write-ColorOutput "🎉 所有测试通过！附件功能正常工作" "Green"
    }
    else {
        Write-ColorOutput "⚠️  有 $failed 个测试失败，需要检查相关功能" "Yellow"
    }
    
    # 清理测试数据
    Write-ColorOutput "清理测试数据..." "Yellow"
    if ($global:testActivityId) {
        $result = Invoke-APIRequest -Method "DELETE" -Endpoint "/activities/$global:testActivityId" -User "admin"
        if ($result.Success) {
            Write-ColorOutput "✓ 测试活动已删除" "Green"
        }
        else {
            Write-ColorOutput "✗ 测试活动删除失败: $($result.Error.message)" "Red"
        }
    }
    
    # 清理临时目录
    if (Test-Path "temp_test_files") {
        Remove-Item "temp_test_files" -Recurse -Force
        Write-ColorOutput "✓ 临时文件已清理" "Green"
    }
}

# 运行测试
Test-AttachmentFeatures
