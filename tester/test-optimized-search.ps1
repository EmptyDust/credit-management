# Test optimized user search API
# Only keep the unified /search/users API, remove all redundant search routes

$baseUrl = "http://localhost:8080/api"
$gatewayUrl = "http://localhost:8080"

Write-Host "=== Optimized User Search API Test ===" -ForegroundColor Green

# 1. Admin login
Write-Host "\n1. Admin login" -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "adminpassword"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$gatewayUrl/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
if ($null -eq $loginResponse.data.token -or $loginResponse.data.token -eq "") {
    Write-Host "Admin login failed. Response:" -ForegroundColor Red
    $loginResponse | ConvertTo-Json
    exit 1
}
$adminToken = $loginResponse.data.token
Write-Host "Admin login success. Token: $($adminToken.Substring(0, 20))..." -ForegroundColor Green

# 2. General user search - all users
Write-Host "\n2. General user search - all users" -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($searchResponse.data.total) users." -ForegroundColor Green

# 3. General user search - filter by user_type
Write-Host "\n3. General user search - filter by user_type" -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users?user_type=student" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($searchResponse.data.total) students." -ForegroundColor Green

$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users?user_type=teacher" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($searchResponse.data.total) teachers." -ForegroundColor Green

# 4. General user search - keyword search
Write-Host "\n4. General user search - keyword search" -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users?query=admin" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($searchResponse.data.total) users with keyword 'admin'." -ForegroundColor Green

# 5. General user search - filter by college
Write-Host "\n5. General user search - filter by college" -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users?college=Computer%20Science" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($searchResponse.data.total) users in college 'Computer Science'." -ForegroundColor Green

# 6. General user search - filter by department
Write-Host "\n6. General user search - filter by department" -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users?department=CS" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($searchResponse.data.total) users in department 'CS'." -ForegroundColor Green

# 7. General user search - combined filters
Write-Host "\n7. General user search - combined filters" -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users?user_type=student&college=Computer%20Science&status=active" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($searchResponse.data.total) active students in 'Computer Science' college." -ForegroundColor Green

# 8. Pagination test
Write-Host "\n8. Pagination test" -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/search/users?page=1&page_size=5" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Page $($searchResponse.data.page), $($searchResponse.data.page_size) per page, $($searchResponse.data.total_pages) total pages." -ForegroundColor Green

# 9. Student list API
Write-Host "\n9. Student list API" -ForegroundColor Yellow
$studentsResponse = Invoke-RestMethod -Uri "$baseUrl/students" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($studentsResponse.data.users.Count) students." -ForegroundColor Green

# 10. Teacher list API
Write-Host "\n10. Teacher list API" -ForegroundColor Yellow
$teachersResponse = Invoke-RestMethod -Uri "$baseUrl/teachers" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Found $($teachersResponse.data.users.Count) teachers." -ForegroundColor Green

# 11. Student stats API
Write-Host "\n11. Student stats API" -ForegroundColor Yellow
$studentStatsResponse = Invoke-RestMethod -Uri "$baseUrl/students/stats" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Student stats: total $($studentStatsResponse.data.total_students), active $($studentStatsResponse.data.active_students)." -ForegroundColor Green

# 12. Teacher stats API
Write-Host "\n12. Teacher stats API" -ForegroundColor Yellow
$teacherStatsResponse = Invoke-RestMethod -Uri "$baseUrl/teachers/stats" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
Write-Host "Teacher stats: total $($teacherStatsResponse.data.total_teachers), active $($teacherStatsResponse.data.active_teachers)." -ForegroundColor Green

# 13. Removed routes should return 404
Write-Host "\n13. Removed routes should return 404" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/students/search" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
    Write-Host "ERROR: /students/search route still exists!" -ForegroundColor Red
} catch {
    Write-Host "OK: /students/search route removed, returns 404." -ForegroundColor Green
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/teachers/search" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
    Write-Host "ERROR: /teachers/search route still exists!" -ForegroundColor Red
} catch {
    Write-Host "OK: /teachers/search route removed, returns 404." -ForegroundColor Green
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/students/college/Computer%20Science" -Method GET -Headers @{Authorization = "Bearer $adminToken"}
    Write-Host "ERROR: /students/college/{college} route still exists!" -ForegroundColor Red
} catch {
    Write-Host "OK: /students/college/{college} route removed, returns 404." -ForegroundColor Green
}

Write-Host "\n=== Optimized User Search API Test Complete ===" -ForegroundColor Green
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "- Removed redundant search routes" -ForegroundColor White
Write-Host "- Kept unified /search/users API" -ForegroundColor White
Write-Host "- Kept necessary list and stats APIs" -ForegroundColor White
Write-Host "- Simplified API structure, improved maintainability" -ForegroundColor White 