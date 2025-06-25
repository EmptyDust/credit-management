# 测试JWT Token内容
Write-Host "=== 测试JWT Token内容 ===" -ForegroundColor Green

# 1. 管理员登录获取token
Write-Host "`n1. 管理员登录获取token..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body @"
{
    "username": "admin",
    "password": "adminpassword"
}
"@

if ($loginResponse.code -eq 0) {
    Write-Host "管理员登录成功，获取到token" -ForegroundColor Green
    $token = $loginResponse.data.token
    
    # 2. 解析JWT token（不验证签名，只查看内容）
    Write-Host "`n2. 解析JWT token内容..." -ForegroundColor Yellow
    $tokenParts = $token.Split('.')
    if ($tokenParts.Length -eq 3) {
        $payload = $tokenParts[1]
        # 添加padding
        $padding = 4 - ($payload.Length % 4)
        if ($padding -ne 4) {
            $payload = $payload + ("=" * $padding)
        }
        
        try {
            $decodedBytes = [System.Convert]::FromBase64String($payload)
            $decodedString = [System.Text.Encoding]::UTF8.GetString($decodedBytes)
            $claims = $decodedString | ConvertFrom-Json
            
            Write-Host "JWT Token Claims:" -ForegroundColor Cyan
            Write-Host "  user_id: $($claims.user_id)" -ForegroundColor White
            Write-Host "  username: $($claims.username)" -ForegroundColor White
            Write-Host "  user_type: $($claims.user_type)" -ForegroundColor White
            Write-Host "  exp: $($claims.exp)" -ForegroundColor White
            Write-Host "  iat: $($claims.iat)" -ForegroundColor White
            
            if ($claims.user_type -eq "admin") {
                Write-Host "`n✓ user_type 字段正确设置为 'admin'" -ForegroundColor Green
            } else {
                Write-Host "`n✗ user_type 字段不正确: $($claims.user_type)" -ForegroundColor Red
            }
        } catch {
            Write-Host "解析JWT token失败: $($_.Exception.Message)" -ForegroundColor Red
        }
    } else {
        Write-Host "JWT token格式不正确" -ForegroundColor Red
    }
    
    # 3. 测试用户服务权限验证
    Write-Host "`n3. 测试用户服务权限验证..." -ForegroundColor Yellow
    try {
        $headers = @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        }
        
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/users/profile" -Method GET -Headers $headers
        Write-Host "获取用户信息成功: $($response.code)" -ForegroundColor Green
    } catch {
        Write-Host "获取用户信息失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    
} else {
    Write-Host "管理员登录失败: $($loginResponse.message)" -ForegroundColor Red
} 