#!/bin/bash

# Redis启动脚本
echo "启动Redis服务..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker未运行，请先启动Docker"
    exit 1
fi

# 启动Redis容器
docker-compose up -d redis

# 等待Redis启动
echo "等待Redis启动..."
sleep 5

# 检查Redis状态
if docker-compose ps redis | grep -q "Up"; then
    echo "Redis启动成功！"
    echo "Redis地址: localhost:6379"
    echo "使用以下命令连接Redis:"
    echo "  docker exec -it credit_management_redis redis-cli"
    echo "  或者使用: ./redis/redis-cli.sh"
else
    echo "Redis启动失败，请检查日志:"
    docker-compose logs redis
    exit 1
fi 