package models

import (
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
)

// CourseDB represents the database model for a Course
type CourseDB struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string         `gorm:"not null;type:varchar(36);index" json:"user_id"`
	Title       string         `gorm:"not null;type:varchar(255)" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relations
	User        *UserDB          `gorm:"constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Lessons     []LessonDB       `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"lessons,omitempty"`
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

// UserDB represents the database model for a User
type UserDB struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Email     string    `gorm:"not null;uniqueIndex;type:varchar(255)" json:"email"`
	Password  string    `gorm:"not null;type:varchar(255)" json:"-"` // Never expose password
	FirstName string    `gorm:"type:varchar(100)" json:"first_name"`
	LastName  string    `gorm:"type:varchar(100)" json:"last_name"`
	Role      string    `gorm:"not null;type:varchar(20);default:'viewer'" json:"role"`
	Active    bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relations
	Preferences   []UserPreferencesDB `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"preferences,omitempty"`
	Sessions      []UserSessionDB      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"sessions,omitempty"`
	Jobs          []JobDB              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"jobs,omitempty"`
	Courses       []CourseDB           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"courses,omitempty"`
}

// UserPreferencesDB represents the database model for UserPreferences
type UserPreferencesDB struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID           string    `gorm:"not null;type:varchar(36);index" json:"user_id"`
	Voice            string    `gorm:"type:varchar(50)" json:"voice"`
	BackgroundStyle  string    `gorm:"type:varchar(50)" json:"background_style"`
	Quality          string    `gorm:"type:varchar(20);default:'standard'" json:"quality"`
	Language         string    `gorm:"type:varchar(10);default:'en'" json:"language"`
	Preferences      string    `gorm:"type:text" json:"preferences"` // JSON string
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relations
	User *UserDB `gorm:"constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// UserSessionDB represents the database model for UserSession
type UserSessionDB struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID       string    `gorm:"not null;type:varchar(36);index" json:"user_id"`
	TokenHash    string    `gorm:"not null;type:varchar(255);index" json:"token_hash"`
	RefreshToken *string   `gorm:"type:varchar(500)" json:"refresh_token,omitempty"`
	IPAddress    string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	ExpiresAt    time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	LastActivity time.Time `gorm:"autoUpdateTime" json:"last_activity"`
	
	// Relations
	User *UserDB `gorm:"constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// JobDB represents the database model for a Job
type JobDB struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string    `gorm:"not null;type:varchar(36);index" json:"user_id"`
	Type        string    `gorm:"not null;type:varchar(50)" json:"type"`
	Status      string    `gorm:"not null;type:varchar(20);default:'pending'" json:"status"`
	Progress    int       `gorm:"default:0" json:"progress"`
	Payload     string    `gorm:"type:text" json:"payload"` // JSON string
	Result      string    `gorm:"type:text" json:"result,omitempty"` // JSON string
	Error       *string   `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relations
	User *UserDB `gorm:"constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// BeforeCreate hooks for user models
func (u *UserDB) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

func (up *UserPreferencesDB) BeforeCreate(tx *gorm.DB) error {
	if up.ID == "" {
		up.ID = uuid.New().String()
	}
	return nil
}

func (us *UserSessionDB) BeforeCreate(tx *gorm.DB) error {
	if us.ID == "" {
		us.ID = uuid.New().String()
	}
	return nil
}

func (j *JobDB) BeforeCreate(tx *gorm.DB) error {
	if j.ID == "" {
		j.ID = uuid.New().String()
	}
	return nil
}

// TableName method overrides for user models
func (UserDB) TableName() string {
	return "users"
}

func (UserPreferencesDB) TableName() string {
	return "user_preferences"
}

func (UserSessionDB) TableName() string {
	return "user_sessions"
}

func (JobDB) TableName() string {
	return "jobs"
}