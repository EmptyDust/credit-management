import axios from 'axios';
import toast from 'react-hot-toast';

const apiClient = axios.create({
  baseURL: '/api', // All requests will be sent to the api-gateway
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 15000, // 15 seconds timeout
});

// Request interceptor for adding auth token
apiClient.interceptors.request.use(
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

// Response interceptor for handling errors and token refresh
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const { response, config } = error;
    
    // 如果是登录请求，不要触发token刷新逻辑
    if (config.url === '/auth/login') {
      return Promise.reject(error);
    }
    
    if (response) {
      const { status, data } = response;
      
      switch (status) {
        case 401:
          // Try to refresh token first
          const refreshToken = localStorage.getItem('refreshToken');
          if (refreshToken && !config._retry) {
            config._retry = true;
            try {
              const refreshResponse = await axios.post('/api/auth/refresh-token', {
                refresh_token: refreshToken
              });
              
              if (refreshResponse.data.token) {
                localStorage.setItem('token', refreshResponse.data.token);
                config.headers.Authorization = `Bearer ${refreshResponse.data.token}`;
                return apiClient(config);
              }
            } catch (refreshError) {
              // Refresh failed, redirect to login
              localStorage.removeItem('token');
              localStorage.removeItem('refreshToken');
              localStorage.removeItem('user');
              window.location.href = '/login';
              toast.error('登录已过期，请重新登录');
              return Promise.reject(refreshError);
            }
          } else {
            // No refresh token or refresh failed
            localStorage.removeItem('token');
            localStorage.removeItem('refreshToken');
            localStorage.removeItem('user');
            window.location.href = '/login';
            toast.error('登录已过期，请重新登录');
          }
          break;
        case 403:
          toast.error('权限不足，无法执行此操作');
          break;
        case 404:
          toast.error('请求的资源不存在');
          break;
        case 409:
          // Conflict error - usually username or email already exists
          const conflictMessage = data?.error || data?.message || '用户名或邮箱已存在';
          toast.error(conflictMessage);
          break;
        case 422:
          // Validation error
          if (data.errors && Array.isArray(data.errors)) {
            const errorMessages = data.errors.map((err: any) => err.message || err.field).join(', ');
            toast.error(`数据验证失败: ${errorMessages}`);
          } else {
            const message = data.message || data.error || '数据验证失败';
            toast.error(message);
          }
          break;
        case 429:
          toast.error('请求过于频繁，请稍后再试');
          break;
        case 500:
          toast.error('服务器内部错误，请稍后再试');
          break;
        default:
          const message = data?.message || data?.error || '请求失败';
          toast.error(message);
      }
    } else if (error.request) {
      // Network error
      toast.error('网络连接失败，请检查网络设置');
    } else {
      // Other error
      toast.error('请求配置错误');
    }
    
    return Promise.reject(error);
  }
);

// Helper functions for common API operations
export const apiHelpers = {
  // Upload file with progress
  uploadFile: async (url: string, file: File, onProgress?: (progress: number) => void) => {
    const formData = new FormData();
    formData.append('file', file);
    
    return apiClient.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  },

  // Download file
  downloadFile: async (url: string, filename?: string) => {
    const response = await apiClient.get(url, {
      responseType: 'blob',
    });
    
    const blob = new Blob([response.data]);
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename || 'download';
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(downloadUrl);
  },

  // Paginated request helper
  getPaginated: async (url: string, page: number = 1, limit: number = 10, filters?: Record<string, any>) => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      ...filters,
    });
    
    return apiClient.get(`${url}?${params.toString()}`);
  },
};

export default apiClient; 