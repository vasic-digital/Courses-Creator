import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { ApiResponse, ErrorResponse, PaginatedResponse } from '@/types';
import { Course, Job } from '@/types';

// API configuration
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const API_VERSION = 'v1';

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: `${API_BASE_URL}/api/${API_VERSION}`,
  timeout: 30000, // 30 seconds
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor to handle errors
apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    // Handle 401 Unauthorized
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Generic API wrapper functions
export const api = {
  get: async <T>(url: string): Promise<T> => {
    const response = await apiClient.get<T>(url);
    return response.data;
  },

  post: async <T>(url: string, data?: any): Promise<T> => {
    const response = await apiClient.post<T>(url, data);
    return response.data;
  },

  put: async <T>(url: string, data?: any): Promise<T> => {
    const response = await apiClient.put<T>(url, data);
    return response.data;
  },

  delete: async <T>(url: string): Promise<T> => {
    const response = await apiClient.delete<T>(url);
    return response.data;
  },

  // For file uploads
  upload: async <T>(url: string, file: File, onProgress?: (progress: number) => void): Promise<T> => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await apiClient.post<T>(url, formData, {
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

    return response.data;
  },
};

// Error handling helper
export const handleApiError = (error: any): ErrorResponse => {
  if (error.response) {
    // The request was made and the server responded with a status code
    // that falls out of the range of 2xx
    return {
      error: error.response.data.error || 'API Error',
      message: error.response.data.message,
      details: error.response.data.details,
    };
  } else if (error.request) {
    // The request was made but no response was received
    return {
      error: 'Network Error',
      message: 'Could not connect to the server. Please check your internet connection.',
    };
  } else {
    // Something happened in setting up the request that triggered an Error
    return {
      error: 'Request Error',
      message: error.message,
    };
  }
};

// Auth service
export const authService = {
  login: async (email: string, password: string) => {
    return api.post<{ token: string; refreshToken: string }>('/auth/login', {
      email,
      password,
    });
  },

  register: async (email: string, password: string, name: string) => {
    return api.post<{ token: string; refreshToken: string }>('/auth/register', {
      email,
      password,
      name,
    });
  },

  refreshToken: async (refreshToken: string) => {
    return api.post<{ token: string; refreshToken: string }>('/auth/refresh', {
      refreshToken,
    });
  },

  logout: async () => {
    return api.post('/auth/logout');
  },

  getProfile: async () => {
    return api.get('/auth/profile');
  },

  updateProfile: async (data: any) => {
    return api.put('/auth/profile', data);
  },
};

// Course service
export const courseService = {
  getCourses: async (page: number = 1, pageSize: number = 20, search?: string) => {
    const params = new URLSearchParams({
      page: page.toString(),
      pageSize: pageSize.toString(),
    });
    
    if (search) {
      params.append('search', search);
    }
    
    const response = await api.get<Course[]>(`/courses?${params.toString()}`);
    // Backend returns array directly, frontend expects PaginatedResponse
    return {
      items: response,
      total: response.length,
      page,
      pageSize,
      hasNext: false,
      hasPrev: page > 1,
    };
  },

  getCourse: async (id: string) => {
    return api.get<Course>(`/courses/${id}`);
  },

  generateCourse: async (markdown: string, options?: any) => {
    return api.post<{ jobId: string }>('/courses/generate', {
      markdown,
      options,
    });
  },
};

// Job service
export const jobService = {
  getJobs: async (status?: string) => {
    const params = status ? `?status=${status}` : '';
    return api.get<Job[]>(`/jobs${params}`);
  },

  getJob: async (id: string) => {
    return api.get<Job>(`/jobs/${id}`);
  },

  cancelJob: async (id: string) => {
    return api.post(`/jobs/${id}/cancel`);
  },
};

export default api;