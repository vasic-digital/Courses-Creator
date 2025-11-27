package unit

import (
	"encoding/json"
	"testing"

	"github.com/course-creator/core-processor/jobs"
	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestJobDBSimple(t *testing.T) {
	// Create database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Configure SQLite
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	
	// Migrate
	err = db.AutoMigrate(&models.UserDB{}, &models.JobDB{})
	require.NoError(t, err)
	
	// Create user
	user := &models.UserDB{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  "creator",
	}
	err = db.Create(user).Error
	require.NoError(t, err)
	
	// Create job using the same DB connection
	job := &jobs.Job{
		ID:        "test-job-id",
		UserID:    "test-user-id",
		Type:      jobs.JobTypeCourseGeneration,
		Status:    jobs.JobStatusPending,
		Progress:  0,
		Priority:  jobs.JobPriorityNormal,
		Payload:   map[string]interface{}{"test": "value"},
		CreatedAt: db.NowFunc(),
		UpdatedAt: db.NowFunc(),
	}
	
	// Convert to DB model manually
	payloadJSON, err := json.Marshal(job.Payload)
	require.NoError(t, err)
	
	jobDB := &models.JobDB{
		ID:        job.ID,
		UserID:    job.UserID,
		Type:      string(job.Type),
		Status:    string(job.Status),
		Progress:  job.Progress,
		Payload:   string(payloadJSON),
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}
	
	// Save directly
	err = db.Save(jobDB).Error
	require.NoError(t, err)
	
	// Retrieve
	var retrieved models.JobDB
	err = db.Where("id = ?", job.ID).First(&retrieved).Error
	require.NoError(t, err)
	
	require.Equal(t, job.ID, retrieved.ID)
	require.Equal(t, job.UserID, retrieved.UserID)
}