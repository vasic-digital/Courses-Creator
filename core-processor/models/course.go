package models

import (
	"time"
)

// CourseMetadata represents metadata for a course
type CourseMetadata struct {
	Author        string   `json:"author"`
	Language      string   `json:"language"`
	Tags          []string `json:"tags"`
	ThumbnailURL  *string  `json:"thumbnail_url,omitempty"`
	TotalDuration int      `json:"total_duration"`
}

// Subtitle represents subtitle information
type Subtitle struct {
	Language   string                   `json:"language"`
	Content    string                   `json:"content"`
	Timestamps []map[string]interface{} `json:"timestamps"`
}

// Timestamp represents a timestamp entry
type Timestamp struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// InteractiveElement represents an interactive element in a lesson
type InteractiveElement struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // 'code', 'quiz', 'exercise'
	Content  string `json:"content"`
	Position int    `json:"position"` // seconds into video
}

// Lesson represents a lesson within a course
type Lesson struct {
	ID                  string               `json:"id"`
	Title               string               `json:"title"`
	Content             string               `json:"content"`
	VideoURL            *string              `json:"video_url,omitempty"`
	AudioURL            *string              `json:"audio_url,omitempty"`
	Subtitles           []Subtitle           `json:"subtitles"`
	InteractiveElements []InteractiveElement `json:"interactive_elements"`
	Duration            int                  `json:"duration"`
	Order               int                  `json:"order"`
}

// Course represents a complete video course
type Course struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Lessons     []Lesson       `json:"lessons"`
	Metadata    CourseMetadata `json:"metadata"`
	CreatedAt   *time.Time     `json:"created_at,omitempty"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty"`
}

// ProcessingOptions represents options for course processing
type ProcessingOptions struct {
	Voice           *string  `json:"voice,omitempty"`
	BackgroundMusic bool     `json:"background_music"`
	Languages       []string `json:"languages"`
	Quality         string   `json:"quality"` // 'standard' or 'high'
}

// ParsedSection represents a parsed section from markdown
type ParsedSection struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
	Order   int      `json:"order"`
}

// ParsedCourse represents parsed course data from markdown
type ParsedCourse struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Sections    []ParsedSection        `json:"sections"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ProcessingResult represents the result of course processing
type ProcessingResult struct {
	Course     Course   `json:"course"`
	OutputPath string   `json:"output_path"`
	Duration   int      `json:"duration"`
	Status     string   `json:"status"` // 'success' or 'failed'
	Errors     []string `json:"errors,omitempty"`
}
