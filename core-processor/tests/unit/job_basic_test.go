package unit

import (
	"context"
	"testing"

	"github.com/course-creator/core-processor/jobs"
	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobQueueBasic(t *testing.T) {
	t.Run("Enqueue without starting", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)
		
		// Create user
		testUser := &models.UserDB{
			ID:    "test-user-id",
			Email: "test@example.com",
			Role:  "creator",
		}
		require.NoError(t, db.Create(testUser).Error)
		
		payload := map[string]interface{}{
			"test_key": "test_value",
		}
			
		// This should save to DB but not process (queue not started)
		job, err := queue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "test-user-id", payload, jobs.JobPriorityNormal)
		require.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, jobs.JobStatusPending, job.Status)
		
		// Verify job is in DB
		var jobDB models.JobDB
		err = db.Where("id = ?", job.ID).First(&jobDB).Error
		require.NoError(t, err)
		assert.Equal(t, job.ID, jobDB.ID)
		assert.Equal(t, "test-user-id", jobDB.UserID)
	})
	
	t.Run("Get job", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)
		
		// Create user
		testUser := &models.UserDB{
			ID:    "test-user-id",
			Email: "test@example.com",
			Role:  "creator",
		}
		require.NoError(t, db.Create(testUser).Error)
		
		payload := map[string]interface{}{
			"test_key": "test_value",
		}
			
		// Create job
		createdJob, err := queue.Enqueue(context.Background(), jobs.JobTypeVideoProcessing, "test-user-id", payload, jobs.JobPriorityHigh)
		require.NoError(t, err)
		
		// Get job
		retrievedJob, err := queue.GetJob(context.Background(), createdJob.ID)
		require.NoError(t, err)
		assert.Equal(t, createdJob.ID, retrievedJob.ID)
		assert.Equal(t, createdJob.Type, retrievedJob.Type)
		assert.Equal(t, createdJob.UserID, retrievedJob.UserID)
		assert.Equal(t, createdJob.Payload, retrievedJob.Payload)
	})
	
	t.Run("Get user jobs", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)
		
		// Create user
		testUser := &models.UserDB{
			ID:    "test-user-id",
			Email: "test@example.com",
			Role:  "creator",
		}
		require.NoError(t, db.Create(testUser).Error)
		
		// Create jobs
		job1, _ := queue.Enqueue(context.Background(), jobs.JobTypeAudioGeneration, "test-user-id", map[string]interface{}{}, jobs.JobPriorityNormal)
		job2, _ := queue.Enqueue(context.Background(), jobs.JobTypeSubtitleGeneration, "test-user-id", map[string]interface{}{}, jobs.JobPriorityNormal)
		
		// Get jobs for user
		userJobs, err := queue.GetUserJobs(context.Background(), "test-user-id", 10, 0)
		require.NoError(t, err)
		assert.Len(t, userJobs, 2)
		
		// Verify job IDs
		jobIDs := make(map[string]bool)
		for _, job := range userJobs {
			jobIDs[job.ID] = true
		}
		assert.True(t, jobIDs[job1.ID])
		assert.True(t, jobIDs[job2.ID])
	})
}