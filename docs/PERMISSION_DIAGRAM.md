# 权限控制可视化图表

## 1. 系统架构图

```mermaid
graph TB
    Client[客户端] --> Gateway[API Gateway]

    Gateway --> JWT[JWT Middleware]
    Gateway --> Auth[Auth Middleware]
    Gateway --> Permission[Permission Middleware]

    JWT --> Auth
    Auth --> Permission

    Permission --> AuthService[认证服务]
    Permission --> UserService[用户服务]
    Permission --> ActivityService[学分活动服务]

    subgraph "权限中间件"
        AllUsers[AllUsers()]
        StudentOnly[StudentOnly()]
        TeacherOrAdmin[TeacherOrAdmin()]
        AdminOnly[AdminOnly()]
    end

    Permission --> AllUsers
    Permission --> StudentOnly
    Permission --> TeacherOrAdmin
    Permission --> AdminOnly
```

## 2. 用户角色层级图

```mermaid
graph TD
    Admin[管理员 Admin] --> Teacher[教师 Teacher]
    Teacher --> Student[学生 Student]

    Admin --> AdminPerms[管理员权限]
    Teacher --> TeacherPerms[教师权限]
    Student --> StudentPerms[学生权限]

    subgraph "管理员权限"
        AdminPerms --> ManageUsers[用户管理]
        AdminPerms --> DeleteActivities[删除活动]
        AdminPerms --> SystemConfig[系统配置]
    end

    subgraph "教师权限"
        TeacherPerms --> ManageStudents[管理学生]
        TeacherPerms --> CreateActivities[创建活动]
        TeacherPerms --> ReviewActivities[审核活动]
        TeacherPerms --> ManageParticipants[管理参与者]
        TeacherPerms --> ViewAllApplications[查看所有申请]
    end

    subgraph "学生权限"
        StudentPerms --> ViewProfile[查看个人信息]
        StudentPerms --> CreateActivities[创建活动]
        StudentPerms --> ViewOwnApplications[查看自己的申请]
        StudentPerms --> LeaveActivity[离开活动]
        StudentPerms --> ViewActivities[查看活动]
    end
```

## 3. 权限控制流程图

```mermaid
flowchart TD
    Start([用户请求]) --> Gateway[API Gateway]
    Gateway --> JWT{JWT Token 验证}

    JWT -->|无效| Error1[返回 401 Unauthorized]
    JWT -->|有效| Extract[提取用户信息]

    Extract --> Permission{权限中间件检查}

    Permission -->|AllUsers| CheckAuth{用户已认证?}
    Permission -->|StudentOnly| CheckStudent{用户是学生?}
    Permission -->|TeacherOrAdmin| CheckTeacher{用户是教师或管理员?}
    Permission -->|AdminOnly| CheckAdmin{用户是管理员?}

    CheckAuth -->|否| Error2[返回 401 Unauthorized]
    CheckAuth -->|是| Forward[转发到微服务]

    CheckStudent -->|否| Error3[返回 403 Forbidden]
    CheckStudent -->|是| Forward

    CheckTeacher -->|否| Error4[返回 403 Forbidden]
    CheckTeacher -->|是| Forward

    CheckAdmin -->|否| Error5[返回 403 Forbidden]
    CheckAdmin -->|是| Forward

    Forward --> Success[请求成功处理]
```

## 4. 功能模块权限矩阵

```mermaid
graph LR
    subgraph "用户管理模块"
        UM[用户管理] --> UM_Student[学生权限]
        UM --> UM_Teacher[教师权限]
        UM --> UM_Admin[管理员权限]
    end

    subgraph "活动管理模块"
        AM[活动管理] --> AM_Student[学生权限]
        AM --> AM_Teacher[教师权限]
        AM --> AM_Admin[管理员权限]
    end

    subgraph "申请管理模块"
        APM[申请管理] --> APM_Student[学生权限]
        APM --> APM_Teacher[教师权限]
        APM --> APM_Admin[管理员权限]
    end

    subgraph "参与者管理模块"
        PM[参与者管理] --> PM_Student[学生权限]
        PM --> PM_Teacher[教师权限]
        PM --> PM_Admin[管理员权限]
    end

    subgraph "附件管理模块"
        FM[附件管理] --> FM_Student[学生权限]
        FM --> FM_Teacher[教师权限]
        FM --> FM_Admin[管理员权限]
    end
```

## 5. 详细权限映射图

