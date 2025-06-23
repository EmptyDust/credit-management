import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const api = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// 请求拦截器 - 添加认证token
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// 响应拦截器 - 处理错误
api.interceptors.response.use(
    (response) => {
        return response;
    },
    (error) => {
        if (error.response?.status === 401) {
            // Token过期，清除本地存储并重定向到登录页
            localStorage.removeItem('token');
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

// 认证相关API
export const authAPI = {
    // 用户登录
    login: (data: { username: string; password: string }) =>
        api.post('/auth/login', data),
    
    // 用户注册
    register: (data: { username: string; password: string; userType: string }) =>
        api.post('/auth/register', data),
    
    // 刷新token
    refresh: (data: { refresh_token: string }) =>
        api.post('/auth/refresh', data),
    
    // 用户登出
    logout: () => api.post('/auth/logout'),
    
    // 验证token
    validate: () => api.get('/auth/validate'),
    
    // 获取用户权限
    getPermissions: () => api.get('/auth/permissions'),
    
    // 获取角色列表
    getRoles: () => api.get('/auth/roles'),
    
    // 创建角色
    createRole: (data: { name: string; description: string; permissions: string[] }) =>
        api.post('/auth/roles', data),
    
    // 更新角色
    updateRole: (id: number, data: { name: string; description: string; permissions: string[] }) =>
        api.put(`/auth/roles/${id}`, data),
    
    // 删除角色
    deleteRole: (id: number) => api.delete(`/auth/roles/${id}`),
    
    // 分配角色给用户
    assignRole: (data: { userID: number; roleID: number }) =>
        api.post('/auth/assign-role', data),
    
    // 移除用户角色
    removeRole: (data: { userID: number; roleID: number }) =>
        api.post('/auth/remove-role', data),
};

// 用户管理相关API
export const userAPI = {
    // 获取用户信息
    getProfile: () => api.get('/users/profile'),
    
    // 更新用户信息
    updateProfile: (data: { name: string; contact: string; email: string }) =>
        api.put('/users/profile', data),
    
    // 获取所有用户
    getUsers: () => api.get('/users'),
    
    // 获取指定用户
    getUser: (id: number) => api.get(`/users/${id}`),
    
    // 创建用户
    createUser: (data: { username: string; password: string; userType: string; name: string; contact: string; email: string }) =>
        api.post('/users', data),
    
    // 更新用户
    updateUser: (id: number, data: { name: string; contact: string; email: string; status: string }) =>
        api.put(`/users/${id}`, data),
    
    // 删除用户
    deleteUser: (id: number) => api.delete(`/users/${id}`),
    
    // 获取通知
    getNotifications: () => api.get('/users/notifications'),
    
    // 标记通知为已读
    markNotificationRead: (id: number) => api.put(`/users/notifications/${id}/read`),
    
    // 删除通知
    deleteNotification: (id: number) => api.delete(`/users/notifications/${id}`),
};

// 学生信息相关API
export const studentAPI = {
    // 获取所有学生
    getStudents: () => api.get('/students'),
    
    // 获取指定学生
    getStudent: (username: string) => api.get(`/students/${username}`),
    
    // 根据学号获取学生
    getStudentByID: (studentID: string) => api.get(`/students/id/${studentID}`),
    
    // 创建学生
    createStudent: (data: { username: string; studentID: string; name: string; college: string; major: string; class: string; contact: string; email: string; grade: string }) =>
        api.post('/students', data),
    
    // 更新学生
    updateStudent: (username: string, data: { name: string; college: string; major: string; class: string; contact: string; email: string; grade: string; status: string }) =>
        api.put(`/students/${username}`, data),
    
    // 删除学生
    deleteStudent: (username: string) => api.delete(`/students/${username}`),
    
    // 根据学院获取学生
    getStudentsByCollege: (college: string) => api.get(`/students/college/${college}`),
    
    // 根据专业获取学生
    getStudentsByMajor: (major: string) => api.get(`/students/major/${major}`),
    
    // 根据班级获取学生
    getStudentsByClass: (className: string) => api.get(`/students/class/${className}`),
    
    // 根据状态获取学生
    getStudentsByStatus: (status: string) => api.get(`/students/status/${status}`),
    
    // 搜索学生
    searchStudents: (query: string) => api.get(`/students/search?q=${query}`),
};

// 教师信息相关API
export const teacherAPI = {
    // 获取所有教师
    getTeachers: () => api.get('/teachers'),
    
    // 获取指定教师
    getTeacher: (username: string) => api.get(`/teachers/${username}`),
    
    // 创建教师
    createTeacher: (data: { username: string; name: string; contact: string; email: string; department: string; title: string; specialty: string }) =>
        api.post('/teachers', data),
    
    // 更新教师
    updateTeacher: (username: string, data: { name: string; contact: string; email: string; department: string; title: string; specialty: string; status: string }) =>
        api.put(`/teachers/${username}`, data),
    
    // 删除教师
    deleteTeacher: (username: string) => api.delete(`/teachers/${username}`),
    
    // 根据院系获取教师
    getTeachersByDepartment: (department: string) => api.get(`/teachers/department/${department}`),
    
    // 根据职称获取教师
    getTeachersByTitle: (title: string) => api.get(`/teachers/title/${title}`),
    
    // 根据状态获取教师
    getTeachersByStatus: (status: string) => api.get(`/teachers/status/${status}`),
    
    // 搜索教师
    searchTeachers: (query: string) => api.get(`/teachers/search?q=${query}`),
    
    // 获取活跃教师
    getActiveTeachers: () => api.get('/teachers/active'),
};

// 申请管理相关API
export const applicationAPI = {
    // 申请类型相关
    getApplicationTypes: () => api.get('/applications/types'),
    getApplicationType: (id: number) => api.get(`/applications/types/${id}`),
    createApplicationType: (data: { name: string; description: string; category: string; maxCredits: number; minCredits: number; isActive: boolean }) =>
        api.post('/applications/types', data),
    updateApplicationType: (id: number, data: { name: string; description: string; category: string; maxCredits: number; minCredits: number; isActive: boolean }) =>
        api.put(`/applications/types/${id}`, data),
    deleteApplicationType: (id: number) => api.delete(`/applications/types/${id}`),
    
    // 申请相关
    getApplications: () => api.get('/applications'),
    getApplication: (id: number) => api.get(`/applications/${id}`),
    createApplication: (data: { typeID: number; title: string; description: string; content: string; credits: number }) =>
        api.post('/applications', data),
    updateApplication: (id: number, data: { title: string; description: string; content: string; credits: number }) =>
        api.put(`/applications/${id}`, data),
    deleteApplication: (id: number) => api.delete(`/applications/${id}`),
    updateApplicationStatus: (id: number, data: { status: string; approvedCredits: number; reviewNote: string }) =>
        api.put(`/applications/${id}/status`, data),
    
    // 用户申请相关
    getUserApplications: (userID: number) => api.get(`/applications/user/${userID}`),
    getApplicationsByType: (typeID: number) => api.get(`/applications/type/${typeID}`),
    getPendingApplications: () => api.get('/applications/pending'),
    getApprovedApplications: () => api.get('/applications/approved'),
    getRejectedApplications: () => api.get('/applications/rejected'),
    getApplicationStats: () => api.get('/applications/stats'),
    
    // 文件相关
    uploadFile: (applicationID: number, file: File, category: string, description: string) => {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('category', category);
        formData.append('description', description);
        return api.post(`/applications/${applicationID}/files`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
    },
    downloadFile: (fileID: number) => api.get(`/applications/files/download/${fileID}`, { responseType: 'blob' }),
    deleteFile: (fileID: number) => api.delete(`/applications/files/${fileID}`),
    getApplicationFiles: (applicationID: number) => api.get(`/applications/${applicationID}/files`),
};

// 事项管理相关API
export const affairAPI = {
    // 事项相关
    getAffairs: () => api.get('/affairs'),
    getAffair: (id: number) => api.get(`/affairs/${id}`),
    createAffair: (data: { name: string; description: string; category: string; maxCredits: number }) =>
        api.post('/affairs', data),
    updateAffair: (id: number, data: { name: string; description: string; category: string; maxCredits: number; status: string }) =>
        api.put(`/affairs/${id}`, data),
    deleteAffair: (id: number) => api.delete(`/affairs/${id}`),
    getAffairsByCategory: (category: string) => api.get(`/affairs/category/${category}`),
    getActiveAffairs: () => api.get('/affairs/active'),
    
    // 事项-学生关系相关
    addStudentToAffair: (data: { affairID: number; studentID: string; isMainResponsible: boolean }) =>
        api.post('/affair-students', data),
    removeStudentFromAffair: (affairID: number, studentID: string) =>
        api.delete(`/affair-students/${affairID}/${studentID}`),
    getStudentsByAffair: (affairID: number) => api.get(`/affair-students/affair/${affairID}`),
    getAffairsByStudent: (studentID: string) => api.get(`/affair-students/student/${studentID}`),
};

export default api; 