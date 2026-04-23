import apiClient from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface ActivityListItem {
  id: string;
  title: string;
  description?: string;
  category: string;
  academic_year: string;
  semester: string;
  start_date?: string;
  end_date?: string;
  status: string;
  owner_id: string;
}

export interface ActivityListResponse {
  items: ActivityListItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export async function listActivities(filters: {
  keyword?: string;
  status?: string;
  category?: string;
} = {}) {
  const response = await apiClient.get<ApiSuccessResponse<ActivityListResponse>>(
    "/activities",
    {
      params: filters,
    }
  );

  return response.data.data;
}
