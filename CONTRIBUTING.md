# Contributing

本仓库当前以 `v2` 作为新系统唯一开发主线。请不要再把 `master` 或 `v1` 当作日常开发基线。

## 分支规则

- `v1`：旧系统冻结分支，只保留历史
- `v2`：当前默认分支，也是所有新开发和 PR 的目标分支
- `master`：历史过渡分支，不再作为开发起点

日常开发要求：

1. 新任务从最新 `v2` 拉出功能分支。
2. 不要直接在 `v2` 上提交日常改动。
3. PR 目标分支统一为 `v2`。

示例：

```bash
git switch v2
git pull --ff-only origin v2
git switch -c feat/<topic>
```

如果需要隔离工作区，优先使用：

```bash
git worktree add .worktrees/<branch> -b <branch> v2
```

## 本地开发

启动整套新系统：

```bash
make up-refactor
```

停止：

```bash
make down-refactor
```

后端开发：

```bash
make dev-backend
```

前端开发：

```bash
cd frontend
npm exec --yes pnpm -- dev
```

## 验证要求

代码改动默认优先执行：

```bash
make test-refactor
```

如果只做局部改动，也至少运行与改动范围匹配的验证，例如：

```bash
make test-backend
make test-e2e
make test-frontend
```

纯文档或元数据改动至少应运行针对性检查，例如：

```bash
git diff --check
docker compose config
```

## AI 协作规范

如果你在本机使用 Codex、Claude Code、Copilot CLI 等开发者 AI，请先让 AI 阅读：

- [docs/ai-collaboration-workflow.md](docs/ai-collaboration-workflow.md)

这份文档规定了：

- `gh` / GitHub 权限检查
- `v2` 分支基线与 PR 目标
- worktree 使用约定
- 发 PR 前的同步与验证顺序
- 何时允许 `rebase`、强推和合并

## 提交与 PR

1. 提交信息描述真实改动结果，不要写空洞标题。
2. PR 标题、PR 正文、PR 评论直接写自然中文即可，不要堆固定机器人模板。
3. PR 正文至少应包含：
   - 本次改动
   - 风险与影响
   - 验证情况
