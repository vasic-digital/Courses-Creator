package api

import (
	"net/http"
	"strconv"

	"github.com/course-creator/core-processor/jobs"
	"github.com/gin-gonic/gin"
)

// JobHandler handles job-related endpoints
type JobHandler struct {
	jobQueue *jobs.JobQueue
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobQueue *jobs.JobQueue) *JobHandler {
	return &JobHandler{
		jobQueue: jobQueue,
	}
}

// CreateJob creates a new job
// @Summary Create a new job
// @Description Create and enqueue a new background job
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Job creation request"
// @Success 201 {object} jobs.Job
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/jobs [post]
func (h *JobHandler) CreateJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req struct {
		Type     string                 `json:"type" binding:"required"`
		Payload  map[string]interface{} `json:"payload" binding:"required"`
		Priority int                    `json:"priority,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate job type
	var jobType jobs.JobType
	switch req.Type {
	case "course_generation":
		jobType = jobs.JobTypeCourseGeneration
	case "video_processing":
		jobType = jobs.JobTypeVideoProcessing
	case "audio_generation":
		jobType = jobs.JobTypeAudioGeneration
	case "subtitle_generation":
		jobType = jobs.JobTypeSubtitleGeneration
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Invalid job type",
			"valid_types": []string{"course_generation", "video_processing", "audio_generation", "subtitle_generation"},
		})
		return
	}

	// Set default priority if not provided
	priority := jobs.JobPriorityNormal
	if req.Priority > 0 {
		priority = jobs.JobPriority(req.Priority)
		if priority < jobs.JobPriorityLow || priority > jobs.JobPriorityCritical {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid priority value (must be 1-4)",
			})
			return
		}
	}

	// Enqueue job
	job, err := h.jobQueue.Enqueue(c.Request.Context(), jobType, userID.(string), req.Payload, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// GetJob retrieves a job by ID
// @Summary Get a job
// @Description Retrieve a specific job by ID
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} jobs.Job
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/jobs/{id} [get]
func (h *JobHandler) GetJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	job, err := h.jobQueue.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Job not found",
			"details": err.Error(),
		})
		return
	}

	// Check if user owns this job or is admin
	userRole, _ := c.Get("user_role")
	userID, _ := c.Get("user_id")

	if job.UserID != userID.(string) && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied to this job",
		})
		return
	}

	c.JSON(http.StatusOK, job)
}

// GetUserJobs retrieves jobs for the current user
// @Summary Get user jobs
// @Description Retrieve jobs for the current authenticated user
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} jobs.Job
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/jobs [get]
func (h *JobHandler) GetUserJobs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse query parameters
	limit := 20 // Default
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := 0 // Default
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	jobs, err := h.jobQueue.GetUserJobs(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve jobs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, jobs)
}

// CancelJob cancels a job
// @Summary Cancel a job
// @Description Cancel a pending or running job
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/jobs/{id}/cancel [post]
func (h *JobHandler) CancelJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	// Get job to check ownership
	job, err := h.jobQueue.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Job not found",
			"details": err.Error(),
		})
		return
	}

	// Check if user owns this job or is admin
	userRole, _ := c.Get("user_role")
	userID, _ := c.Get("user_id")

	if job.UserID != userID.(string) && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied to this job",
		})
		return
	}

	// Cancel job
	if err := h.jobQueue.CancelJob(c.Request.Context(), jobID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to cancel job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job cancelled successfully",
	})
}

// UpdateJobProgress updates job progress (for internal use by workers)
// @Summary Update job progress
// @Description Update job progress (internal endpoint for workers)
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body map[string]int true "Progress update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/jobs/{id}/progress [put]
func (h *JobHandler) UpdateJobProgress(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	var req struct {
		Progress int `json:"progress" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Update job progress
	if err := h.jobQueue.UpdateProgress(jobID, req.Progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update job progress",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job progress updated successfully",
	})
}

// GetJobTypes returns available job types
// @Summary Get job types
// @Description Get list of available job types
// @Tags jobs
// @Produce json
// @Success 200 {object} map[string][]string
// @Router /api/v1/jobs/types [get]
func (h *JobHandler) GetJobTypes(c *gin.Context) {
	jobTypes := map[string]string{
		"course_generation":   "Generate complete course from markdown",
		"video_processing":    "Process video components",
		"audio_generation":    "Generate audio from text",
		"subtitle_generation": "Generate subtitles from audio",
	}

	c.JSON(http.StatusOK, gin.H{
		"job_types": jobTypes,
	})
}

// GetSystemJobs retrieves all jobs (admin only)
// @Summary Get all jobs (Admin only)
// @Description Retrieve all jobs in the system (admin only)
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Param status query string false "Filter by job status"
// @Param type query string false "Filter by job type"
// @Success 200 {array} jobs.Job
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/jobs [get]
func (h *JobHandler) GetSystemJobs(c *gin.Context) {
	// This endpoint requires admin permissions
	// For simplicity, we'll use the existing job queue methods
	// In a production system, you might want to add direct database queries for admin

	// Parse query parameters
	limit := 50 // Default for admin
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 200 {
			limit = val
		}
	}

	offset := 0 // Default
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	// In a real implementation, you would query the database with filters
	// For now, return a simplified response
	c.JSON(http.StatusOK, gin.H{
		"message": "System jobs endpoint - implement with proper database filtering",
		"limit":   limit,
		"offset":  offset,
		"filters": map[string]string{
			"status": c.Query("status"),
			"type":   c.Query("type"),
		},
	})
}
