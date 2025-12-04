package unit

import (
	"context"
	"testing"
	"time"

	"github.com/course-creator/core-processor/jobs"
	"github.com/course-creator/core-processor/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobQueueBasicOperations(t *testing.T) {
	t.Run("Create Queue", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)

		// Should not be running initially
		assert.False(t, queue.IsRunning())
	})

	t.Run("Start and Stop Queue", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)

		// Register a dummy handler
		queue.RegisterHandler(jobs.JobTypeCourseGeneration, func(ctx context.Context, job *jobs.Job) error {
			return nil
		})

		// Start queue
		err := queue.Start()
		require.NoError(t, err)
		assert.True(t, queue.IsRunning())

		// Stop queue
		queue.Stop()
		assert.False(t, queue.IsRunning())
	})

	t.Run("Create and Convert Job", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)

		// Create a job in memory
		jobID := uuid.New().String()
		job := &jobs.Job{
			ID:        jobID,
			UserID:    "test-user-id",
			Type:      jobs.JobTypeCourseGeneration,
			Status:    jobs.JobStatusPending,
			Priority:  jobs.JobPriorityNormal,
			Payload:   map[string]interface{}{"test": "data"},
			CreatedAt: time.Now(),
		}

		// Convert to DB model using method on queue
		jobDB, err := queue.ConvertToDBModel(job)
		require.NoError(t, err)

		// Check conversion
		assert.Equal(t, job.ID, jobDB.ID)
		assert.Equal(t, job.UserID, jobDB.UserID)
		assert.Equal(t, string(job.Type), jobDB.Type)
		assert.Equal(t, string(job.Status), jobDB.Status)
		assert.Equal(t, job.Progress, jobDB.Progress)
	})

	t.Run("Convert DB Job to Job", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)

		// Create a DB job
		jobDB := &models.JobDB{
			ID:        uuid.New().String(),
			UserID:    "test-user-id",
			Type:      string(jobs.JobTypeCourseGeneration),
			Status:    string(jobs.JobStatusPending),
			Progress:  0,
			Payload:   `{"test": "data"}`,
			CreatedAt: time.Now(),
		}

		// Convert to job using method on queue
		job, err := queue.ConvertFromDBModel(jobDB)
		require.NoError(t, err)

		// Check conversion
		assert.Equal(t, jobDB.ID, job.ID)
		assert.Equal(t, jobDB.UserID, job.UserID)
		assert.Equal(t, jobs.JobType(jobDB.Type), job.Type)
		assert.Equal(t, jobs.JobStatus(jobDB.Status), job.Status)
		assert.Equal(t, jobDB.Progress, job.Progress)
		assert.Equal(t, "data", job.Payload["test"])
	})
}

func TestJobTypes(t *testing.T) {
	// Test all job types are defined
	t.Run("JobType Constants", func(t *testing.T) {
		assert.Equal(t, jobs.JobType("course_generation"), jobs.JobTypeCourseGeneration)
		assert.Equal(t, jobs.JobType("video_processing"), jobs.JobTypeVideoProcessing)
		assert.Equal(t, jobs.JobType("audio_generation"), jobs.JobTypeAudioGeneration)
		assert.Equal(t, jobs.JobType("subtitle_generation"), jobs.JobTypeSubtitleGeneration)
	})

	t.Run("JobStatus Constants", func(t *testing.T) {
		assert.Equal(t, jobs.JobStatus("pending"), jobs.JobStatusPending)
		assert.Equal(t, jobs.JobStatus("running"), jobs.JobStatusRunning)
		assert.Equal(t, jobs.JobStatus("completed"), jobs.JobStatusCompleted)
		assert.Equal(t, jobs.JobStatus("failed"), jobs.JobStatusFailed)
		assert.Equal(t, jobs.JobStatus("cancelled"), jobs.JobStatusCancelled)
	})
}
