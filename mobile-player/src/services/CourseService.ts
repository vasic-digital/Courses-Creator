import axios from 'axios';
import { Course } from '../types';

const API_BASE_URL = 'http://localhost:8080/api/v1';

export class CourseService {
  static async getCourses(): Promise<Course[]> {
    try {
      const response = await axios.get(`${API_BASE_URL}/courses`);
      return response.data.courses || [];
    } catch (error) {
      console.error('Error fetching courses:', error);
      return [];
    }
  }

  static async getCourse(courseId: string): Promise<Course | null> {
    try {
      const response = await axios.get(`${API_BASE_URL}/courses/${courseId}`);
      return response.data.course;
    } catch (error) {
      console.error('Error fetching course:', error);
      return null;
    }
  }

  static async generateCourse(markdownPath: string, outputDir: string, options: any): Promise<any> {
    try {
      const response = await axios.post(`${API_BASE_URL}/courses/generate`, {
        markdown_path: markdownPath,
        output_dir: outputDir,
        options
      });
      return response.data;
    } catch (error) {
      console.error('Error generating course:', error);
      throw error;
    }
  }
}