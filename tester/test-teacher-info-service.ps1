# Teacher Info Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

Write-Host "=== Teacher Info Service API Tests ===" -ForegroundColor Cyan

# --- Create Teacher ---
$teacherUsername = "TEACHER_$(Get-Random)"
$createBody = @{ username = $teacherUsername; name = "Test Teacher"; department = "CS"; title = "Lecturer"; specialty = "AI"; contact = "9876543210"; email = "$teacherUsername@test.com" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/teachers" -Method Post -Headers @{"Content-Type" = "application/json" } -Body $createBody
    $teacherUuid = $response.teacher.id
    Write-Host "PASS: Teacher created successfully. UUID: $teacherUuid" -ForegroundColor Green
    
    # 验证返回的ID是UUID格式
    if ($teacherUuid -match "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$") {
        Write-Host "PASS: Teacher ID is valid UUID format" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Teacher ID is not valid UUID format: $teacherUuid" -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Create teacher failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Get Teacher by UUID ---
try {
    $teacher = Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUuid" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($teacher.id -eq $teacherUuid) {
        Write-Host "PASS: Get teacher by UUID successful." -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Teacher ID does not match." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get teacher failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Teacher by UUID ---
$updateBody = @{ department = "Math" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUuid" -Method Put -Headers @{"Content-Type" = "application/json" } -Body $updateBody
    Write-Host "PASS: Update teacher by UUID successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Update teacher failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get All Teachers ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/teachers" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.teachers -and $response.teachers.Count -gt 0) {
        Write-Host "PASS: Successfully fetched all teachers. Count: $($response.teachers.Count)" -ForegroundColor Green
        
        # 验证返回的教师都有UUID格式的ID
        $validUuids = $response.teachers | Where-Object { $_.id -match "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$" }
        if ($validUuids.Count -eq $response.teachers.Count) {
            Write-Host "PASS: All teachers have valid UUID format IDs" -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Some teachers do not have valid UUID format IDs" -ForegroundColor Red
        }
    }
    else {
        Write-Host "FAIL: No teachers returned or invalid response format." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get all teachers failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Teachers by Department ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/teachers/department/CS" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.teachers -and $response.teachers.Count -ge 0) {
        Write-Host "PASS: Successfully fetched teachers by department. Count: $($response.teachers.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Invalid response for teachers by department." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get teachers by department failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Teachers by Title ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/teachers/title/Lecturer" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.teachers -and $response.teachers.Count -ge 0) {
        Write-Host "PASS: Successfully fetched teachers by title. Count: $($response.teachers.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Invalid response for teachers by title." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get teachers by title failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Search Teachers ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/teachers/search?q=Test" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.teachers -and $response.teachers.Count -ge 0) {
        Write-Host "PASS: Successfully searched teachers. Count: $($response.teachers.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Invalid response for teacher search." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Search teachers failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Active Teachers ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/teachers/active" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.teachers -and $response.teachers.Count -ge 0) {
        Write-Host "PASS: Successfully fetched active teachers. Count: $($response.teachers.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Invalid response for active teachers." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get active teachers failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Delete Teacher by UUID ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUuid" -Method Delete -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Delete teacher by UUID successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Delete teacher failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "=== Teacher Info Service API Tests Completed ===" -ForegroundColor Cyan 