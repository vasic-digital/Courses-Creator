package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
)

func setupTestDB(t *testing.T) *database.DB {
	// Create a test database
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models
	err = gormDB.AutoMigrate(
		&models.CourseDB{},
		&models.CourseMetadataDB{},
		&models.LessonDB{},
		&models.SubtitleDB{},
		&models.InteractiveElementDB{},
		&models.ProcessingJobDB{},
		&models.UserDB{},
		&models.JobDB{},
	)
	require.NoError(t, err)

	// Create database wrapper
	db := &database.DB{DB: gormDB}

	return db
}

func stringPtr(s string) *string {
	return &s
}

func TestProcessingJobRepository_CreateJob(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	job := &models.ProcessingJobDB{
		ID:         "test-job-1",
		Status:     "pending",
		CourseID:   stringPtr("course123"),
		Progress:   0,
		InputPath:  "/input/test.md",
		OutputPath: stringPtr("/output/test"),
		Options:    `{"voice":"en","quality":"high"}`,
	}

	result, err := repo.CreateJob(job)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, job.ID, result.ID)
	assert.Equal(t, job.Status, result.Status)
	assert.Equal(t, job.Progress, result.Progress)
	assert.Equal(t, job.InputPath, result.InputPath)
	assert.NotZero(t, result.CreatedAt)
}

func TestProcessingJobRepository_GetJobByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create a job first
	job := &models.ProcessingJobDB{
		ID:        "test-get-job",
		Status:    "running",
		Progress:  50,
		InputPath: "/input/test.md",
	}

	_, err := repo.CreateJob(job)
	require.NoError(t, err)

	tests := []struct {
		name        string
		jobID       string
		expectError bool
	}{
		{
			name:        "successful job retrieval",
			jobID:       "test-get-job",
			expectError: false,
		},
		{
			name:        "job not found",
			jobID:       "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetJobByID(tt.jobID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.jobID, result.ID)
				assert.Equal(t, "running", result.Status)
			}
		})
	}
}

