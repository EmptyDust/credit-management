#!/bin/bash

# 数据库恢复脚本

set -e

# 配置
DB_NAME="credit_management"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"
BACKUP_DIR="./backups"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# 显示可用的备份文件
show_backups() {
    print_header "可用的备份文件"
    
    if [ ! -d "$BACKUP_DIR" ]; then
        print_error "备份目录不存在: $BACKUP_DIR"
        exit 1
    fi
    
    local backups=($(ls -t "$BACKUP_DIR"/credit_management_*.sql.gz 2>/dev/null))
    
    if [ ${#backups[@]} -eq 0 ]; then
        print_warning "没有找到备份文件"
        exit 1
    fi
    
    echo "可用的备份文件:"
    for i in "${!backups[@]}"; do
        local file=$(basename "${backups[$i]}")
        local size=$(du -h "${backups[$i]}" | cut -f1)
        local date=$(echo "$file" | sed 's/credit_management_\(.*\)\.sql\.gz/\1/')
        echo "  $((i+1)). $file ($size) - $date"
    done
}

# 恢复数据库
restore_database() {
    local backup_file="$1"
    
    if [ ! -f "$backup_file" ]; then
        print_error "备份文件不存在: $backup_file"
        exit 1
    fi
    
    print_header "开始恢复数据库"
    print_warning "这将覆盖现有的数据库数据！"
    read -p "确认继续吗？(y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_message "恢复操作已取消"
        exit 0
    fi
    
    print_message "恢复文件: $backup_file"
    
    # 解压备份文件
    local temp_file=$(mktemp)
    print_message "解压备份文件..."
    gunzip -c "$backup_file" > "$temp_file"
    
    # 执行恢复
    print_message "执行数据库恢复..."
    PGPASSWORD=password psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d postgres \
        -f "$temp_file"
    
    if [ $? -eq 0 ]; then
        print_message "✓ 数据库恢复成功"
    else
        print_error "✗ 数据库恢复失败"
        rm -f "$temp_file"
        exit 1
    fi
    
    # 清理临时文件
    rm -f "$temp_file"
}

# 主函数
main() {
    print_header "数据库恢复工具"
    
    # 检查参数
    if [ $# -eq 0 ]; then
        show_backups
        echo
        read -p "请选择要恢复的备份文件编号: " choice
        
        if [[ ! "$choice" =~ ^[0-9]+$ ]]; then
            print_error "无效的选择"
            exit 1
        fi
        
        local backups=($(ls -t "$BACKUP_DIR"/credit_management_*.sql.gz 2>/dev/null))
        local selected_index=$((choice-1))
        
        if [ $selected_index -lt 0 ] || [ $selected_index -ge ${#backups[@]} ]; then
            print_error "无效的备份文件编号"
            exit 1
        fi
        
        restore_database "${backups[$selected_index]}"
    else
        # 直接指定备份文件
        restore_database "$1"
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 