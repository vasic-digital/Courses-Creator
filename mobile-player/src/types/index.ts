// Mobile player types
export interface Course {
  id: string;
  title: string;
  description: string;
  lessons: Lesson[];
  metadata: CourseMetadata;
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
  position: number;
}

export interface CourseMetadata {
  author: string;
  language: string;
  tags: string[];
  thumbnailUrl?: string;
  totalDuration: number;
}

export interface PlayerState {
  currentLesson: number;
  isPlaying: boolean;
  currentTime: number;
  volume: number;
  subtitlesEnabled: boolean;
}