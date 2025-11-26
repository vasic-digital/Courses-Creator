// Shared TypeScript types for Course Creator

export interface Course {
  id: string;
  title: string;
  description: string;
  lessons: Lesson[];
  metadata: CourseMetadata;
  createdAt: Date;
  updatedAt: Date;
}

export interface Lesson {
  id: string;
  title: string;
  content: string;
  videoUrl?: string;
  audioUrl?: string;
  subtitles: Subtitle[];
  interactiveElements: InteractiveElement[];
  duration: number;
  order: number;
}

export interface Subtitle {
  language: string;
  content: string;
  timestamps: Timestamp[];
}

export interface Timestamp {
  start: number;
  end: number;
  text: string;
}

export interface InteractiveElement {
  id: string;
  type: 'code' | 'quiz' | 'exercise';
  content: string;
  position: number; // seconds into video
}

export interface CourseMetadata {
  author: string;
  language: string;
  tags: string[];
  thumbnailUrl?: string;
  totalDuration: number;
}

export interface ProcessingOptions {
  voice?: string;
  backgroundMusic?: boolean;
  languages: string[];
  quality: 'standard' | 'high';
}

export interface ProcessingResult {
  course: Course;
  outputPath: string;
  duration: number;
  status: 'success' | 'failed';
  errors?: string[];
}