#!/bin/bash

# Redis CLI工具脚本
# 使用方法: ./redis-cli.sh [command]

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-}

if [ -n "$REDIS_PASSWORD" ]; then
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD "$@"
else
    redis-cli -h $REDIS_HOST -p $REDIS_PORT "$@"
fi 