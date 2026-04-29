# 双创分申请平台 - 前端

这是双创分申请平台的前端应用，基于 React 18 + TypeScript + Tailwind CSS 构建。

## 功能特性

### 🎯 核心功能

- **用户认证**: 登录、注册、JWT token 管理、权限控制
- **仪表板**: 系统概览、统计数据、最近活动、快速操作
- **活动管理**: 创建活动、编辑活动、提交审核、审核活动、批量操作、导入导出
- **参与者管理**: 添加参与者、设置学分、批量操作、退出活动
- **申请管理**: 查看申请列表、申请详情、申请统计、导出申请数据
- **附件管理**: 上传附件、批量上传、预览下载、管理附件
- **用户管理**: 学生信息 CRUD、教师信息管理、搜索筛选、批量操作
- **搜索功能**: 高级搜索活动、申请、参与者、附件
- **个人资料**: 用户信息管理、密码修改

### 🎨 用户体验

- **响应式设计**: 支持桌面端和移动端
- **现代化 UI**: 基于 shadcn/ui 组件库
- **深色模式**: 支持主题切换
- **实时通知**: Toast 通知系统
- **加载状态**: 优雅的加载动画
- **错误处理**: 友好的错误提示

### 🔐 权限控制

- **角色权限**: 学生、教师、管理员不同权限
- **路由保护**: 基于权限的页面访问控制
- **API 拦截**: 自动 token 管理和错误处理

## 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **样式**: Tailwind CSS
- **组件库**: shadcn/ui
- **路由**: React Router DOM
- **状态管理**: React Context
- **表单**: React Hook Form + Zod
- **HTTP 客户端**: Axios
- **通知**: React Hot Toast
- **图标**: Lucide React

## 快速开始

### 环境要求

- Node.js 18+
- pnpm (推荐) 或 npm

### 安装依赖

```bash
cd frontend
pnpm install
```

### 开发模式

```bash
pnpm dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
pnpm build
```

### 预览生产版本

```bash
pnpm preview
```

## 项目结构

```
src/
├── components/          # 组件
│   ├── ui/             # UI组件库
│   ├── Layout.tsx      # 布局组件
│   ├── ProtectedRoute.tsx # 路由保护
│   └── ThemeToggle.tsx # 主题切换
├── contexts/           # 上下文
│   ├── AuthContext.tsx # 认证上下文
│   └── ThemeContext.tsx # 主题上下文
├── lib/                # 工具库
│   ├── api.ts          # API客户端
│   └── utils.ts        # 工具函数
├── pages/              # 页面组件
│   ├── Dashboard.tsx   # 仪表板
│   ├── Login.tsx       # 登录
│   ├── Register.tsx    # 注册
│   ├── Applications.tsx # 申请管理
│   ├── Affairs.tsx     # 事务管理
│   ├── Students.tsx    # 学生管理
│   ├── Teachers.tsx    # 教师管理
│   └── Profile.tsx     # 个人资料
├── App.tsx             # 应用入口
└── main.tsx            # 主入口
```

## 页面功能

### 登录页面 (`/login`)

- 用户名/密码登录
- 表单验证
- 错误提示
- JWT token 管理

### 注册页面 (`/register`)

- 用户注册
- 角色选择（学生/教师）
- 表单验证
- 密码强度检查

### 仪表板 (`/dashboard`)

- 系统统计概览
- 用户统计卡片
- 活动统计
- 申请统计
- 最近活动
- 快速操作

### 活动管理 (`/activities`)

- 活动列表展示（支持分页）
- 状态筛选（draft/pending_review/approved/rejected）
- 类别筛选
- 搜索功能
- 创建活动（支持模板）
- 编辑活动（仅草稿状态）
- 提交审核
- 撤回活动
- 审核活动（教师/管理员）
- 批量操作（批量创建、更新、删除）
- 导入导出（CSV/Excel）
- 活动详情查看

### 活动详情 (`/activities/:id`)

- 活动基本信息展示
- 参与者管理
- 附件管理
- 审核操作（教师/管理员）
- 活动统计

### 申请管理 (`/applications`)

- 申请列表展示（学生只看自己的，教师/管理员看全部）
- 状态筛选
- 搜索功能
- 申请详情查看
- 申请统计
- 导出申请数据

### 学生管理 (`/students`)

- 学生信息列表（支持分页）
- 搜索和筛选（按姓名、用户名、学号、学部、专业等）
- 信息编辑
- 批量操作（批量删除、批量更新状态）
- 学生统计

### 教师管理 (`/teachers`)

- 教师信息列表（支持分页）
- 搜索和筛选
- 部门管理
- 职称管理
- 状态管理
- 教师统计

