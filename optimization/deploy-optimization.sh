#!/bin/bash
# ========================================
# Debian内核优化部署脚本
# 自动应用系统级和容器级优化配置
# ========================================

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "此脚本需要root权限运行"
        log_info "请使用: sudo $0"
        exit 1
    fi
}

# 检查系统版本
check_system() {
    log_info "检查系统版本..."

    if [ ! -f /etc/os-release ]; then
        log_error "无法检测系统版本"
        exit 1
    fi

    . /etc/os-release

    if [ "$ID" != "debian" ]; then
        log_warn "此脚本针对Debian系统优化，当前系统: $ID"
        read -p "是否继续? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    log_info "系统: $PRETTY_NAME"
    log_info "内核: $(uname -r)"
}

# 备份现有配置
backup_configs() {
    log_info "备份现有配置..."

    BACKUP_DIR="/root/kernel-optimization-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"

    # 备份sysctl配置
    if [ -f /etc/sysctl.conf ]; then
        cp /etc/sysctl.conf "$BACKUP_DIR/"
    fi

    # 备份sysctl.d目录
    if [ -d /etc/sysctl.d ]; then
        cp -r /etc/sysctl.d "$BACKUP_DIR/"
    fi

    # 备份limits配置
    if [ -f /etc/security/limits.conf ]; then
        cp /etc/security/limits.conf "$BACKUP_DIR/"
    fi

    log_info "配置已备份到: $BACKUP_DIR"
}

# 应用sysctl优化
apply_sysctl() {
    log_info "应用内核参数优化..."

    SYSCTL_FILE="/etc/sysctl.d/99-credit-management.conf"

    if [ ! -f "./optimization/sysctl-optimization.conf" ]; then
        log_error "找不到优化配置文件: ./optimization/sysctl-optimization.conf"
        exit 1
    fi

    # 复制配置文件
    cp ./optimization/sysctl-optimization.conf "$SYSCTL_FILE"

    # 应用配置
    sysctl -p "$SYSCTL_FILE"

    log_info "内核参数优化已应用"
}

# 配置透明大页
configure_thp() {
    log_info "配置透明大页（THP）..."

    # PostgreSQL推荐禁用透明大页
    if [ -f /sys/kernel/mm/transparent_hugepage/enabled ]; then
        echo madvise > /sys/kernel/mm/transparent_hugepage/enabled
        echo madvise > /sys/kernel/mm/transparent_hugepage/defrag
        log_info "透明大页已设置为 madvise 模式"

        # 永久生效（添加到rc.local或systemd服务）
        if [ ! -f /etc/systemd/system/disable-thp.service ]; then
            cat > /etc/systemd/system/disable-thp.service <<EOF
[Unit]
Description=Disable Transparent Huge Pages (THP)
DefaultDependencies=no
After=sysinit.target local-fs.target
Before=postgresql.service

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'echo madvise > /sys/kernel/mm/transparent_hugepage/enabled'
ExecStart=/bin/sh -c 'echo madvise > /sys/kernel/mm/transparent_hugepage/defrag'

[Install]
WantedBy=basic.target
EOF
            systemctl daemon-reload
            systemctl enable disable-thp.service
            log_info "透明大页配置已永久生效"
        fi
    else
        log_warn "系统不支持透明大页配置"
    fi
}

# 配置文件描述符限制
configure_limits() {
    log_info "配置文件描述符限制..."

    LIMITS_FILE="/etc/security/limits.d/99-credit-management.conf"

    cat > "$LIMITS_FILE" <<EOF
# 学分管理系统 - 文件描述符限制配置
*               soft    nofile          65536
*               hard    nofile          65536
*               soft    nproc           32768
*               hard    nproc           32768
root            soft    nofile          65536
root            hard    nofile          65536
EOF

    log_info "文件描述符限制已配置"
}

# 创建数据目录
create_data_dirs() {
    log_info "创建数据目录..."

    mkdir -p ./data/{postgres,redis,attachments,avatars}
    chmod 700 ./data/postgres ./data/redis
    chmod 755 ./data/attachments ./data/avatars

    log_info "数据目录已创建"
}

# 验证Docker配置
verify_docker() {
    log_info "验证Docker配置..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装"
        exit 1
    fi

    # 检查Docker存储驱动
    STORAGE_DRIVER=$(docker info --format '{{.Driver}}')
    log_info "Docker存储驱动: $STORAGE_DRIVER"

    if [ "$STORAGE_DRIVER" != "overlay2" ]; then
        log_warn "推荐使用overlay2存储驱动，当前: $STORAGE_DRIVER"
    fi

    # 检查cgroup版本
    CGROUP_VERSION=$(docker info --format '{{.CgroupVersion}}')
    log_info "Cgroup版本: $CGROUP_VERSION"

    if [ "$CGROUP_VERSION" != "2" ]; then
        log_warn "推荐使用cgroup v2，当前: v$CGROUP_VERSION"
    fi
}

# 应用Docker Compose优化
apply_docker_compose() {
    log_info "准备Docker Compose优化配置..."

    if [ ! -f "./optimization/docker-compose.optimized.yml" ]; then
        log_error "找不到优化的Docker Compose配置文件"
        exit 1
    fi

    log_info "优化的Docker Compose配置已准备就绪"
    log_info "使用以下命令启动优化后的服务:"
    log_info "  docker-compose -f optimization/docker-compose.optimized.yml up -d"
}

# 显示优化摘要
show_summary() {
    log_info "========================================="
    log_info "优化配置已完成！"
    log_info "========================================="
    echo
    log_info "已应用的优化:"
    echo "  ✓ 内核参数优化 (sysctl)"
    echo "  ✓ 透明大页配置"
    echo "  ✓ 文件描述符限制"
    echo "  ✓ 数据目录创建"
    echo "  ✓ Docker配置验证"
    echo
    log_info "下一步操作:"
    echo "  1. 重启系统以确保所有配置生效（可选但推荐）:"
    echo "     sudo reboot"
    echo
    echo "  2. 或者，重新加载配置并启动服务:"
    echo "     sudo sysctl --system"
    echo "     docker-compose -f optimization/docker-compose.optimized.yml up -d"
    echo
    echo "  3. 监控系统性能:"
    echo "     docker stats"
    echo "     htop"
    echo
    log_info "配置文件位置:"
    echo "  - 内核参数: /etc/sysctl.d/99-credit-management.conf"
    echo "  - 文件限制: /etc/security/limits.d/99-credit-management.conf"
    echo "  - THP服务: /etc/systemd/system/disable-thp.service"
    echo "  - 备份目录: $BACKUP_DIR"
    echo
}

# 主函数
main() {
    log_info "========================================="
    log_info "Debian内核优化部署脚本"
    log_info "========================================="
    echo

    check_root
    check_system
    backup_configs
    apply_sysctl
    configure_thp
    configure_limits
    create_data_dirs
    verify_docker
    apply_docker_compose

    echo
    show_summary
}

# 运行主函数
main
