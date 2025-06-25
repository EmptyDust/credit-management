# 测试用户搜索API
Write-Host "=== 测试用户搜索API ===" -ForegroundColor Green

# 1. 注册测试用户
Write-Host "`n1. 注册测试用户..." -ForegroundColor Yellow

# 注册学生A
$studentA = @{
    username = "searchstudent1"
    password = "Password123"
    email = "search_student1@example.com"
    phone = "13800138101"
    real_name = "搜索学生1"
    user_type = "student"
    student_id = "20230101"
    college = "计算机学院"
    major = "软件工程"
    class = "软件2301"
    grade = "2023"
}

$registerResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/register" -Method POST -ContentType "application/json" -Body ($studentA | ConvertTo-Json)

if ($registerResponse.code -eq 0) {
    Write-Host "学生A注册成功" -ForegroundColor Green
} else {
    Write-Host "学生A注册失败: $($registerResponse.message)" -ForegroundColor Red
    exit 1
}

# 注册学生B
$studentB = @{
    username = "searchstudent2"
    password = "Password123"
    email = "search_student2@example.com"
    phone = "13800138102"
    real_name = "搜索学生2"
    user_type = "student"
    student_id = "20230102"
    college = "信息学院"
    major = "计算机科学"
    class = "计科2301"
    grade = "2023"
}

$registerResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/users/register" -Method POST -ContentType "application/json" -Body ($studentB | ConvertTo-Json)

if ($registerResponse.code -eq 0) {
    Write-Host "学生B注册成功" -ForegroundColor Green
} else {
    Write-Host "学生B注册失败: $($registerResponse.message)" -ForegroundColor Red
    exit 1
}

# 2. 登录获取token
Write-Host "`n2. 登录获取token..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body @"
{
    "username": "searchstudent1",
    "password": "Password123"
}
"@

if ($loginResponse.code -eq 0) {
    $token = $loginResponse.data.token
    Write-Host "登录成功，获取到token" -ForegroundColor Green
} else {
    Write-Host "登录失败: $($loginResponse.message)" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# 3. 测试通用用户搜索
Write-Host "`n3. 测试通用用户搜索..." -ForegroundColor Yellow
$searchResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/search/users?query=搜索学生&user_type=student" -Method GET -Headers $headers

if ($searchResponse.code -eq 0) {
    Write-Host "通用用户搜索成功" -ForegroundColor Green
    Write-Host "找到 $($searchResponse.data.total) 个用户" -ForegroundColor Cyan
    foreach ($user in $searchResponse.data.users) {
        Write-Host "  - $($user.username): $($user.real_name)" -ForegroundColor Cyan
    }
} else {
    Write-Host "通用用户搜索失败: $($searchResponse.message)" -ForegroundColor Red
    exit 1
}

# 4. 测试学生搜索
Write-Host "`n4. 测试学生搜索..." -ForegroundColor Yellow
$studentSearchResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/students/search?query=搜索学生" -Method GET -Headers $headers

if ($studentSearchResponse.code -eq 0) {
    Write-Host "学生搜索成功" -ForegroundColor Green
    Write-Host "找到 $($studentSearchResponse.data.total) 个学生" -ForegroundColor Cyan
    foreach ($student in $studentSearchResponse.data.users) {
        Write-Host "  - $($student.username): $($student.real_name) ($($student.college))" -ForegroundColor Cyan
    }
} else {
    Write-Host "学生搜索失败: $($studentSearchResponse.message)" -ForegroundColor Red
    exit 1
}

# 5. 测试按学院搜索
Write-Host "`n5. 测试按学院搜索..." -ForegroundColor Yellow
$collegeSearchResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/students/college/计算机学院" -Method GET -Headers $headers

if ($collegeSearchResponse.code -eq 0) {
    Write-Host "按学院搜索成功" -ForegroundColor Green
    Write-Host "计算机学院有 $($collegeSearchResponse.data.total) 个学生" -ForegroundColor Cyan
    foreach ($student in $collegeSearchResponse.data.users) {
        Write-Host "  - $($student.username): $($student.real_name)" -ForegroundColor Cyan
    }
} else {
    Write-Host "按学院搜索失败: $($collegeSearchResponse.message)" -ForegroundColor Red
    exit 1
}

# 6. 测试按专业搜索
Write-Host "`n6. 测试按专业搜索..." -ForegroundColor Yellow
$majorSearchResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/students/major/软件工程" -Method GET -Headers $headers

if ($majorSearchResponse.code -eq 0) {
    Write-Host "按专业搜索成功" -ForegroundColor Green
    Write-Host "软件工程专业有 $($majorSearchResponse.data.total) 个学生" -ForegroundColor Cyan
    foreach ($student in $majorSearchResponse.data.users) {
        Write-Host "  - $($student.username): $($student.real_name)" -ForegroundColor Cyan
    }
} else {
    Write-Host "按专业搜索失败: $($majorSearchResponse.message)" -ForegroundColor Red
    exit 1
}

# 7. 测试分页功能
Write-Host "`n7. 测试分页功能..." -ForegroundColor Yellow
$pageSearchResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/search/users?user_type=student&page=1&page_size=5" -Method GET -Headers $headers

if ($pageSearchResponse.code -eq 0) {
    Write-Host "分页搜索成功" -ForegroundColor Green
    Write-Host "总用户数: $($pageSearchResponse.data.total)" -ForegroundColor Cyan
    Write-Host "当前页: $($pageSearchResponse.data.page)" -ForegroundColor Cyan
    Write-Host "每页大小: $($pageSearchResponse.data.page_size)" -ForegroundColor Cyan
    Write-Host "总页数: $($pageSearchResponse.data.total_pages)" -ForegroundColor Cyan
    Write-Host "当前页用户数: $($pageSearchResponse.data.users.Count)" -ForegroundColor Cyan
} else {
    Write-Host "分页搜索失败: $($pageSearchResponse.message)" -ForegroundColor Red
    exit 1
}

# 8. 测试权限控制（学生搜索教师）
Write-Host "`n8. 测试权限控制（学生搜索教师）..." -ForegroundColor Yellow
$teacherSearchResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/search/users?user_type=teacher" -Method GET -Headers $headers

if ($teacherSearchResponse.code -eq 0) {
    Write-Host "教师搜索成功" -ForegroundColor Green
    Write-Host "找到 $($teacherSearchResponse.data.total) 个教师" -ForegroundColor Cyan
    foreach ($teacher in $teacherSearchResponse.data.users) {
        Write-Host "  - $($teacher.username): $($teacher.real_name)" -ForegroundColor Cyan
    }
} else {
    Write-Host "教师搜索失败: $($teacherSearchResponse.message)" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ 用户搜索API测试通过！" -ForegroundColor Green
Write-Host "`n=== 测试完成 ===" -ForegroundColor Green 