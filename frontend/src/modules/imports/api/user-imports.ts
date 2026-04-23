import apiClient, { apiHelpers } from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface ImportValidationError {
  field: string;
  code: string;
  message: string;
}

export interface ImportRowResult {
  row_no: number;
  status: string;
  errors: ImportValidationError[];
}

export interface ImportSummary {
  total_rows: number;
  valid_rows: number;
  invalid_rows: number;
}

export interface ImportPrecheckResult {
  job_id: string;
  job_type: string;
  file_name: string;
  file_hash: string;
  expires_at: string;
  summary: ImportSummary;
  rows: ImportRowResult[];
}

export interface ImportConfirmResult {
  job_id: string;
  stage: string;
  summary: ImportSummary;
  created_user_ids: string[];
}

export async function downloadUserImportTemplate(userType: "student" | "teacher") {
  const path =
    userType === "student"
      ? "/imports/templates/students"
      : "/imports/templates/teachers";
  const filename =
    userType === "student" ? "students-template.csv" : "teachers-template.csv";

  return apiHelpers.downloadFile(path, filename);
}

export async function precheckUserImport(
  userType: "student" | "teacher",
  file: File
) {
  const formData = new FormData();
  formData.append("file", file);

  const path =
    userType === "student"
      ? "/imports/students/precheck"
      : "/imports/teachers/precheck";

  const response = await apiClient.post<ApiSuccessResponse<ImportPrecheckResult>>(
    path,
    formData
  );

  return response.data.data;
}

export async function confirmUserImport(
  userType: "student" | "teacher",
  payload: { job_id: string; file_hash: string }
) {
  const path =
    userType === "student"
      ? "/imports/students/confirm"
      : "/imports/teachers/confirm";

  const response = await apiClient.post<ApiSuccessResponse<ImportConfirmResult>>(
    path,
    payload
  );

  return response.data.data;
}
