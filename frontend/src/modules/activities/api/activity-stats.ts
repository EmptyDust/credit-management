import apiClient, { apiHelpers } from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface ActivityStatsFilters {
  keyword?: string;
  status?: string;
  category?: string;
  academic_year?: string;
  semester?: string;
}

export interface ActivityKeyLabelCount {
  key: string;
  label: string;
  count: number;
}

export interface ActivityValueCount {
  value: string;
  count: number;
}

export interface ActivityStats {
  status_counts: Record<string, number>;
  category_counts: ActivityKeyLabelCount[];
  academic_year_counts: ActivityValueCount[];
  semester_counts: ActivityValueCount[];
  mine_count: number;
  participated_count: number;
}

function compactFilters(filters: ActivityStatsFilters) {
  return Object.fromEntries(
    Object.entries(filters).filter(([, value]) => value !== undefined && value !== "")
  );
}

function exportPath(filters: ActivityStatsFilters) {
  const params = new URLSearchParams(
    compactFilters(filters) as Record<string, string>
  );
  const query = params.toString();

  return query ? `/activities/export?${query}` : "/activities/export";
}

export async function getActivityStats(filters: ActivityStatsFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<ActivityStats>>(
    "/activities/stats",
    {
      params: compactFilters(filters),
    }
  );

  return response.data.data;
}

export async function exportActivities(filters: ActivityStatsFilters = {}) {
  return apiHelpers.downloadFile(exportPath(filters), "activities.csv");
}
