package models

import (
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
)

// CourseDB represents the database model for a Course
type CourseDB struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Title       string         `gorm:"not null;type:varchar(255)" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relations
	Lessons     []LessonDB     `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"lessons,omitempty"`
	Metadata    CourseMetadataDB `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"metadata,omitempty"`
}

// CourseMetadataDB represents the database model for CourseMetadata
type CourseMetadataDB struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CourseID      string    `gorm:"not null;type:varchar(36);index" json:"course_id"`
	Author        string    `gorm:"type:varchar(255)" json:"author"`
	Language      string    `gorm:"type:varchar(10)" json:"language"`
	Tags          string    `gorm:"type:text" json:"tags"` // JSON string
	ThumbnailURL  *string   `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	TotalDuration int       `gorm:"default:0" json:"total_duration"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relations
	Course *CourseDB `gorm:"constraint:OnDelete:CASCADE" json:"course,omitempty"`
}

// LessonDB represents the database model for a Lesson
type LessonDB struct {
	ID       string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CourseID string    `gorm:"not null;type:varchar(36);index" json:"course_id"`
	Title    string    `gorm:"not null;type:varchar(255)" json:"title"`
	Content  string    `gorm:"type:longtext" json:"content"`
	VideoURL *string   `gorm:"type:varchar(500)" json:"video_url,omitempty"`
	AudioURL *string   `gorm:"type:varchar(500)" json:"audio_url,omitempty"`
	Duration int       `gorm:"default:0" json:"duration"`
	Order    int       `gorm:"not null" json:"order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relations
	Course                *CourseDB          `gorm:"constraint:OnDelete:CASCADE" json:"course,omitempty"`
	Subtitles            []SubtitleDB       `gorm:"foreignKey:LessonID;constraint:OnDelete:CASCADE" json:"subtitles,omitempty"`
	InteractiveElements  []InteractiveElementDB `gorm:"foreignKey:LessonID;constraint:OnDelete:CASCADE" json:"interactive_elements,omitempty"`
}

// SubtitleDB represents the database model for a Subtitle
type SubtitleDB struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	LessonID  string    `gorm:"not null;type:varchar(36);index" json:"lesson_id"`
	Language  string    `gorm:"not null;type:varchar(10)" json:"language"`
	Content   string    `gorm:"type:longtext" json:"content"`
	Timestamps string  `gorm:"type:longtext" json:"timestamps"` // JSON string
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relations
	Lesson *LessonDB `gorm:"constraint:OnDelete:CASCADE" json:"lesson,omitempty"`
}

// InteractiveElementDB represents the database model for an InteractiveElement
type InteractiveElementDB struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	LessonID  string    `gorm:"not null;type:varchar(36);index" json:"lesson_id"`
	Type      string    `gorm:"not null;type:varchar(50)" json:"type"`
	Content   string    `gorm:"type:text" json:"content"`
	Position  int       `gorm:"not null" json:"position"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relations
	Lesson *LessonDB `gorm:"constraint:OnDelete:CASCADE" json:"lesson,omitempty"`
}

// ProcessingJobDB represents a processing job in the database
type ProcessingJobDB struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CourseID    *string   `gorm:"type:varchar(36);index" json:"course_id,omitempty"`
	Status      string    `gorm:"not null;type:varchar(20);default:'pending'" json:"status"`
	InputPath   string    `gorm:"type:varchar(500)" json:"input_path"`
	OutputPath  *string   `gorm:"type:varchar(500)" json:"output_path,omitempty"`
	Options     string    `gorm:"type:text" json:"options"` // JSON string
	Error       *string   `gorm:"type:text" json:"error,omitempty"`
	Progress    int       `gorm:"default:0" json:"progress"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUIDs
func (c *CourseDB) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

func (cm *CourseMetadataDB) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		cm.ID = uuid.New().String()
	}
	return nil
}

func (l *LessonDB) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

func (s *SubtitleDB) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

func (ie *InteractiveElementDB) BeforeCreate(tx *gorm.DB) error {
	if ie.ID == "" {
		ie.ID = uuid.New().String()
	}
	return nil
}

func (pj *ProcessingJobDB) BeforeCreate(tx *gorm.DB) error {
	if pj.ID == "" {
		pj.ID = uuid.New().String()
	}
	return nil
}

// TableName method overrides
func (CourseDB) TableName() string {
	return "courses"
}

func (CourseMetadataDB) TableName() string {
	return "course_metadata"
}

func (LessonDB) TableName() string {
	return "lessons"
}

func (SubtitleDB) TableName() string {
	return "subtitles"
}

func (InteractiveElementDB) TableName() string {
	return "interactive_elements"
}

func (ProcessingJobDB) TableName() string {
	return "processing_jobs"
}