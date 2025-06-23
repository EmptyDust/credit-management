#!/bin/bash

echo "🚀 启动创新创业学分管理平台..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 停止现有容器
echo "🛑 停止现有容器..."
docker-compose down

# 构建并启动服务
echo "🔨 构建并启动服务..."
docker-compose up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose ps

echo ""
echo "✅ 服务启动完成！"
echo ""
echo "🌐 访问地址："
echo "   前端界面: http://lab.emptydust.com"
echo "   API网关: http://lab.emptydust.com/api"
echo ""
echo "📋 服务端口："
echo "   前端 (Nginx): 80"
echo "   API网关: 8080"
echo "   用户管理: 8081"
echo "   事项管理: 8083"
echo "   学生信息: 8084"
echo "   教师信息: 8085"
echo "   申请管理: 8086"
echo "   PostgreSQL: 5432"
echo ""
echo "🔧 管理命令："
echo "   查看状态: ./status.sh"
echo "   停止服务: ./stop.sh"
echo "   查看日志: docker-compose logs -f"
echo ""
echo "⚠️  请确保已将 lab.emptydust.com 域名解析到本服务器IP" 