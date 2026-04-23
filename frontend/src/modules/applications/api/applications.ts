import apiClient from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface ApplicationListItem {
  id: string;
  status: string;
  applied_credits: number;
  awarded_credits: number;
  academic_year?: string;
  semester?: string;
  activity: {
    id?: string;
    title?: string;
    category?: string;
    owner_id?: string;
  };
}

export interface ApplicationStats {
  total: number;
  approved_count: number;
  total_credits: number;
}

export interface ApplicationListResponse {
  items: ApplicationListItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export async function listApplications() {
  const response = await apiClient.get<ApiSuccessResponse<ApplicationListResponse>>(
    "/applications"
  );

  return response.data.data;
}

export async function getApplicationStats() {
  const response = await apiClient.get<ApiSuccessResponse<ApplicationStats>>(
    "/applications/stats"
  );

  return response.data.data;
}

export async function getApplicationDetail(id: string) {
  const response = await apiClient.get<
    ApiSuccessResponse<{ application: ApplicationListItem }>
  >(`/applications/${id}`);

  return response.data.data.application;
}
