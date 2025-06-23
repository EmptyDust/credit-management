# Auth Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Register User First ---
$guid = [guid]::NewGuid().ToString().Substring(0, 8)
$username = "testuser_$guid"
$password = "password123"
$registerBody = @{ username = $username; password = $password; user_type = "student"; email = "$($username)@test.com"; real_name = "Test User" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerBody -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: User registered successfully." -ForegroundColor Green
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

# --- Validate Token ---
$validationBody = @{ "token" = $jwtToken } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/validate-token" -Method Post -Headers $authHeaders -Body $validationBody
    if ($response.valid) {
        Write-Host "PASS: Token validation successful." -ForegroundColor Green
    } else {
        Write-Host "FAIL: Token validation failed." -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL: Token validation request failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Refresh Token ---
$refreshBody = @{"refresh_token" = $jwtToken } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/refresh-token" -Method Post -Headers $authHeaders -Body $refreshBody
    if ($response.token) {
        $jwtToken = $response.token
        $authHeaders["Authorization"] = "Bearer $jwtToken"
        Write-Host "PASS: Token refresh successful." -ForegroundColor Green
    } else {
        Write-Host "FAIL: Token refresh failed." -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL: Token refresh request failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Logout ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/auth/logout" -Method Post -Headers $authHeaders
    Write-Host "PASS: Logout successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Logout failed: $($_.Exception.Message)" -ForegroundColor Red
} 