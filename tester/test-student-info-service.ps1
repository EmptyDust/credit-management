# Student Info Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

Write-Host "=== Student Info Service API Tests ===" -ForegroundColor Cyan

# --- Create Student ---
$studentId = "STU_$(Get-Random)"
$createBody = @{ username = $studentId; student_id = $studentId; name = "Test Student"; college = "Engineering"; major = "CS"; class = "2025"; contact = "1234567890"; email = "$studentId@test.com"; grade = "2021" } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/students" -Method Post -Headers @{"Content-Type" = "application/json" } -Body $createBody
    $studentUuid = $response.student.id
    Write-Host "PASS: Student created successfully. UUID: $studentUuid" -ForegroundColor Green
    
    # 验证返回的ID是UUID格式
    if ($studentUuid -match "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$") {
        Write-Host "PASS: Student ID is valid UUID format" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Student ID is not valid UUID format: $studentUuid" -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Create student failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Get Student by UUID ---
try {
    $student = Invoke-RestMethod -Uri "$baseUrl/api/students/$studentUuid" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($student.id -eq $studentUuid) {
        Write-Host "PASS: Get student by UUID successful." -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Student ID does not match." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get student failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Student by UUID ---
$updateBody = @{ college = "Science" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/students/$studentUuid" -Method Put -Headers @{"Content-Type" = "application/json" } -Body $updateBody
    Write-Host "PASS: Update student by UUID successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Update student failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get All Students ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/students" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.students -and $response.students.Count -gt 0) {
        Write-Host "PASS: Successfully fetched all students. Count: $($response.students.Count)" -ForegroundColor Green
        
        # 验证返回的学生都有UUID格式的ID
        $validUuids = $response.students | Where-Object { $_.id -match "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$" }
        if ($validUuids.Count -eq $response.students.Count) {
            Write-Host "PASS: All students have valid UUID format IDs" -ForegroundColor Green
        }
        else {
            Write-Host "FAIL: Some students do not have valid UUID format IDs" -ForegroundColor Red
        }
    }
    else {
        Write-Host "FAIL: No students returned or invalid response format." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get all students failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Students by College ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/students/college/Engineering" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.students -and $response.students.Count -ge 0) {
        Write-Host "PASS: Successfully fetched students by college. Count: $($response.students.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Invalid response for students by college." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get students by college failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Students by Major ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/students/major/CS" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.students -and $response.students.Count -ge 0) {
        Write-Host "PASS: Successfully fetched students by major. Count: $($response.students.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Invalid response for students by major." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Get students by major failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Search Students ---
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/students/search?q=Test" -Method Get -Headers @{"Content-Type" = "application/json" }
    if ($response -and $response.students -and $response.students.Count -ge 0) {
        Write-Host "PASS: Successfully searched students. Count: $($response.students.Count)" -ForegroundColor Green
    }
    else {
        Write-Host "FAIL: Invalid response for student search." -ForegroundColor Red
    }
}
catch {
    Write-Host "FAIL: Search students failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Delete Student by UUID ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/students/$studentUuid" -Method Delete -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Delete student by UUID successful." -ForegroundColor Green
}
catch {
    Write-Host "FAIL: Delete student failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "=== Student Info Service API Tests Completed ===" -ForegroundColor Cyan 