func TestProcessingJobRepository_GetAllJobs(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create test jobs
	jobs := []*models.ProcessingJobDB{
		{ID: "job-1", Status: "pending", InputPath: "/input/1.md"},
		{ID: "job-2", Status: "running", InputPath: "/input/2.md"},
		{ID: "job-3", Status: "completed", InputPath: "/input/3.md"},
	}

	for _, job := range jobs {
		_, err := repo.CreateJob(job)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		offset        int
		limit         int
		expectedLen   int
		expectedTotal int64
	}{
		{
			name:          "get all jobs",
			offset:        0,
			limit:         10,
			expectedLen:   3,
			expectedTotal: 3,
		},
		{
			name:          "get jobs with pagination",
			offset:        1,
			limit:         2,
			expectedLen:   2,
			expectedTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := repo.GetAllJobs(tt.offset, tt.limit)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestProcessingJobRepository_GetJobsByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create test jobs with different statuses
	jobs := []*models.ProcessingJobDB{
		{ID: "job-pending-1", Status: "pending", InputPath: "/input/1.md"},
		{ID: "job-pending-2", Status: "pending", InputPath: "/input/2.md"},
		{ID: "job-running", Status: "running", InputPath: "/input/3.md"},
		{ID: "job-completed", Status: "completed", InputPath: "/input/4.md"},
	}

	for _, job := range jobs {
		_, err := repo.CreateJob(job)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		status        string
		expectedLen   int
		expectedTotal int64
	}{
		{
			name:          "get pending jobs",
			status:        "pending",
			expectedLen:   2,
			expectedTotal: 2,
		},
		{
			name:          "get running jobs",
			status:        "running",
			expectedLen:   1,
			expectedTotal: 1,
		},
		{
			name:          "get completed jobs",
			status:        "completed",
			expectedLen:   1,
			expectedTotal: 1,
		},
		{
			name:          "get failed jobs",
			status:        "failed",
			expectedLen:   0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := repo.GetJobsByStatus(tt.status, 0, 10)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestProcessingJobRepository_UpdateJob(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create initial job
	job := &models.ProcessingJobDB{
		ID:        "update-test",
		Status:    "pending",
		Progress:  0,
		InputPath: "/input/test.md",
	}

	_, err := repo.CreateJob(job)
	require.NoError(t, err)

	// Update job
	updates := map[string]interface{}{
		"status":   "running",
		"progress": 50,
	}

	result, err := repo.UpdateJob("update-test", updates)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "update-test", result.ID)
	assert.Equal(t, "running", result.Status)
	assert.Equal(t, 50, result.Progress)
	assert.True(t, result.UpdatedAt.After(result.CreatedAt))
}

func TestProcessingJobRepository_UpdateJobStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create job
	job := &models.ProcessingJobDB{
		ID:        "status-test",
		Status:    "pending",
		InputPath: "/input/test.md",
	}

	_, err := repo.CreateJob(job)
	require.NoError(t, err)

	// Update to running
	result, err := repo.UpdateJobStatus("status-test", "running")
	assert.NoError(t, err)
	assert.Equal(t, "running", result.Status)
	assert.NotNil(t, result.StartedAt)

	// Update to completed
	result, err = repo.UpdateJobStatus("status-test", "completed")
	assert.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.NotNil(t, result.CompletedAt)
}

func TestProcessingJobRepository_UpdateJobProgress(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create job
	job := &models.ProcessingJobDB{
		ID:        "progress-test",
		Status:    "running",
		Progress:  0,
		InputPath: "/input/test.md",
	}

	_, err := repo.CreateJob(job)
	require.NoError(t, err)

	// Update progress
	result, err := repo.UpdateJobProgress("progress-test", 75)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 75, result.Progress)
}

func TestProcessingJobRepository_UpdateJobError(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create job
	job := &models.ProcessingJobDB{
		ID:        "error-test",
		Status:    "running",
		Progress:  50,
		InputPath: "/input/test.md",
	}

	_, err := repo.CreateJob(job)
	require.NoError(t, err)

	// Update with error
	result, err := repo.UpdateJobError("error-test", "processing failed")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "failed", result.Status)
	assert.Equal(t, 0, result.Progress)
	assert.Equal(t, "processing failed", *result.Error)
}

func TestProcessingJobRepository_SetJobCourseID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create job without course ID
	job := &models.ProcessingJobDB{
		ID:        "course-id-test",
		Status:    "pending",
		InputPath: "/input/test.md",
	}

	_, err := repo.CreateJob(job)
	require.NoError(t, err)

	// Set course ID
	result, err := repo.SetJobCourseID("course-id-test", "course-456")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "course-456", *result.CourseID)
}

func TestProcessingJobRepository_DeleteJob(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create job
	job := &models.ProcessingJobDB{
		ID:        "delete-test",
		Status:    "pending",
		InputPath: "/input/test.md",
	}

	_, err := repo.CreateJob(job)
	require.NoError(t, err)

	// Verify job exists
	_, err = repo.GetJobByID("delete-test")
	assert.NoError(t, err)

	// Delete job
	err = repo.DeleteJob("delete-test")
	assert.NoError(t, err)

	// Verify job is deleted
	_, err = repo.GetJobByID("delete-test")
	assert.Error(t, err)
}

func TestProcessingJobRepository_GetPendingJobs(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create jobs with different statuses
	jobs := []*models.ProcessingJobDB{
		{ID: "pending-1", Status: "pending", InputPath: "/input/1.md"},
		{ID: "pending-2", Status: "pending", InputPath: "/input/2.md"},
		{ID: "running-1", Status: "running", InputPath: "/input/3.md"},
	}

	for _, job := range jobs {
		_, err := repo.CreateJob(job)
		require.NoError(t, err)
	}

	result, err := repo.GetPendingJobs()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	for _, job := range result {
		assert.Equal(t, "pending", job.Status)
	}
}

func TestProcessingJobRepository_GetRunningJobs(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProcessingJobRepository(db)

	// Create jobs with different statuses
	jobs := []*models.ProcessingJobDB{
		{ID: "running-1", Status: "running", InputPath: "/input/1.md"},
		{ID: "running-2", Status: "running", InputPath: "/input/2.md"},
		{ID: "pending-1", Status: "pending", InputPath: "/input/3.md"},
	}

	for _, job := range jobs {
		_, err := repo.CreateJob(job)
		require.NoError(t, err)
	}

	result, err := repo.GetRunningJobs()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	for _, job := range result {
		assert.Equal(t, "running", job.Status)
	}
}

