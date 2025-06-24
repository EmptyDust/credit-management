# Affair Management Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Register and Login User for Testing ---
$guid = [guid]::NewGuid().ToString().Substring(0, 8)
$username = "user_$guid"
$password = "password123"

# Register user
$registerBody = @{ username = $username; password = $password; user_type = "student"; email = "$($username)@test.com"; real_name = "Test User" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerBody -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: User registered successfully." -ForegroundColor Green
} catch {
    Write-Host "FAIL: User registration failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Login to get token
$loginBody = @{ username = $username; password = $password } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -Headers @{"Content-Type" = "application/json" }
    $jwtToken = $response.token
    $authHeaders = @{ "Authorization" = "Bearer $jwtToken"; "Content-Type" = "application/json"; "X-User-Id" = $username }
    Write-Host "PASS: Login successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Create Affair with New Interface ---
$affairName = "affair_$(Get-Random)"
$createBody = @{ 
    name = $affairName; 
    description = "Test affair description"; 
    creator_id = $username; 
    participants = @($username, "student_001", "student_002"); 
    attachments = "[{`"name`":`"test.pdf`",`"url`":`"/uploads/test.pdf`"}]" 
} | ConvertTo-Json -Depth 5
try {
    $resp = Invoke-RestMethod -Uri "$baseUrl/api/affairs" -Method Post -Headers $authHeaders -Body $createBody
    $affairId = $resp.id
    Write-Host "PASS: Affair created with new interface. ID: $affairId" -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create affair failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Affair with Participants ---
try {
    $affair = Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId" -Method Get -Headers $authHeaders
    if ($affair.affair -and $affair.participants) {
        Write-Host "PASS: Get affair with participants successful. Participants: $($affair.participants.Count)" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Get affair response format incorrect" -ForegroundColor Red
    }
} catch {
    Write-Host "FAIL: Get affair failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Affair Participants ---
try {
    $participants = Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId/participants" -Method Get -Headers $authHeaders
    Write-Host "PASS: Get affair participants successful. Count: $($participants.Count)" -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get affair participants failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Affair Applications ---
try {
    $applications = Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId/applications" -Method Get -Headers $authHeaders
    Write-Host "PASS: Get affair applications successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get affair applications failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Update Affair (Creator Only) ---
$updateBody = @{ 
    name = "updated_$affairName"; 
    description = "Updated description"; 
    attachments = "[{`"name`":`"updated.pdf`",`"url`":`"/uploads/updated.pdf`"}]" 
} | ConvertTo-Json -Depth 5
try {
    Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId" -Method Put -Headers $authHeaders -Body $updateBody
    Write-Host "PASS: Update affair successful (creator)." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Update affair failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Update Affair with Different User (Should Fail) ---
# Register another user
$guid2 = [guid]::NewGuid().ToString().Substring(0, 8)
$username2 = "user_$guid2"
$registerBody2 = @{ username = $username2; password = $password; user_type = "student"; email = "$($username2)@test.com"; real_name = "Test User 2" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerBody2 -Headers @{"Content-Type" = "application/json" }
    $loginBody2 = @{ username = $username2; password = $password } | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody2 -Headers @{"Content-Type" = "application/json" }
    $jwtToken2 = $response.token
    $authHeaders2 = @{ "Authorization" = "Bearer $jwtToken2"; "Content-Type" = "application/json"; "X-User-Id" = $username2 }
    
    # Try to update with different user
    Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId" -Method Put -Headers $authHeaders2 -Body $updateBody
    Write-Host "FAIL: Update affair should have failed for non-creator" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 403) {
        Write-Host "PASS: Update affair correctly rejected for non-creator (403 Forbidden)" -ForegroundColor Green
    } else {
        Write-Host "FAIL: Update affair failed with unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# --- Delete Affair ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId" -Method Delete -Headers $authHeaders
    Write-Host "PASS: Delete affair successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Delete affair failed: $($_.Exception.Message)" -ForegroundColor Red
} 