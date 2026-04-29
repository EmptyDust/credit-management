# 开发者 AI 协作与 PR 流程（`v2`）

本文参考并本地化自 [QLHazyCoder/codex-oauth-automation-extension 的《开发者AI开发与PR提交流程》](https://github.com/QLHazyCoder/codex-oauth-automation-extension/blob/master/%E5%BC%80%E5%8F%91%E8%80%85AI%E5%BC%80%E5%8F%91%E4%B8%8EPR%E6%8F%90%E4%BA%A4%E6%B5%81%E7%A8%8B.md)，但已按本仓库当前分支策略、测试要求、运行入口和部署方式重写。

本文面向“开发者自己电脑上的 AI”。目标是让 AI 在本仓库里开发、整理改动、发起 PR、更新 PR、补充说明时，始终遵守同一套可执行流程。

## 1. 适用范围

适用于：

- 新功能开发
- Bug 修复
- 文档与流程治理
- 重构后的 `v2` 主线维护
- 发起或更新 PR
- 在开发者明确授权后合并自己的 PR

不适用于：

- 在未阅读代码上下文前凭猜测写结论
- 直接在 `v1` 上继续开发新系统
- 把日常开发基线设为 `master`

## 2. 仓库分支模型

本仓库当前采用双代际分支模型：

- `v1`：旧系统冻结分支，只用于保留历史，不再作为新开发基线。
- `v2`：当前默认分支，也是新系统唯一的日常开发基线。
- `master`：历史过渡分支，不再作为开发起点，也不再作为 PR 目标分支。

硬规则：

1. 新任务必须从最新 `v2` 拉出功能分支。
2. PR 目标分支只能是 `v2`。
3. AI 不得直接在 `v2` 上提交日常开发改动。
4. AI 不得把 `master` 作为开发基线，也不得把 PR 目标设为 `master`。

## 3. 硬性规则

1. 任何结论都不能猜，必须基于真实命令输出、真实 diff、真实代码上下文。
2. 在执行任何 GitHub 操作前，必须确认 `gh` 可用、已登录、账号正确、权限足够。
3. 如果当前工作区存在无法确认归属的脏改动，必须先停止并告知开发者，不能偷偷混入本次任务。
4. 代码改动默认应在独立功能分支或独立 worktree 中完成；本仓库优先使用 `.worktrees/`。
5. 如果旧逻辑已经明显无用，且确认无其他依赖，可在本次改动内一并清理，不要为了“看起来兼容”保留陈旧代码。
6. 代码改动完成后必须进行验证。默认优先使用 `make test-refactor`；如果只做局部改动，可以运行更小的验证集，但必须在最终反馈中明确说明验证范围。
7. 纯文档或元数据改动至少要做针对性校验，例如 `git diff --check`、`docker compose config`、相关命令可用性检查，不能完全不验证。
8. 不允许使用 `git push --force`；如确有必要改写已推送分支历史，只能在开发者允许 `rebase` 的前提下使用 `git push --force-with-lease`。
9. 没有开发者明确授权时，AI 不得擅自合并 PR、关闭 PR、删除远端分支或重写共享历史。
10. PR 标题、PR 正文、PR 评论都使用自然中文，不要写“自动回复”“AI 结果如下”之类的固定机器人腔。

## 4. 使用前准备

开发者本机至少应满足：

1. 已安装 `git`
2. 已安装 GitHub CLI，也就是 `gh`
3. `gh auth login` 已完成
4. 对目标仓库至少有读取权限；如果要推分支、改 PR、合并 PR，还需要对应写权限

如果 AI 检查到 `gh` 不可用、未登录、登录到错误账号，或者权限不足，必须先停止并明确告知开发者。

## 5. 标准执行顺序

### 阶段 1：环境确认与仓库现状检查

开始前先执行：

```bash
gh --version
gh auth status
git status --short --branch
git remote -v
git branch --show-current
git fetch origin
git symbolic-ref refs/remotes/origin/HEAD
```

要求：

1. 必须确认当前仓库远端是正确仓库。
2. 必须确认默认远端分支是否是 `origin/v2`。
3. 必须确认当前工作区是否有未提交改动。
4. 不能在没看当前分支状态的情况下直接开始写代码或发 PR。

### 阶段 2：先对齐最新 `v2`

#### 场景 A：这是一个新任务

新任务必须从最新 `v2` 开始：

```bash
git switch v2
git pull --ff-only origin v2
git switch -c <feature-branch>
```

如果需要隔离工作区，优先用 worktree：

```bash
git worktree add .worktrees/<feature-branch> -b <feature-branch> v2
```

规则：

1. 不允许直接在 `v2` 上开发。
2. 不允许从 `master` 或 `v1` 拉出新系统功能分支。
3. 如果 `.worktrees/` 不存在或未被忽略，应先修正仓库配置，再创建 worktree。

#### 场景 B：这是一个已有分支上的继续开发

如果开发已经在某个功能分支上进行，先判断自己相对 `origin/v2` 的状态：

```bash
git fetch origin
git rev-list --left-right --count origin/v2...HEAD
git log --oneline HEAD..origin/v2
```

要求：

1. 必须真实判断当前分支是否落后于 `origin/v2`。
2. 如果当前分支落后，可以继续开发，但在发起 PR 前必须再次补齐最新 `origin/v2`。
3. 如果预期冲突复杂，必须尽早告诉开发者，不要拖到最后才爆冲突。

### 阶段 3：开发与本地整理

开发时至少遵守下面这些要求：

1. 先读相关代码上下文，再改文件。
2. 不只改表面调用点，要检查相关状态流、配置、页面流程、脚本和文档。
3. 如果本次改动改变了分支策略、运行入口、命令或目录结构，要同步更新对应文档和自动化配置。
4. 改完后必须自己检查：
   - `git diff --stat`
   - `git diff`
   - 是否残留冲突标记
   - 是否混入无关改动

冲突标记可用下面的方式检查：

```bash
rg -n '^(<<<<<<<|=======|>>>>>>>)' .
```

### 阶段 4：发起 PR 前再次同步最新 `v2`

无论这个分支什么时候开始开发，只要准备发起 PR，就必须再次获取远端最新 `v2`：

```bash
git fetch origin
git log --oneline HEAD..origin/v2
```

如果 `origin/v2` 有新提交：

- 优先使用 `git rebase origin/v2`
- 如果开发者明确禁止改写历史，或分支已被多人共享，再改用 `git merge origin/v2`

处理规则：

1. 执行 `rebase` 后必须重新检查 diff，确认逻辑没有被改坏。
2. 如果执行了 `rebase` 且分支已推送过远端，后续只能使用 `git push --force-with-lease origin <feature-branch>`。
3. 不能只删冲突标记就结束，必须继续核对冲突两边的真实逻辑。

### 阶段 5：验证、提交与推送

PR 前至少做这些检查：

```bash
git status --short
git diff --stat origin/v2...HEAD
git diff origin/v2...HEAD
git diff --check
```

验证规则：

1. 代码改动默认优先执行：

```bash
make test-refactor
```

2. 如果只跑了局部命令，例如：
   - `make test-backend`
   - `make test-e2e`
   - `make test-frontend`
   - 某个模块的单测

   则必须在最终反馈和 PR 正文里说明原因与覆盖范围。

3. 文档或仓库元数据改动至少要跑相关的目标检查，例如：
   - `docker compose config`
   - `git diff --check`
   - 受影响的命令或脚本自检

推送前确认：

1. 当前分支不是 `v2`、`v1`、`master`
2. diff 只包含本次任务相关改动
3. 提交信息描述真实改动结果，不要写空洞标题

推送示例：

```bash
git push -u origin <feature-branch>
```

### 阶段 6：创建或更新 PR

PR 只能指向 `v2`。

创建 PR：

```bash
gh pr create --base v2 --head <feature-branch> --title "<PR标题>" --body-file <PR正文文件>
```

检查已有 PR：

```bash
gh pr view <PR_NUMBER> --json number,title,baseRefName,headRefName,state,isDraft,url
```

如果已有 PR 的目标分支不是 `v2`，必须改正：

```bash
gh pr edit <PR_NUMBER> --base v2
```

PR 标题要求：

1. 直接描述功能结果或修复结果。
2. 不要把标题写成 Git 动作描述。
3. 不要写空洞标题。

PR 正文建议结构：

```markdown
## 本次改动
- ...

## 风险与影响
- ...

## 验证情况
- 已执行 ...
- 未执行 ...
```

要求：

1. 正文必须基于真实改动来写。
2. 不要写固定“自动回复”抬头。
3. 不要写与代码或验证情况不相符的夸大表述。

### 阶段 7：只有在明确授权时，才允许合并 PR

开发者明确要求 AI 继续合并自己的 PR 时，先再次确认：

```bash
gh pr view <PR_NUMBER> --json number,title,baseRefName,headRefName,state,isDraft,mergeable,mergeStateStatus,url
git fetch origin
git log --oneline HEAD..origin/v2
```

合并前必须满足：

1. PR 目标分支是 `v2`
2. PR 不是 draft
3. PR 仍然是 open
4. 当前分支已吸收最新 `origin/v2`
5. 没有尚未处理的明确问题
6. 开发者已经明确授权合并

示例：

```bash
gh pr merge <PR_NUMBER> --merge --delete-branch
```

## 6. 最终反馈给开发者时必须说明

AI 完成后，至少要明确反馈：

1. 实际执行了哪些命令和动作
2. 当前分支是否已经同步最新 `origin/v2`
3. 是否创建了新 PR，或者更新了已有 PR
4. PR 编号和链接是什么
5. PR 目标分支是否确认是 `v2`
6. 是否执行了 `rebase` 或 `merge`
7. 实际运行了哪些验证；如果没有全量验证，要明确说明
8. 是否已经合并；如果已合并，要明确说明是合并到 `v2`

## 7. 一句话执行要求

AI 在本仓库里做开发与提 PR 时，必须按“先确认环境与当前工作区，再对齐最新 `v2`，再开发与整理改动，再次同步最新 `v2`，然后完成验证，最后只向 `v2` 发起或更新 PR；只有在开发者明确授权时，才允许把自己的 PR 合并到 `v2`”的顺序执行，不能跳步，不能猜，不能偷懒。
