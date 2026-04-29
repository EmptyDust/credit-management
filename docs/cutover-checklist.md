# Cutover Checklist

## Goal

将默认运行入口从旧微服务栈切换到：

- `postgres`
- `redis`
- `backend`
- `frontend`

## Before Cutover

- 确认 [docs/refactor-requirements.md](/Users/fengling/github/credit-management/docs/refactor-requirements.md) 已锁定业务规则
- 确认前端路由和 API 全部面向 `/api/v1`
- 确认 `backend/tests/e2e` 覆盖最小 acceptance：
  - 登录 / 刷新 / 登出后 refresh 失效
  - draft -> submit -> approve 自动生成申请
  - 导入 precheck -> confirm 正常完成

## Cutover Steps

1. 执行 `make test-refactor`
2. 执行 `docker compose config`，确认默认服务只依赖新栈
3. 执行 `make up-refactor`
4. 验证 `frontend` 首页可访问
5. 验证 `backend` 的 `/healthz` 与 `/api/v1/healthz`
6. 验证前端登录、活动创建、审核、申请查看、用户导入
7. 若需要旧栈，仅通过 `docker compose --profile legacy ...` 手动启动

## Acceptance Checklist

- 前端容器通过 Nginx 代理访问新 backend
- `docker compose up -d` 不再默认启动旧网关和旧三服务
- `make test-refactor` 通过
- 新 backend e2e 通过
- 前端构建通过

## Rollback

若切换后需要临时回看旧链路：

1. 停止默认新栈：`make down-refactor`
2. 显式启动 legacy profile
3. 不回滚业务规则文档，只把运行入口切回旧服务进行排查
