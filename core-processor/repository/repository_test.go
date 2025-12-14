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
