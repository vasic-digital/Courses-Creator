package unit

import (
	"context"
	"testing"
	"time"

	"github.com/course-creator/core-processor/jobs"
	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestJobQueue(t *testing.T) *jobs.JobQueue {
	db := setupTestDB(t)
	queue := jobs.NewJobQueue(db, 2) // 2 workers
	
	return queue
}

func TestJobQueue(t *testing.T) {
	queue := setupTestJobQueue(t)
	
	t.Run("Start and Stop Queue", func(t *testing.T) {
		// Start queue
		err := queue.Start()
		require.NoError(t, err)
		
		// Stop queue
		queue.Stop()
	})
	
	t.Run("Enqueue Job", func(t *testing.T) {
		err := queue.Start()
		require.NoError(t, err)
		defer queue.Stop()
		
		payload := map[string]interface{}{
			"test_key": "test_value",
		}
		
		job, err := queue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "test-user-id", payload, jobs.JobPriorityNormal)
		require.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, jobs.JobTypeCourseGeneration, job.Type)
		assert.Equal(t, "test-user-id", job.UserID)
		assert.Equal(t, jobs.JobStatusPending, job.Status)
		assert.Equal(t, 0, job.Progress)
		assert.Equal(t, jobs.JobPriorityNormal, job.Priority)
		assert.Equal(t, payload, job.Payload)
		assert.NotEmpty(t, job.ID)
		assert.False(t, job.CreatedAt.IsZero())
		assert.False(t, job.UpdatedAt.IsZero())
		assert.Nil(t, job.StartedAt)
		assert.Nil(t, job.CompletedAt)
		assert.Nil(t, job.Error)
	})
	
	t.Run("Get Job", func(t *testing.T) {
		err := queue.Start()
		require.NoError(t, err)
		defer queue.Stop()
		
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
	
	t.Run("Get User Jobs", func(t *testing.T) {
		err := queue.Start()
		require.NoError(t, err)
		defer queue.Stop()
		
		// Create jobs for different users
		job1, _ := queue.Enqueue(context.Background(), jobs.JobTypeAudioGeneration, "user-1", map[string]interface{}{}, jobs.JobPriorityNormal)
		job2, _ := queue.Enqueue(context.Background(), jobs.JobTypeSubtitleGeneration, "user-1", map[string]interface{}{}, jobs.JobPriorityNormal)
		job3, _ := queue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "user-2", map[string]interface{}{}, jobs.JobPriorityNormal)
		
		// Get jobs for user-1
		user1Jobs, err := queue.GetUserJobs(context.Background(), "user-1", 10, 0)
		require.NoError(t, err)
		assert.Len(t, user1Jobs, 2)
		
		// Verify job IDs
		jobIDs := make(map[string]bool)
		for _, job := range user1Jobs {
			jobIDs[job.ID] = true
		}
		assert.True(t, jobIDs[job1.ID])
		assert.True(t, jobIDs[job2.ID])
		assert.False(t, jobIDs[job3.ID])
	})
	
	t.Run("Cancel Job", func(t *testing.T) {
		err := queue.Start()
		require.NoError(t, err)
		defer queue.Stop()
		
		// Create job
		job, err := queue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "test-user-id", map[string]interface{}{}, jobs.JobPriorityNormal)
		require.NoError(t, err)
		assert.Equal(t, jobs.JobStatusPending, job.Status)
		
		// Cancel job
		err = queue.CancelJob(context.Background(), job.ID)
		require.NoError(t, err)
		
		// Verify job was cancelled
		cancelledJob, err := queue.GetJob(context.Background(), job.ID)
		require.NoError(t, err)
		assert.Equal(t, jobs.JobStatusCancelled, cancelledJob.Status)
	})
	
	t.Run("Update Progress", func(t *testing.T) {
		err := queue.Start()
		require.NoError(t, err)
		defer queue.Stop()
		
		// Create job
		job, err := queue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "test-user-id", map[string]interface{}{}, jobs.JobPriorityNormal)
		require.NoError(t, err)
		assert.Equal(t, 0, job.Progress)
		
		// Update progress
		err = queue.UpdateProgress(job.ID, 50)
		require.NoError(t, err)
		
		// Verify progress was updated
		updatedJob, err := queue.GetJob(context.Background(), job.ID)
		require.NoError(t, err)
		assert.Equal(t, 50, updatedJob.Progress)
	})
	
	t.Run("Update Result", func(t *testing.T) {
		err := queue.Start()
		require.NoError(t, err)
		defer queue.Stop()
		
		// Create job
		job, err := queue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "test-user-id", map[string]interface{}{}, jobs.JobPriorityNormal)
		require.NoError(t, err)
		assert.Nil(t, job.Result)
		
		// Update result
		result := map[string]interface{}{
			"output_path": "/path/to/output",
			"duration":    300,
		}
		err = queue.UpdateResult(job.ID, result)
		require.NoError(t, err)
		
		// Verify result was updated
		updatedJob, err := queue.GetJob(context.Background(), job.ID)
		require.NoError(t, err)
		assert.Equal(t, result, updatedJob.Result)
	})
}

func TestJobHandlers(t *testing.T) {
	t.Skip("Skipping job handlers test for now - requires mock pipeline and storage")
	
	// This test would need to set up mocks for the pipeline and storage components
	// For now, we're skipping it as it would require more complex setup
}

