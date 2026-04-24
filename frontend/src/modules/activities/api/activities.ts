import apiClient from "@/lib/api";
import { apiHelpers } from "@/lib/api";
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

export interface ActivityFile {
  id: string;
  activity_id: string;
  name: string;
  content_type?: string;
  size: number;
  uploaded_by: string;
  created_at: string;
}

export interface ActivityListResponse {
  items: ActivityListItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
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

export async function listActivities(filters: {
  keyword?: string;
  status?: string;
  category?: string;
  academic_year?: string;
  semester?: string;
} = {}) {
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

export async function addActivityParticipants(
  id: string,
  input: { user_ids: string[]; credits: number }
) {
  const response = await apiClient.post<
    ApiSuccessResponse<{ items: ActivityParticipant[] }>
  >(`/activities/${id}/participants`, input);

  return response.data.data.items;
}

export async function leaveActivity(id: string) {
  await apiClient.post(`/activities/${id}/participants/leave`);
}

export async function listActivityFiles(id: string) {
  const response = await apiClient.get<
    ApiSuccessResponse<{
      items: ActivityFile[];
      total: number;
      page: number;
      page_size: number;
      total_pages: number;
    }>
  >(`/activities/${id}/files`);

  return response.data.data.items;
}

export async function uploadActivityFile(id: string, file: File) {
  const formData = new FormData();
  formData.append("file", file);

  const response = await apiClient.post<
    ApiSuccessResponse<{ file: ActivityFile }>
  >(`/activities/${id}/files`, formData);

  return response.data.data.file;
}

export async function deleteActivityFile(activityID: string, fileID: string) {
  await apiClient.delete(`/activities/${activityID}/files/${fileID}`);
}

export async function downloadActivityFile(fileID: string, filename?: string) {
  await apiHelpers.downloadFile(`/files/${fileID}/download`, filename);
}
