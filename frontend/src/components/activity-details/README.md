# 活动详情组件系统

## 概述

活动详情组件系统为五种不同的活动类型提供了专门的显示界面，每种活动类型都有独特的 UI 设计和信息展示方式。

## 活动类型

系统支持以下五种活动类型：

1. **创新创业实践活动** - 创新实践、实习项目等
2. **学科竞赛** - 各类学科竞赛和比赛
3. **大学生创业项目** - 创业项目相关活动
4. **创业实践项目** - 创业实践和公司运营
5. **论文专利** - 论文发表、专利申请等

## 组件结构

```
activity-details/
├── index.ts                           # 组件导出文件
├── ActivityDetailContainer.tsx        # 活动详情容器（路由组件）
├── InnovationActivityDetail.tsx       # 创新创业活动详情
├── CompetitionActivityDetail.tsx      # 学科竞赛活动详情
├── EntrepreneurshipProjectDetail.tsx  # 大学生创业项目详情
├── EntrepreneurshipPracticeDetail.tsx # 创业实践项目详情
├── PaperPatentDetail.tsx              # 论文专利详情
└── README.md                          # 本文档
```

## 核心组件

### ActivityDetailContainer

活动详情容器组件，根据活动类别自动路由到对应的详情组件。

```tsx
import { ActivityDetailContainer } from "@/components/activity-details";

function MyComponent() {
  const activity = {
    /* 活动数据 */
  };

  return <ActivityDetailContainer activity={activity} />;
}
```

### 各类型详情组件

每个活动类型都有专门的详情组件，包含：

- **ActivityBasicInfo** - 活动基本信息
- **ActivityActions** - 活动操作按钮
- **ActivityParticipants** - 参与者列表
- **ActivityApplications** - 申请列表
- **特定详情信息** - 根据活动类型显示特定字段

## 数据结构

### 基础活动信息

```typescript
interface Activity {
  id: string;
  title: string;
  description: string;
  start_date: string;
  end_date: string;
  status: ActivityStatus;
  category: ActivityCategory;
  requirements: string;
  owner_id: string;
  // ... 其他字段
}
```

### 活动详情类型

每种活动类型都有对应的详情数据结构：

```typescript
// 创新创业实践活动详情
interface InnovationActivityDetail {
  item: string; // 实践事项
  company: string; // 实习公司
  project_no: string; // 课题编号
  issuer: string; // 发证机构
  date: string; // 实践日期
  total_hours: number; // 累计学时
}

// 学科竞赛详情
interface CompetitionActivityDetail {
  level: string; // 竞赛级别
  competition: string; // 竞赛名称
  award_level: string; // 获奖等级
  rank: string; // 排名
}

// 大学生创业项目详情
interface EntrepreneurshipProjectDetail {
  project_name: string; // 项目名称
  project_level: string; // 项目等级
  project_rank: string; // 项目排名
}

// 创业实践项目详情
interface EntrepreneurshipPracticeDetail {
  company_name: string; // 公司名称
  legal_person: string; // 公司法人
  share_percent: number; // 占股比例
}

// 论文专利详情
interface PaperPatentDetail {
  name: string; // 名称
  category: string; // 类别
  rank: string; // 排名
}
```

## 使用方式

### 基本使用

```tsx
import { ActivityDetailContainer } from "@/components/activity-details";
import { ActivityWithDetails } from "@/types/activity";

function MyComponent() {
  const activity: ActivityWithDetails = {
    /* 活动数据 */
  };

  return <ActivityDetailContainer activity={activity} />;
}
```

### 直接使用特定组件

```tsx
import { InnovationActivityDetail } from "@/components/activity-details";

function MyComponent() {
  const activity = {
    /* 活动数据 */
  };
  const detail = {
    /* 创新创业详情数据 */
  };

  return <InnovationActivityDetail activity={activity} detail={detail} />;
}
```

## 样式和主题

### 设计原则

1. **一致性** - 所有组件使用统一的设计语言
2. **可访问性** - 支持键盘导航和屏幕阅读器
3. **响应式** - 适配不同屏幕尺寸
4. **主题化** - 支持明暗主题切换

### 颜色方案

每种活动类型都有独特的颜色标识：

- **创新创业** - 黄色系 (`text-yellow-600`)
- **学科竞赛** - 黄色系 (`text-yellow-600`)
- **创业项目** - 蓝色系 (`text-blue-600`)
- **创业实践** - 绿色系 (`text-green-600`)
- **论文专利** - 紫色系 (`text-purple-600`)

## 扩展指南

### 添加新的活动类型

1. 在 `types/activity.ts` 中定义新的详情接口
2. 创建对应的详情组件
3. 在 `ActivityDetailContainer` 中添加路由逻辑
4. 更新 `index.ts` 导出文件

### 自定义样式

每个组件都支持通过 CSS 类名进行样式自定义：

```tsx
<InnovationActivityDetail
  activity={activity}
  detail={detail}
  className="custom-styles"
/>
```

## 最佳实践

1. **数据安全** - 所有组件都会检查数据是否存在再显示
2. **性能优化** - 使用条件渲染避免不必要的计算
3. **可访问性** - 使用语义化的 HTML 标签和 ARIA 属性
4. **国际化** - 所有文本都使用中文，便于后续国际化

## 注意事项

1. 确保传入的 `activity` 对象包含完整的活动信息
2. 详情数据是可选的，组件会优雅地处理缺失数据
3. 所有组件都支持 TypeScript 类型检查
4. 组件会自动处理加载状态和错误状态

## 更新日志

### v2.0.0

- 重构活动类型系统，支持五种标准活动类型
- 删除重复字段，优化数据结构
- 统一组件接口和样式
- 改进类型安全性

### v1.0.0

- 初始版本，支持基础活动详情显示
