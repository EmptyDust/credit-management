import apiClient from './api';

// 基础用户信息（基本信息视图）
export interface UserBasicInfo {
  id: string;
  username: string;
  real_name: string;
  avatar?: string;
}

// 学生基本信息
export interface StudentBasicInfo extends UserBasicInfo {
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
  grade?: string;
}

// 教师基本信息
export interface TeacherBasicInfo extends UserBasicInfo {
  department?: string;
  title?: string;
}

// 用户详细信息（详细信息视图）
export interface UserDetailInfo extends UserBasicInfo {
  email: string;
  phone?: string;
  status: 'active' | 'inactive' | 'suspended';
  last_login_at?: string;
  register_time: string;
}

// 学生详细信息
export interface StudentDetailInfo extends UserDetailInfo {
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
  grade?: string;
}

// 教师详细信息
export interface TeacherDetailInfo extends UserDetailInfo {
  department?: string;
  title?: string;
}

// 用户完整信息（完整信息视图）
export interface UserCompleteInfo extends UserDetailInfo {
  user_type: 'student' | 'teacher' | 'admin';
  created_at: string;
  updated_at: string;
}

// 学生完整信息
export interface StudentCompleteInfo extends UserCompleteInfo {
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
  grade?: string;
}

// 教师完整信息
export interface TeacherCompleteInfo extends UserCompleteInfo {
  department?: string;
  title?: string;
}

// 兼容性接口（保持向后兼容）
export interface UserInfo extends UserCompleteInfo {
  // 保持原有字段，但实际使用时根据权限返回不同级别的信息
}

export interface StudentInfo extends StudentCompleteInfo {
  // 保持原有字段，但实际使用时根据权限返回不同级别的信息
}

export interface TeacherInfo extends TeacherCompleteInfo {
  // 保持原有字段，但实际使用时根据权限返回不同级别的信息
}

export interface UserSearchParams {
  page: number;
  page_size: number;
  user_type: 'student' | 'teacher';
  query?: string;
  college?: string;
  major?: string;
  class?: string;
  grade?: string;
  department?: string;
  title?: string;
  status?: 'active' | 'inactive' | 'suspended';
}

export interface ViewBasedSearchResponse {
  users: any[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  view_type: string;
}

export interface UserStats {
  total: number;
  students: number;
  teachers: number;
  admins: number;
  active: number;
  inactive: number;
}

class UserService {
  /**
   * 获取当前用户信息
   */
  async getCurrentUser(): Promise<UserInfo> {
    const response = await apiClient.get('/users/profile');
    return response.data.data;
  }

  /**
   * 更新当前用户信息
   */
  async updateCurrentUser(userData: Partial<UserInfo>): Promise<UserInfo> {
    const response = await apiClient.put('/users/profile', userData);
    return response.data.data;
  }

  /**
   * 获取用户列表（使用视图进行权限控制）
   */
  async searchUsers(params: UserSearchParams): Promise<ViewBasedSearchResponse> {
    const response = await apiClient.get('/search/users', { params });
    return response.data.data;
  }

  /**
   * 根据ID获取用户信息
   */
  async getUserById(userId: string): Promise<UserInfo> {
    const response = await apiClient.get(`/users/${userId}`);
    return response.data.data;
  }

  /**
   * 获取学生列表（使用视图进行权限控制）
   */
  async getStudents(params?: {
    page?: number;
    page_size?: number;
    college?: string;
    major?: string;
    class?: string;
    grade?: string;
    status?: string;
    query?: string;
  }): Promise<ViewBasedSearchResponse> {
    const searchParams: UserSearchParams = {
      page: params?.page || 1,
      page_size: params?.page_size || 10,
      user_type: 'student',
      query: params?.query,
      college: params?.college,
      major: params?.major,
      class: params?.class,
      grade: params?.grade,
      status: params?.status as any,
    };
    
    return this.searchUsers(searchParams);
  }

  /**
   * 获取教师列表（使用视图进行权限控制）
   */
  async getTeachers(params?: {
    page?: number;
    page_size?: number;
    department?: string;
    title?: string;
    status?: string;
    query?: string;
  }): Promise<ViewBasedSearchResponse> {
    const searchParams: UserSearchParams = {
      page: params?.page || 1,
      page_size: params?.page_size || 10,
      user_type: 'teacher',
      query: params?.query,
      department: params?.department,
      title: params?.title,
      status: params?.status as any,
    };
    
    return this.searchUsers(searchParams);
  }

  /**
   * 获取用户统计信息
   */
  async getUserStats(): Promise<UserStats> {
    const response = await apiClient.get('/users/stats');
    return response.data.data;
  }

  /**
   * 获取学生统计信息
   */
  async getStudentStats(): Promise<{
    total: number;
    byCollege: Record<string, number>;
    byMajor: Record<string, number>;
    byStatus: Record<string, number>;
  }> {
    const response = await apiClient.get('/users/stats/students');
    return response.data.data;
  }

  /**
   * 获取教师统计信息
   */
  async getTeacherStats(): Promise<{
    total: number;
    byDepartment: Record<string, number>;
    byTitle: Record<string, number>;
    byStatus: Record<string, number>;
  }> {
    const response = await apiClient.get('/users/stats/teachers');
    return response.data.data;
  }

  /**
   * 格式化用户名
   */
  formatUserName(user: UserInfo): string {
    return user.real_name || user.username;
  }

  /**
   * 获取用户类型显示名称
   */
  getUserTypeDisplayName(userType: string): string {
    const typeMap: Record<string, string> = {
      student: '学生',
      teacher: '教师',
      admin: '管理员',
    };
    return typeMap[userType] || userType;
  }

  /**
   * 获取状态显示名称
   */
  getStatusDisplayName(status: string): string {
    const statusMap: Record<string, string> = {
      active: '活跃',
      inactive: '停用',
      suspended: '暂停',
    };
    return statusMap[status] || status;
  }
}

export default new UserService(); 