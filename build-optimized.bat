@echo off
echo ========================================
echo 开始构建信用管理系统 (优化版本)
echo ========================================

REM 设置Docker构建参数
set DOCKER_BUILDKIT=1
set COMPOSE_DOCKER_CLI_BUILD=1

REM 检查Docker是否运行
echo 检查Docker服务状态...
docker info >nul 2>&1
if errorlevel 1 (
    echo 错误: Docker服务未运行，请启动Docker Desktop
    pause
    exit /b 1
)

REM 清理旧的构建缓存
echo 清理旧的构建缓存...
docker builder prune -f

REM 设置构建超时和资源限制
echo 设置构建环境...
docker system prune -f

REM 逐个构建服务（避免并行构建导致的资源竞争）
echo.
echo 构建PostgreSQL数据库...
docker pull postgres:15-alpine

echo.
echo 构建API网关...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_api_gateway ./api-gateway
if errorlevel 1 (
    echo 错误: API网关构建失败
    pause
    exit /b 1
)

echo.
echo 构建用户管理服务...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_user_management ./user-management-service
if errorlevel 1 (
    echo 错误: 用户管理服务构建失败
    pause
    exit /b 1
)

echo.
echo 构建认证服务...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_auth_service ./auth-service
if errorlevel 1 (
    echo 错误: 认证服务构建失败
    pause
    exit /b 1
)

echo.
echo 构建申请管理服务...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_application_management ./application-management-service
if errorlevel 1 (
    echo 错误: 申请管理服务构建失败
    pause
    exit /b 1
)

echo.
echo 构建事务管理服务...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_affair_management ./affair-management-service
if errorlevel 1 (
    echo 错误: 事务管理服务构建失败
    pause
    exit /b 1
)

echo.
echo 构建学生信息服务...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_student_info ./student-info-service
if errorlevel 1 (
    echo 错误: 学生信息服务构建失败
    pause
    exit /b 1
)

echo.
echo 构建教师信息服务...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_teacher_info ./teacher-info-service
if errorlevel 1 (
    echo 错误: 教师信息服务构建失败
    pause
    exit /b 1
)

echo.
echo 构建前端应用...
docker build --no-cache --memory=2g --cpus=1.0 -t credit_management_frontend ./frontend
if errorlevel 1 (
    echo 错误: 前端应用构建失败
    pause
    exit /b 1
)

echo.
echo ========================================
echo 所有服务构建完成！
echo ========================================

echo.
echo 启动系统...
docker-compose up -d

echo.
echo 等待服务启动...
timeout /t 30 /nobreak >nul

echo.
echo 检查服务状态...
docker-compose ps

echo.
echo ========================================
echo 系统启动完成！
echo 前端访问地址: http://localhost:3000
echo API网关地址: http://localhost:8000
echo ========================================

pause 