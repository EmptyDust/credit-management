import apiClient from "@/lib/api";
import type { ActivityCategorySchemaEnvelope } from "@/modules/activities/schema/types";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface ActivityCategorySummary {
  key: string;
  label: string;
  sort_order: number;
}

export async function listActivityCategories() {
  const response = await apiClient.get<
    ApiSuccessResponse<{ items: ActivityCategorySummary[] }>
  >("/activity-categories");

  return response.data.data.items;
}

export async function getActivityCategorySchema(categoryKey: string) {
  const response = await apiClient.get<
    ApiSuccessResponse<ActivityCategorySchemaEnvelope>
  >(`/activity-categories/${encodeURIComponent(categoryKey)}/schema`);

  return response.data.data;
}
