# 统一用户服务 (User Service)

这是合并后的统一用户服务，整合了原有的用户管理、学生信息和教师信息三个服务的功能。

## 功能特性

- 用户注册和管理
- 学生信息管理
- 教师信息管理
- 用户搜索和统计
- 权限控制

## 快速开始

### 环境要求

- Go 1.24+
- PostgreSQL
- Docker (可选)

### 本地运行

```bash
# 安装依赖
go mod download

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=credit_management
export JWT_SECRET=your-secret-key

# 运行服务
go run main.go
```

### Docker 运行

```bash
# 构建镜像
docker build -t user-service .

# 运行容器
docker run -d \
  --name user-service \
  -p 8084:8084 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  -e JWT_SECRET=your-jwt-secret \
  user-service
```

## API 文档

详细的API文档请参考：[API_DOCS/user-service-API.md](../API_DOCS/user-service-API.md)

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DB_HOST` | `localhost` | 数据库主机 |
| `DB_PORT` | `5432` | 数据库端口 |
| `DB_USER` | `postgres` | 数据库用户名 |
| `DB_PASSWORD` | `password` | 数据库密码 |
| `DB_NAME` | `credit_management` | 数据库名称 |
| `DB_SSLMODE` | `disable` | 数据库SSL模式 |
| `JWT_SECRET` | `your-secret-key` | JWT密钥 |
| `PORT` | `8084` | 服务端口 |

## 健康检查

```
GET /health
```

## 许可证

MIT 