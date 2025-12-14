package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCourse_JSONMarshaling(t *testing.T) {
	now := time.Now()
	course := Course{
		ID:          "test-course-id",
		Title:       "Test Course",
		Description: "A test course",
		Lessons: []Lesson{
			{
				ID:       "lesson-1",
				Title:    "Lesson 1",
				Content:  "Lesson content",
				Duration: 300,
				Order:    1,
			},
		},
		Metadata: CourseMetadata{
			Author:        "Test Author",
			Language:      "en",
			Tags:          []string{"test", "course"},
			TotalDuration: 300,
		},
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	data, err := json.Marshal(course)
	require.NoError(t, err)

	var unmarshaled Course
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, course.ID, unmarshaled.ID)
	assert.Equal(t, course.Title, unmarshaled.Title)
	assert.Equal(t, course.Description, unmarshaled.Description)
	assert.Len(t, unmarshaled.Lessons, 1)
	assert.Equal(t, course.Metadata.Author, unmarshaled.Metadata.Author)
}

func TestLesson_JSONMarshaling(t *testing.T) {
	lesson := Lesson{
		ID:       "lesson-id",
		Title:    "Test Lesson",
		Content:  "Lesson content",
		Duration: 300,
		Order:    1,
		Subtitles: []Subtitle{
			{
				Language: "en",
				Content:  "English subtitles",
			},
		},
		InteractiveElements: []InteractiveElement{
			{
				ID:       "element-1",
				Type:     "quiz",
				Content:  "Quiz content",
				Position: 150,
			},
		},
	}

	data, err := json.Marshal(lesson)
	require.NoError(t, err)

	var unmarshaled Lesson
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, lesson.ID, unmarshaled.ID)
	assert.Equal(t, lesson.Title, unmarshaled.Title)
	assert.Len(t, unmarshaled.Subtitles, 1)
	assert.Len(t, unmarshaled.InteractiveElements, 1)
}

func TestUser_JSONMarshaling(t *testing.T) {
	now := time.Now()
	user := User{
		ID:        "user-id",
		Email:     "test@example.com",
		Password:  "hashed-password",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "creator",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	// Password should not be in JSON
	assert.NotContains(t, string(data), "hashed-password")

	var unmarshaled User
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, user.ID, unmarshaled.ID)
	assert.Equal(t, user.Email, unmarshaled.Email)
	assert.Empty(t, unmarshaled.Password) // Should be empty due to json:"-" tag
	assert.Equal(t, user.FirstName, unmarshaled.FirstName)
	assert.Equal(t, user.Role, unmarshaled.Role)
}

func TestJob_JSONMarshaling(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now

	job := Job{
		ID:          "job-id",
		UserID:      "user-id",
		Type:        "course_generation",
		Status:      "completed",
		Progress:    100,
		Payload:     map[string]interface{}{"course_id": "course-123"},
		Result:      map[string]interface{}{"output_path": "/path/to/output"},
		CreatedAt:   now,
		UpdatedAt:   now,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}

	data, err := json.Marshal(job)
	require.NoError(t, err)

	var unmarshaled Job
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, job.ID, unmarshaled.ID)
	assert.Equal(t, job.Type, unmarshaled.Type)
	assert.Equal(t, job.Status, unmarshaled.Status)
	assert.Equal(t, job.Progress, unmarshaled.Progress)
}

func TestCourseDB_BeforeCreate(t *testing.T) {
	course := &CourseDB{
		UserID:      "user-123",
		Title:       "Test Course",
		Description: "Test description",
	}

	err := course.BeforeCreate(&gorm.DB{})
	require.NoError(t, err)

	// Should generate a UUID
	assert.NotEmpty(t, course.ID)
	_, err = uuid.Parse(course.ID)
	assert.NoError(t, err)
}

func TestLessonDB_BeforeCreate(t *testing.T) {
	lesson := &LessonDB{
		CourseID: "course-123",
		Title:    "Test Lesson",
		Content:  "Test content",
		Order:    1,
	}

	err := lesson.BeforeCreate(&gorm.DB{})
	require.NoError(t, err)

	assert.NotEmpty(t, lesson.ID)
	_, err = uuid.Parse(lesson.ID)
	assert.NoError(t, err)
}

func TestUserDB_BeforeCreate(t *testing.T) {
	user := &UserDB{
		Email:     "test@example.com",
		Password:  "hashed-password",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "creator",
	}

	err := user.BeforeCreate(&gorm.DB{})
	require.NoError(t, err)

	assert.NotEmpty(t, user.ID)
	_, err = uuid.Parse(user.ID)
	assert.NoError(t, err)
}

func TestJobDB_BeforeCreate(t *testing.T) {
	job := &JobDB{
		UserID:  "user-123",
		Type:    "course_generation",
		Status:  "pending",
		Payload: "{}",
	}

	err := job.BeforeCreate(&gorm.DB{})
	require.NoError(t, err)

	assert.NotEmpty(t, job.ID)
	_, err = uuid.Parse(job.ID)
	assert.NoError(t, err)
}

func TestTableNames(t *testing.T) {
	assert.Equal(t, "courses", CourseDB{}.TableName())
	assert.Equal(t, "course_metadata", CourseMetadataDB{}.TableName())
	assert.Equal(t, "lessons", LessonDB{}.TableName())
	assert.Equal(t, "subtitles", SubtitleDB{}.TableName())
	assert.Equal(t, "interactive_elements", InteractiveElementDB{}.TableName())
	assert.Equal(t, "processing_jobs", ProcessingJobDB{}.TableName())
	assert.Equal(t, "users", UserDB{}.TableName())
	assert.Equal(t, "user_preferences", UserPreferencesDB{}.TableName())
	assert.Equal(t, "user_sessions", UserSessionDB{}.TableName())
	assert.Equal(t, "jobs", JobDB{}.TableName())
}

func TestProcessingOptions_JSONMarshaling(t *testing.T) {
	voice := "en-US-Wavenet-A"
	options := ProcessingOptions{
		Voice:           &voice,
		BackgroundMusic: true,
		BackgroundStyle: "nature",
		Languages:       []string{"en", "es"},
		Quality:         "high",
	}

	data, err := json.Marshal(options)
	require.NoError(t, err)

	var unmarshaled ProcessingOptions
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, *options.Voice, *unmarshaled.Voice)
	assert.Equal(t, options.BackgroundMusic, unmarshaled.BackgroundMusic)
	assert.Equal(t, options.BackgroundStyle, unmarshaled.BackgroundStyle)
	assert.Equal(t, options.Languages, unmarshaled.Languages)
	assert.Equal(t, options.Quality, unmarshaled.Quality)
}

func TestParsedCourse_JSONMarshaling(t *testing.T) {
	parsed := ParsedCourse{
		Title:       "Parsed Course",
		Description: "Parsed description",
		Sections: []ParsedSection{
			{
				Title:   "Section 1",
				Content: "Section content",
				Order:   1,
			},
		},
		Metadata: map[string]interface{}{
			"author": "Test Author",
		},
	}

	data, err := json.Marshal(parsed)
	require.NoError(t, err)

	var unmarshaled ParsedCourse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, parsed.Title, unmarshaled.Title)
	assert.Equal(t, parsed.Description, unmarshaled.Description)
	assert.Len(t, unmarshaled.Sections, 1)
	assert.Equal(t, "Test Author", unmarshaled.Metadata["author"])
}
