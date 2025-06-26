import apiClient from './api';

export interface UserInfo {
  id: string;
  username: string;
  real_name: string;
  user_type: 'student' | 'teacher' | 'admin';
  email: string;
  phone?: string;
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
  department?: string;
  title?: string;
  status: 'active' | 'inactive' | 'suspended';
  avatar?: string;
  created_at: string;
  updated_at: string;
}

export interface StudentInfo {
  id: string;
  username: string;
  real_name: string;
  student_id: string;
  college: string;
  major: string;
  class: string;
  grade?: string;
  email: string;
  phone?: string;
  status: string;
}

export interface TeacherInfo {
  id: string;
  username: string;
  real_name: string;
  department: string;
  title: string;
  specialty?: string;
  email: string;
  phone?: string;
  status: string;
}

export interface UserSearchParams {
  page?: number;
  limit?: number;
  userType?: 'student' | 'teacher' | 'admin';
  status?: 'active' | 'inactive' | 'suspended';
  search?: string;
  college?: string;
  major?: string;
  department?: string;
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
   * 获取用户列表
   */
  async getUsers(params?: UserSearchParams): Promise<{
    data: UserInfo[];
    total: number;
    page: number;
    limit: number;
  }> {
    const response = await apiClient.get('/users', { params });
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
   * 根据用户名搜索用户
   */
  async searchUsersByUsername(username: string): Promise<UserInfo[]> {
    const response = await apiClient.get('/users/search/username', {
      params: { username }
    });
    return response.data.data;
  }

  /**
   * 获取学生列表
   */
  async getStudents(params?: {
    page?: number;
    limit?: number;
    college?: string;
    major?: string;
    class?: string;
    status?: string;
    search?: string;
  }): Promise<{
    data: StudentInfo[];
    total: number;
    page: number;
    limit: number;
  }> {
    const response = await apiClient.get('/students', { params });
    return response.data.data;
  }

  /**
   * 根据学号搜索学生
   */
  async searchStudentsByStudentId(studentId: string): Promise<StudentInfo[]> {
    const response = await apiClient.get('/students/search', {
      params: { student_id: studentId }
    });
    return response.data.data;
  }

  /**
   * 获取教师列表
   */
  async getTeachers(params?: {
    page?: number;
    limit?: number;
    department?: string;
    title?: string;
    status?: string;
    search?: string;
  }): Promise<{
    data: TeacherInfo[];
    total: number;
    page: number;
    limit: number;
  }> {
    const response = await apiClient.get('/teachers', { params });
    return response.data.data;
  }

  /**
   * 根据用户名搜索教师
   */
  async searchTeachersByUsername(username: string): Promise<TeacherInfo[]> {
    const response = await apiClient.get('/teachers/search', {
      params: { username }
    });
    return response.data.data;
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
    const response = await apiClient.get('/students/stats');
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
    const response = await apiClient.get('/teachers/stats');
    return response.data.data;
  }

  /**
   * 统一搜索用户（支持学生、教师、管理员）
   */
  async searchUsers(searchTerm: string, userType?: 'student' | 'teacher' | 'admin'): Promise<UserInfo[]> {
    const response = await apiClient.get('/search/users', {
      params: { 
        q: searchTerm,
        user_type: userType 
      }
    });
    return response.data.data;
  }

  /**
   * 根据学院获取学生
   */
  async getStudentsByCollege(college: string): Promise<StudentInfo[]> {
    const response = await apiClient.get(`/students/college/${encodeURIComponent(college)}`);
    return response.data.data;
  }

  /**
   * 根据专业获取学生
   */
  async getStudentsByMajor(major: string): Promise<StudentInfo[]> {
    const response = await apiClient.get(`/students/major/${encodeURIComponent(major)}`);
    return response.data.data;
  }

  /**
   * 根据班级获取学生
   */
  async getStudentsByClass(className: string): Promise<StudentInfo[]> {
    const response = await apiClient.get(`/students/class/${encodeURIComponent(className)}`);
    return response.data.data;
  }

  /**
   * 根据部门获取教师
   */
  async getTeachersByDepartment(department: string): Promise<TeacherInfo[]> {
    const response = await apiClient.get(`/teachers/department/${encodeURIComponent(department)}`);
    return response.data.data;
  }

  /**
   * 根据职称获取教师
   */
  async getTeachersByTitle(title: string): Promise<TeacherInfo[]> {
    const response = await apiClient.get(`/teachers/title/${encodeURIComponent(title)}`);
    return response.data.data;
  }

  /**
   * 获取活跃教师列表
   */
  async getActiveTeachers(): Promise<TeacherInfo[]> {
    const response = await apiClient.get('/teachers/active');
    return response.data.data;
  }

  /**
   * 批量获取用户信息（用于补充申请列表中的用户信息）
   */
  async getUsersByIds(userIds: string[]): Promise<UserInfo[]> {
    if (userIds.length === 0) return [];
    
    // 由于API可能不支持批量查询，这里使用Promise.all并发请求
    const userPromises = userIds.map(id => this.getUserById(id).catch(() => null));
    const users = await Promise.all(userPromises);
    return users.filter(user => user !== null) as UserInfo[];
  }

  /**
   * 格式化用户显示名称
   */
  formatUserName(user: UserInfo): string {
    if (user.user_type === 'student' && user.student_id) {
      return `${user.real_name} (${user.student_id})`;
    }
    return user.real_name || user.username;
  }

  /**
   * 获取用户类型显示名称
   */
  getUserTypeDisplayName(userType: string): string {
    const typeMap = {
      student: '学生',
      teacher: '教师',
      admin: '管理员'
    };
    return typeMap[userType as keyof typeof typeMap] || userType;
  }

  /**
   * 获取状态显示名称
   */
  getStatusDisplayName(status: string): string {
    const statusMap = {
      active: '活跃',
      inactive: '非活跃',
      suspended: '已暂停'
    };
    return statusMap[status as keyof typeof statusMap] || status;
  }
}

export const userService = new UserService();
export default userService; 