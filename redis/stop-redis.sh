#!/bin/bash

# Redis停止脚本
echo "停止Redis服务..."

# 停止Redis容器
docker-compose stop redis

# 检查Redis状态
if docker-compose ps redis | grep -q "Up"; then
    echo "Redis停止失败"
    exit 1
else
    echo "Redis已停止"
fi 