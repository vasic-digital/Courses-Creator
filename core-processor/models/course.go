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
	Diagrams            []Diagram            `json:"diagrams,omitempty"`
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
	BackgroundStyle string   `json:"background_style,omitempty"`
	Languages       []string `json:"languages"`
	Quality         string   `json:"quality"` // 'standard' or 'high'
}

// Diagram represents a diagram in a course
type Diagram struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	ImageURL    *string                `json:"image_url,omitempty"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
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

// User represents a user in the system
type User struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Password  string                 `json:"-"` // Never expose password in JSON
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Role      string                 `json:"role"` // 'admin', 'creator', 'viewer'
	Active    bool                   `json:"active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	CustomData map[string]interface{} `json:"custom_data,omitempty"`
}

// UserPreferences represents user-specific preferences
type UserPreferences struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Voice        string            `json:"voice"`
	BackgroundStyle string         `json:"background_style"`
	Quality      string            `json:"quality"` // 'standard' or 'high'
	Language     string            `json:"language"`
	Preferences  map[string]string `json:"preferences"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// UserSession represents a user session
type UserSession struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	TokenHash    string    `json:"token_hash"`
	RefreshToken *string   `json:"refresh_token,omitempty"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Type        string                 `json:"type"` // 'course_generation', 'video_processing', etc.
	Status      string                 `json:"status"` // 'pending', 'running', 'completed', 'failed'
	Progress    int                    `json:"progress"` // 0-100
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       *string                `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}
