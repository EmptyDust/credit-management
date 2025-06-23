#!/bin/bash

echo "========================================"
echo "创新学分管理系统测试脚本"
echo "========================================"

echo ""
echo "检查系统状态..."

echo ""
echo "1. 检查Docker容器状态..."
docker-compose ps

echo ""
echo "2. 检查服务健康状态..."

echo "检查API网关..."
if curl -s http://localhost:8000/health > /dev/null; then
    echo "[成功] API网关运行正常"
else
    echo "[警告] API网关可能未启动"
fi

echo "检查用户管理服务..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "[成功] 用户管理服务运行正常"
else
    echo "[警告] 用户管理服务可能未启动"
fi

echo "检查认证服务..."
if curl -s http://localhost:8081/health > /dev/null; then
    echo "[成功] 认证服务运行正常"
else
    echo "[警告] 认证服务可能未启动"
fi

echo "检查申请管理服务..."
if curl -s http://localhost:8082/health > /dev/null; then
    echo "[成功] 申请管理服务运行正常"
else
    echo "[警告] 申请管理服务可能未启动"
fi

echo "检查事务管理服务..."
if curl -s http://localhost:8083/health > /dev/null; then
    echo "[成功] 事务管理服务运行正常"
else
    echo "[警告] 事务管理服务可能未启动"
fi

echo "检查学生信息服务..."
if curl -s http://localhost:8084/health > /dev/null; then
    echo "[成功] 学生信息服务运行正常"
else
    echo "[警告] 学生信息服务可能未启动"
fi

echo "检查教师信息服务..."
if curl -s http://localhost:8085/health > /dev/null; then
    echo "[成功] 教师信息服务运行正常"
else
    echo "[警告] 教师信息服务可能未启动"
fi

echo ""
echo "3. 检查数据库连接..."
if docker-compose exec postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo "[成功] 数据库连接正常"
else
    echo "[错误] 数据库连接失败"
fi

echo ""
echo "4. 检查前端服务..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "[成功] 前端服务运行正常"
else
    echo "[警告] 前端服务可能未启动"
fi

echo ""
echo "========================================"
echo "测试完成"
echo "========================================"
echo ""
echo "如果所有服务都显示正常，系统运行良好"
echo "如果有警告或错误，请检查相应的服务日志"
echo ""
echo "查看详细日志: docker-compose logs [服务名]"
echo "" 