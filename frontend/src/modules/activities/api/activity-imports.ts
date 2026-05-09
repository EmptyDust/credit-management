import apiClient from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface ActivityImportValidationError {
  field: string;
  code: string;
  message: string;
}

export interface ActivityImportSummary {
  total_rows: number;
  valid_rows: number;
  invalid_rows: number;
}

export interface ActivityImportRowResult {
  row_no: number;
  status: string;
  payload: Record<string, unknown>;
  errors: ActivityImportValidationError[];
}

export interface ActivityImportPrecheckResult {
  job_id: string;
  job_type: string;
  file_name: string;
  file_hash: string;
  expires_at: string;
  summary: ActivityImportSummary;
  rows: ActivityImportRowResult[];
}

export interface ActivityImportConfirmResult {
  job_id: string;
  stage: string;
  summary: ActivityImportSummary;
  created_activity_ids: string[];
}

export async function precheckActivityImport(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  const response = await apiClient.post<
    ApiSuccessResponse<ActivityImportPrecheckResult>
  >("/imports/activities/precheck", formData);

  return response.data.data;
}

export async function confirmActivityImport(payload: {
  job_id: string;
  file_hash: string;
}) {
  const response = await apiClient.post<
    ApiSuccessResponse<ActivityImportConfirmResult>
  >("/imports/activities/confirm", payload);

  return response.data.data;
}
