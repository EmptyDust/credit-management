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
}
catch {
    Write-Host "FAIL: Student registration failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Login to get token
$loginBody = @{ username = $username; password = $password } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -Headers @{"Content-Type" = "application/json" }
    $jwtToken = $response.token
    $authHeaders = @{ "Authorization" = "Bearer $jwtToken"; "Content-Type" = "application/json"; "X-User-Id" = $username }
    Write-Host "PASS: Student login successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Student login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Create Affair First ---
$affairName = "test_affair_$(Get-Random)"
$createAffairBody = @{ 
    name         = $affairName; 
    description  = "Test affair for application testing"; 
    creator_id   = $username; 
    participants = @($username, "student_001", "student_002"); 
    attachments  = "[{`"name`":`"test.pdf`",`"url`":`"/uploads/test.pdf`"}]" 
} | ConvertTo-Json -Depth 5
try {
    $affairResp = Invoke-RestMethod -Uri "$baseUrl/api/affairs" -Method Post -Headers $authHeaders -Body $createAffairBody
    $affairId = $affairResp.id
    Write-Host "PASS: Affair created. ID: $affairId" -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Create affair failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Test Batch Create Applications ---
$batchCreateBody = @{ 
    affair_id    = $affairId; 
    creator_id   = $username; 
    participants = @($username, "student_001", "student_002") 
} | ConvertTo-Json -Depth 5
try {
    $batchResp = Invoke-RestMethod -Uri "$baseUrl/api/applications/batch" -Method Post -Headers $authHeaders -Body $batchCreateBody
    Write-Host "PASS: Batch create applications successful. Count: $($batchResp.count)" -ForegroundColor Green
    $appId = $batchResp.applications[0].id
}
catch {
    Write-Host "FAIL: Batch create applications failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Test Get Application Detail ---
try {
    $appDetail = Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId/detail" -Method Get -Headers $authHeaders
    if ($appDetail.application) {
        Write-Host "PASS: Get application detail successful. Status: $($appDetail.application.status)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Get application detail response format incorrect" -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get application detail failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Update Application Details ---
$updateDetailsBody = @{ 
    applied_credits = 3.0; 
    details         = @{ 
        level   = "国家级"; 
        name    = "全国大学生数学竞赛"; 
        award   = "一等奖"; 
        ranking = 1 
    } 
} | ConvertTo-Json -Depth 5
try {
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId/details" -Method Put -Headers $authHeaders -Body $updateDetailsBody
    Write-Host "PASS: Update application details successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Update application details failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Submit Application ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId/submit" -Method Post -Headers $authHeaders
    Write-Host "PASS: Submit application successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Submit application failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Register and Login Teacher User for Review ---
$teacherGuid = [guid]::NewGuid().ToString().Substring(0, 8)
$teacherUsername = "teacher_$teacherGuid"
$teacherPassword = "password123"
$registerTeacherBody = @{ username = $teacherUsername; password = $teacherPassword; user_type = "teacher"; email = "$($teacherUsername)@test.com"; real_name = "Test Teacher" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerTeacherBody -Headers @{"Content-Type" = "application/json" }
    $loginTeacherBody = @{ username = $teacherUsername; password = $teacherPassword } | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginTeacherBody -Headers @{"Content-Type" = "application/json" }
    $teacherToken = $response.token
    $teacherAuthHeaders = @{ "Authorization" = "Bearer $teacherToken"; "Content-Type" = "application/json"; "X-User-Id" = $teacherUsername }
    Write-Host "PASS: Teacher user registered and logged in successfully." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Teacher registration or login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Test Update Application Status (as Teacher/Admin) ---
$updateStatusBody = @{ 
    status           = "approved"; 
    review_comment   = "材料完整，符合要求"; 
    approved_credits = 3.0 
} | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId/status" -Method Put -Headers $teacherAuthHeaders -Body $updateStatusBody
    Write-Host "PASS: Update application status successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Update application status failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Update Application Status to Rejected (as Teacher/Admin) ---
# Create a new application for rejected test
$batchCreateBody2 = @{ 
    affair_id    = $affairId; 
    creator_id   = $username; 
    participants = @("student_003", "student_004") 
} | ConvertTo-Json -Depth 5
try {
    $batchResp2 = Invoke-RestMethod -Uri "$baseUrl/api/applications/batch" -Method Post -Headers $authHeaders -Body $batchCreateBody2
    $appId2 = $batchResp2.applications[0].id
    
    # Submit the new application
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId2/submit" -Method Post -Headers $authHeaders
    
    # Now test rejected status
    $updateStatusBodyRejected = @{ 
        status           = "rejected"; 
        review_comment   = "Not qualified"; 
        approved_credits = 0.0 
    } | ConvertTo-Json
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId2/status" -Method Put -Headers $teacherAuthHeaders -Body $updateStatusBodyRejected
    Write-Host "PASS: Update application status to rejected successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Update application status to rejected failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Test Permission Control (Different User) ---
# Register another student
$guid2 = [guid]::NewGuid().ToString().Substring(0, 8)
$username2 = "student_$guid2"
$registerBody2 = @{ username = $username2; password = $password; user_type = "student"; email = "$($username2)@test.com"; real_name = "Test Student 2" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerBody2 -Headers @{"Content-Type" = "application/json" }
    $loginBody2 = @{ username = $username2; password = $password } | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody2 -Headers @{"Content-Type" = "application/json" }
    $jwtToken2 = $response.token
    $authHeaders2 = @{ "Authorization" = "Bearer $jwtToken2"; "Content-Type" = "application/json"; "X-User-Id" = $username2 }
    
    # Try to update application details with different user
    Invoke-RestMethod -Uri "$baseUrl/api/applications/$appId/details" -Method Put -Headers $authHeaders2 -Body $updateDetailsBody
    Write-Host "FAIL: Update application details should have failed for different user" -ForegroundColor Red
}
catch {
    if ($_.Exception.Response.StatusCode -eq 403) {
        Write-Host "PASS: Update application details correctly rejected for different user (403 Forbidden)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Update application details failed with unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# --- Get Applications by User ---
try {
    $apps = Invoke-RestMethod -Uri "$baseUrl/api/applications/user/$username" -Method Get -Headers $authHeaders
    Write-Host "PASS: Get applications by user successful. Count: $($apps.Count)" -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Get applications by user failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get All Applications ---
try {
    $allApps = Invoke-RestMethod -Uri "$baseUrl/api/applications" -Method Get -Headers $authHeaders
    $appsCount = 0
    if ($allApps.applications) { $appsCount = $allApps.applications.Count }
    elseif ($allApps.Count) { $appsCount = $allApps.Count }
    if ($appsCount -ge 1) {
        Write-Host "PASS: Get all applications successful. Count: $appsCount" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Get all applications returned empty list." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get all applications failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== Application Management Service Tests Completed ===" -ForegroundColor Cyan 