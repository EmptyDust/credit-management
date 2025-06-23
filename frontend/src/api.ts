import axios from 'axios';

// 使用相对路径，让nginx代理处理API请求
const API_BASE_URL = '/api';

const api = axios.create({
    baseURL: API_BASE_URL,
});

// Add a request interceptor to include the token in headers
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

// 通用请求函数
async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const response = await fetch(url, {
        headers: {
            'Content-Type': 'application/json',
            ...options.headers,
        },
        ...options,
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
}

// 用户管理API
export const userAPI = {
    register: (data: { username: string; password: string; email: string; role: string }) =>
        request('/api/users/register', {
            method: 'POST',
            body: JSON.stringify(data),
        }),

    login: (data: { username: string; password: string }) =>
        request('/api/users/login', {
            method: 'POST',
            body: JSON.stringify(data),
        }),

    getUser: (id: string) => request(`/api/users/${id}`),

    updateUser: (id: string, data: any) =>
        request(`/api/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        }),

    deleteUser: (id: string) =>
        request(`/api/users/${id}`, {
            method: 'DELETE',
        }),
};

// 学生信息API
export const studentAPI = {
    createStudent: (data: any) =>
        request('/api/students', {
            method: 'POST',
            body: JSON.stringify(data),
        }),

    getStudents: () => request('/api/students'),

    getStudent: (id: string) => request(`/api/students/${id}`),

    getStudentByUser: (userID: string) => request(`/api/students/user/${userID}`),

    updateStudent: (id: string, data: any) =>
        request(`/api/students/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        }),

    deleteStudent: (id: string) =>
        request(`/api/students/${id}`, {
            method: 'DELETE',
        }),
};

// 教师信息API
export const teacherAPI = {
    createTeacher: (data: any) =>
        request('/api/teachers', {
            method: 'POST',
            body: JSON.stringify(data),
        }),

    getTeachers: () => request('/api/teachers'),

    getTeacher: (id: string) => request(`/api/teachers/${id}`),

    getTeacherByUser: (userID: string) => request(`/api/teachers/user/${userID}`),

    updateTeacher: (id: string, data: any) =>
        request(`/api/teachers/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        }),

    deleteTeacher: (id: string) =>
        request(`/api/teachers/${id}`, {
            method: 'DELETE',
        }),
};

// 事项管理API
export const affairAPI = {
    createAffair: (data: any) =>
        request('/api/affairs', {
            method: 'POST',
            body: JSON.stringify(data),
        }),

    getAffairs: () => request('/api/affairs'),

    getAffair: (id: string) => request(`/api/affairs/${id}`),

    updateAffair: (id: string, data: any) =>
        request(`/api/affairs/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        }),

    deleteAffair: (id: string) =>
        request(`/api/affairs/${id}`, {
            method: 'DELETE',
        }),
};

// 申请管理API
export const applicationAPI = {
    createApplication: (data: any) =>
        request('/api/applications', {
            method: 'POST',
            body: JSON.stringify(data),
        }),

    getApplications: () => request('/api/applications'),

    getApplication: (id: string) => request(`/api/applications/${id}`),

    getApplicationsByUser: (userID: string) => request(`/api/applications/user/${userID}`),

    getApplicationsByStudent: (studentID: string) => request(`/api/applications/student/${studentID}`),

    updateApplication: (id: string, data: any) =>
        request(`/api/applications/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        }),

    reviewApplication: (id: string, data: any) =>
        request(`/api/applications/${id}/review`, {
            method: 'POST',
            body: JSON.stringify(data),
        }),

    deleteApplication: (id: string) =>
        request(`/api/applications/${id}`, {
            method: 'DELETE',
        }),
};

export default api; 