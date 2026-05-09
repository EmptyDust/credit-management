# Credit Management System

学分活动管理系统的当前默认运行形态已经切换到：

- `frontend/`: React + TypeScript + Vite
- `backend/`: Go 模块化单体
- `postgres` + `redis`: 基础运行依赖

当前 `v2` 分支只保留重构后的新系统实现，不再包含旧微服务运行入口。

当前默认运行时基于持久化依赖：

- `postgres`: 持久化业务数据
- `redis`: 缓存、会话与运行时辅助状态
- `backend`: 默认以 Postgres + Redis 作为运行时后端，不再以内存存储作为本地默认形态

## Branch Model

- `v2`: 当前默认分支，也是新系统唯一开发主线
- `v1`: 旧系统冻结分支，仅保留历史
- `master`: 历史过渡分支，不再作为新开发基线

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
docker compose up -d postgres redis
make dev-backend
```

初始化数据库：

```bash
make migrate-backend
make seed-backend
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

持久化运行时验证：

```bash
make migrate-backend
make seed-backend
make test-integration
```

`make test-integration` 会在本地默认端点不可用时拉起 `postgres` 和 `redis`，然后以本地 compose 默认值运行：

- `TEST_DATABASE_DSN=postgres://postgres:password@localhost:5433/credit_management?sslmode=disable`
- `TEST_DATABASE_SCHEMA=integration_runtime`
- `TEST_REDIS_ADDR=localhost:6379`
- `TEST_REDIS_DB=15`

这个目标会重建 `TEST_DATABASE_SCHEMA` 对应的 schema，并清空 `TEST_REDIS_DB` 对应的 Redis DB。它只适合一次性验证环境，不应指向生产、共享开发库，或你正在手工使用的 Redis DB。

如需改连其他测试实例，可在执行前覆盖这些环境变量，但仍要保证目标库和目标 Redis DB 是可销毁的测试隔离环境。

单独运行：

```bash
make test-backend
make test-e2e
make test-integration
make test-frontend
```

验证预期：

- `make test-refactor` 仍是快速基线检查，覆盖后端单测、后端 e2e 与前端构建
- 持久化基础是否真实接入 Postgres + Redis，以及应用重建后的状态是否可继续读取，需要看 `make test-integration`
- 运行时冒烟检查仍应包含 `http://localhost:3000`、`http://localhost:8082/healthz` 与 `http://localhost:8082/api/v1/healthz`

## Current Implementation Notes

- 前端 API 统一走 `/api/v1`
- 登录、注册、活动工作台、活动详情、申请列表、用户页、导入向导、Devtools 已迁入 `frontend/src/modules/**`
- 活动工作台包含 `/activities`、`/activities/new`、`/activities/:id/edit`，支持分类 schema 驱动草稿创建/编辑、统计卡、筛选、CSV 导入预校验/确认与 CSV 导出
- 后端 acceptance e2e 位于 `backend/tests/e2e`
- 活动工作台端到端契约覆盖位于 `backend/tests/e2e/activity_workspace_flow_test.go`
- 活动审核通过后会自动生成只读申请
- 默认 compose 直接运行新 backend；前端 Nginx 代理也已指向 `backend:8082`
- `make migrate-backend` 和 `make seed-backend` 分别对应 `backend/cmd/migrate` 与 `backend/cmd/seed`

## Core Docs

- `docs/superpowers/plans/` 不再作为运行时基线的一部分保留。
- 当前仓库内仅保留仍与 `v2` 直接相关的交付与运维文档。
- 协作与提交流程见 [CONTRIBUTING.md](CONTRIBUTING.md)。
- 开发者 AI 专用流程见 [docs/ai-collaboration-workflow.md](docs/ai-collaboration-workflow.md)。
