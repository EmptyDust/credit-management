#!/bin/bash

# Docker构建脚本
# 用于构建所有服务的Docker镜像

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# 检查Docker是否运行
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker未运行或无法访问"
        exit 1
    fi
    print_message "Docker检查通过"
}

# 构建单个服务
build_service() {
    local service_name=$1
    local dockerfile_path=$2
    local image_name="credit-management-${service_name}"
    
    print_message "构建服务: $service_name"
    
    if [ ! -f "$dockerfile_path/Dockerfile" ]; then
        print_error "Dockerfile不存在: $dockerfile_path/Dockerfile"
        return 1
    fi
    
    cd "$dockerfile_path"
    docker build -t "$image_name:latest" .
    
    if [ $? -eq 0 ]; then
        print_message "✓ $service_name 构建成功"
    else
        print_error "✗ $service_name 构建失败"
        return 1
    fi
    
    cd - > /dev/null
}

# 主函数
main() {
    print_header "开始构建Docker镜像"
    
    # 检查Docker
    check_docker
    
    # 服务列表（按依赖顺序）
    services=(
        "database:./database"
        "api-gateway:./api-gateway"
        "auth-service:./auth-service"
        "user-service:./user-service"
        "credit-activity-service:./credit-activity-service"
        "frontend:./frontend"
    )
    
    # 构建所有服务
    for service in "${services[@]}"; do
        IFS=':' read -r service_name dockerfile_path <<< "$service"
        build_service "$service_name" "$dockerfile_path"
    done
    
    print_header "构建完成"
    print_message "所有服务构建成功！"
    print_message "使用以下命令启动服务："
    echo "docker-compose up -d"
    echo ""
    print_message "数据库初始化信息："
    echo "- 数据库名: credit_management"
    echo "- 用户名: postgres"
    echo "- 密码: password"
    echo "- 端口: 5432"
    echo ""
    print_message "初始用户："
    echo "- 管理员: admin / password"
    echo "- 学生: student1 / password"
    echo "- 教师: teacher1 / password"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 