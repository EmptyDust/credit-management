#!/bin/bash

echo "🔍 检查域名配置..."

# 获取服务器IP
SERVER_IP=$(curl -s ifconfig.me)
echo "🌐 服务器公网IP: $SERVER_IP"

# 检查域名解析
echo "🔍 检查域名解析..."
if nslookup lab.emptydust.com > /dev/null 2>&1; then
    DOMAIN_IP=$(nslookup lab.emptydust.com | grep -A1 "Name:" | tail -1 | awk '{print $2}')
    echo "✅ 域名解析正常: lab.emptydust.com -> $DOMAIN_IP"
    
    if [ "$DOMAIN_IP" = "$SERVER_IP" ]; then
        echo "✅ 域名解析指向正确IP"
    else
        echo "⚠️  域名解析IP ($DOMAIN_IP) 与服务器IP ($SERVER_IP) 不匹配"
        echo "   请检查DNS配置"
    fi
else
    echo "❌ 域名解析失败，请检查DNS配置"
fi

# 检查端口80是否开放
echo "🔍 检查端口80..."
if netstat -tlnp | grep :80 > /dev/null; then
    echo "✅ 端口80已开放"
else
    echo "❌ 端口80未开放，请检查防火墙配置"
fi

# 检查Docker服务状态
echo "🔍 检查Docker服务..."
if docker info > /dev/null 2>&1; then
    echo "✅ Docker服务运行正常"
else
    echo "❌ Docker服务未运行"
fi

# 检查容器状态
echo "🔍 检查容器状态..."
if docker-compose ps | grep -q "Up"; then
    echo "✅ 容器运行正常"
    docker-compose ps
else
    echo "❌ 容器未运行，请执行 ./start.sh"
fi

echo ""
echo "📋 配置检查完成"
echo "🌐 如果一切正常，请访问: http://lab.emptydust.com" 