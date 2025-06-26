#!/bin/bash

# 学分活动服务统一检索API测试脚本
# 使用curl测试所有检索功能

BASE_URL="http://localhost:8080/api"
SERVICE_URL="http://localhost:8083"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 打印测试结果
print_result() {
    local test_name="$1"
    local status="$2"
    local response="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓ PASS${NC} - $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAIL${NC} - $test_name"
        echo -e "${YELLOW}Response:${NC} $response"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# 打印标题
print_title() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

# 测试健康检查
test_health_check() {
    print_title "健康检查测试"
    
    # 测试API网关健康检查
    response=$(curl -s -w "%{http_code}" "$BASE_URL/../health")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_result "API网关健康检查" "PASS" "$body"
    else
        print_result "API网关健康检查" "FAIL" "HTTP $http_code: $body"
    fi
    
    # 测试活动服务健康检查
    response=$(curl -s -w "%{http_code}" "$SERVICE_URL/health")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_result "活动服务健康检查" "PASS" "$body"
    else
        print_result "活动服务健康检查" "FAIL" "HTTP $http_code: $body"
    fi
}

# 测试活动搜索API
test_activity_search() {
    print_title "活动搜索API测试"
    
    # 1. 基础搜索测试
    echo -e "${YELLOW}测试1: 基础活动搜索${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/activities?page=1&page_size=5")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "基础活动搜索（未认证）" "PASS" "正确返回401未认证"
    else
        print_result "基础活动搜索（未认证）" "FAIL" "HTTP $http_code: $body"
    fi
    
    # 2. 带认证的搜索测试（需要有效的token）
    echo -e "${YELLOW}测试2: 带认证的活动搜索${NC}"
    # 注意：这里需要有效的token，暂时跳过
    print_result "带认证的活动搜索" "SKIP" "需要有效token"
    
    # 3. 参数验证测试
    echo -e "${YELLOW}测试3: 参数验证${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/activities?page=0&page_size=1000")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "参数验证测试" "PASS" "正确返回401未认证"
    else
        print_result "参数验证测试" "FAIL" "HTTP $http_code: $body"
    fi
}

# 测试申请搜索API
test_application_search() {
    print_title "申请搜索API测试"
    
    # 1. 基础搜索测试
    echo -e "${YELLOW}测试1: 基础申请搜索${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/applications?page=1&page_size=5")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "基础申请搜索（未认证）" "PASS" "正确返回401未认证"
    else
        print_result "基础申请搜索（未认证）" "FAIL" "HTTP $http_code: $body"
    fi
    
    # 2. 带条件的搜索测试
    echo -e "${YELLOW}测试2: 带条件的申请搜索${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/applications?min_credits=1.0&max_credits=5.0&sort_by=submitted_at&sort_order=desc")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "带条件的申请搜索" "PASS" "正确返回401未认证"
    else
        print_result "带条件的申请搜索" "FAIL" "HTTP $http_code: $body"
    fi
}

# 测试参与者搜索API
test_participant_search() {
    print_title "参与者搜索API测试"
    
    # 1. 基础搜索测试
    echo -e "${YELLOW}测试1: 基础参与者搜索${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/participants?page=1&page_size=5")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "基础参与者搜索（未认证）" "PASS" "正确返回401未认证"
    else
        print_result "基础参与者搜索（未认证）" "FAIL" "HTTP $http_code: $body"
    fi
    
    # 2. 按活动搜索参与者
    echo -e "${YELLOW}测试2: 按活动搜索参与者${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/participants?activity_id=test-activity-id&min_credits=1.0")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "按活动搜索参与者" "PASS" "正确返回401未认证"
    else
        print_result "按活动搜索参与者" "FAIL" "HTTP $http_code: $body"
    fi
}

# 测试附件搜索API
test_attachment_search() {
    print_title "附件搜索API测试"
    
    # 1. 基础搜索测试
    echo -e "${YELLOW}测试1: 基础附件搜索${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/attachments?page=1&page_size=5")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "基础附件搜索（未认证）" "PASS" "正确返回401未认证"
    else
        print_result "基础附件搜索（未认证）" "FAIL" "HTTP $http_code: $body"
    fi
    
    # 2. 按文件类型搜索
    echo -e "${YELLOW}测试2: 按文件类型搜索附件${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/attachments?file_type=.pdf&file_category=document&min_size=1000000")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "按文件类型搜索附件" "PASS" "正确返回401未认证"
    else
        print_result "按文件类型搜索附件" "FAIL" "HTTP $http_code: $body"
    fi
}

# 测试错误处理
test_error_handling() {
    print_title "错误处理测试"
    
    # 1. 测试不存在的路由
    echo -e "${YELLOW}测试1: 不存在的搜索路由${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/nonexistent")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "404" ]; then
        print_result "不存在的搜索路由" "PASS" "正确返回404"
    else
        print_result "不存在的搜索路由" "FAIL" "HTTP $http_code: $body"
    fi
    
    # 2. 测试无效的查询参数
    echo -e "${YELLOW}测试2: 无效的查询参数${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/activities?invalid_param=value")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ]; then
        print_result "无效的查询参数" "PASS" "正确返回401未认证"
    else
        print_result "无效的查询参数" "FAIL" "HTTP $http_code: $body"
    fi
}

# 测试API网关路由
test_gateway_routes() {
    print_title "API网关路由测试"
    
    # 1. 测试API网关根路径
    echo -e "${YELLOW}测试1: API网关根路径${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/../")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_result "API网关根路径" "PASS" "成功获取API信息"
    else
        print_result "API网关根路径" "FAIL" "HTTP $http_code: $body"
    fi
    
    # 2. 测试搜索路由是否存在
    echo -e "${YELLOW}测试2: 搜索路由存在性${NC}"
    response=$(curl -s -w "%{http_code}" "$BASE_URL/search/activities")
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" = "401" ] || [ "$http_code" = "200" ]; then
        print_result "搜索路由存在性" "PASS" "路由存在且可访问"
    else
        print_result "搜索路由存在性" "FAIL" "HTTP $http_code: $body"
    fi
}

# 生成测试报告
generate_report() {
    print_title "测试报告"
    
    echo -e "${BLUE}总测试数:${NC} $TOTAL_TESTS"
    echo -e "${GREEN}通过测试:${NC} $PASSED_TESTS"
    echo -e "${RED}失败测试:${NC} $FAILED_TESTS"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 所有测试通过！${NC}"
    else
        echo -e "\n${RED}❌ 有 $FAILED_TESTS 个测试失败${NC}"
    fi
}

# 主函数
main() {
    echo -e "${BLUE}开始测试学分活动服务统一检索API${NC}"
    echo -e "${BLUE}测试时间: $(date)${NC}"
    echo -e "${BLUE}API网关地址: $BASE_URL${NC}"
    echo -e "${BLUE}活动服务地址: $SERVICE_URL${NC}"
    
    # 执行所有测试
    test_health_check
    test_gateway_routes
    test_activity_search
    test_application_search
    test_participant_search
    test_attachment_search
    test_error_handling
    
    # 生成报告
    generate_report
}

# 运行主函数
main 