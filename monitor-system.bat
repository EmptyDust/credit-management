@echo off
echo ========================================
echo 创新学分管理系统监控脚本
echo ========================================

:monitor_loop
cls
echo 系统监控 - 按 Ctrl+C 退出
echo ========================================
echo 时间: %date% %time%
echo.

echo 1. 容器状态:
docker-compose ps
echo.

echo 2. 系统资源使用情况:
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
echo.

echo 3. 服务响应时间:
echo API网关: 
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8000/health' -TimeoutSec 5; Write-Host '正常' } catch { Write-Host '异常' }"

echo 用户管理服务: 
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8080/health' -TimeoutSec 5; Write-Host '正常' } catch { Write-Host '异常' }"

echo 认证服务: 
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8081/health' -TimeoutSec 5; Write-Host '正常' } catch { Write-Host '异常' }"

echo 申请管理服务: 
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8082/health' -TimeoutSec 5; Write-Host '正常' } catch { Write-Host '异常' }"

echo 事务管理服务: 
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8083/health' -TimeoutSec 5; Write-Host '正常' } catch { Write-Host '异常' }"

echo 学生信息服务: 
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8084/health' -TimeoutSec 5; Write-Host '正常' } catch { Write-Host '异常' }"

echo 教师信息服务: 
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8085/health' -TimeoutSec 5; Write-Host '正常' } catch { Write-Host '异常' }"
echo.

echo 4. 数据库连接状态:
docker-compose exec postgres pg_isready -U postgres >nul 2>&1
if %errorlevel% equ 0 (
    echo 数据库: 正常
) else (
    echo 数据库: 异常
)
echo.

echo 5. 最近日志 (最后10行):
echo API网关日志:
docker-compose logs --tail=10 api-gateway
echo.

echo 6. 磁盘使用情况:
docker system df
echo.

echo 刷新间隔: 30秒
timeout /t 30 /nobreak >nul
goto monitor_loop 