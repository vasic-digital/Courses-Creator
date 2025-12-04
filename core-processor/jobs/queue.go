package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/course-creator/core-processor/metrics"
	"github.com/course-creator/core-processor/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeCourseGeneration   JobType = "course_generation"
	JobTypeVideoProcessing    JobType = "video_processing"
	JobTypeAudioGeneration    JobType = "audio_generation"
	JobTypeSubtitleGeneration JobType = "subtitle_generation"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobPriority represents the priority of a job
type JobPriority int

const (
	JobPriorityLow      JobPriority = 1
	JobPriorityNormal   JobPriority = 2
	JobPriorityHigh     JobPriority = 3
	JobPriorityCritical JobPriority = 4
)

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	Progress    int                    `json:"progress"` // 0-100
	Priority    JobPriority            `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       *string                `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job *Job) error

// JobQueue represents the job queue system
type JobQueue struct {
	db         *gorm.DB
	handlers   map[JobType]JobHandler
	workers    int
	jobQueue   chan *Job
	resultChan chan *Job
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool
}

// NewJobQueue creates a new job queue
func NewJobQueue(db *gorm.DB, workers int) *JobQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &JobQueue{
		db:         db,
		handlers:   make(map[JobType]JobHandler),
		workers:    workers,
		jobQueue:   make(chan *Job, 1000),
		resultChan: make(chan *Job, 100),
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
	}
}

// RegisterHandler registers a job handler for a specific job type
func (jq *JobQueue) RegisterHandler(jobType JobType, handler JobHandler) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	jq.handlers[jobType] = handler
}

// Start starts the job queue workers
func (jq *JobQueue) Start() error {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	if jq.running {
		return fmt.Errorf("job queue is already running")
	}

	jq.running = true

	// Start worker goroutines
	for i := 0; i < jq.workers; i++ {
		jq.wg.Add(1)
		go jq.worker(i)
	}

	// Start result processor
	jq.wg.Add(1)
	go jq.resultProcessor()

	// Load pending jobs from database
	if err := jq.loadPendingJobs(); err != nil {
		log.Printf("Warning: failed to load pending jobs: %v", err)
	}

	log.Printf("Job queue started with %d workers", jq.workers)
	return nil
}

// Stop stops the job queue
func (jq *JobQueue) Stop() {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	if !jq.running {
		return
	}

	jq.running = false
	jq.cancel()
	jq.wg.Wait()

	log.Println("Job queue stopped")
}

// IsRunning returns whether the job queue is currently running
func (jq *JobQueue) IsRunning() bool {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	return jq.running
}

