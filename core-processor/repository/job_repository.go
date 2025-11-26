package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
)

// ProcessingJobRepository handles processing job data operations
type ProcessingJobRepository struct {
	db *database.DB
}

// NewProcessingJobRepository creates a new processing job repository
func NewProcessingJobRepository(db *database.DB) *ProcessingJobRepository {
	return &ProcessingJobRepository{db: db}
}

// CreateJob creates a new processing job
func (r *ProcessingJobRepository) CreateJob(job *models.ProcessingJobDB) (*models.ProcessingJobDB, error) {
	if err := r.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to create processing job: %w", err)
	}
	return r.GetJobByID(job.ID)
}

// GetJobByID retrieves a processing job by ID
func (r *ProcessingJobRepository) GetJobByID(id string) (*models.ProcessingJobDB, error) {
	var job models.ProcessingJobDB
	if err := r.db.First(&job, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get processing job: %w", err)
	}
	return &job, nil
}

// GetAllJobs retrieves all processing jobs with pagination
func (r *ProcessingJobRepository) GetAllJobs(offset, limit int) ([]models.ProcessingJobDB, int64, error) {
	var jobs []models.ProcessingJobDB
	var total int64

	// Count total jobs
	if err := r.db.Model(&models.ProcessingJobDB{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count processing jobs: %w", err)
	}

	// Get jobs with pagination (ordered by created_at desc)
	if err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get processing jobs: %w", err)
	}

	return jobs, total, nil
}

// GetJobsByStatus retrieves processing jobs by status
func (r *ProcessingJobRepository) GetJobsByStatus(status string, offset, limit int) ([]models.ProcessingJobDB, int64, error) {
	var jobs []models.ProcessingJobDB
	var total int64

	// Count jobs by status
	if err := r.db.Model(&models.ProcessingJobDB{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count processing jobs: %w", err)
	}

	// Get jobs by status with pagination
	if err := r.db.Where("status = ?", status).Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get processing jobs: %w", err)
	}

	return jobs, total, nil
}

// GetJobsByCourseID retrieves processing jobs for a specific course
func (r *ProcessingJobRepository) GetJobsByCourseID(courseID string, offset, limit int) ([]models.ProcessingJobDB, int64, error) {
	var jobs []models.ProcessingJobDB
	var total int64

	// Count jobs by course ID
	if err := r.db.Model(&models.ProcessingJobDB{}).Where("course_id = ?", courseID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count processing jobs: %w", err)
	}

	// Get jobs by course ID with pagination
	if err := r.db.Where("course_id = ?", courseID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get processing jobs: %w", err)
	}

	return jobs, total, nil
}

// UpdateJob updates a processing job
func (r *ProcessingJobRepository) UpdateJob(id string, updates map[string]interface{}) (*models.ProcessingJobDB, error) {
	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	if err := r.db.Model(&models.ProcessingJobDB{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update processing job: %w", err)
	}
	return r.GetJobByID(id)
}

// UpdateJobStatus updates the status of a processing job
func (r *ProcessingJobRepository) UpdateJobStatus(id, status string) (*models.ProcessingJobDB, error) {
	updates := map[string]interface{}{"status": status}
	
	// Add timestamps based on status
	now := time.Now()
	switch status {
	case "running":
		updates["started_at"] = &now
	case "completed", "failed":
		updates["completed_at"] = &now
	}

	return r.UpdateJob(id, updates)
}

// UpdateJobProgress updates the progress of a processing job
func (r *ProcessingJobRepository) UpdateJobProgress(id string, progress int) (*models.ProcessingJobDB, error) {
	updates := map[string]interface{}{
		"progress": progress,
	}
	return r.UpdateJob(id, updates)
}

// UpdateJobError updates the error message of a processing job
func (r *ProcessingJobRepository) UpdateJobError(id, errorMessage string) (*models.ProcessingJobDB, error) {
	updates := map[string]interface{}{
		"status":    "failed",
		"error":     &errorMessage,
		"progress":  0,
	}
	return r.UpdateJob(id, updates)
}

// SetJobCourseID sets the course ID for a processing job
func (r *ProcessingJobRepository) SetJobCourseID(jobID, courseID string) (*models.ProcessingJobDB, error) {
	updates := map[string]interface{}{
		"course_id": &courseID,
	}
	return r.UpdateJob(jobID, updates)
}

// DeleteJob deletes a processing job by ID
func (r *ProcessingJobRepository) DeleteJob(id string) error {
	if err := r.db.Delete(&models.ProcessingJobDB{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete processing job: %w", err)
	}
	return nil
}

// GetPendingJobs retrieves jobs that are pending
func (r *ProcessingJobRepository) GetPendingJobs() ([]models.ProcessingJobDB, error) {
	var jobs []models.ProcessingJobDB
	if err := r.db.Where("status = ?", "pending").Order("created_at ASC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	return jobs, nil
}

// GetRunningJobs retrieves jobs that are currently running
func (r *ProcessingJobRepository) GetRunningJobs() ([]models.ProcessingJobDB, error) {
	var jobs []models.ProcessingJobDB
	if err := r.db.Where("status = ?", "running").Order("started_at ASC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to get running jobs: %w", err)
	}
	return jobs, nil
}

// JobOptions represents the options stored in JSON format
type JobOptions struct {
	Voice           *string  `json:"voice,omitempty"`
	BackgroundMusic bool     `json:"background_music"`
	Languages       []string `json:"languages"`
	Quality         string   `json:"quality"`
}

// ParseJobOptions parses the options JSON from a processing job
func ParseJobOptions(optionsJSON string) (*JobOptions, error) {
	if optionsJSON == "" {
		return &JobOptions{
			BackgroundMusic: false,
			Languages:       []string{"en"},
			Quality:         "standard",
		}, nil
	}
	
	var options JobOptions
	if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
		return nil, fmt.Errorf("failed to parse job options: %w", err)
	}
	return &options, nil
}

// SerializeJobOptions serializes job options to JSON
func SerializeJobOptions(options *JobOptions) (string, error) {
	if options == nil {
		return "", nil
	}
	
	data, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("failed to serialize job options: %w", err)
	}
	return string(data), nil
}