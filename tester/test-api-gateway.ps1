# 该文件将被拆分为多个微服务测试脚本。请参考新生成的各微服务测试脚本。
# 主脚本内容将被替换为调用各子脚本的逻辑。

# Example: 调用各子脚本
& "$PSScriptRoot/test-auth-service.ps1"
& "$PSScriptRoot/test-user-management-service.ps1"
& "$PSScriptRoot/test-permission-service.ps1"
& "$PSScriptRoot/test-student-info-service.ps1"
& "$PSScriptRoot/test-teacher-info-service.ps1"
& "$PSScriptRoot/test-affair-management-service.ps1"
& "$PSScriptRoot/test-application-management-service.ps1"

Write-Host "`n`n=== All Microservice Tests Completed ===" -ForegroundColor Cyan 