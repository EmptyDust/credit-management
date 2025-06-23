#!/bin/bash

echo "📊 创新创业学分管理平台状态检查"
echo "=================================="

# 检查Docker服务状态
echo ""
echo "🐳 Docker服务状态："
if docker info >/dev/null 2>&1; then
    echo "✅ Docker运行正常"
else
    echo "❌ Docker未运行"
    exit 1
fi

# 检查容器状态
echo ""
echo "📦 容器状态："
docker-compose ps

# 检查端口占用
echo ""
echo "🔌 端口占用情况："
echo "   8080 (API网关): $(netstat -tlnp 2>/dev/null | grep :8080 || echo '未占用')"
echo "   8081 (用户服务): $(netstat -tlnp 2>/dev/null | grep :8081 || echo '未占用')"
echo "   8083 (事项服务): $(netstat -tlnp 2>/dev/null | grep :8083 || echo '未占用')"
echo "   8084 (学生服务): $(netstat -tlnp 2>/dev/null | grep :8084 || echo '未占用')"
echo "   8085 (教师服务): $(netstat -tlnp 2>/dev/null | grep :8085 || echo '未占用')"
echo "   8086 (申请服务): $(netstat -tlnp 2>/dev/null | grep :8086 || echo '未占用')"
echo "   3000 (前端):     $(netstat -tlnp 2>/dev/null | grep :3000 || echo '未占用')"
echo "   5432 (数据库):   $(netstat -tlnp 2>/dev/null | grep :5432 || echo '未占用')"

# 检查服务健康状态
echo ""
echo "🏥 服务健康检查："
services=("api-gateway" "user-management-service" "student-info-service" "teacher-info-service" "affair-management-service" "general-application-service" "frontend")

for service in "${services[@]}"; do
    if docker-compose ps | grep -q "$service.*Up"; then
        echo "✅ $service: 运行中"
    else
        echo "❌ $service: 未运行"
    fi
done

echo ""
echo "🌐 访问地址："
echo "   前端应用: http://localhost:3000"
echo "   API网关:  http://localhost:8080"
echo "   数据库:   localhost:5432"

echo ""
echo "👤 默认账号："
echo "   管理员: admin / password"
echo "   学生:    student1 / password"
echo "   教师:    teacher1 / password" 