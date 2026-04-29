#!/bin/bash

# 数据库备份脚本

set -e

# 配置
DB_NAME="credit_management"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/credit_management_${DATE}.sql"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 检查备份目录
if [ ! -d "$BACKUP_DIR" ]; then
    print_message "创建备份目录: $BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"
fi

# 执行备份
print_message "开始备份数据库: $DB_NAME"
print_message "备份文件: $BACKUP_FILE"

# 使用pg_dump进行备份
PGPASSWORD=password pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --verbose \
    --clean \
    --if-exists \
    --create \
    --no-owner \
    --no-privileges \
    > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
    print_message "✓ 数据库备份成功"
    print_message "备份文件大小: $(du -h "$BACKUP_FILE" | cut -f1)"
    
    # 压缩备份文件
    print_message "压缩备份文件..."
    gzip "$BACKUP_FILE"
    print_message "✓ 备份文件已压缩: ${BACKUP_FILE}.gz"
    
    # 清理旧备份（保留最近7天的备份）
    print_message "清理旧备份文件..."
    find "$BACKUP_DIR" -name "credit_management_*.sql.gz" -mtime +7 -delete
    print_message "✓ 旧备份文件清理完成"
    
else
    print_error "✗ 数据库备份失败"
    exit 1
fi 