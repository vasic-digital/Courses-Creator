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

func setupJobTestQueue(t *testing.T) *jobs.JobQueue {
	// This function is no longer needed since tests use isolated databases
	// but keeping for reference
	db := setupTestDB(t)
	queue := jobs.NewJobQueue(db, 2)

	// Register dummy handlers for all job types to prevent hanging
	queue.RegisterHandler(jobs.JobTypeCourseGeneration, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})
	queue.RegisterHandler(jobs.JobTypeVideoProcessing, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})
	queue.RegisterHandler(jobs.JobTypeAudioGeneration, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})
	queue.RegisterHandler(jobs.JobTypeSubtitleGeneration, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})

	// Create test users to satisfy foreign key constraints
	testUser1 := &models.UserDB{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "creator",
	}
	testUser2 := &models.UserDB{
		ID:    "user-1",
		Email: "user1@example.com",
		Role:  "creator",
	}
	testUser3 := &models.UserDB{
		ID:    "user-2",
		Email: "user2@example.com",
		Role:  "creator",
	}

	require.NoError(t, db.Create(testUser1).Error)
	require.NoError(t, db.Create(testUser2).Error)
	require.NoError(t, db.Create(testUser3).Error)

	return queue
}

// registerAllHandlers registers dummy handlers for all job types
func registerAllHandlers(queue *jobs.JobQueue) {
	queue.RegisterHandler(jobs.JobTypeCourseGeneration, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})
	queue.RegisterHandler(jobs.JobTypeVideoProcessing, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})
	queue.RegisterHandler(jobs.JobTypeAudioGeneration, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})
	queue.RegisterHandler(jobs.JobTypeSubtitleGeneration, func(ctx context.Context, job *jobs.Job) error {
		return nil
	})
}

func TestJobQueue(t *testing.T) {
	t.Run("Start and Stop Queue", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)

		// Start queue
		err := queue.Start()
		require.NoError(t, err)

		// Stop queue
		queue.Stop()
	})

	t.Run("Enqueue Job", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)
		// Don't start the queue - just test enqueueing

		// Create user
		testUser := &models.UserDB{
			ID:    "test-user-id",
			Email: "test@example.com",
			Role:  "creator",
		}
		require.NoError(t, db.Create(testUser).Error)

		// Use a timeout context to avoid hanging
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		payload := map[string]interface{}{
			"test_key": "test_value",
		}

		job, err := queue.Enqueue(ctx, jobs.JobTypeCourseGeneration, "test-user-id", payload, jobs.JobPriorityNormal)
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
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)
		// Don't start the queue - just test basic operations

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

	t.Run("Get User Jobs", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)
		// Don't start the queue - just test basic operations

		// Create users
		user1 := &models.UserDB{
			ID:    "user-1",
			Email: "user1@example.com",
			Role:  "creator",
		}
		user2 := &models.UserDB{
			ID:    "user-2",
			Email: "user2@example.com",
			Role:  "creator",
		}
		require.NoError(t, db.Create(user1).Error)
		require.NoError(t, db.Create(user2).Error)

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
		assert.Equal(t, result["output_path"], updatedJob.Result["output_path"])
		assert.Equal(t, float64(result["duration"].(int)), updatedJob.Result["duration"])
	})
}

func TestJobHandlers(t *testing.T) {
	t.Skip("Skipping job handlers test for now - requires mock pipeline and storage")

	// This test would need to set up mocks for the pipeline and storage components
	// For now, we're skipping it as it would require more complex setup
}

func TestJobConversion(t *testing.T) {
	// Create fresh queue to isolate test
	db := setupTestDB(t)
	queue := jobs.NewJobQueue(db, 2)
	defer queue.Stop()
	registerAllHandlers(queue)

	// Create test user to satisfy foreign key constraints
	testUser := &models.UserDB{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "creator",
	}
	require.NoError(t, db.Create(testUser).Error)

	t.Run("Convert To DB Model", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)

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

		// Test conversion using the public method
		jobDB, err := queue.ConvertToDBModel(job)
		require.NoError(t, err)
		require.NotNil(t, jobDB)

		assert.Equal(t, job.ID, jobDB.ID)
		assert.Equal(t, job.UserID, jobDB.UserID)
		assert.Equal(t, string(job.Type), jobDB.Type)
		assert.Equal(t, string(job.Status), jobDB.Status)
		assert.Equal(t, job.Progress, jobDB.Progress)
		assert.NotEmpty(t, jobDB.Payload)
		assert.NotEmpty(t, jobDB.Result)
	})

	t.Run("Convert From DB Model", func(t *testing.T) {
		// Create fresh queue to isolate test
		db := setupTestDB(t)
		queue := jobs.NewJobQueue(db, 2)

		jobDB := &models.JobDB{
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

		// Test conversion using the public method
		job, err := queue.ConvertFromDBModel(jobDB)
		require.NoError(t, err)
		require.NotNil(t, job)

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
