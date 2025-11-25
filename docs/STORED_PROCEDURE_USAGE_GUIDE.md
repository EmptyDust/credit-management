# 存储过程使用指南（已废弃）

> 最后更新：2025-11-25  
> 所有 PostgreSQL 存储过程 / 函数（如 `delete_activity_with_permission_check`、`batch_delete_activities`、`restore_deleted_activity` 等）已被移除，逻辑下沉至 Go 微服务。

## 现状

- 活动删除 / 批量删除：`credit-activity-service/handlers/activity_crud.go` 与 `activity_batch.go`；
- 权限校验：各服务的 HTTP handler 负责（结合 JWT / Header 信息）；
- 数据恢复、软删除：统一使用 GORM 的软删除能力。

如需扩展业务逻辑，请在对应服务新增方法与测试，而非回退到数据库层函数。*** End Patch}*** End Patch to=functions.apply_patch JSON parse error: Invalid control character at: line 2 column 1 (char 1) 

