#!/bin/bash

echo "🛑 停止创新创业学分管理平台..."

# 停止并删除容器
docker-compose down

echo "✅ 服务已停止！"
echo ""
echo "💡 提示："
echo "   重新启动: ./start.sh"
echo "   查看日志: docker-compose logs -f"
echo "   清理数据: docker-compose down -v" 