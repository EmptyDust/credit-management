# Credit Management System

学分活动管理系统的当前默认运行形态已经切换到：

- `frontend/`: React + TypeScript + Vite
- `backend/`: Go 模块化单体
- `postgres` + `redis`: 基础运行依赖

当前 `v2` 分支只保留重构后的新系统实现，不再包含旧微服务运行入口。

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

单独运行：

```bash
make test-backend
make test-e2e
make test-frontend
```

## Current Implementation Notes

- 前端 API 统一走 `/api/v1`
- 登录、注册、活动列表、活动详情、申请列表、用户页、导入向导、Devtools 已迁入 `frontend/src/modules/**`
- 后端 acceptance e2e 位于 `backend/tests/e2e`
- 活动审核通过后会自动生成只读申请
- 默认 compose 直接运行新 backend；前端 Nginx 代理也已指向 `backend:8082`
- `make migrate-backend` 和 `make seed-backend` 分别对应 `backend/cmd/migrate` 与 `backend/cmd/seed`

## Core Docs

- `docs/superpowers/plans/` 不再作为运行时基线的一部分保留。
- 当前仓库内仅保留仍与 `v2` 直接相关的交付与运维文档。
- 协作与提交流程见 [CONTRIBUTING.md](CONTRIBUTING.md)。
- 开发者 AI 专用流程见 [docs/ai-collaboration-workflow.md](docs/ai-collaboration-workflow.md)。
