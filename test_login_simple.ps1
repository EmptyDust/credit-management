# 简单的登录测试脚本

Write-Host "=== 简单登录测试 ===" -ForegroundColor Green

$baseUrl = "http://localhost:8080"

# 测试登录
$loginBody = @{
    username = "admin"
    password = "adminpassword"
} | ConvertTo-Json

Write-Host "发送登录请求..." -ForegroundColor Yellow
Write-Host "请求体: $loginBody" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "${baseUrl}/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "登录成功!" -ForegroundColor Green
    Write-Host "响应: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
}
catch {
    Write-Host "登录失败: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "状态码: $statusCode" -ForegroundColor Red
    }
} 