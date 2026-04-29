import apiClient, { apiHelpers } from "@/lib/api";
import type { ApiSuccessResponse } from "@/shared/types/api";

export interface AdminUser {
  id: string;
  username: string;
  real_name: string;
  user_type: string;
  status: string;
  department_id?: string;
  student_id?: string;
  teacher_id?: string;
  grade?: string;
  title?: string;
}

export interface UserListResponse {
  items: AdminUser[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface Profile {
  id: string;
  username: string;
  email?: string;
  phone?: string;
  real_name: string;
  user_type: string;
  status: string;
  department_id?: string;
  student_id?: string;
  teacher_id?: string;
  grade?: string;
  title?: string;
  avatar_url?: string;
}

export async function listStudents(filters: { keyword?: string; status?: string } = {}) {
  const response = await apiClient.get<ApiSuccessResponse<UserListResponse>>(
    "/students",
    {
      params: filters,
    }
  );

  return response.data.data;
}

export async function listTeachers(filters: { keyword?: string; status?: string } = {}) {
  const response = await apiClient.get<ApiSuccessResponse<UserListResponse>>(
    "/teachers",
    {
      params: filters,
    }
  );

  return response.data.data;
}

export async function getProfile() {
  const response = await apiClient.get<ApiSuccessResponse<Profile>>("/me/profile");
  return response.data.data;
}

export async function updateProfile(payload: {
  email?: string;
  phone?: string;
  real_name?: string;
  grade?: string;
  title?: string;
  department_id?: string;
}) {
  const response = await apiClient.patch<ApiSuccessResponse<Profile>>(
    "/me/profile",
    payload
  );

  return response.data.data;
}

export async function exportStudents() {
  return apiHelpers.downloadFile("/students/export", "students.csv");
}

export async function exportTeachers() {
  return apiHelpers.downloadFile("/teachers/export", "teachers.csv");
}
