@echo off
echo ========================================
echo 创新学分管理系统启动脚本
echo ========================================

echo.
echo 检查Docker Desktop状态...
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] Docker Desktop未运行或未安装
    echo 请确保Docker Desktop已启动
    echo 如果未安装，请访问: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo [成功] Docker Desktop正在运行

echo.
echo 检查Docker Compose...
docker-compose version >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] Docker Compose未安装
    pause
    exit /b 1
)

echo [成功] Docker Compose可用

echo.
echo 停止现有容器...
docker-compose down

echo.
echo 清理旧镜像...
docker system prune -f

echo.
echo 构建并启动所有服务...
docker-compose up --build -d

if %errorlevel% neq 0 (
    echo [错误] 启动失败
    echo 请检查错误信息并重试
    pause
    exit /b 1
)

echo.
echo [成功] 所有服务已启动
echo.
echo 服务访问地址:
echo - 前端应用: http://localhost:3000
echo - API网关: http://localhost:8000
echo - 用户管理服务: http://localhost:8080
echo - 认证服务: http://localhost:8081
echo - 申请管理服务: http://localhost:8082
echo - 事务管理服务: http://localhost:8083
echo - 学生信息服务: http://localhost:8084
echo - 教师信息服务: http://localhost:8085
echo - PostgreSQL数据库: localhost:5432
echo.
echo 查看服务状态: docker-compose ps
echo 查看日志: docker-compose logs -f [服务名]
echo 停止服务: docker-compose down
echo.
echo 系统启动完成！
pause 