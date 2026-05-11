import apiClient, { apiHelpers } from "@/lib/api";
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

export interface ActivityDetail extends ActivityListItem {
  reviewer_id?: string;
  review_comments?: string;
  reviewed_at?: string;
  details?: Record<string, unknown>;
}

export interface ActivityParticipant {
  id: string;
  activity_id: string;
  user_id: string;
  credits: number;
  joined_at: string;
}

export interface ActivityParticipantStats {
  total_count: number;
  total_credits: number;
  average_credits: number;
  recent_new_count: number;
}

export interface ActivityFile {
  id: string;
  activity_id: string;
  name: string;
  original_name?: string;
  content_type?: string;
  file_category?: string;
  description?: string;
  size: number;
  uploaded_by: string;
  created_at: string;
  download_count?: number;
  md5_hash?: string;
}

export interface ActivityFileCount {
  value: string;
  count: number;
}

export interface ActivityFileListResponse {
  items: ActivityFile[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  total_size: number;
  category_counts: ActivityFileCount[];
  type_counts: ActivityFileCount[];
}

export interface ActivityListResponse {
  items: ActivityListItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ActivityListFilters {
  keyword?: string;
  status?: string;
  category?: string;
  academic_year?: string;
  semester?: string;
  page?: number;
  page_size?: number;
}

export interface ActivityMutationInput {
  title: string;
  description?: string;
  category: string;
  academic_year: string;
  semester: string;
  start_date?: string;
  end_date?: string;
  details?: Record<string, unknown>;
}

export interface ReviewActivityInput {
  target_status: "approved" | "rejected";
  review_comments: string;
}

export async function listActivities(filters: ActivityListFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<ActivityListResponse>>(
    "/activities",
    {
      params: filters,
    }
  );

  return response.data.data;
}

export async function getActivityDetail(id: string) {
  const response = await apiClient.get<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >(`/activities/${id}`);

  return response.data.data.activity;
}

export async function createActivity(input: ActivityMutationInput) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >("/activities", input);

  return response.data.data.activity;
}

export async function updateActivity(id: string, input: ActivityMutationInput) {
  const response = await apiClient.patch<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >(`/activities/${id}`, input);

  return response.data.data.activity;
}

export async function submitActivity(id: string) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >(`/activities/${id}/submit`);

  return response.data.data.activity;
}

export async function withdrawActivity(id: string) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >(`/activities/${id}/withdraw`);

  return response.data.data.activity;
}

export async function reviewActivity(id: string, input: ReviewActivityInput) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >(`/activities/${id}/review`, input);

  return response.data.data.activity;
}

export async function copyActivity(id: string) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >(`/activities/${id}/copy`);

  return response.data.data.activity;
}

export async function transferActivityOwner(id: string, newOwnerID: string) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ activity: ActivityDetail }>
  >(`/activities/${id}/transfer-owner`, {
    new_owner_id: newOwnerID,
  });

  return response.data.data.activity;
}

export async function deleteActivity(id: string) {
  await apiClient.delete(`/activities/${id}`);
}

export async function listActivityParticipants(id: string) {
  const response = await apiClient.get<
    ApiSuccessResponse<{
      items: ActivityParticipant[];
      total: number;
      page: number;
      page_size: number;
      total_pages: number;
    }>
  >(`/activities/${id}/participants`);

  return response.data.data.items;
}

export async function getActivityParticipantStats(id: string) {
  const response = await apiClient.get<ApiSuccessResponse<ActivityParticipantStats>>(
    `/activities/${id}/participants/stats`
  );

  return response.data.data;
}

export async function exportActivityParticipants(
  id: string,
  options: { format?: "csv" | "xlsx" } = {}
) {
  const format = options.format === "xlsx" ? "xlsx" : "csv";
  const query = format === "xlsx" ? "?format=xlsx" : "";

  return apiHelpers.downloadFile(
    `/activities/${id}/participants/export${query}`,
    `participants.${format}`
  );
}

export async function addActivityParticipants(
  id: string,
  input: { user_ids: string[]; credits: number }
) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ items: ActivityParticipant[] }>
  >(`/activities/${id}/participants`, input);

  return response.data.data.items;
}

export async function updateActivityParticipantCredits(
  activityID: string,
  participantID: string,
  credits: number
) {
  const response = await apiClient.patch<
    ApiSuccessResponse<{ participant: ActivityParticipant }>
  >(`/activities/${activityID}/participants/${participantID}`, { credits });

  return response.data.data.participant;
}

export async function batchSetActivityParticipantCredits(
  activityID: string,
  participantIDs: string[],
  credits: number
) {
  const response = await apiClient.post<ApiSuccessResponse<{ affected_count: number }>>(
    `/activities/${activityID}/participants/batch-set-credits`,
    {
      participant_ids: participantIDs,
      credits,
    }
  );

  return response.data.data;
}

export async function deleteActivityParticipant(
  activityID: string,
  participantID: string
) {
  await apiClient.delete(`/activities/${activityID}/participants/${participantID}`);
}

export async function batchDeleteActivityParticipants(
  activityID: string,
  participantIDs: string[]
) {
  const response = await apiClient.post<ApiSuccessResponse<{ affected_count: number }>>(
    `/activities/${activityID}/participants/batch-delete`,
    {
      participant_ids: participantIDs,
    }
  );

  return response.data.data;
}

export async function leaveActivity(id: string) {
  await apiClient.post(`/activities/${id}/participants/leave`);
}

export async function listActivityFiles(id: string) {
  const response = await apiClient.get<ApiSuccessResponse<ActivityFileListResponse>>(
    `/activities/${id}/files`
  );

  return response.data.data;
}

export async function uploadActivityFile(id: string, file: File) {
  const items = await uploadActivityFiles(id, [file]);

  return items[0];
}

export async function uploadActivityFiles(
  id: string,
  files: File[],
  metadata: { file_category?: string; description?: string } = {}
) {
  const formData = new FormData();
  files.forEach((file) => {
    formData.append("files[]", file);
  });
  if (metadata.file_category) {
    formData.append("file_category", metadata.file_category);
  }
  if (metadata.description) {
    formData.append("description", metadata.description);
  }

  const response = await apiClient.post<
    ApiSuccessResponse<{ items: ActivityFile[]; file?: ActivityFile }>
  >(`/activities/${id}/files`, formData);

  return response.data.data.items;
}

export async function deleteActivityFile(activityID: string, fileID: string) {
  await apiClient.delete(`/activities/${activityID}/files/${fileID}`);
}

export async function downloadActivityFile(fileID: string, filename?: string) {
  await apiHelpers.downloadFile(`/files/${fileID}/download`, filename);
}

export async function previewActivityFile(fileID: string, filename?: string) {
  await apiHelpers.downloadFile(`/files/${fileID}/preview`, filename);
}