// Enqueue adds a new job to the queue
func (jq *JobQueue) Enqueue(ctx context.Context, jobType JobType, userID string, payload map[string]interface{}, priority JobPriority) (*Job, error) {
	job := &Job{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      jobType,
		Status:    JobStatusPending,
		Progress:  0,
		Priority:  priority,
		Payload:   payload,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := jq.saveJob(job); err != nil {
		return nil, fmt.Errorf("failed to save job to database: %w", err)
	}

	// Add to queue
	select {
	case jq.jobQueue <- job:
		return job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Queue is full, return error
		return nil, fmt.Errorf("job queue is full")
	}
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	var jobDB models.JobDB
	if err := jq.db.Where("id = ?", jobID).First(&jobDB).Error; err != nil {
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	return jq.ConvertFromDBModel(&jobDB)
}

// GetUserJobs retrieves jobs for a specific user
func (jq *JobQueue) GetUserJobs(ctx context.Context, userID string, limit, offset int) ([]*Job, error) {
	var jobsDB []models.JobDB
	if err := jq.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobsDB).Error; err != nil {
		return nil, fmt.Errorf("failed to find user jobs: %w", err)
	}

	jobs := make([]*Job, len(jobsDB))
	for i, jobDB := range jobsDB {
		job, err := jq.ConvertFromDBModel(&jobDB)
		if err != nil {
			return nil, fmt.Errorf("failed to convert job %d: %w", i, err)
		}
		jobs[i] = job
	}

	return jobs, nil
}

// CancelJob cancels a job
func (jq *JobQueue) CancelJob(ctx context.Context, jobID string) error {
	job, err := jq.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status != JobStatusPending && job.Status != JobStatusRunning {
		return fmt.Errorf("cannot cancel job in %s status", job.Status)
	}

	job.Status = JobStatusCancelled
	job.UpdatedAt = time.Now()

	if err := jq.saveJob(job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// worker processes jobs from the queue
func (jq *JobQueue) worker(id int) {
	defer jq.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case <-jq.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		case job := <-jq.jobQueue:
			jq.processJob(job)
		}
	}
}

// processJob processes a single job
func (jq *JobQueue) processJob(job *Job) {
	log.Printf("Processing job %s of type %s for user %s", job.ID, job.Type, job.UserID)
	startTime := time.Now()

	// Update job status to running
	job.Status = JobStatusRunning
	now := time.Now()
	job.StartedAt = &now
	job.UpdatedAt = now

	if err := jq.saveJob(job); err != nil {
		log.Printf("Failed to update job status to running: %v", err)
		return
	}

	// Get handler for job type
	jq.mu.Lock()
	handler, exists := jq.handlers[job.Type]
	jq.mu.Unlock()

	if !exists {
		errorMsg := fmt.Sprintf("no handler registered for job type: %s", job.Type)
		jq.markJobFailed(job, errorMsg)
		metrics.RecordJobCompletion("failed", string(job.Type), time.Since(startTime))
		return
	}

	// Execute job handler
	err := handler(jq.ctx, job)
	duration := time.Since(startTime)

	// Send job to result processor for final update
	select {
	case jq.resultChan <- job:
	default:
		log.Printf("Result channel is full, dropping result for job %s", job.ID)
	}

	if err != nil {
		jq.markJobFailed(job, err.Error())
		metrics.RecordJobCompletion("failed", string(job.Type), duration)
	} else {
		jq.markJobCompleted(job)
		metrics.RecordJobCompletion("completed", string(job.Type), duration)
	}
}

// resultProcessor processes job results
func (jq *JobQueue) resultProcessor() {
	defer jq.wg.Done()

	for {
		select {
		case <-jq.ctx.Done():
			return
		case job := <-jq.resultChan:
			// Final job processing is already done in processJob
			// This just ensures final database save
			if err := jq.saveJob(job); err != nil {
				log.Printf("Failed to save final job state: %v", err)
			}
		}
	}
}

// markJobFailed marks a job as failed
func (jq *JobQueue) markJobFailed(job *Job, errorMsg string) {
	job.Status = JobStatusFailed
	job.Error = &errorMsg
	job.UpdatedAt = time.Now()

	now := time.Now()
	job.CompletedAt = &now

	if err := jq.saveJob(job); err != nil {
		log.Printf("Failed to mark job %s as failed: %v", job.ID, err)
	}
}

// markJobCompleted marks a job as completed
func (jq *JobQueue) markJobCompleted(job *Job) {
	job.Status = JobStatusCompleted
	job.Progress = 100
	job.UpdatedAt = time.Now()

	now := time.Now()
	job.CompletedAt = &now

	if err := jq.saveJob(job); err != nil {
		log.Printf("Failed to mark job %s as completed: %v", job.ID, err)
	}
}

// saveJob saves a job to the database
func (jq *JobQueue) saveJob(job *Job) error {
	jobDB, err := jq.ConvertToDBModel(job)
	if err != nil {
		return fmt.Errorf("failed to convert job to DB model: %w", err)
	}

	return jq.db.Save(jobDB).Error
}

// loadPendingJobs loads pending jobs from database and adds them to the queue
func (jq *JobQueue) loadPendingJobs() error {
	var jobsDB []models.JobDB
	if err := jq.db.Where("status = ?", JobStatusPending).
		Order("created_at ASC").
		Find(&jobsDB).Error; err != nil {
		return fmt.Errorf("failed to load pending jobs: %w", err)
	}

	for _, jobDB := range jobsDB {
		job, err := jq.ConvertFromDBModel(&jobDB)
		if err != nil {
			log.Printf("Failed to convert pending job %s: %v", jobDB.ID, err)
			continue
		}

		select {
		case jq.jobQueue <- job:
			log.Printf("Loaded pending job %s", job.ID)
		default:
			log.Printf("Queue is full, skipping pending job %s", job.ID)
		}
	}

	return nil
}

// ConvertToDBModel converts a Job to models.JobDB
func (jq *JobQueue) ConvertToDBModel(job *Job) (*models.JobDB, error) {
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var resultJSON []byte
	if job.Result != nil {
		resultJSON, err = json.Marshal(job.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
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

// ConvertFromDBModel converts models.JobDB to Job
func (jq *JobQueue) ConvertFromDBModel(jobDB *models.JobDB) (*Job, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jobDB.Payload), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var result map[string]interface{}
	if jobDB.Result != "" {
		if err := json.Unmarshal([]byte(jobDB.Result), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return &Job{
		ID:          jobDB.ID,
		UserID:      jobDB.UserID,
		Type:        JobType(jobDB.Type),
		Status:      JobStatus(jobDB.Status),
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

// UpdateProgress updates job progress
func (jq *JobQueue) UpdateProgress(jobID string, progress int) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}

	return jq.db.Model(&models.JobDB{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"progress":   progress,
			"updated_at": time.Now(),
		}).Error
}

// UpdateResult updates job result
func (jq *JobQueue) UpdateResult(jobID string, result map[string]interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	return jq.db.Model(&models.JobDB{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"result":     string(resultJSON),
			"updated_at": time.Now(),
		}).Error
}
