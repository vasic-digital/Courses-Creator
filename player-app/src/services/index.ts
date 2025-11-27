// Re-export all API services
export { api, handleApiError, authService, courseService, jobService } from './api';

// Import Course and Job types for service
import { Course, Job } from '@/types';

// Create a service object that matches what the components expect
export const apiService = {
  getCourses: async (params?: any) => {
    const { courseService } = await import('./api');
    return courseService.getCourses(params?.page, params?.pageSize, params?.search);
  },
  getCourse: async (id: string) => {
    const { courseService } = await import('./api');
    return courseService.getCourse(id);
  }
};