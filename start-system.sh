#!/bin/bash

echo "========================================"
echo "创新创业学分管理系统启动脚本"
echo "========================================"
echo

echo "检查Docker状态..."
if ! command -v docker &> /dev/null; then
    echo "错误: Docker未安装或未在PATH中"
    echo "请先安装Docker并启动"
    exit 1
fi

echo "检查Docker是否运行..."
if ! docker ps &> /dev/null; then
    echo "错误: Docker未运行"
    echo "请先启动Docker服务"
    echo "然后重新运行此脚本"
    exit 1
fi

echo "Docker状态正常，开始启动系统..."
echo

echo "1. 停止现有容器..."
docker-compose down

echo "2. 构建镜像..."
docker-compose build

echo "3. 启动所有服务..."
docker-compose up -d

echo
echo "========================================"
echo "系统启动完成！"
echo "========================================"
echo
echo "服务访问地址:"
echo "- 前端应用: http://localhost:3000"
echo "- API网关: http://localhost:8080"
echo "- 数据库: localhost:5432"
echo
echo "服务端口:"
echo "- auth-service: 8081"
echo "- user-management-service: 8084"
echo "- application-management-service: 8085"
echo "- student-info-service: 8082"
echo "- teacher-info-service: 8083"
echo "- affair-management-service: 8087"
echo
echo "查看服务状态: docker-compose ps"
echo "查看日志: docker-compose logs -f"
echo "停止服务: docker-compose down"
echo 