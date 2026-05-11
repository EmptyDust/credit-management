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

export interface UserListFilters {
  keyword?: string;
  status?: string;
  grade?: string;
  title?: string;
  page?: number;
  page_size?: number;
}

export interface UserStatsCount {
  key: string;
  label: string;
  count: number;
}

export interface UserStats {
  total: number;
  status_counts: Record<string, number>;
  grade_counts?: UserStatsCount[];
  title_counts?: UserStatsCount[];
}

export interface UserMutationInput {
  username?: string;
  password?: string;
  real_name?: string;
  email?: string;
  phone?: string;
  department_id?: string;
  student_id?: string;
  teacher_id?: string;
  grade?: string;
  title?: string;
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

export async function listStudents(filters: UserListFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<UserListResponse>>(
    "/students",
    {
      params: filters,
    }
  );

  return response.data.data;
}

export async function listTeachers(filters: UserListFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<UserListResponse>>(
    "/teachers",
    {
      params: filters,
    }
  );

  return response.data.data;
}

export async function getStudentStats(filters: UserListFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<UserStats>>(
    "/students/stats",
    {
      params: filters,
    }
  );

  return response.data.data;
}

export async function getTeacherStats(filters: UserListFilters = {}) {
  const response = await apiClient.get<ApiSuccessResponse<UserStats>>(
    "/teachers/stats",
    {
      params: filters,
    }
  );

  return response.data.data;
}

export async function createStudent(input: UserMutationInput) {
  const response = await apiClient.post<ApiSuccessResponse<{ user: AdminUser }>>(
    "/students",
    input
  );

  return response.data.data.user;
}

export async function createTeacher(input: UserMutationInput) {
  const response = await apiClient.post<ApiSuccessResponse<{ user: AdminUser }>>(
    "/teachers",
    input
  );

  return response.data.data.user;
}

export async function updateStudent(id: string, input: UserMutationInput) {
  const response = await apiClient.patch<ApiSuccessResponse<{ user: AdminUser }>>(
    `/students/${id}`,
    input
  );

  return response.data.data.user;
}

export async function updateTeacher(id: string, input: UserMutationInput) {
  const response = await apiClient.patch<ApiSuccessResponse<{ user: AdminUser }>>(
    `/teachers/${id}`,
    input
  );

  return response.data.data.user;
}

export async function deleteStudent(id: string) {
  await apiClient.delete(`/students/${id}`);
}

export async function deleteTeacher(id: string) {
  await apiClient.delete(`/teachers/${id}`);
}

export async function updateUserStatus(id: string, status: string) {
  const response = await apiClient.patch<ApiSuccessResponse<{ user: AdminUser }>>(
    `/users/${id}/status`,
    { status }
  );

  return response.data.data.user;
}

export async function resetUserPassword(id: string, newPassword: string) {
  await apiClient.post(`/users/${id}/reset-password`, { new_password: newPassword });
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

function compactFilters(filters: UserListFilters) {
  return Object.fromEntries(
    Object.entries(filters).filter(([, value]) => value !== undefined && value !== "")
  );
}

function exportPath(path: string, filters: UserListFilters) {
  const params = new URLSearchParams(
    compactFilters(filters) as Record<string, string>
  );
  const query = params.toString();

  return query ? `${path}?${query}` : path;
}

export async function exportStudents(filters: UserListFilters = {}) {
  return apiHelpers.downloadFile(
    exportPath("/students/export", filters),
    "students.csv"
  );
}

export async function exportTeachers(filters: UserListFilters = {}) {
  return apiHelpers.downloadFile(
    exportPath("/teachers/export", filters),
    "teachers.csv"
  );
}