func TestParseJobOptions(t *testing.T) {
	tests := []struct {
		name        string
		optionsJSON string
		expected    *JobOptions
		expectError bool
	}{
		{
			name:        "empty options",
			optionsJSON: "",
			expected: &JobOptions{
				BackgroundMusic: false,
				Languages:       []string{"en"},
				Quality:         "standard",
			},
			expectError: false,
		},
		{
			name:        "valid options",
			optionsJSON: `{"voice":"en","background_music":true,"languages":["en","es"],"quality":"high"}`,
			expected: &JobOptions{
				Voice:           stringPtr("en"),
				BackgroundMusic: true,
				Languages:       []string{"en", "es"},
				Quality:         "high",
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			optionsJSON: `{"invalid": json}`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJobOptions(tt.optionsJSON)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Voice, result.Voice)
				assert.Equal(t, tt.expected.BackgroundMusic, result.BackgroundMusic)
				assert.Equal(t, tt.expected.Languages, result.Languages)
				assert.Equal(t, tt.expected.Quality, result.Quality)
			}
		})
	}
}

func TestSerializeJobOptions(t *testing.T) {
	options := &JobOptions{
		Voice:           stringPtr("en"),
		BackgroundMusic: true,
		Languages:       []string{"en", "es"},
		Quality:         "high",
	}

	result, err := SerializeJobOptions(options)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Verify it can be parsed back
	parsed, err := ParseJobOptions(result)
	assert.NoError(t, err)
	assert.Equal(t, options.Voice, parsed.Voice)
	assert.Equal(t, options.BackgroundMusic, parsed.BackgroundMusic)
	assert.Equal(t, options.Languages, parsed.Languages)
	assert.Equal(t, options.Quality, parsed.Quality)
}

func TestCourseRepository_CreateCourse(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseRepository(db)

	course := &models.Course{
		ID:          "test-course-1",
		Title:       "Test Course",
		Description: "A test course description",
		Metadata: models.CourseMetadata{
			Author:        "Test Author",
			Language:      "en",
			Tags:          []string{"test", "golang"},
			ThumbnailURL:  stringPtr("https://example.com/thumb.jpg"),
			TotalDuration: 3600,
		},
		Lessons: []models.Lesson{
			{
				ID:       "lesson-1",
				Title:    "Introduction",
				Content:  "Welcome to the course",
				VideoURL: stringPtr("https://example.com/video1.mp4"),
				AudioURL: stringPtr("https://example.com/audio1.mp3"),
				Duration: 600,
				Order:    1,
				Subtitles: []models.Subtitle{
					{
						Language: "en",
						Content:  "Welcome to the course",
						Timestamps: []map[string]interface{}{
							{"start": 0.0, "end": 5.0, "text": "Welcome"},
						},
					},
				},
				InteractiveElements: []models.InteractiveElement{
					{
						ID:       "element-1",
						Type:     "quiz",
						Content:  `{"question": "What is Go?", "options": ["Language", "Food"], "answer": 0}`,
						Position: 300,
					},
				},
			},
		},
	}

	result, err := repo.CreateCourse(course)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, course.ID, result.ID)
	assert.Equal(t, course.Title, result.Title)
	assert.Equal(t, course.Description, result.Description)
	assert.NotZero(t, result.CreatedAt)
	assert.NotZero(t, result.UpdatedAt)

	// Verify metadata was created
	assert.Equal(t, course.Metadata.Author, result.Metadata.Author)
	assert.Equal(t, course.Metadata.Language, result.Metadata.Language)
	assert.Equal(t, course.Metadata.ThumbnailURL, result.Metadata.ThumbnailURL)
	assert.Equal(t, course.Metadata.TotalDuration, result.Metadata.TotalDuration)

	// Verify lessons were created
	assert.Len(t, result.Lessons, 1)
	assert.Equal(t, course.Lessons[0].ID, result.Lessons[0].ID)
	assert.Equal(t, course.Lessons[0].Title, result.Lessons[0].Title)
	assert.Equal(t, course.Lessons[0].Content, result.Lessons[0].Content)
	assert.Equal(t, course.Lessons[0].VideoURL, result.Lessons[0].VideoURL)
	assert.Equal(t, course.Lessons[0].AudioURL, result.Lessons[0].AudioURL)
	assert.Equal(t, course.Lessons[0].Duration, result.Lessons[0].Duration)
	assert.Equal(t, course.Lessons[0].Order, result.Lessons[0].Order)

	// Verify subtitles were created
	assert.Len(t, result.Lessons[0].Subtitles, 1)
	assert.Equal(t, course.Lessons[0].Subtitles[0].Language, result.Lessons[0].Subtitles[0].Language)
	assert.Equal(t, course.Lessons[0].Subtitles[0].Content, result.Lessons[0].Subtitles[0].Content)

	// Verify interactive elements were created
	assert.Len(t, result.Lessons[0].InteractiveElements, 1)
	assert.Equal(t, course.Lessons[0].InteractiveElements[0].ID, result.Lessons[0].InteractiveElements[0].ID)
	assert.Equal(t, course.Lessons[0].InteractiveElements[0].Type, result.Lessons[0].InteractiveElements[0].Type)
	assert.Equal(t, course.Lessons[0].InteractiveElements[0].Content, result.Lessons[0].InteractiveElements[0].Content)
	assert.Equal(t, course.Lessons[0].InteractiveElements[0].Position, result.Lessons[0].InteractiveElements[0].Position)
}

