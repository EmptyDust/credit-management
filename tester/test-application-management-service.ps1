# Application Management Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Register and Login Student User ---
$guid = [guid]::NewGuid().ToString().Substring(0, 8)
$username = "student_$guid"
$password = "password123"
$studentNumber = "STU_$guid"

# Register student user
$registerBody = @{ username = $username; password = $password; user_type = "student"; email = "$($username)@test.com"; real_name = "Test Student" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerBody -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Student user registered successfully." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Student registration failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Login to get token
$loginBody = @{ username = $username; password = $password } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -Headers @{"Content-Type" = "application/json" }
    $jwtToken = $response.token
    $authHeaders = @{ "Authorization" = "Bearer $jwtToken"; "Content-Type" = "application/json" }
    Write-Host "PASS: Student login successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Student login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Create Affair First ---
$affairName = "test_affair_$(Get-Random)"
$createAffairBody = @{ name = $affairName } | ConvertTo-Json
try {
    $affairResp = Invoke-RestMethod -Uri "$baseUrl/api/affairs" -Method Post -Headers $authHeaders -Body $createAffairBody
    $affairId = $affairResp.id
    Write-Host "PASS: Affair created. ID: $affairId" -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create affair failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Create Application ---
$createBody = @{ 
    affair_id = $affairId; 
    student_number = $studentNumber; 
    details = @{ 
        name = "Test Application"; 
        applied_credits = 5.0;
        level = "National";
        award = "1st Prize";
        ranking = 1
    } 
} | ConvertTo-Json -Depth 5
try {
    $resp = Invoke-RestMethod -Uri "$baseUrl/api/applications" -Method Post -Headers $authHeaders -Body $createBody
    $appId = $resp.id
    Write-Host "PASS: Application created. ID: $appId" -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create application failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Application ---
try {
    $app = Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId" -Method Get -Headers $authHeaders
    Write-Host "PASS: Get application successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get application failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Application Status ---
$updateBody = @{ status = "审核通过"; review_comment = "Test"; approved_credits = 5.0 } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId/status" -Method Put -Headers $authHeaders -Body $updateBody
    Write-Host "PASS: Update application status successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Update application status failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Applications by User ---
try {
    $apps = Invoke-RestMethod -Uri "$baseUrl/api/applications/user/$studentNumber" -Method Get -Headers $authHeaders
    Write-Host "PASS: Get applications by user successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get applications by user failed: $($_.Exception.Message)" -ForegroundColor Red
} 