export type UserRole = "student" | "teacher" | "admin";

export interface ApiSuccessResponse<T> {
  code: string;
  message: string;
  data: T;
  request_id?: string;
}

export interface ApiErrorResponse {
  code: string;
  message: string;
  request_id?: string;
  details?: unknown;
}

export interface SessionUser {
  id: string;
  uuid: string;
  username: string;
  userType: UserRole;
  user_type: UserRole;
  status: "active" | "inactive" | "suspended";
  email?: string;
  phone?: string;
  fullName?: string;
  real_name?: string;
  student_id?: string;
  teacher_id?: string;
  department_id?: string;
  grade?: string;
  title?: string;
  avatar?: string;
  avatar_url?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in?: number;
  refresh_expires_in?: number;
}

export interface AuthPayload extends AuthTokens {
  user: SessionUser;
}
