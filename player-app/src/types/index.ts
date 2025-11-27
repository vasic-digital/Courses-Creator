// Core types for the web player

export interface Course {
  id: string;
  title: string;
  description: string;
  author: string;
  thumbnailUrl?: string;
  duration: number; // in seconds
  lessons: Lesson[];
  metadata: CourseMetadata;
  createdAt: string;
  updatedAt: string;
}

export interface Lesson {
  id: string;
  courseId: string;
  title: string;
  description?: string;
  videoUrl: string;
  duration: number; // in seconds
  transcript?: string;
  subtitles?: Subtitle[];
  order: number;
  interactiveElements?: InteractiveElement[];
  createdAt: string;
  updatedAt: string;
}

export interface Subtitle {
  id: string;
  language: string;
  format: 'vtt' | 'srt' | 'ass';
  url: string;
  label: string;
}

export interface InteractiveElement {
  id: string;
  type: 'quiz' | 'code' | 'link' | 'note';
  timestamp: number; // seconds from start
  data: any;
  title?: string;
}

export interface CourseMetadata {
  language: string;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  tags: string[];
  category?: string;
  estimatedHours: number;
}

// API types
export interface ApiResponse<T> {
  data: T;
  message?: string;
  status: 'success' | 'error';
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface ErrorResponse {
  error: string;
  message?: string;
  details?: any;
}

// Player state types
export interface PlaybackState {
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  volume: number;
  playbackRate: number;
  quality: string;
  isFullscreen: boolean;
  isMuted: boolean;
  buffered: number[];
}

export interface VideoQuality {
  label: string;
  value: string;
  width: number;
  height: number;
  bitrate?: number;
}

// UI state types
export interface LoadingState {
  isLoading: boolean;
  message?: string;
}

export interface ErrorState {
  hasError: boolean;
  error?: Error | ErrorResponse;
}

// Component props types
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

export interface ButtonProps extends BaseComponentProps {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  onClick?: (e: React.MouseEvent) => void;
  type?: 'button' | 'submit' | 'reset';
}

export interface InputProps extends BaseComponentProps {
  type?: string;
  value?: string;
  placeholder?: string;
  disabled?: boolean;
  error?: string;
  onChange?: (value: string) => void;
  onBlur?: (e: React.FocusEvent) => void;
  onFocus?: (e: React.FocusEvent) => void;
}

// Hook types
export interface UseApiResult<T> {
  data: T | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => void;
}

export interface UseLocalStorageResult<T> {
  value: T;
  setValue: (value: T) => void;
  removeValue: () => void;
}

// Navigation types
export interface NavigationItem {
  id: string;
  label: string;
  path: string;
  icon?: string;
  badge?: number;
  external?: boolean;
}

// Job types
export interface Job {
  id: string;
  userId: string;
  type: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number; // 0-100
  payload: any;
  result?: any;
  error?: string;
  createdAt: string;
  updatedAt: string;
  startedAt?: string;
  completedAt?: string;
}