package api

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/course-creator/core-processor/repository"
	"github.com/gin-gonic/gin"
)

// CourseHandler handles course-related API endpoints
type CourseHandler struct {
	courseGenerator *pipeline.CourseGenerator
	db             *database.DB
	courseRepo     *repository.CourseRepository
	jobRepo        *repository.ProcessingJobRepository
}

// NewCourseHandler creates a new course handler
func NewCourseHandler(db *database.DB) *CourseHandler {
	return &CourseHandler{
		courseGenerator: pipeline.NewCourseGenerator(),
		db:             db,
		courseRepo:     repository.NewCourseRepository(db),
		jobRepo:        repository.NewProcessingJobRepository(db),
	}
}

// GenerateCourse handles POST /api/courses/generate
func (h *CourseHandler) GenerateCourse(c *gin.Context) {
	var req GenerateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if req.MarkdownPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "markdown_path is required"})
		return
	}

	// Set default output directory if not provided
	outputDir := req.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(req.MarkdownPath)
	}

	// Set default options
	options := req.Options
	if options.Quality == "" {
		options.Quality = "standard"
	}
	if options.Languages == nil {
		options.Languages = []string{"en"}
	}

	// Create processing job
	jobOptions := &repository.JobOptions{
		Voice:           options.Voice,
		BackgroundMusic: options.BackgroundMusic,
		Languages:       options.Languages,
		Quality:         options.Quality,
	}

	optionsJSON, err := repository.SerializeJobOptions(jobOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize options: " + err.Error()})
		return
	}

	job := &models.ProcessingJobDB{
		InputPath: req.MarkdownPath,
		OutputPath: &outputDir,
		Options:   optionsJSON,
		Status:    "pending",
		Progress:  0,
	}

	job, err = h.jobRepo.CreateJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processing job: " + err.Error()})
		return
	}

	// Start async processing
	go h.processCourseAsync(job.ID, req.MarkdownPath, outputDir, options)

	response := GenerateCourseResponse{
		JobID:      job.ID,
		Status:     "pending",
		Message:    "Course generation started",
	}

	c.JSON(http.StatusAccepted, response)
}

// GetCourse handles GET /api/courses/:id
func (h *CourseHandler) GetCourse(c *gin.Context) {
	courseID := c.Param("id")

	course, err := h.courseRepo.GetCourseByID(courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found", "id": courseID})
		return
	}

	c.JSON(http.StatusOK, course)
}

// ListCourses handles GET /api/courses
func (h *CourseHandler) ListCourses(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var courses []models.CourseDB
	var total int64
	var err error

	if search != "" {
		courses, total, err = h.courseRepo.SearchCourses(search, offset, limit)
	} else {
		courses, total, err = h.courseRepo.GetAllCourses(offset, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := ListCoursesResponse{
		Courses: courses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles GET /api/health
func (h *CourseHandler) HealthCheck(c *gin.Context) {
	// Check database connection
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// GetJob handles GET /api/jobs/:id
func (h *CourseHandler) GetJob(c *gin.Context) {
	jobID := c.Param("id")

	job, err := h.jobRepo.GetJobByID(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found", "id": jobID})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs handles GET /api/jobs
func (h *CourseHandler) ListJobs(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	courseID := c.Query("course_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var jobs []models.ProcessingJobDB
	var total int64
	var err error

	if courseID != "" {
		jobs, total, err = h.jobRepo.GetJobsByCourseID(courseID, offset, limit)
	} else if status != "" {
		jobs, total, err = h.jobRepo.GetJobsByStatus(status, offset, limit)
	} else {
		jobs, total, err = h.jobRepo.GetAllJobs(offset, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := ListJobsResponse{
		Jobs:  jobs,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

// API request/response types

type GenerateCourseRequest struct {
	MarkdownPath string                   `json:"markdown_path" binding:"required"`
	OutputDir    string                   `json:"output_dir,omitempty"`
	Options      models.ProcessingOptions `json:"options,omitempty"`
}

type GenerateCourseResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ListCoursesResponse struct {
	Courses []models.CourseDB `json:"courses"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

type ListJobsResponse struct {
	Jobs  []models.ProcessingJobDB `json:"jobs"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
}

// processCourseAsync processes a course asynchronously
func (h *CourseHandler) processCourseAsync(jobID, inputPath, outputDir string, options models.ProcessingOptions) {
	// Update job status to running
	_, err := h.jobRepo.UpdateJobStatus(jobID, "running")
	if err != nil {
		// Log error but continue
		return
	}

	// Update progress
	h.jobRepo.UpdateJobProgress(jobID, 10)

	// Generate course
	course, err := h.courseGenerator.GenerateCourse(inputPath, outputDir, options)
	if err != nil {
		// Update job with error
		h.jobRepo.UpdateJobError(jobID, err.Error())
		return
	}

	// Update progress
	h.jobRepo.UpdateJobProgress(jobID, 80)

	// Save course to database
	courseDB, err := h.courseRepo.CreateCourse(course)
	if err != nil {
		h.jobRepo.UpdateJobError(jobID, "Failed to save course: "+err.Error())
		return
	}

	// Link job to course
	h.jobRepo.SetJobCourseID(jobID, courseDB.ID)

	// Update job status to completed
	h.jobRepo.UpdateJobProgress(jobID, 100)
	h.jobRepo.UpdateJobStatus(jobID, "completed")
}
