# 数据库触发器使用指南（已废弃）

> 最后更新：2025-11-25

平台已彻底移除数据库触发器，所有业务规则、数据校验与文件清理均由各微服务实现。此文档仅说明迁移背景，详细逻辑请查阅对应服务代码。

## 迁移原因

- 触发器难以调试，且与业务代码割裂；
- CI/CD 中无法方便地模拟触发器副作用；
- 需要在 Go 服务内统一记录日志与埋点。

## 现有实现位置

| 功能 | 负责模块 |
| --- | --- |
| `updated_at` 自动维护 | GORM `autoCreateTime/autoUpdateTime` |
| 用户数据校验（邮箱、学号、工号等） | `user-service/utils/validator.go` 与 `handlers` |
| 活动审批生成/撤回申请 | `credit-activity-service/handlers/activity_review.go` + `activity_side_effects.go` |
| 附件孤立文件检测与删除 | `credit-activity-service/handlers/activity_side_effects.go` 及 `handlers/attachment.go` |

## 建议

- 若需新增校验或联动逻辑，请在对应微服务新增单元测试；
- 数据库仅保留必要的约束、索引与视图，避免再次引入触发器。

