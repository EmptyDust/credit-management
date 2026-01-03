# 学分活动服务设计文档

## 1. 服务合并背景

### 当前问题

- 事务管理和申请管理服务分离，但数据高度耦合
- 缺乏真正的服务间通信，使用模拟数据
- 流程复杂，用户体验不佳
- 权限控制分散，维护困难

### 合并优势

- 简化架构，减少服务间通信
- 统一数据模型，保证一致性
- 简化业务流程，提升用户体验
- 统一权限控制，便于维护

## 2. 业务需求确认

### 2.1 核心业务规则

1. **学分分配灵活性**：活动中各个参与同学获得学分可能不同，支持批量设置和单独设置
2. **申请查看功能**：学生可以查看自己的所有学分申请项
3. **活动管理权限**：
   - 只有活动创建者可以提交和修改活动
   - 只有活动创建者可以新增删除活动参与成员
   - 活动参与者可以退出活动
   - **活动创建者不一定是活动参与者**
   - **教师用户也有活动的创建和自己创建活动的活动编辑、修改权限**
4. **活动状态简化**：只有草稿-待审核-通过/拒绝状态，无需完成或其他状态
5. **自动申请生成**：活动通过后自动从活动导出申请，使用数据库触发器或存储过程
6. **申请流程简化**：学生无需修改申请，教师也无需审核申请，仅需查看和导出功能
7. **修改重提机制**：拒绝/通过状态应允许学生修改后再次提交审核
8. **管理员权限**：管理员具有编辑和管理的所有权限
9. **撤回机制**：
   - **学生提交活动后/活动被拒绝/活动已经通过的情况下，活动创建者可以撤回**
   - **撤回后即到草稿阶段**
   - **撤回时同时删除所有相关申请**
10. **导出功能**：
    - **学生可以导出自己的申请数据**
    - 教师/管理员可以导出所有申请数据
11. **参与者限制**：**只有学生可以参与活动**

## 3. 新的业务流程

### 3.1 学生创建活动流程

```
学生创建活动(草稿) → 设置参与者和学分 → 提交审核 → 教师审核 → 通过/拒绝 → 通过后自动生成申请
```

### 3.2 详细流程说明

#### 阶段 1：活动创建（草稿状态）

1. 学生/教师创建学分活动

   - 填写活动信息：标题、描述、时间、要求等
   - 状态：`draft`（草稿）
   - 可以随时修改活动信息

2. 设置参与者和学分分配

   - 添加/删除参与者（仅限学生）
   - 为每个参与者设置不同的学分值
   - 支持批量设置和单独设置

3. 提交活动审核
   - 状态：`pending_review`（待审核）
   - 只有草稿状态的活动可以提交审核

#### 阶段 2：活动审核

4. 教师/管理员审核活动

   - 审核活动内容和学分设置
   - 可以修改活动信息
   - 状态：`approved`（已通过）或 `rejected`（已拒绝）

5. 活动审核通过后，自动为参与者生成申请
   - 状态：`approved`（固定状态，自动生成）
   - 申请学分直接从活动参与者设置中继承

#### 阶段 3：活动撤回

6. 活动创建者可以撤回活动
   - **只能撤回待审核状态的活动**
   - 撤回后状态回到：`draft`（草稿）
   - 撤回时自动删除所有相关申请（如果活动之前已通过并生成了申请）

#### 阶段 4：申请查看和导出

7. 学生查看自己的申请

   - 可以查看所有已通过的申请
   - 可以导出自己的申请数据

8. 教师/管理员查看和导出申请
   - 可以查看所有申请
   - 可以导出所有申请数据

## 4. 数据模型设计

### 4.1 核心实体

#### CreditActivity（学分活动）

```go
type CreditActivity struct {
    ID             string         `json:"id" gorm:"primaryKey;type:uuid"`
    Title          string         `json:"title" gorm:"not null"`
    Description    string         `json:"description"`
    StartDate      time.Time      `json:"start_date"`
    EndDate        time.Time      `json:"end_date"`
    Status         string         `json:"status" gorm:"default:'draft';index"`
    Category       string         `json:"category"`
    Requirements   string         `json:"requirements"`
    OwnerID        string         `json:"owner_id" gorm:"type:uuid;not null;index"`
    ReviewerID     string         `json:"reviewer_id" gorm:"type:uuid"`
    ReviewComments string         `json:"review_comments"`
    ReviewedAt     *time.Time     `json:"reviewed_at"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
