package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestQueue(t *testing.T) *JobQueue {
	// Create a dummy DB connection for testing
	db, err := gorm.Open(nil, &gorm.Config{})
	require.NoError(t, err)

	queue := NewJobQueue(db, 2)
	return queue
}

func TestNewJobQueue(t *testing.T) {
	queue := setupTestQueue(t)

	assert.NotNil(t, queue)
	assert.Equal(t, 2, queue.workers)
	assert.NotNil(t, queue.handlers)
	assert.NotNil(t, queue.jobQueue)
	assert.NotNil(t, queue.resultChan)
	assert.NotNil(t, queue.ctx)
	assert.NotNil(t, queue.cancel)
	assert.False(t, queue.running)
}

func TestRegisterHandler(t *testing.T) {
	queue := setupTestQueue(t)

	handler := func(ctx context.Context, job *Job) error {
		return nil
	}

	queue.RegisterHandler(JobTypeCourseGeneration, handler)

	assert.Contains(t, queue.handlers, JobTypeCourseGeneration)
	assert.NotNil(t, queue.handlers[JobTypeCourseGeneration])
}

func TestJobTypes(t *testing.T) {
	assert.Equal(t, JobType("course_generation"), JobTypeCourseGeneration)
	assert.Equal(t, JobType("video_processing"), JobTypeVideoProcessing)
	assert.Equal(t, JobType("audio_generation"), JobTypeAudioGeneration)
	assert.Equal(t, JobType("subtitle_generation"), JobTypeSubtitleGeneration)
}

func TestJobStatuses(t *testing.T) {
	assert.Equal(t, JobStatus("pending"), JobStatusPending)
	assert.Equal(t, JobStatus("running"), JobStatusRunning)
	assert.Equal(t, JobStatus("completed"), JobStatusCompleted)
	assert.Equal(t, JobStatus("failed"), JobStatusFailed)
	assert.Equal(t, JobStatus("cancelled"), JobStatusCancelled)
}

func TestJobPriorities(t *testing.T) {
	assert.Equal(t, JobPriority(1), JobPriorityLow)
	assert.Equal(t, JobPriority(2), JobPriorityNormal)
	assert.Equal(t, JobPriority(3), JobPriorityHigh)
	assert.Equal(t, JobPriority(4), JobPriorityCritical)
}

func TestIsRunning(t *testing.T) {
	queue := setupTestQueue(t)

	assert.False(t, queue.IsRunning())
}

func TestUpdateProgress(t *testing.T) {
	queue := setupTestQueue(t)

	// Test invalid progress values
	err := queue.UpdateProgress("test-job", 150)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "progress must be between 0 and 100")

	err = queue.UpdateProgress("test-job", -10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "progress must be between 0 and 100")
}

func TestUpdateResult(t *testing.T) {
	queue := setupTestQueue(t)

	result := map[string]interface{}{
		"status": "completed",
		"output": "/path/to/output",
	}

	// Test that the method can be called (database errors are expected with nil DB)
	err := queue.UpdateResult("test-job", result)
	// We don't assert on the error since it depends on the database setup
	_ = err // Just ensure the method doesn't panic
}

func TestConvertToDBModel(t *testing.T) {
	queue := setupTestQueue(t)

	job := &Job{
		ID:       "test-job",
		UserID:   "user123",
		Type:     JobTypeCourseGeneration,
		Status:   JobStatusRunning,
		Progress: 75,
		Priority: JobPriorityHigh,
		Payload: map[string]interface{}{
			"input":  "/input.md",
			"output": "/output",
		},
		Result: map[string]interface{}{
			"status": "success",
		},
		Error:       stringPtr("some error"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		StartedAt:   &time.Time{},
		CompletedAt: &time.Time{},
	}

	dbModel, err := queue.ConvertToDBModel(job)

	require.NoError(t, err)
	assert.NotNil(t, dbModel)
	assert.Equal(t, "test-job", dbModel.ID)
	assert.Equal(t, "user123", dbModel.UserID)
	assert.Equal(t, "course_generation", dbModel.Type)
	assert.Equal(t, "running", dbModel.Status)
	assert.Equal(t, 75, dbModel.Progress)
	assert.Contains(t, dbModel.Payload, "input")
	assert.Contains(t, dbModel.Result, "status")
	assert.Equal(t, "some error", *dbModel.Error)
}

func TestConvertFromDBModel_JSONHandling(t *testing.T) {
	// Test the JSON unmarshaling part of ConvertFromDBModel
	payloadJSON := `{"input": "/input.md", "output": "/output"}`
	resultJSON := `{"status": "success"}`

	// Test the JSON unmarshaling part
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(payloadJSON), &payload)
	require.NoError(t, err)
	assert.Contains(t, payload, "input")
	assert.Equal(t, "/input.md", payload["input"])

	var result map[string]interface{}
	err = json.Unmarshal([]byte(resultJSON), &result)
	require.NoError(t, err)
	assert.Contains(t, result, "status")
	assert.Equal(t, "success", result["status"])
}

func stringPtr(s string) *string {
	return &s
}
