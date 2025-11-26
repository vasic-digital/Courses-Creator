// Shared types for the creator app
export interface ProcessingOptions {
  voice?: string;
  backgroundMusic: boolean;
  languages: string[];
  quality: 'standard' | 'high';
}

export interface CourseGenerationRequest {
  markdown_path: string;
  output_dir: string;
  options: ProcessingOptions;
}

export interface CourseGenerationResponse {
  course: any; // Using shared types
  status: string;
  output_path: string;
}