func TestCourseRepository_GetCourseByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseRepository(db)

	// Create a course first
	course := &models.Course{
		ID:          "get-test-course",
		Title:       "Get Test Course",
		Description: "Course for get testing",
		Metadata: models.CourseMetadata{
			Author:   "Test Author",
			Language: "en",
		},
		Lessons: []models.Lesson{
			{
				ID:      "get-lesson-1",
				Title:   "Lesson 1",
				Content: "Content 1",
				Order:   1,
			},
		},
	}

	created, err := repo.CreateCourse(course)
	require.NoError(t, err)

	tests := []struct {
		name        string
		courseID    string
		expectError bool
	}{
		{
			name:        "successful course retrieval",
			courseID:    "get-test-course",
			expectError: false,
		},
		{
			name:        "course not found",
			courseID:    "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetCourseByID(tt.courseID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.courseID, result.ID)
				assert.Equal(t, created.Title, result.Title)
				assert.Equal(t, created.Metadata.Author, result.Metadata.Author)
				assert.Len(t, result.Lessons, 1)
			}
		})
	}
}

func TestCourseRepository_GetAllCourses(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseRepository(db)

	// Create test courses
	courses := []*models.Course{
		{
			ID:          "course-1",
			Title:       "Course One",
			Description: "First course",
		},
		{
			ID:          "course-2",
			Title:       "Course Two",
			Description: "Second course",
		},
		{
			ID:          "course-3",
			Title:       "Course Three",
			Description: "Third course",
		},
	}

	for _, course := range courses {
		_, err := repo.CreateCourse(course)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		offset        int
		limit         int
		expectedLen   int
		expectedTotal int64
	}{
		{
			name:          "get all courses",
			offset:        0,
			limit:         10,
			expectedLen:   3,
			expectedTotal: 3,
		},
		{
			name:          "get courses with pagination",
			offset:        1,
			limit:         2,
			expectedLen:   2,
			expectedTotal: 3,
		},
		{
			name:          "get courses with small limit",
			offset:        0,
			limit:         1,
			expectedLen:   1,
			expectedTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := repo.GetAllCourses(tt.offset, tt.limit)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestCourseRepository_UpdateCourse(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseRepository(db)

	// Create initial course
	course := &models.Course{
		ID:          "update-test-course",
		Title:       "Original Title",
		Description: "Original description",
		Metadata: models.CourseMetadata{
			Author:   "Original Author",
			Language: "en",
		},
		Lessons: []models.Lesson{
			{
				ID:      "original-lesson",
				Title:   "Original Lesson",
				Content: "Original content",
				Order:   1,
			},
		},
	}

	_, err := repo.CreateCourse(course)
	require.NoError(t, err)

	// Update course
	updatedCourse := &models.Course{
		ID:          "update-test-course",
		Title:       "Updated Title",
		Description: "Updated description",
		Metadata: models.CourseMetadata{
			Author:   "Updated Author",
			Language: "es",
		},
		Lessons: []models.Lesson{
			{
				ID:      "updated-lesson",
				Title:   "Updated Lesson",
				Content: "Updated content",
				Order:   1,
			},
		},
	}

	result, err := repo.UpdateCourse("update-test-course", updatedCourse)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "update-test-course", result.ID)
	assert.Equal(t, "Updated Title", result.Title)
	assert.Equal(t, "Updated description", result.Description)
	assert.Equal(t, "Updated Author", result.Metadata.Author)
	assert.Equal(t, "es", result.Metadata.Language)
	assert.True(t, result.UpdatedAt.After(result.CreatedAt))

	// Verify lessons were updated
	assert.Len(t, result.Lessons, 1)
	assert.Equal(t, "updated-lesson", result.Lessons[0].ID)
	assert.Equal(t, "Updated Lesson", result.Lessons[0].Title)
	assert.Equal(t, "Updated content", result.Lessons[0].Content)
}

func TestCourseRepository_DeleteCourse(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseRepository(db)

	// Create course
	course := &models.Course{
		ID:          "delete-test-course",
		Title:       "Delete Test Course",
		Description: "Course to be deleted",
	}

	_, err := repo.CreateCourse(course)
	require.NoError(t, err)

	// Verify course exists
	_, err = repo.GetCourseByID("delete-test-course")
	assert.NoError(t, err)

	// Delete course
	err = repo.DeleteCourse("delete-test-course")
	assert.NoError(t, err)

	// Verify course is deleted
	_, err = repo.GetCourseByID("delete-test-course")
	assert.Error(t, err)
}

func TestCourseRepository_SearchCourses(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseRepository(db)

	// Create test courses
	courses := []*models.Course{
		{
			ID:          "search-1",
			Title:       "Golang Programming",
			Description: "Learn Go programming language",
		},
		{
			ID:          "search-2",
			Title:       "Python Basics",
			Description: "Introduction to Python programming",
		},
		{
			ID:          "search-3",
			Title:       "Advanced Golang",
			Description: "Advanced Go concepts",
		},
	}

	for _, course := range courses {
		_, err := repo.CreateCourse(course)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		query         string
		offset        int
		limit         int
		expectedLen   int
		expectedTotal int64
	}{
		{
			name:          "search by title",
			query:         "Golang",
			offset:        0,
			limit:         10,
			expectedLen:   2,
			expectedTotal: 2,
		},
		{
			name:          "search by description",
			query:         "programming",
			offset:        0,
			limit:         10,
			expectedLen:   2,
			expectedTotal: 2,
		},
		{
			name:          "search with no results",
			query:         "nonexistent",
			offset:        0,
			limit:         10,
			expectedLen:   0,
			expectedTotal: 0,
		},
		{
			name:          "search with pagination",
			query:         "Go",
			offset:        0,
			limit:         1,
			expectedLen:   1,
			expectedTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := repo.SearchCourses(tt.query, tt.offset, tt.limit)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestCourseRepository_ErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseRepository(db)

	t.Run("create course with duplicate ID should fail", func(t *testing.T) {
		course := &models.Course{
			ID:          "duplicate-test",
			Title:       "Test Course",
			Description: "A test course",
		}

		// Create first course
		_, err := repo.CreateCourse(course)
		assert.NoError(t, err)

		// Try to create duplicate - this should fail due to unique constraint
		_, err = repo.CreateCourse(course)
		assert.Error(t, err)
	})

	t.Run("update non-existent course should fail", func(t *testing.T) {
		course := &models.Course{
			ID:          "nonexistent-update",
			Title:       "Updated Title",
			Description: "Updated description",
		}

		_, err := repo.UpdateCourse("nonexistent-update", course)
		assert.Error(t, err)
	})

	t.Run("delete non-existent course should not error", func(t *testing.T) {
		// GORM delete is idempotent - deleting non-existent record doesn't error
		err := repo.DeleteCourse("nonexistent-delete")
		assert.NoError(t, err)
	})
}
