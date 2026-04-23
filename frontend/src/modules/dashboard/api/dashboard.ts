import apiClient from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface DashboardOverview {
  user_count: number;
  activity_count: number;
  application_count: number;
}

export async function getDashboardOverview() {
  const response = await apiClient.get<ApiSuccessResponse<DashboardOverview>>(
    "/dashboard/overview"
  );

  return response.data.data;
}
