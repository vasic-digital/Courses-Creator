package integration

import (
	"testing"
	"time"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_Integration(t *testing.T) {
	// Setup in-memory database
	db, err := database.NewDatabase(&database.Config{
		Path:  ":memory:",
		Debug: false,
		Env:   "test",
	})
	require.NoError(t, err)
	defer db.Close()

	// Get repositories
	courseRepo := repository.NewCourseRepository(db)
	jobRepo := repository.NewProcessingJobRepository(db)

	t.Run("Course CRUD Operations", func(t *testing.T) {
		// First create a system user (matching the hardcoded UserID in CreateCourse)
		systemUser := &models.UserDB{
			ID:        "system",
			Email:     "system@example.com",
			Password:  "hashed_password",
			FirstName: "System",
			LastName:  "User",
			Role:      "admin",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Create system user in database
		err = db.GetGormDB().Create(systemUser).Error
		require.NoError(t, err, "Should create system user")

		// Create a course using the API model
		course := &models.Course{
			ID:          uuid.New().String(),
			Title:       "Test Course",
			Description: "Test course description",
		}

		// Test Create
		t.Log("Creating course...")
		createdCourse, err := courseRepo.CreateCourse(course)
		if err != nil {
			t.Logf("CreateCourse error: %v", err)
		} else {
			t.Logf("CreateCourse succeeded, createdCourse: %v", createdCourse)
		}
		require.NoError(t, err, "CreateCourse should succeed")
		require.NotNil(t, createdCourse, "Created course should not be nil")
		assert.Equal(t, course.ID, createdCourse.ID)
		assert.Equal(t, "Test Course", createdCourse.Title)

		// Direct database check
		var directCourse models.CourseDB
		err = db.GetGormDB().First(&directCourse, "id = ?", course.ID).Error
		if err != nil {
			t.Logf("Direct database check failed: %v", err)
			// Count all courses
			var count int64
			err = db.GetGormDB().Model(&models.CourseDB{}).Count(&count).Error
			t.Logf("Total courses in database: %d", count)
		} else {
			t.Logf("Direct database check passed: ID=%s, Title=%s, DeletedAt.Valid=%v, DeletedAt.Time=%v", directCourse.ID, directCourse.Title, directCourse.DeletedAt.Valid, directCourse.DeletedAt.Time)
		}

		// Test Get by ID
		retrievedCourse, err := courseRepo.GetCourseByID(course.ID)
		if err != nil {
			t.Logf("GetCourseByID error: %v", err)
			// Try to get the course directly without preloads
			var directCourse models.CourseDB
			err = db.GetGormDB().First(&directCourse, "id = ?", course.ID).Error
			if err != nil {
				t.Logf("Direct query with db.GetGormDB() error: %v", err)
			} else {
				t.Logf("Direct query with db.GetGormDB() found course: ID=%s, Title=%s, UserID=%s, DeletedAt=%v", directCourse.ID, directCourse.Title, directCourse.UserID, directCourse.DeletedAt)
			}
			// Fail the test
			t.FailNow()
		}
		require.NoError(t, err)
		assert.Equal(t, course.ID, retrievedCourse.ID)
		assert.Equal(t, "Test Course", retrievedCourse.Title)

		// Test Update
		updateCourse := &models.Course{
			ID:          retrievedCourse.ID,
			Title:       "Updated Test Course",
			Description: retrievedCourse.Description,
		}
		updatedCourse, err := courseRepo.UpdateCourse(course.ID, updateCourse)
		require.NoError(t, err)
		assert.Equal(t, "Updated Test Course", updatedCourse.Title)

		// Test Get All
		courses, total, err := courseRepo.GetAllCourses(0, 10)
		if err != nil {
			t.Logf("GetAllCourses error: %v", err)
		}
		t.Logf("GetAllCourses returned %d courses, total=%d", len(courses), total)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(courses), 1)
		assert.GreaterOrEqual(t, total, 1)

		// Test Search
		searchedCourses, total, err := courseRepo.SearchCourses("Test", 0, 10)
		if err != nil {
			t.Logf("SearchCourses error: %v", err)
		}
		t.Logf("SearchCourses returned %d courses, total=%d", len(searchedCourses), total)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(searchedCourses), 1)
		assert.GreaterOrEqual(t, total, 1)

		// Test Delete
		err = courseRepo.DeleteCourse(course.ID)
		require.NoError(t, err)

		// Verify deletion
		deletedCourse, err := courseRepo.GetCourseByID(course.ID)
		assert.Error(t, err)
		assert.Nil(t, deletedCourse)
	})

	t.Run("Job CRUD Operations", func(t *testing.T) {
		// Create a job using the correct type
		job := &models.ProcessingJobDB{
			ID:        uuid.New().String(),
			Status:    "pending",
			InputPath: "/test/input.md",
			Options:   "{}",
			Progress:  0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Test Create
		createdJob, err := jobRepo.CreateJob(job)
		require.NoError(t, err)
		assert.Equal(t, job.ID, createdJob.ID)
		assert.Equal(t, "pending", createdJob.Status)

		// Test Get by ID
		retrievedJob, err := jobRepo.GetJobByID(job.ID)
		require.NoError(t, err)
		assert.Equal(t, job.ID, retrievedJob.ID)

		// Test Update Status
		updatedJob, err := jobRepo.UpdateJobStatus(job.ID, "processing")
		require.NoError(t, err)
		assert.Equal(t, "processing", updatedJob.Status)

		// Test Update Progress
		updatedJob, err = jobRepo.UpdateJobProgress(job.ID, 50)
		require.NoError(t, err)
		assert.Equal(t, 50, updatedJob.Progress)

		// Test Get All Jobs
		jobs, total, err := jobRepo.GetAllJobs(1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(jobs), 1)
		assert.GreaterOrEqual(t, total, 1)

		// Test Get Jobs by Status
		processingJobs, total, err := jobRepo.GetJobsByStatus("processing", 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(processingJobs), 1)
		assert.GreaterOrEqual(t, total, 1)

		// Test Get Pending Jobs
		pendingJobs, err := jobRepo.GetPendingJobs()
		require.NoError(t, err)
		// Our job is now "processing", not "pending"
		assert.Equal(t, 0, len(pendingJobs))

		// Test Delete
		err = jobRepo.DeleteJob(job.ID)
		require.NoError(t, err)

		// Verify deletion
		deletedJob, err := jobRepo.GetJobByID(job.ID)
		assert.Error(t, err)
		assert.Nil(t, deletedJob)
	})

	t.Run("Database Connection Health", func(t *testing.T) {
		// Test Ping
		err := db.Ping()
		assert.NoError(t, err)

		// Test Close (already deferred)
		assert.NotNil(t, db.GetGormDB())
	})
}
