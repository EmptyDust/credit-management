# Permission Management Service API Test Script
$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8000"

# --- Login as Admin User ---
$adminUsername = "admin"
$adminPassword = "adminpassword"

$loginBody = @{ username = $adminUsername; password = $adminPassword } | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -Headers @{"Content-Type" = "application/json" }
    $jwtToken = $response.token
    $authHeaders = @{ "Authorization" = "Bearer $jwtToken"; "Content-Type" = "application/json" }
    Write-Host "PASS: Admin login successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Admin login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# --- Initialize Permissions First ---
try {
    Invoke-RestMethod -Uri "$baseUrl/api/init-permissions" -Method Post -Headers $authHeaders
    Write-Host "PASS: Permissions initialized successfully." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Initialize permissions failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Role Management ---
# Create Role
$roleBody = @{ name = "testrole_$(Get-Random)"; description = "Test role" } | ConvertTo-Json
try {
    $roleResp = Invoke-RestMethod -Uri "$baseUrl/api/permissions/roles" -Method Post -Headers $authHeaders -Body $roleBody
    $roleId = $roleResp.id
    Write-Host "PASS: Role created. ID: $roleId" -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create role failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Get All Roles
try {
    $roles = Invoke-RestMethod -Uri "$baseUrl/api/permissions/roles" -Method Get -Headers $authHeaders
    Write-Host "PASS: Get all roles successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get all roles failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Update Role
$updateRoleBody = @{ name = "updatedrole_$(Get-Random)"; description = "Updated test role" } | ConvertTo-Json
try {
    Invoke-RestMethod -Uri "$baseUrl/api/permissions/roles/$roleId" -Method Put -Headers $authHeaders -Body $updateRoleBody
    Write-Host "PASS: Update role successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Update role failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Delete Role
try {
    Invoke-RestMethod -Uri "$baseUrl/api/permissions/roles/$roleId" -Method Delete -Headers $authHeaders
    Write-Host "PASS: Delete role successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Delete role failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Permission Management ---
# Create Permission
$permBody = @{ name = "testperm_$(Get-Random)"; description = "Test permission"; resource = "test"; action = "read" } | ConvertTo-Json
try {
    $permResp = Invoke-RestMethod -Uri "$baseUrl/api/permissions" -Method Post -Headers $authHeaders -Body $permBody
    $permId = $permResp.id
    Write-Host "PASS: Permission created. ID: $permId" -ForegroundColor Green
} catch {
    Write-Host "FAIL: Create permission failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Get All Permissions
try {
    $perms = Invoke-RestMethod -Uri "$baseUrl/api/permissions" -Method Get -Headers $authHeaders
    Write-Host "PASS: Get all permissions successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Get all permissions failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Delete Permission
try {
    Invoke-RestMethod -Uri "$baseUrl/api/permissions/$permId" -Method Delete -Headers $authHeaders
    Write-Host "PASS: Delete permission successful." -ForegroundColor Green
} catch {
    Write-Host "FAIL: Delete permission failed: $($_.Exception.Message)" -ForegroundColor Red
}

# --- Assignment Management & Query Endpoints ---
# (可补充：分配角色/权限给用户、角色分配权限、查询等) 