import { getActivityOptions } from "@/lib/options";
import type { SelectOption } from "@/lib/options";

// 动态活动类型定义 - 从API配置获取
export type ActivityCategory = string;

// 动态活动状态定义 - 从API配置获取  
export type ActivityStatus = string;

// 活动配置类型
export interface ActivityConfig {
  categories: SelectOption[];
  statuses: SelectOption[];
  review_actions: SelectOption[];
  category_fields: Record<string, Array<{
    name: string;
    label: string;
    type: string;
    required?: boolean;
    options?: SelectOption[];
    min?: number;
    max?: number;
    maxLength?: number;
    filterable?: boolean;
  }>>;
}

// 获取活动配置的工具函数
export async function getActivityConfig(): Promise<ActivityConfig> {
  const options = await getActivityOptions();
  return {
    categories: options.categories || [],
    statuses: options.statuses || [],
    review_actions: options.review_actions || [],
    category_fields: options.category_fields || {},
  };
}

// 获取活动分类选项
export async function getActivityCategories(): Promise<SelectOption[]> {
  const config = await getActivityConfig();
  return config.categories;
}

// 获取活动状态选项
export async function getActivityStatuses(): Promise<SelectOption[]> {
  const config = await getActivityConfig();
  return config.statuses;
}

// 获取审核操作选项
export async function getReviewActions(): Promise<SelectOption[]> {
  const config = await getActivityConfig();
  return config.review_actions;
}

// 获取特定分类的字段配置
export async function getCategoryFields(category: string) {
  const config = await getActivityConfig();
  return config.category_fields[category] || [];
}

// 基础活动信息
export interface Activity {
  id: string;
  title: string;
  description: string;
  start_date: string;
  end_date: string;
  status: ActivityStatus;
  category: ActivityCategory;
  owner_id: string;
  owner_info?: UserInfo;
  reviewer_id?: string;
  review_comments?: string;
  reviewed_at?: string;
  created_at: string;
  updated_at: string;
  // 列表场景下由后端返回的聚合字段
  participants_count?: number;
  applications_count?: number;
  participants?: Participant[];
  applications?: Application[];
  // 配置驱动下的通用详情
  details?: Record<string, any>;
}


// 参与者信息
export interface Participant {
  id: string;
  credits: number;
  joined_at: string;
  user_info?: UserInfo;
}

// 申请信息
export interface Application {
  id: string;
  activity_id: string;
  user_id?: string;
  status: string;
  applied_credits: number;
  awarded_credits: number;
  submitted_at: string;
  created_at: string;
  updated_at: string;
  activity?: ActivityInfo;
  user_info?: UserInfo;
}

// 活动信息（用于申请）
export interface ActivityInfo {
  id: string;
  title: string;
  description: string;
  category: ActivityCategory;
  start_date: string;
  end_date: string;
}

// 用户信息（简化版，用于活动相关组件）
export interface UserInfo {
  id: string;
  username: string;
  real_name: string;
  user_type?: 'student' | 'teacher' | 'admin';
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
  grade?: string;
  department?: string;
  title?: string;
  // 向后兼容字段
  name?: string;
  role?: string;
}

// 活动统计
export interface ActivityStats {
  total_activities: number;
  draft_count: number;
  pending_count: number;
  approved_count: number;
  rejected_count: number;
  total_participants: number;
  total_credits: number;
}

// 申请统计
export interface ApplicationStats {
  total_applications: number;
  total_credits: number;
  awarded_credits: number;
}

// 分页响应
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

// 标准响应
export interface StandardResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 活动详情组件属性
export interface ActivityDetailProps {
  activity: Activity;
  isOwner: boolean;
  isTeacherOrAdmin: boolean;
  onRefresh: () => void;
}

// 活动类型详情组件属性 - 使用通用的details字段
export interface ActivityTypeDetailProps {
  activity: Activity;
  detail: Record<string, any>; // 使用通用的details字段，支持动态配置
}

// 附件类型定义
export interface Attachment {
  id: string;
  activity_id: string;
  file_name: string;
  original_name: string;
  file_size: number;
  file_type: string;
  file_category: string;
  description: string;
  uploaded_by: string;
  uploaded_at: string;
  download_count: number;
  download_url: string;
  uploader?: UserInfo;
} 