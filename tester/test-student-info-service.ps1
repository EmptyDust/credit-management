# Student Info Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Create Student ---
$studentId = "STU_$(Get-Random)"
$createBody = @{ username = $studentId; student_id = $studentId; name = "Test Student"; college = "Engineering"; major = "CS"; class = "2025"; contact = "1234567890"; email = "$studentId@test.com"; grade = "2021" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/students" -Method Post -Headers @{"Content-Type" = "application/json" } -Body $createBody
    Write-Host "PASS: Student created." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create student failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Student ---
try {
    $student = Invoke-RestMethod -Uri "$baseUrl/api/students/$studentId" -Method Get -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Get student successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get student failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Student ---
$updateBody = @{ college = "Science" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/students/$studentId" -Method Put -Headers @{"Content-Type" = "application/json" } -Body $updateBody
    Write-Host "PASS: Update student successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Update student failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Delete Student ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/students/$studentId" -Method Delete -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Delete student successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Delete student failed: $($_.Exception.Message)" -ForegroundColor Red
} 