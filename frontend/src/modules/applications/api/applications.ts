import apiClient, { apiHelpers } from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface ApplicationFilters {
  keyword?: string;
  activity_id?: string;
  user_id?: string;
  status?: string;
  academic_year?: string;
  semester?: string;
  page?: number;
  page_size?: number;
}

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

function compactFilters(filters: ApplicationFilters) {
  return Object.fromEntries(
    Object.entries(filters)
      .filter(([, value]) => value !== undefined && value !== "")
      .map(([key, value]) => [key, String(value)])
  );
}

function exportPath(filters: ApplicationFilters) {
  const params = new URLSearchParams(
    compactFilters(filters) as Record<string, string>
  );
  const query = params.toString();

  return query ? `/applications/export?${query}` : "/applications/export";
}

export async function listApplications(filters: ApplicationFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<ApplicationListResponse>>(
    "/applications",
    {
      params: compactFilters(filters),
    }
  );

  return response.data.data;
}

export async function getApplicationStats(filters: ApplicationFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<ApplicationStats>>(
    "/applications/stats",
    {
      params: compactFilters(filters),
    }
  );

  return response.data.data;
}

export async function exportApplications(filters: ApplicationFilters = {}) {
  return apiHelpers.downloadFile(exportPath(filters), "applications.csv");
}

export async function getApplicationDetail(id: string) {
  const response = await apiClient.get<
    ApiSuccessResponse<{ application: ApplicationListItem }>
  >(`/applications/${id}`);

  return response.data.data.application;
}
