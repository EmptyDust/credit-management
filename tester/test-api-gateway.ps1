# API Gateway E2E Test Script

# ==============================================================================
# 1. Global Variables and Initial Setup
# ==============================================================================
$ErrorActionPreference = "Stop" 
$baseUrl = "http://localhost:8000"
$global:jwtToken = $null
$global:studentId = $null
$global:studentNumber = "TEST_STUDENT_001" # Hardcoded for testing
$global:adminId = $null
$global:authHeaders = $null
$global:adminAuthHeaders = $null

function Initial-Setup {
    Write-Host "============== 1. Initial Setup ==============" -ForegroundColor Green
    
    # --- 1.1 Health Check ---
    Write-Host "`n--- 1.1 Health Check ---" -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/health" -Method Get
        Write-Host "PASS: Health check successful." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Health check failed: $($_.Exception.Message)" -ForegroundColor Red
        Exit 1
    }

    # --- 1.2 Register Student ---
    Write-Host "`n--- 1.2 Register Student User ---" -ForegroundColor Yellow
    $guid = [guid]::NewGuid().ToString().Substring(0, 8)
    $studentUsername = "student_$guid"
    $studentPassword = "password123"
    $registerStudentBody = @{
        username   = $studentUsername
        password   = $studentPassword
        user_type  = "student"
        student_id = $global:studentNumber # This is likely for the student-info service, not user-management
        email      = "$($studentUsername)@test.com"
        real_name  = "Test Student User"
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerStudentBody -Headers @{"Content-Type" = "application/json" }
        $global:studentId = $response.user.id
        Write-Host "PASS: Student '$studentUsername' registered successfully. ID: $($global:studentId)" -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Student registration failed: $($_.Exception.Message)" -ForegroundColor Red
        Exit 1
    }

    # --- 1.3 Login Student ---
    Write-Host "`n--- 1.3 Login Student and Get Token ---" -ForegroundColor Yellow
    $loginBody = @{ username = $studentUsername; password = $studentPassword } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -Headers @{"Content-Type" = "application/json" }
        $global:jwtToken = $response.token
        $global:authHeaders = @{ "Authorization" = "Bearer $($global:jwtToken)"; "Content-Type" = "application/json" }
        Write-Host "PASS: Student login successful." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Student login failed: $($_.Exception.Message)" -ForegroundColor Red
        Exit 1
    }
}

# ==============================================================================
# 2. Test Functions
# ==============================================================================

