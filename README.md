# Credit Management System

学分活动管理系统的当前默认运行形态已经切换到：

- `frontend/`: React + TypeScript + Vite
- `backend/`: Go 模块化单体
- `postgres` + `redis`: 基础运行依赖

旧 `api-gateway`、`auth-service`、`credit-activity-service`、`user-service` 仍保留在仓库中作为历史参考，但默认 `docker compose` 流程不再依赖它们；如需启动旧栈，需显式启用 `legacy` profile。

## Default Stack

| Component | Port | Purpose |
| --- | --- | --- |
| `frontend` | `3000` | 用户界面，Nginx 代理 `/api/*` 到新 backend |
| `backend` | `8082` | 模块化单体 API |
| `postgres` | `5433` | 数据库 |
| `redis` | `6379` | 缓存与会话辅助 |

健康检查：

- `http://localhost:3000`
- `http://localhost:8082/healthz`
- `http://localhost:8082/api/v1/healthz`

## Quick Start

```bash
git clone https://github.com/EmptyDust/credit-management
cd credit-management

make up-refactor
```

停止：

```bash
make down-refactor
```

查看日志：

```bash
make logs-refactor
```

## Local Development

后端：

```bash
make dev-backend
```

前端：

```bash
cd frontend
npm exec --yes pnpm -- dev
```

## Verification

完整重构栈验证：

```bash
make test-refactor
```

单独运行：

```bash
make test-backend
make test-e2e
make test-frontend
```

## Legacy Services

如果你需要临时启动旧微服务进行比对，请显式启用 `legacy` profile，例如：

```bash
docker compose --profile legacy up -d api-gateway auth-service credit-activity-service user-service
```

旧测试入口仍保留：

```bash
make legacy-test
```

## Current Implementation Notes

- 前端 API 统一走 `/api/v1`
- 登录、注册、活动列表、活动详情、申请列表、用户页、导入向导、Devtools 已迁入 `frontend/src/modules/**`
- 后端 acceptance e2e 位于 `backend/tests/e2e`
- 活动审核通过后会自动生成只读申请
- 默认 compose 已切到新 backend；前端 Nginx 代理也已指向 `backend:8082`
