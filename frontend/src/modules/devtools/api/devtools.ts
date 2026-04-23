import apiClient, { apiHelpers } from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface DevtoolsHealth {
  database: { status: string };
  redis: { status: string };
  storage: { status: string };
  version: string;
}

export interface DevtoolsLogEntry {
  id?: string;
  level?: string;
  message?: string;
  created_at?: string;
}

export interface AuditLogEntry {
  id: string;
  actor_id?: string;
  actor_role?: string;
  target_type?: string;
  target_id?: string;
  action: string;
  result?: string;
  request_id?: string;
  created_at: string;
}

export async function getDevtoolsHealth() {
  const response = await apiClient.get<ApiSuccessResponse<DevtoolsHealth>>(
    "/devtools/health"
  );
  return response.data.data;
}

export async function getDevtoolsLogs() {
  const response = await apiClient.get<
    ApiSuccessResponse<{
      items: DevtoolsLogEntry[];
      total: number;
    }>
  >("/devtools/logs");
  return response.data.data;
}

export async function getAuditLogs() {
  const response = await apiClient.get<
    ApiSuccessResponse<{
      items: AuditLogEntry[];
      total: number;
    }>
  >("/devtools/audit-logs");
  return response.data.data;
}

export async function exportAuditLogs() {
  return apiHelpers.downloadFile("/devtools/audit-logs/export", "audit-logs.csv");
}
