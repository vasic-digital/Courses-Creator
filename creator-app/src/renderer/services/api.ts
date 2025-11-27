import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';

// API base configuration
const API_BASE_URL = 'http://localhost:8080/api/v1';

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
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
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle auth errors
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  login: async (email: string, password: string) => {
    const response = await apiClient.post('/auth/login', { email, password });
    const { token } = response.data;
    localStorage.setItem('auth_token', token);
    return response.data;
  },

  register: async (name: string, email: string, password: string) => {
    const response = await apiClient.post('/auth/register', { name, email, password });
    const { token } = response.data;
    localStorage.setItem('auth_token', token);
    return response.data;
  },

  logout: () => {
    localStorage.removeItem('auth_token');
  },

  getCurrentUser: async () => {
    const response = await apiClient.get('/auth/me');
    return response.data;
  },
};

// Public API (no auth required)
export const publicAPI = {
  getCourses: async () => {
    const response = await apiClient.get('/public/courses');
    return response.data;
  },

  getCourse: async (id: string) => {
    const response = await apiClient.get(`/public/courses/${id}`);
    return response.data;
  },

  healthCheck: async () => {
    const response = await apiClient.get('/health');
    return response.data;
  },
};

// Protected Course API (auth required)
export const courseAPI = {
  getCourses: async () => {
    const response = await apiClient.get('/courses');
    return response.data;
  },

  getCourse: async (id: string) => {
    const response = await apiClient.get(`/courses/${id}`);
    return response.data;
  },

  generateCourse: async (markdownPath: string, outputDir: string, options: any) => {
    const response = await apiClient.post('/courses/generate', {
      markdown_path: markdownPath,
      output_dir: outputDir,
      options,
    });
    return response.data;
  },

  deleteCourse: async (id: string) => {
    const response = await apiClient.delete(`/courses/${id}`);
    return response.data;
  },

  updateCourse: async (id: string, data: any) => {
    const response = await apiClient.put(`/courses/${id}`, data);
    return response.data;
  },
};

// Job API
export const jobAPI = {
  getJobs: async () => {
    const response = await apiClient.get('/jobs');
    return response.data;
  },

  getJob: async (id: string) => {
    const response = await apiClient.get(`/jobs/${id}`);
    return response.data;
  },
};

export default apiClient;