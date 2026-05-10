import apiClient from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface DashboardOverview {
  user_count: number;
  activity_count: number;
  application_count: number;
  user_stats: {
    total: number;
    student_count: number;
    teacher_count: number;
    admin_count: number;
    active_count: number;
    disabled_count: number;
    role_counts: Record<string, number>;
    status_counts: Record<string, number>;
  };
  activity_stats: {
    status_counts: Record<string, number>;
    category_counts: Array<KeyLabelCount>;
    academic_year_counts: Array<ValueCount>;
    semester_counts: Array<ValueCount>;
    mine_count: number;
    participated_count: number;
  };
  application_stats: {
    total: number;
    approved_count: number;
    total_credits: number;
  };
  category_distribution: Array<KeyLabelCount>;
  recent_activities: Array<ActivitySummary>;
  popular_activities: Array<ActivitySummary>;
}

export interface KeyLabelCount {
  key: string;
  label: string;
  count: number;
}

export interface ValueCount {
  value: string;
  count: number;
}

export interface ActivitySummary {
  id: string;
  title: string;
  category: string;
  academic_year: string;
  semester: string;
  status: string;
  owner_id: string;
  participant_count: number;
}

export async function getDashboardOverview() {
  const response = await apiClient.get<ApiSuccessResponse<DashboardOverview>>(
    "/dashboard/overview"
  );

  return response.data.data;
}