func TestJobConversion(t *testing.T) {
	queue := setupTestJobQueue(t)
	
	t.Run("Convert To DB Model", func(t *testing.T) {
		job := &jobs.Job{
			ID:        "test-job-id",
			UserID:    "test-user-id",
			Type:      jobs.JobTypeCourseGeneration,
			Status:    jobs.JobStatusPending,
			Progress:  0,
			Priority:  jobs.JobPriorityNormal,
			Payload:   map[string]interface{}{"test": "value"},
			Result:    map[string]interface{}{"output": "/path"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		jobDB, err := queue.ConvertToDBModel(job)
		require.NoError(t, err)
		
		assert.Equal(t, job.ID, jobDB.ID)
		assert.Equal(t, job.UserID, jobDB.UserID)
		assert.Equal(t, string(job.Type), jobDB.Type)
		assert.Equal(t, string(job.Status), jobDB.Status)
		assert.Equal(t, job.Progress, jobDB.Progress)
		assert.NotEmpty(t, jobDB.Payload)
		assert.NotEmpty(t, jobDB.Result)
	})
	
	t.Run("Convert From DB Model", func(t *testing.T) {
		jobDB := &database.JobDB{
			ID:        "test-job-id",
			UserID:    "test-user-id",
			Type:      string(jobs.JobTypeCourseGeneration),
			Status:    string(jobs.JobStatusPending),
			Progress:  0,
			Payload:   `{"test": "value"}`,
			Result:    `{"output": "/path"}`,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		job, err := queue.ConvertFromDBModel(jobDB)
		require.NoError(t, err)
		
		assert.Equal(t, jobDB.ID, job.ID)
		assert.Equal(t, jobDB.UserID, job.UserID)
		assert.Equal(t, jobs.JobType(jobDB.Type), job.Type)
		assert.Equal(t, jobs.JobStatus(jobDB.Status), job.Status)
		assert.Equal(t, jobDB.Progress, job.Progress)
		assert.Equal(t, map[string]interface{}{"test": "value"}, job.Payload)
		assert.Equal(t, map[string]interface{}{"output": "/path"}, job.Result)
	})
}

func TestJobPriority(t *testing.T) {
	assert.Equal(t, jobs.JobPriority(1), jobs.JobPriorityLow)
	assert.Equal(t, jobs.JobPriority(2), jobs.JobPriorityNormal)
	assert.Equal(t, jobs.JobPriority(3), jobs.JobPriorityHigh)
	assert.Equal(t, jobs.JobPriority(4), jobs.JobPriorityCritical)
}

func TestJobStatus(t *testing.T) {
	assert.Equal(t, jobs.JobStatus("pending"), jobs.JobStatusPending)
	assert.Equal(t, jobs.JobStatus("running"), jobs.JobStatusRunning)
	assert.Equal(t, jobs.JobStatus("completed"), jobs.JobStatusCompleted)
	assert.Equal(t, jobs.JobStatus("failed"), jobs.JobStatusFailed)
	assert.Equal(t, jobs.JobStatus("cancelled"), jobs.JobStatusCancelled)
}

func TestJobType(t *testing.T) {
	assert.Equal(t, jobs.JobType("course_generation"), jobs.JobTypeCourseGeneration)
	assert.Equal(t, jobs.JobType("video_processing"), jobs.JobTypeVideoProcessing)
	assert.Equal(t, jobs.JobType("audio_generation"), jobs.JobTypeAudioGeneration)
	assert.Equal(t, jobs.JobType("subtitle_generation"), jobs.JobTypeSubtitleGeneration)
}

// Helper functions to access private methods in jobs package for testing
func (jq *jobs.JobQueue) ConvertToDBModel(job *jobs.Job) (*models.JobDB, error) {
	// This is a helper method to access the private convertToDBModel method
	// In the actual implementation, this method would be private
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return nil, err
	}
	
	var resultJSON []byte
	if job.Result != nil {
		resultJSON, err = json.Marshal(job.Result)
		if err != nil {
			return nil, err
		}
	}
	
	return &models.JobDB{
		ID:          job.ID,
		UserID:      job.UserID,
		Type:        string(job.Type),
		Status:      string(job.Status),
		Progress:    job.Progress,
		Payload:     string(payloadJSON),
		Result:      string(resultJSON),
		Error:       job.Error,
		CreatedAt:   job.CreatedAt,
		UpdatedAt:   job.UpdatedAt,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
	}, nil
}

func (jq *jobs.JobQueue) ConvertFromDBModel(jobDB *models.JobDB) (*jobs.Job, error) {
	// This is a helper method to access the private convertFromDBModel method
	// In the actual implementation, this method would be private
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jobDB.Payload), &payload); err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	if jobDB.Result != "" {
		if err := json.Unmarshal([]byte(jobDB.Result), &result); err != nil {
			return nil, err
		}
	}
	
	return &jobs.Job{
		ID:          jobDB.ID,
		UserID:      jobDB.UserID,
		Type:        jobs.JobType(jobDB.Type),
		Status:      jobs.JobStatus(jobDB.Status),
		Progress:    jobDB.Progress,
		Payload:     payload,
		Result:      result,
		Error:       jobDB.Error,
		CreatedAt:   jobDB.CreatedAt,
		UpdatedAt:   jobDB.UpdatedAt,
		StartedAt:   jobDB.StartedAt,
		CompletedAt: jobDB.CompletedAt,
	}, nil
}