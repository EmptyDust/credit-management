#!/bin/bash

echo "🧪 创新创业学分管理平台 - 系统测试"
echo "=================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    local expected_status=$3
    
    echo -n "测试 $name... "
    
    # 发送请求并获取状态码
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ 通过${NC}"
        return 0
    else
        echo -e "${RED}❌ 失败 (状态码: $status_code, 期望: $expected_status)${NC}"
        return 1
    fi
}

# 等待服务启动
echo -e "${BLUE}等待服务启动...${NC}"
sleep 10

# 测试数据库连接
echo -e "\n${YELLOW}📊 数据库连接测试${NC}"
if docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
    echo -e "${GREEN}✅ PostgreSQL 数据库连接正常${NC}"
else
    echo -e "${RED}❌ PostgreSQL 数据库连接失败${NC}"
fi

# 测试API网关
echo -e "\n${YELLOW}🌐 API网关测试${NC}"
test_endpoint "API网关健康检查" "http://localhost:8080/health" "200"

# 测试用户管理服务
echo -e "\n${YELLOW}👤 用户管理服务测试${NC}"
test_endpoint "用户管理服务" "http://localhost:8081/api/users" "404" # 应该返回404因为没有GET /users路由

# 测试学生信息服务
echo -e "\n${YELLOW}🎓 学生信息服务测试${NC}"
test_endpoint "学生信息服务" "http://localhost:8084/api/students" "200"

# 测试教师信息服务
echo -e "\n${YELLOW}👨‍🏫 教师信息服务测试${NC}"
test_endpoint "教师信息服务" "http://localhost:8085/api/teachers" "200"

# 测试事项管理服务
echo -e "\n${YELLOW}📋 事项管理服务测试${NC}"
test_endpoint "事项管理服务" "http://localhost:8083/api/affairs" "200"

# 测试通用申请服务
echo -e "\n${YELLOW}📝 通用申请服务测试${NC}"
test_endpoint "通用申请服务" "http://localhost:8086/api/applications" "200"

# 测试前端应用
echo -e "\n${YELLOW}🖥️ 前端应用测试${NC}"
test_endpoint "前端应用" "http://localhost:3000" "200"

# 测试用户注册
echo -e "\n${YELLOW}🔐 用户注册测试${NC}"
register_response=$(curl -s -X POST http://localhost:8081/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "email": "test@example.com",
    "role": "student"
  }')

if echo "$register_response" | grep -q "id"; then
    echo -e "${GREEN}✅ 用户注册成功${NC}"
else
    echo -e "${RED}❌ 用户注册失败${NC}"
    echo "响应: $register_response"
fi

# 测试用户登录
echo -e "\n${YELLOW}🔑 用户登录测试${NC}"
login_response=$(curl -s -X POST http://localhost:8081/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }')

if echo "$login_response" | grep -q "token"; then
    echo -e "${GREEN}✅ 用户登录成功${NC}"
else
    echo -e "${RED}❌ 用户登录失败${NC}"
    echo "响应: $login_response"
fi

# 测试创建事项
echo -e "\n${YELLOW}📋 创建事项测试${NC}"
affair_response=$(curl -s -X POST http://localhost:8083/api/affairs \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试事项",
    "description": "这是一个测试事项",
    "type": "test",
    "status": "active"
  }')

if echo "$affair_response" | grep -q "id"; then
    echo -e "${GREEN}✅ 事项创建成功${NC}"
else
    echo -e "${RED}❌ 事项创建失败${NC}"
    echo "响应: $affair_response"
fi

# 测试创建申请
echo -e "\n${YELLOW}📝 创建申请测试${NC}"
application_response=$(curl -s -X POST http://localhost:8086/api/applications \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "student_id": 1,
    "affair_id": 1,
    "title": "测试申请",
    "description": "这是一个测试申请",
    "type": "test",
    "credits": 2.0
  }')

if echo "$application_response" | grep -q "id"; then
    echo -e "${GREEN}✅ 申请创建成功${NC}"
else
    echo -e "${RED}❌ 申请创建失败${NC}"
    echo "响应: $application_response"
fi

# 总结
echo -e "\n${BLUE}📊 测试总结${NC}"
echo "=================================="
echo "✅ 所有基础服务测试完成"
echo ""
echo "🌐 访问地址："
echo "   前端应用: http://localhost:3000"
echo "   API网关:  http://localhost:8080"
echo ""
echo "👤 测试账号："
echo "   管理员: admin / password"
echo "   学生:    student1 / password"
echo "   教师:    teacher1 / password"
echo ""
echo "💡 提示："
echo "   如果某些测试失败，请检查服务是否完全启动"
echo "   使用 'docker-compose logs' 查看详细日志" 