## API 集成

### 认证相关

- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出
- `POST /api/users/register` - 用户注册

### 活动管理

- `GET /api/activities` - 获取活动列表（支持分页、搜索、筛选）
- `POST /api/activities` - 创建活动
- `GET /api/activities/:id` - 获取活动详情
- `PUT /api/activities/:id` - 更新活动
- `DELETE /api/activities/:id` - 删除活动
- `POST /api/activities/:id/submit` - 提交活动审核
- `POST /api/activities/:id/withdraw` - 撤回活动
- `POST /api/activities/:id/review` - 审核活动
- `POST /api/activities/:id/copy` - 复制活动
- `GET /api/activities/categories` - 获取活动类别
- `GET /api/activities/templates` - 获取活动模板

### 参与者管理

- `GET /api/activities/:id/participants` - 获取参与者列表
- `POST /api/activities/:id/participants` - 添加参与者
- `PUT /api/activities/:id/participants/:uuid/credits` - 设置单个学分
- `PUT /api/activities/:id/participants/batch-credits` - 批量设置学分
- `DELETE /api/activities/:id/participants/:uuid` - 删除参与者
- `POST /api/activities/:id/participants/batch-remove` - 批量删除参与者
- `POST /api/activities/:id/leave` - 退出活动（仅学生）
- `GET /api/activities/:id/participants/stats` - 获取参与者统计
- `GET /api/activities/:id/participants/export` - 导出参与者列表

### 申请管理

- `GET /api/applications` - 获取用户申请列表
- `GET /api/applications/:id` - 获取申请详情
- `GET /api/applications/all` - 获取所有申请（教师/管理员）
- `GET /api/applications/stats` - 获取申请统计
- `GET /api/applications/export` - 导出申请数据

### 附件管理

- `GET /api/activities/:id/attachments` - 获取附件列表
- `POST /api/activities/:id/attachments` - 上传附件
- `POST /api/activities/:id/attachments/batch` - 批量上传附件
- `GET /api/activities/:id/attachments/:attachment_id/download` - 下载附件
- `GET /api/activities/:id/attachments/:attachment_id/preview` - 预览附件
- `PUT /api/activities/:id/attachments/:attachment_id` - 更新附件信息
- `DELETE /api/activities/:id/attachments/:attachment_id` - 删除附件

## 权限系统

### 学生权限

- 创建和编辑自己的活动
- 提交和撤回活动审核
- 查看自己的申请
- 导出自己的申请数据
- 参与活动（通过被添加）
- 退出活动
- 上传附件（仅在自己创建的活动）
- 查看个人资料

### 教师权限

- 学生所有权限
- 审核所有活动
- 查看所有申请
- 导出所有申请数据
- 管理学生信息
- 批量操作活动
- 导入导出活动

### 管理员权限

- 教师所有权限
- 删除活动
- 管理教师信息
- 用户管理
- 系统配置
- 完整的数据统计

## 开发指南

### 添加新页面

1. 在 `src/pages/` 创建新组件
2. 在 `src/App.tsx` 添加路由
3. 在 `src/components/Layout.tsx` 添加导航项

### 添加新组件

1. 在 `src/components/` 创建组件
2. 使用 TypeScript 定义接口
3. 添加必要的样式

### API 调用

使用 `src/lib/api.ts` 中的 `apiClient`：

```typescript
import apiClient from "@/lib/api";

// GET请求
const response = await apiClient.get("/endpoint");

// POST请求
const response = await apiClient.post("/endpoint", data);

// PUT请求
const response = await apiClient.put("/endpoint", data);

// DELETE请求
const response = await apiClient.delete("/endpoint");
```

### 表单处理

使用 React Hook Form + Zod：

```typescript
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

const schema = z.object({
  name: z.string().min(1, "名称不能为空"),
  email: z.string().email("邮箱格式不正确"),
});

const form = useForm({
  resolver: zodResolver(schema),
  defaultValues: { name: "", email: "" },
});
```

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t credit-management-frontend .

# 运行容器
docker run -p 3000:3000 credit-management-frontend
```

### 静态部署

```bash
# 构建
pnpm build

# 部署到Nginx
cp -r dist/* /var/www/html/
```

## 故障排除

### 常见问题

1. **API 连接失败**

   - 检查后端服务是否运行
   - 确认前端代理与后端地址配置
   - 检查网络连接

2. **权限问题**

   - 确认用户角色设置
   - 检查权限配置
   - 重新登录

3. **文件上传失败**
   - 检查文件大小限制
   - 确认文件格式支持
   - 检查存储空间

### 调试模式

```bash
# 开启详细日志
DEBUG=* pnpm dev
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request
