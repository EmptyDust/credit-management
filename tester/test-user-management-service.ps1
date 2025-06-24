# User Management Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

Write-Host "=== User Management Service API Tests ===" -ForegroundColor Cyan

# --- Register User ---
$guid = [guid]::NewGuid().ToString().Substring(0, 8)
$username = "user_$guid"
$password = "password123"
$registerBody = @{ username = $username; password = $password; user_type = "student"; email = "$($username)@test.com"; real_name = "Test User" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerBody -Headers @{"Content-Type" = "application/json" }
    $userId = $response.user.id
    Write-Host "PASS: User '$username' registered successfully. ID: $userId" -ForegroundColor Green
    
    # 验证返回的ID是UUID格式
    if ($userId -match "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$") {
        Write-Host "PASS: User ID is valid UUID format" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: User ID is not valid UUID format: $userId" -ForegroundColor Red
    }
}
catch {
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
}
catch {
    Write-Host "FAIL: Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Get Profile ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/profile" -Method Get -Headers $authHeaders
    if ($response.username -eq $username) {
        Write-Host "PASS: Successfully fetched user profile." -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Profile username does not match." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get user profile failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Profile ---
$updateBody = @{ real_name = "Updated User" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/users/profile" -Method Put -Headers $authHeaders -Body $updateBody
    Write-Host "PASS: Successfully updated user profile." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Update user profile failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Get User by UUID ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/$userId" -Method Get -Headers $authHeaders
    if ($response.id -eq $userId) {
        Write-Host "PASS: Successfully fetched user by UUID." -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: User ID does not match." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get user by UUID failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Update User by UUID ---
$updateUserBody = @{ real_name = "Updated User by UUID" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/users/$userId" -Method Put -Headers $authHeaders -Body $updateUserBody
    Write-Host "PASS: Successfully updated user by UUID." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Update user by UUID failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Get All Users ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users" -Method Get -Headers $authHeaders
    if ($response.users -and $response.users.Count -gt 0) {
        Write-Host "PASS: Successfully fetched all users. Count: $($response.users.Count)" -ForegroundColor Green
        
        # 验证返回的用户都有UUID格式的ID
        $validUuids = $response.users | Where-Object { $_.id -match "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$" }
        if ($validUuids.Count -eq $response.users.Count) {
            Write-Host "PASS: All users have valid UUID format IDs" -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Some users do not have valid UUID format IDs" -ForegroundColor Red
        }
    }
    else {
        Write-Host "FAIL: No users returned or invalid response format." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get all users failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Get Users by Type ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/type/student" -Method Get -Headers $authHeaders
    if ($response -and $response.Count -gt 0) {
        Write-Host "PASS: Successfully fetched students. Count: $($response.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: No students returned or invalid response format." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get users by type failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Get User Stats ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/stats" -Method Get -Headers $authHeaders
    if ($response.total_users -ge 0) {
        Write-Host "PASS: Successfully fetched user stats." -ForegroundColor Green
        Write-Host "  Total Users: $($response.total_users)" -ForegroundColor Yellow
        Write-Host "  Active Users: $($response.active_users)" -ForegroundColor Yellow
        Write-Host "  Student Users: $($response.student_users)" -ForegroundColor Yellow
    }
    else {
        Write-Host "FAIL: Invalid user stats response." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get user stats failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "=== User Management Service API Tests Completed ===" -ForegroundColor Cyan 