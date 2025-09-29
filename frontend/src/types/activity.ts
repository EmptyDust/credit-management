// 活动类型定义
export type ActivityCategory = 
  | "创新创业实践活动"
  | "学科竞赛"
  | "大学生创业项目"
  | "创业实践项目"
  | "论文专利";

// 活动类别常量
export const ACTIVITY_CATEGORIES: ActivityCategory[] = [
  "创新创业实践活动",
  "学科竞赛", 
  "大学生创业项目",
  "创业实践项目",
  "论文专利"
];

// 活动状态
export type ActivityStatus = 
  | "draft"           // 草稿
  | "pending_review"  // 待审核
  | "approved"        // 已通过
  | "rejected";       // 已拒绝

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

// 活动类型详情组件属性
export interface ActivityTypeDetailProps {
  activity: Activity;
  detail: InnovationActivityDetail | CompetitionActivityDetail | EntrepreneurshipProjectDetail | EntrepreneurshipPracticeDetail | PaperPatentDetail;
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