function Test-UserManagement {
    Write-Host "`n============== 2. User Management Service Tests ==============" -ForegroundColor Green

    # --- 2.1 Get User Profile ---
    Write-Host "`n--- 2.1 Get User Profile ---" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/users/profile" -Method Get -Headers $global:authHeaders
        if ($response.id -eq $global:studentId) {
            Write-Host "PASS: Successfully fetched user profile." -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Fetched profile ID does not match registered student ID." -ForegroundColor Red
        }
    }
    catch {
        Write-Host "FAIL: Get user profile failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

function Test-AuthService {
    Write-Host "`n============== 3. Auth Service Tests ==============" -ForegroundColor Green

    # --- 3.1 Validate Token ---
    Write-Host "`n--- 3.1 Validate Token ---" -ForegroundColor Yellow
    try {
        $validationBody = @{ "token" = $global:jwtToken } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/validate-token" -Method Post -Headers $global:authHeaders -Body $validationBody
        if ($response.valid) {
            Write-Host "PASS: Token validation successful." -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Token validation failed." -ForegroundColor Red
        }
    }
    catch {
        Write-Host "FAIL: Token validation request failed: $($_.Exception.Message)" -ForegroundColor Red
    }

    # --- 3.2 Refresh Token ---
    Write-Host "`n--- 3.2 Refresh Token ---" -ForegroundColor Yellow
    try {
        $refreshBody = @{"refresh_token" = $global:jwtToken } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/refresh-token" -Method Post -Headers $global:authHeaders -Body $refreshBody
        if ($response.token) {
            $global:jwtToken = $response.token
            $global:authHeaders["Authorization"] = "Bearer $($global:jwtToken)"
            Write-Host "PASS: Token refresh successful." -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Token refresh failed." -ForegroundColor Red
        }
    }
    catch {
        Write-Host "FAIL: Token refresh request failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

function Test-AffairAndApplicationServices {
    Write-Host "`n============== 4. Affair and Application Service Tests ==============" -ForegroundColor Green
    
    # --- 4.1 Create Affairs if they don't exist ---
    Write-Host "`n--- 4.1 Ensure Affairs Exist ---" -ForegroundColor Yellow
    $requiredAffairs = @("discipline_competition", "innovation_practice", "research_innovation", "paper_patent", "social_service")
    $affairDetails = @{}

    # Get existing affairs first
    try {
        $existingAffairs = Invoke-RestMethod -Uri "$baseUrl/api/affairs" -Method Get -Headers $global:authHeaders
        foreach ($affair in $existingAffairs) {
            $affairDetails[$affair.name] = $affair.id
        }
        Write-Host "PASS: Fetched existing affairs." -ForegroundColor Green
    }
    catch {
        Write-Host "WARN: Could not fetch existing affairs. Will attempt to create all. $($_.Exception.Message)" -ForegroundColor Yellow
    }

    foreach ($typeName in $requiredAffairs) {
        if (-not $affairDetails.ContainsKey($typeName)) {
            $body = @{ name = $typeName } | ConvertTo-Json
            try {
                $response = Invoke-RestMethod -Uri "$baseUrl/api/affairs" -Method Post -Headers $global:authHeaders -Body $body
                $affairDetails[$typeName] = $response.id
                Write-Host "PASS: Created affair '$typeName' with ID: $($response.id)" -ForegroundColor Green
            }
            catch {
                Write-Host "FAIL: Failed to create affair '$typeName': $($_.Exception.Message)" -ForegroundColor Red
            }
        }
        else {
            Write-Host "INFO: Affair '$typeName' already exists with ID: $($affairDetails[$typeName])" -ForegroundColor Cyan
        }
    }

    # --- 4.2 Create Applications for each Affair ---
    Write-Host "`n--- 4.2 Create Applications ---" -ForegroundColor Yellow
    $applicationDetails = @{
        discipline_competition = @{ level = "National"; name = "NC"; award = "1st Prize"; ranking = 1; applied_credits = 5.0 }
        innovation_practice    = @{ internship = "Google"; project_id = "Project-X"; date = "2023-10-01"; hours = 100; applied_credits = 8.0 }
        research_innovation    = @{ project_name = "AI Research"; project_type = "Fundamental"; role = "Lead"; applied_credits = 10.0 }
        paper_patent           = @{ name = "My Paper"; category = "Journal"; ranking = 1; applied_credits = 15.0 }
        social_service         = @{ activity_name = "Volunteering"; organization = "Red Cross"; hours = 40; applied_credits = 3.0 }
    }
    $createdAppIDs = New-Object System.Collections.ArrayList

    foreach ($typeName in $affairDetails.Keys) {
        $affairId = $affairDetails[$typeName]
        $details = $applicationDetails[$typeName]

        $body = @{
            affair_id      = $affairId
            student_number = $global:studentNumber
            details        = $details
        } | ConvertTo-Json -Depth 5

        try {
            $response = Invoke-RestMethod -Uri "$baseUrl/api/applications" -Method Post -Headers $global:authHeaders -Body $body
            if ($response.id) {
                Write-Host "PASS: Created application for affair '$typeName' (ID: $affairId). App ID: $($response.id)" -ForegroundColor Green
                $null = $createdAppIDs.Add($response.id)
            }
            else {
                Write-Host "FAIL: Create application for affair '$typeName' returned no ID." -ForegroundColor Red
            }
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            $errorBody = $_.Exception.Response.Content
            Write-Host "FAIL: Create application for affair '$typeName' failed. Status: $statusCode" -ForegroundColor Red
            Write-Host "Error Details: $errorBody" -ForegroundColor Red
        }
    }

    # --- 4.3 Get User's Applications ---
    Write-Host "`n--- 4.3 Get User's Applications ---" -ForegroundColor Yellow
    if ($createdAppIDs.Count -gt 0) {
        try {
            $userApps = Invoke-RestMethod -Uri "$baseUrl/api/applications/user/$($global:studentNumber)" -Method Get -Headers $global:authHeaders
            if ($userApps.Count -ge $createdAppIDs.Count) {
                Write-Host "PASS: Successfully fetched $($userApps.Count) applications for the user." -ForegroundColor Green
            }
            else {
                Write-Host "FAIL: Expected at least $($createdAppIDs.Count) applications, but got $($userApps.Count)." -ForegroundColor Red
            }
        }
        catch {
            Write-Host "FAIL: Failed to get user applications: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "SKIP: No applications were created, skipping get user applications test." -ForegroundColor Gray
    }


    # --- 4.4 Update Application Status ---
    Write-Host "`n--- 4.4 Update Application Status ---" -ForegroundColor Yellow
    if ($createdAppIDs.Count -gt 0) {
        $appToUpdateId = $createdAppIDs[0]
        $updateBody = @{
            status           = "审核通过"
            review_comment   = "Test comment"
            approved_credits = 5.0
        } | ConvertTo-Json

        try {
            Invoke-RestMethod -Uri "$baseUrl/api/applications/$appToUpdateId/status" -Method Put -Headers $global:authHeaders -Body $updateBody
            Write-Host "PASS: Successfully sent update status request for App ID: $appToUpdateId" -ForegroundColor Green
        }
        catch {
            Write-Host "FAIL: Failed to update application status for App ID: $appToUpdateId. $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "SKIP: No applications were created, skipping update status test." -ForegroundColor Gray
    }

    # --- 4.5 Verify Application Status Update ---
    Write-Host "`n--- 4.5 Verify Application Status Update ---" -ForegroundColor Yellow
    if ($createdAppIDs.Count -gt 0) {
        $appToVerifyId = $createdAppIDs[0]
        try {
            $app = Invoke-RestMethod -Uri "$baseUrl/api/applications/$appToVerifyId" -Method Get -Headers $global:authHeaders
            if ($app.application.status -eq "审核通过") {
                Write-Host "PASS: Application status successfully updated to '审核通过' for App ID: $appToVerifyId" -ForegroundColor Green
            }
            else {
                Write-Host "FAIL: Application status was not updated correctly. Current status: $($app.application.status)" -ForegroundColor Red
            }
        }
        catch {
            Write-Host "FAIL: Failed to verify application status for App ID: $appToVerifyId. $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "SKIP: No applications were created, skipping verify status test." -ForegroundColor Gray
    }
}

function Test-StudentInfoService {
    Write-Host "`n============== 5. Student Info Service Tests ==============" -ForegroundColor Green

    # --- 5.1 Create Student Info ---
    Write-Host "`n--- 5.1 Create Student Info ---" -ForegroundColor Yellow
    $createStudentBody = @{
        username   = "student_$(($global:studentId))"
        student_id = $global:studentNumber
        name       = "Test Student"
        college    = "College of Engineering"
        major      = "Computer Science"
        class      = "Class of 2025"
        contact    = "1234567890"
        email      = "student@test.com"
        grade      = "2021"
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$baseUrl/api/students" -Method Post -Headers $global:authHeaders -Body $createStudentBody
        Write-Host "PASS: Successfully created student info for student number: $($global:studentNumber)" -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Failed to create student info: $($_.Exception.Message)" -ForegroundColor Red
    }

    # --- 5.2 Get Student Info ---
    Write-Host "`n--- 5.2 Get Student Info ---" -ForegroundColor Yellow
    try {
        $studentInfo = Invoke-RestMethod -Uri "$baseUrl/api/students/$($global:studentNumber)" -Method Get -Headers $global:authHeaders
        if ($studentInfo.college -eq "College of Engineering") {
            Write-Host "PASS: Successfully fetched student info and verified college." -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Fetched student info college '$($studentInfo.college)' does not match expected." -ForegroundColor Red
        }
    }
    catch {
        Write-Host "FAIL: Failed to get student info: $($_.Exception.Message)" -ForegroundColor Red
    }

    # --- 5.3 Update Student Info ---
    Write-Host "`n--- 5.3 Update Student Info ---" -ForegroundColor Yellow
    $updateStudentBody = @{
        college = "College of Science"
    } | ConvertTo-Json
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/students/$($global:studentNumber)" -Method Put -Headers $global:authHeaders -Body $updateStudentBody
        Write-Host "PASS: Successfully updated student info." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Failed to update student info: $($_.Exception.Message)" -ForegroundColor Red
    }

    # --- 5.4 Verify Student Info Update ---
    Write-Host "`n--- 5.4 Verify Student Info Update ---" -ForegroundColor Yellow
    try {
        $studentInfo = Invoke-RestMethod -Uri "$baseUrl/api/students/$($global:studentNumber)" -Method Get -Headers $global:authHeaders
        if ($studentInfo.college -eq "College of Science") {
            Write-Host "PASS: Student info college successfully updated to '$($studentInfo.college)'." -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Student info college was not updated. Current: $($studentInfo.college)" -ForegroundColor Red
        }
    }
    catch {
        Write-Host "FAIL: Failed to get student info after update: $($_.Exception.Message)" -ForegroundColor Red
    }
}

function Test-TeacherInfoService {
    Write-Host "`n============== 6. Teacher Info Service Tests ==============" -ForegroundColor Green

    # --- 6.1 Register and Login Teacher ---
    Write-Host "`n--- 6.1 Register and Login Teacher ---" -ForegroundColor Yellow
    $teacherUsername = "teacher_$(($global:testRunId))"
    $teacherPassword = "password123"
    $teacherAuthToken = $null
    $teacherAuthHeaders = $null

    $registerTeacherBody = @{
        username  = $teacherUsername
        password  = $teacherPassword
        user_type = "teacher"
        email     = "$($teacherUsername)@test.com"
        real_name = "Test Teacher User"
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$baseUrl/api/users/register" -Method Post -Body $registerTeacherBody -ContentType "application/json"
        Write-Host "PASS: Registered teacher user '$teacherUsername'." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Could not register teacher user: $($_.Exception.Message)" -ForegroundColor Red
        return
    }

    $loginBody = @{
        username = $teacherUsername
        password = $teacherPassword
    } | ConvertTo-Json

    try {
        $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
        $teacherAuthToken = $loginResponse.token
        $teacherAuthHeaders = @{ Authorization = "Bearer $teacherAuthToken" }
        Write-Host "PASS: Teacher '$teacherUsername' logged in successfully." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Could not log in as teacher: $($_.Exception.Message)" -ForegroundColor Red
        return
    }

    # --- 6.2 Create Teacher Info ---
    Write-Host "`n--- 6.2 Create Teacher Info ---" -ForegroundColor Yellow
    $createTeacherBody = @{
        username   = $teacherUsername
        name       = "Professor Test"
        department = "Computer Science"
        title      = "Professor"
        specialty  = "Distributed Systems"
        contact    = "0987654321"
        email      = "prof.test@test.com"
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$baseUrl/api/teachers" -Method Post -Headers $teacherAuthHeaders -Body $createTeacherBody -ContentType "application/json"
        Write-Host "PASS: Successfully created teacher info for user '$teacherUsername'." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Failed to create teacher info: $($_.Exception.Message)" -ForegroundColor Red
    }

    # --- 6.3 Get Teacher Info ---
    Write-Host "`n--- 6.3 Get Teacher Info ---" -ForegroundColor Yellow
    try {
        $teacherInfo = Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUsername" -Method Get -Headers $teacherAuthHeaders
        if ($teacherInfo.department -eq "Computer Science") {
            Write-Host "PASS: Successfully fetched teacher info and verified department." -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Fetched teacher department '$($teacherInfo.department)' did not match." -ForegroundColor Red
        }
    }
    catch {
        Write-Host "FAIL: Failed to get teacher info: $($_.Exception.Message)" -ForegroundColor Red
    }

    # --- 6.4 Update Teacher Info ---
    Write-Host "`n--- 6.4 Update Teacher Info ---" -ForegroundColor Yellow
    $updateTeacherBody = @{
        department = "Electrical Engineering"
    } | ConvertTo-Json
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUsername" -Method Put -Headers $teacherAuthHeaders -Body $updateTeacherBody -ContentType "application/json"
        Write-Host "PASS: Successfully updated teacher info." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Failed to update teacher info: $($_.Exception.Message)" -ForegroundColor Red
    }

    # --- 6.5 Verify Teacher Info Update ---
    Write-Host "`n--- 6.5 Verify Teacher Info Update ---" -ForegroundColor Yellow
    try {
        $teacherInfo = Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUsername" -Method Get -Headers $teacherAuthHeaders
        if ($teacherInfo.department -eq "Electrical Engineering") {
            Write-Host "PASS: Teacher info department updated to '$($teacherInfo.department)'." -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Teacher info department was not updated. Current: $($teacherInfo.department)" -ForegroundColor Red
        }
    }
    catch {
        Write-Host "FAIL: Failed to get teacher info after update: $($_.Exception.Message)" -ForegroundColor Red
    }
}

function Final-Logout {
    Write-Host "`n============== 5. Final Logout ==============" -ForegroundColor Green
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/auth/logout" -Method Post -Headers $global:authHeaders
        Write-Host "PASS: Logout successful." -ForegroundColor Green
    }
    catch {
        Write-Host "FAIL: Logout failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# ==============================================================================
# 3. Test Execution
# ==============================================================================
Initial-Setup
Test-UserManagement
Test-AuthService
Test-AffairAndApplicationServices
Test-StudentInfoService
Test-TeacherInfoService
Final-Logout

Write-Host "`n`n=== All Tests Completed ===" -ForegroundColor Cyan 