```

#### ActivityParticipant（活动参与者）

```go
type ActivityParticipant struct {
    ActivityID string    `json:"activity_id" gorm:"primaryKey;type:uuid"`
    UserID     string    `json:"id" gorm:"primaryKey;type:uuid"`
    Credits    float64   `json:"credits" gorm:"not null;default:0"`
    JoinedAt   time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

#### Application（学分申请）

```go
type Application struct {
    ID             string         `json:"id" gorm:"primaryKey;type:uuid"`
    ActivityID     string         `json:"activity_id" gorm:"type:uuid;not null;index"`
    UserID         string         `json:"id" gorm:"type:uuid;not null;index"`
    Status         string         `json:"status" gorm:"default:'approved';index"`
    AppliedCredits float64        `json:"applied_credits" gorm:"not null"`
    AwardedCredits float64        `json:"awarded_credits" gorm:"not null"`
    SubmittedAt    time.Time      `json:"submitted_at" gorm:"default:CURRENT_TIMESTAMP"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
```

### 4.2 状态定义

#### 活动状态

- `draft`：草稿状态，可以修改
- `pending_review`：待审核状态，等待教师审核
- `approved`：已通过，自动生成申请
- `rejected`：已拒绝，可以修改后重新提交

#### 申请状态

- `approved`：已通过（固定状态，自动生成）

## 5. 权限控制设计

### 5.1 角色权限矩阵

| 功能           | 学生 | 教师 | 管理员 |
| -------------- | ---- | ---- | ------ |
| 创建活动       | ✓    | ✓    | ✓      |
| 编辑自己的活动 | ✓    | ✓    | ✓      |
| 删除自己的活动 | ✓    | ✓    | ✓      |
| 提交活动审核   | ✓    | ✓    | ✓      |
| 撤回活动       | ✓    | ✓    | ✓      |
| 审核活动       | ✗    | ✓    | ✓      |
| 添加参与者     | ✓    | ✓    | ✓      |
| 删除参与者     | ✓    | ✓    | ✓      |
| 设置学分       | ✓    | ✓    | ✓      |
| 退出活动       | ✓    | ✗    | ✗      |
| 查看自己的申请 | ✓    | ✓    | ✓      |
| 导出自己的申请 | ✓    | ✓    | ✓      |
| 查看所有申请   | ✗    | ✓    | ✓      |
| 导出所有申请   | ✗    | ✓    | ✓      |

### 5.2 特殊权限说明

1. **活动创建者权限**：

   - 可以编辑、删除、提交、撤回自己创建的活动
   - 可以管理自己活动的参与者
   - 不一定是活动的参与者

2. **教师权限**：

   - 可以创建和管理自己的活动
   - 可以审核所有活动
   - 可以查看和导出所有申请

3. **参与者限制**：
   - 只有学生可以参与活动
   - 教师和管理员不能参与活动

## 6. API 设计

### 6.1 活动管理 API

#### 创建活动

- **POST** `/api/activities`
- **权限**：所有认证用户
- **说明**：学生和教师都可以创建活动

#### 获取活动列表

- **GET** `/api/activities`
- **权限**：所有认证用户
- **说明**：学生只能看到自己创建或参与的活动，教师可以看到所有活动

#### 获取活动详情

- **GET** `/api/activities/{id}`
- **权限**：所有认证用户
- **说明**：学生只能看到自己创建或参与的活动

#### 更新活动

- **PUT** `/api/activities/{id}`
- **权限**：活动创建者、管理员
- **说明**：只有草稿状态的活动可以修改

#### 删除活动

- **DELETE** `/api/activities/{id}`
- **权限**：活动创建者、管理员

#### 提交活动审核

- **POST** `/api/activities/{id}/submit`
- **权限**：活动创建者
- **说明**：只有草稿状态的活动可以提交

#### 撤回活动

- **POST** `/api/activities/{id}/withdraw`
- **权限**：活动创建者
- **说明**：**只能撤回待审核状态的活动**，撤回后状态变为草稿

#### 审核活动

- **POST** `/api/activities/{id}/review`
- **权限**：教师、管理员
- **说明**：可以审核待审核、已通过或已拒绝状态的活动（支持修改审核状态）

### 6.2 参与者管理 API

#### 添加参与者

- **POST** `/api/activities/{id}/participants`
- **权限**：活动创建者、管理员
- **说明**：只能添加学生用户

#### 批量设置学分

- **PUT** `/api/activities/{id}/participants/batch-credits`
- **权限**：活动创建者、管理员

#### 设置单个学分

- **PUT** `/api/activities/{id}/participants/{id}/credits`
- **权限**：活动创建者、管理员

#### 删除参与者

- **DELETE** `/api/activities/{id}/participants/{id}`
- **权限**：活动创建者、管理员

#### 退出活动

- **POST** `/api/activities/{id}/leave`
- **权限**：活动参与者（学生）
- **说明**：只有学生可以退出活动

### 6.3 申请管理 API

#### 获取用户申请列表

- **GET** `/api/applications`
- **权限**：所有认证用户
- **说明**：学生只能看到自己的申请，支持分页和筛选

#### 获取申请详情

- **GET** `/api/applications/{id}`
- **权限**：所有认证用户
- **说明**：学生只能查看自己的申请详情

#### 获取申请统计

- **GET** `/api/applications/stats`
- **权限**：所有认证用户
- **说明**：返回申请统计数据

#### 获取所有申请

- **GET** `/api/applications/all`
- **权限**：教师、管理员
- **说明**：获取所有用户的申请列表

#### 导出申请数据

- **GET** `/api/applications/export`
- **权限**：所有认证用户
- **说明**：学生只能导出自己的申请，教师/管理员可以导出所有申请，支持 CSV 格式

### 6.4 附件管理 API

#### 获取附件列表

- **GET** `/api/activities/{id}/attachments`
- **权限**：所有认证用户
- **说明**：获取活动的所有附件列表

#### 上传附件

- **POST** `/api/activities/{id}/attachments`
- **权限**：活动创建者、教师、管理员
- **说明**：上传单个附件文件，支持多种文件格式

#### 批量上传附件

- **POST** `/api/activities/{id}/attachments/batch`
- **权限**：活动创建者、教师、管理员
- **说明**：批量上传多个附件文件

#### 下载附件

- **GET** `/api/activities/{id}/attachments/{attachment_id}/download`
- **权限**：所有认证用户
- **说明**：下载附件文件

#### 预览附件

- **GET** `/api/activities/{id}/attachments/{attachment_id}/preview`
- **权限**：所有认证用户
- **说明**：在线预览附件内容

#### 更新附件信息

- **PUT** `/api/activities/{id}/attachments/{attachment_id}`
- **权限**：活动创建者、教师、管理员
- **说明**：更新附件的描述等信息

#### 删除附件

- **DELETE** `/api/activities/{id}/attachments/{attachment_id}`
- **权限**：活动创建者、教师、管理员
- **说明**：删除附件并清理物理文件（如果文件未被其他附件引用）

### 6.5 搜索 API

#### 搜索活动

- **GET** `/api/search/activities`
- **权限**：所有认证用户
- **说明**：高级搜索活动，支持多条件组合

#### 搜索申请

- **GET** `/api/search/applications`
- **权限**：所有认证用户
- **说明**：搜索申请记录

#### 搜索参与者

- **GET** `/api/search/participants`
- **权限**：所有认证用户
- **说明**：搜索活动参与者

#### 搜索附件

- **GET** `/api/search/attachments`
- **权限**：所有认证用户
- **说明**：搜索附件文件

### 6.4 统计 API

#### 获取活动统计

- **GET** `/api/activities/stats`
- **权限**：所有认证用户

#### 获取申请统计

- **GET** `/api/applications/stats`
- **权限**：所有认证用户

## 7. 数据库设计

### 7.1 表结构

#### credit_activities

```sql
CREATE TABLE credit_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    status TEXT DEFAULT 'draft',
    category TEXT,
    requirements TEXT,
    owner_id UUID NOT NULL,
    reviewer_id UUID,
    review_comments TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_credit_activities_status ON credit_activities(status);
CREATE INDEX idx_credit_activities_owner_id ON credit_activities(owner_id);
CREATE INDEX idx_credit_activities_deleted_at ON credit_activities(deleted_at);
```

#### activity_participants

```sql
CREATE TABLE activity_participants (
    activity_id UUID NOT NULL,
    id UUID NOT NULL,
    credits DECIMAL(5,2) NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (activity_id, id),
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE
);
```

#### applications

```sql
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL,
    id UUID NOT NULL,
    status TEXT DEFAULT 'approved',
    applied_credits DECIMAL(5,2) NOT NULL,
    awarded_credits DECIMAL(5,2) NOT NULL,
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE
);

