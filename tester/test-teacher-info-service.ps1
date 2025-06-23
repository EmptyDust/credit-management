# Teacher Info Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Create Teacher ---
$teacherUsername = "TEACHER_$(Get-Random)"
$createBody = @{ username = $teacherUsername; name = "Test Teacher"; department = "CS"; title = "Lecturer"; specialty = "AI"; contact = "9876543210"; email = "$teacherUsername@test.com" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/teachers" -Method Post -Headers @{"Content-Type" = "application/json" } -Body $createBody
    Write-Host "PASS: Teacher created." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create teacher failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Teacher ---
try {
    $teacher = Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUsername" -Method Get -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Get teacher successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get teacher failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Teacher ---
$updateBody = @{ department = "Math" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUsername" -Method Put -Headers @{"Content-Type" = "application/json" } -Body $updateBody
    Write-Host "PASS: Update teacher successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Update teacher failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Delete Teacher ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/teachers/$teacherUsername" -Method Delete -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Delete teacher successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Delete teacher failed: $($_.Exception.Message)" -ForegroundColor Red
} 