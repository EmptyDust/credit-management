# Affair Management Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Create Affair ---
$affairName = "affair_$(Get-Random)"
$createBody = @{ name = $affairName } | ConvertTo-Json
try {
    $resp = Invoke-RestMethod -Uri "$baseUrl/api/affairs" -Method Post -Headers @{"Content-Type" = "application/json" } -Body $createBody
    $affairId = $resp.id
    Write-Host "PASS: Affair created. ID: $affairId" -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create affair failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Get Affair ---
try {
    $affair = Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId" -Method Get -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Get affair successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get affair failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Update Affair ---
$updateBody = @{ name = "updated_$affairName" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId" -Method Put -Headers @{"Content-Type" = "application/json" } -Body $updateBody
    Write-Host "PASS: Update affair successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Update affair failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Delete Affair ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/affairs/$affairId" -Method Delete -Headers @{"Content-Type" = "application/json" }
    Write-Host "PASS: Delete affair successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Delete affair failed: $($_.Exception.Message)" -ForegroundColor Red
} 