# 学分管理系统 - 前端

这是学分管理系统的前端应用，基于React 18 + TypeScript + Tailwind CSS构建。

## 功能特性

### 🎯 核心功能
- **用户认证**: 登录、注册、权限管理
- **仪表板**: 系统概览、统计数据、最近活动
- **申请管理**: 学生提交申请、教师审核、文件上传
- **事务管理**: 学分事务类型管理、分类统计
- **学生管理**: 学生信息CRUD、搜索筛选
- **教师管理**: 教师信息CRUD、部门管理
- **个人资料**: 用户信息管理、密码修改

### 🎨 用户体验
- **响应式设计**: 支持桌面端和移动端
- **现代化UI**: 基于shadcn/ui组件库
- **深色模式**: 支持主题切换
- **实时通知**: Toast通知系统
- **加载状态**: 优雅的加载动画
- **错误处理**: 友好的错误提示

### 🔐 权限控制
- **角色权限**: 学生、教师、管理员不同权限
- **路由保护**: 基于权限的页面访问控制
- **API拦截**: 自动token管理和错误处理

## 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **样式**: Tailwind CSS
- **组件库**: shadcn/ui
- **路由**: React Router DOM
- **状态管理**: React Context
- **表单**: React Hook Form + Zod
- **HTTP客户端**: Axios
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
- 记住登录状态

### 注册页面 (`/register`)
- 用户注册
- 角色选择（学生/教师）
- 表单验证
- 密码强度检查

### 仪表板 (`/dashboard`)
- 系统统计概览
- 用户统计卡片
- 申请统计
- 最近活动
- 快速操作

### 申请管理 (`/applications`)
- 申请列表展示
- 状态筛选（pending/approved/rejected）
- 搜索功能
- 申请详情查看
- 文件上传下载
- 审核功能（教师/管理员）

### 事务管理 (`/affairs`)
- 事务类型列表
- 分类筛选
- 状态管理
- 统计信息
- CRUD操作

### 学生管理 (`/students`)
- 学生信息列表
- 搜索和筛选
- 信息编辑
- 批量操作

### 教师管理 (`/teachers`)
- 教师信息列表
- 部门管理
- 职称管理
- 状态管理

## API集成

### 认证相关
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出
- `POST /api/users/register` - 用户注册

### 申请管理
- `GET /api/applications` - 获取申请列表
- `POST /api/applications` - 创建申请
- `PUT /api/applications/:id/status` - 更新申请状态
- `POST /api/applications/:id/files` - 上传文件

### 事务管理
- `GET /api/affairs` - 获取事务列表
- `POST /api/affairs` - 创建事务
- `PUT /api/affairs/:id` - 更新事务
- `DELETE /api/affairs/:id` - 删除事务

## 权限系统

### 学生权限
- 查看自己的申请
- 提交新申请
- 上传文件
- 查看个人资料

### 教师权限
- 查看所有申请
- 审核申请
- 查看学生信息
- 管理事务

### 管理员权限
- 所有权限
- 用户管理
- 系统配置
- 数据统计

## 开发指南

### 添加新页面
1. 在 `src/pages/` 创建新组件
2. 在 `src/App.tsx` 添加路由
3. 在 `src/components/Layout.tsx` 添加导航项

### 添加新组件
1. 在 `src/components/` 创建组件
2. 使用TypeScript定义接口
3. 添加必要的样式

### API调用
使用 `src/lib/api.ts` 中的 `apiClient`：
```typescript
import apiClient from '@/lib/api';

// GET请求
const response = await apiClient.get('/endpoint');

// POST请求
const response = await apiClient.post('/endpoint', data);

// PUT请求
const response = await apiClient.put('/endpoint', data);

// DELETE请求
const response = await apiClient.delete('/endpoint');
```

### 表单处理
使用React Hook Form + Zod：
```typescript
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';

const schema = z.object({
  name: z.string().min(1, '名称不能为空'),
  email: z.string().email('邮箱格式不正确'),
});

const form = useForm({
  resolver: zodResolver(schema),
  defaultValues: { name: '', email: '' },
});
```

## 部署

### Docker部署
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

1. **API连接失败**
   - 检查后端服务是否运行
   - 确认API网关配置
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

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 许可证

MIT License