CREATE INDEX idx_applications_activity_id ON applications(activity_id);
CREATE INDEX idx_applications_id ON applications(id);
CREATE INDEX idx_applications_status ON applications(status);
CREATE INDEX idx_applications_deleted_at ON applications(deleted_at);
```

### 7.2 应用层自动化逻辑

**注意**：系统已移除数据库触发器，所有业务逻辑在应用层实现（`credit-activity-service/handlers/activity_side_effects.go`）。

#### 申请自动生成逻辑

- **触发时机**：活动审核通过时（状态从非 `approved` 变为 `approved`）
- **实现位置**：`ActivityHandler.handleStatusSideEffects()` 和 `generateApplicationsForParticipants()`
- **逻辑说明**：
  - 在数据库事务中执行
  - 查询活动的所有参与者（排除已删除的）
  - 为每个参与者创建申请记录（如果不存在）
  - 申请的学分直接继承参与者的学分设置

#### 申请自动删除逻辑

- **触发时机**：活动从已通过状态变为其他状态时
- **实现位置**：`ActivityHandler.handleStatusSideEffects()` 和 `softDeleteApplications()`
- **逻辑说明**：
  - 当活动从 `approved` 变为其他状态（如 `rejected`、`draft` 等）时
  - 自动软删除该活动的所有相关申请
  - 保持数据一致性，避免活动被拒绝但申请仍显示"已通过"

#### 文件清理逻辑

- **触发时机**：活动删除时
- **实现位置**：`ActivityHandler.softDeleteActivityRelations()` 和 `cleanupAttachmentFiles()`
- **逻辑说明**：
  - 检测孤立的附件文件（基于 MD5 哈希）
  - 如果文件没有被其他附件引用，则删除物理文件
  - 保持存储空间清洁

## 8. 部署和运维

### 8.1 环境配置

- **数据库**：PostgreSQL 15+
- **应用服务器**：Go 1.21+
- **端口**：8083
- **健康检查**：`/health`

### 8.2 环境变量

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=credit_management
PORT=8083
```

### 8.3 监控指标

- 活动创建数量
- 活动审核通过率
- 申请生成数量
- API 响应时间
- 错误率

## 9. 测试策略

### 9.1 单元测试

- 模型验证
- 业务逻辑测试
- 权限控制测试

### 9.2 集成测试

- API 接口测试
- 数据库操作测试
- 触发器测试

### 9.3 端到端测试

- 完整业务流程测试
- 权限边界测试
- 性能测试

## 10. 总结

这个设计完全符合您的业务需求，主要特点包括：

1. **简化的业务流程**：从草稿到审核到通过的清晰流程
2. **灵活的权限控制**：支持学生和教师创建活动，但只有学生可以参与
3. **自动化的申请生成**：通过数据库触发器自动生成申请
4. **撤回机制**：支持活动创建者撤回活动并删除相关申请
5. **导出功能**：支持学生导出自己的申请数据
6. **清晰的 API 设计**：完整的 RESTful API 设计
7. **可扩展的架构**：支持未来的功能扩展

这个设计既满足了当前的业务需求，又为未来的发展留下了空间。