```mermaid
graph TB
    subgraph "学生权限"
        S1[查看个人信息] --> S2[修改个人信息]
        S2 --> S3[查看活动列表]
        S3 --> S4[提交申请]
        S4 --> S5[查看自己的申请]
        S5 --> S6[离开活动]
    end

    subgraph "教师权限"
        T1[学生所有权限] --> T2[管理学生]
        T2 --> T3[创建活动]
        T3 --> T4[编辑活动]
        T4 --> T5[审核申请]
        T5 --> T6[管理参与者]
        T6 --> T7[上传附件]
    end

    subgraph "管理员权限"
        A1[教师所有权限] --> A2[管理教师]
        A2 --> A3[删除活动]
        A3 --> A4[系统配置]
        A4 --> A5[用户管理]
    end

    S6 --> T1
    T7 --> A1
```

## 6. API 路由权限分布

```mermaid
graph LR
    subgraph "公开路由"
        Public[无需认证]
    end

    subgraph "认证路由"
        Auth[需要认证]
    end

    subgraph "学生专用路由"
        Student[仅学生]
    end

    subgraph "教师路由"
        Teacher[教师或管理员]
    end

    subgraph "管理员路由"
        Admin[仅管理员]
    end

    Public --> Auth
    Auth --> Student
    Auth --> Teacher
    Teacher --> Admin
```

## 7. 权限检查时序图

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant Gateway as API Gateway
    participant Auth as 认证服务
    participant Service as 微服务

    Client->>Gateway: 发送请求 (带JWT Token)
    Gateway->>Gateway: JWT Token 验证
    Gateway->>Gateway: 提取用户信息
    Gateway->>Gateway: 权限中间件检查

    alt 权限验证通过
        Gateway->>Service: 转发请求 (带用户信息)
        Service->>Service: 处理业务逻辑
        Service->>Gateway: 返回响应
        Gateway->>Client: 返回结果
    else 权限验证失败
        Gateway->>Client: 返回 403 Forbidden
    end
```

## 8. 权限继承关系图

```mermaid
graph TD
    Base[基础权限] --> Student[学生权限]
    Student --> Teacher[教师权限]
    Teacher --> Admin[管理员权限]

    subgraph "基础权限"
        Base --> ViewProfile[查看个人信息]
        Base --> ViewActivities[查看活动]
        Base --> ViewApplications[查看申请]
    end

    subgraph "学生特有权限"
        Student --> LeaveActivity[离开活动]
        Student --> SubmitApplication[提交申请]
    end

    subgraph "教师特有权限"
        Teacher --> CreateActivity[创建活动]
        Teacher --> ReviewApplication[审核申请]
        Teacher --> ManageStudents[管理学生]
    end

    subgraph "管理员特有权限"
        Admin --> DeleteActivity[删除活动]
        Admin --> ManageTeachers[管理教师]
        Admin --> SystemConfig[系统配置]
    end
```

## 9. 安全控制点图

```mermaid
graph TB
    subgraph "前端安全控制"
        F1[UI权限控制]
        F2[路由守卫]
        F3[API调用控制]
    end

    subgraph "网关安全控制"
        G1[JWT验证]
        G2[权限中间件]
        G3[请求转发]
    end

    subgraph "微服务安全控制"
        M1[用户信息验证]
        M2[业务权限检查]
        M3[数据访问控制]
    end

    F1 --> F2
    F2 --> F3
    F3 --> G1
    G1 --> G2
    G2 --> G3
    G3 --> M1
    M1 --> M2
    M2 --> M3
```

## 10. 权限监控图

```mermaid
graph LR
    subgraph "权限监控"
        Monitor[权限监控系统]
        Alert[告警系统]
        Log[日志系统]
        Audit[审计系统]
    end

    subgraph "监控指标"
        M1[权限拒绝次数]
        M2[异常访问模式]
        M3[权限使用统计]
        M4[安全事件记录]
    end

    Monitor --> M1
    Monitor --> M2
    Monitor --> M3
    Monitor --> M4

    M1 --> Alert
    M2 --> Alert
    M3 --> Log
    M4 --> Audit
```

## 图表说明

### 1. 系统架构图

展示了整个系统的权限控制架构，从客户端到微服务的完整流程。

### 2. 用户角色层级图

显示了三种用户角色的权限继承关系，管理员拥有所有权限。

### 3. 权限控制流程图

详细描述了权限验证的决策流程，包括各种权限检查分支。

### 4. 功能模块权限矩阵

展示了不同功能模块对不同用户类型的权限分配。

### 5. 详细权限映射图

具体列出了每种用户类型可以执行的操作。

### 6. API 路由权限分布

显示了 API 路由的权限层级分布。

### 7. 权限检查时序图

展示了权限检查的时序流程。

### 8. 权限继承关系图

详细展示了权限的继承关系。

### 9. 安全控制点图

标识了系统中的关键安全控制点。

### 10. 权限监控图

展示了权限监控和审计的架构。

这些图表提供了完整的权限控制可视化，帮助理解系统的权限架构和安全机制。
