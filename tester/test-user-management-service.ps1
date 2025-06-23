# User Management Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Register User ---
$guid = [guid]::NewGuid().ToString().Substring(0, 8)
$username = "user_$guid"
$password = "password123"
$registerBody = @{ username = $username; password = $password; user_type = "student"; email = "$($username)@test.com"; real_name = "Test User" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerBody -Headers @{"Content-Type" = "application/json" }
    $userId = $response.user.id
    Write-Host "PASS: User '$username' registered successfully. ID: $userId" -ForegroundColor Green
} catch {
    Write-Host "FAIL: User registration failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Login ---
$loginBody = @{ username = $username; password = $password } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -Headers @{"Content-Type" = "application/json" }
    $jwtToken = $response.token
    $authHeaders = @{ "Authorization" = "Bearer $jwtToken"; "Content-Type" = "application/json" }
    Write-Host "PASS: Login successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Get Profile ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/profile" -Method Get -Headers $authHeaders
    if ($response.username -eq $username) {
        Write-Host "PASS: Successfully fetched user profile." -ForegroundColor Green
    } else {
        Write-Host "FAIL: Profile username does not match." -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL: Get user profile failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Profile ---
$updateBody = @{ real_name = "Updated User" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/users/profile" -Method Put -Headers $authHeaders -Body $updateBody
    Write-Host "PASS: Successfully updated user profile." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Update user profile failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get All Users (Admin Required) ---
# (需要admin token，略，实际测试时应补充admin登录和操作)

# --- Get Users by Type (Admin Required) ---
# (同上) 