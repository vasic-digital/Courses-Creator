package api

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/repository"
	"github.com/course-creator/core-processor/services"
	"github.com/gin-gonic/gin"
)

// CourseAPIService provides API endpoints that match what the frontend expects
type CourseAPIService struct {
	handler *CourseHandler
}

// NewCourseAPIService creates a new course API service
func NewCourseAPIService(handler *CourseHandler) *CourseAPIService {
	return &CourseAPIService{
		handler: handler,
	}
}

// GetCoursesAPI handles GET /api/v1/courses with frontend-compatible response
func (s *CourseAPIService) GetCoursesAPI(c *gin.Context) {
	// Parse query parameters
	params := make(map[string]interface{})
	
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			params["page"] = p
		}
	}
	
	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			params["pageSize"] = ps
		}
	}
	
	if search := c.Query("search"); search != "" {
		params["search"] = search
	}

	courses, err := s.handler.GetCourses(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return in the format the frontend expects
	c.JSON(http.StatusOK, courses)
}

// GetCourseAPI handles GET /api/v1/courses/:id with frontend-compatible response
func (s *CourseAPIService) GetCourseAPI(c *gin.Context) {
	courseID := c.Param("id")

	course, err := s.handler.GetCourseByID(courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found", "id": courseID})
		return
	}

	c.JSON(http.StatusOK, course)
}

// GenerateCourseAPI handles POST /api/v1/courses/generate with frontend-compatible request
func (s *CourseAPIService) GenerateCourseAPI(c *gin.Context) {
	var req GenerateCourseRequestAPI
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if req.Markdown == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "markdown is required"})
		return
	}
	
	// Validate markdown content for security issues
	if !services.ValidateContent(req.Markdown) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content detected"})
		return
	}

	// Set default options
	if req.Options.Quality == "" {
		req.Options.Quality = "standard"
	}
	if req.Options.Languages == nil {
		req.Options.Languages = []string{"en"}
	}

	// Create a temporary file for the markdown
	tempDir, err := ioutil.TempDir("", "course-gen")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp directory"})
		return
	}
	
	markdownFile := filepath.Join(tempDir, "content.md")
	if err := ioutil.WriteFile(markdownFile, []byte(req.Markdown), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write markdown file"})
		return
	}
	
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create output directory"})
		return
	}

	// Get user ID from context (for future use in user-specific job tracking)
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create a processing job
	jobOptions := &repository.JobOptions{
		Voice:           req.Options.Voice,
		BackgroundMusic: req.Options.BackgroundMusic,
		Languages:       req.Options.Languages,
		Quality:         req.Options.Quality,
	}

	optionsJSON, err := repository.SerializeJobOptions(jobOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize options: " + err.Error()})
		return
	}

	job := &models.ProcessingJobDB{
		InputPath:  markdownFile,
		OutputPath: &outputDir,
		Options:    optionsJSON,
		Status:     "pending",
		Progress:   0,
	}

	job, err = s.handler.jobRepo.CreateJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create processing job: " + err.Error()})
		return
	}

	// Start async processing
	go s.handler.processCourseAsync(job.ID, markdownFile, outputDir, req.Options)

	response := GenerateCourseResponse{
		JobID:   job.ID,
		Status:  "pending",
		Message: "Course generation started",
	}

	c.JSON(http.StatusAccepted, response)
}

// RegisterCourseAPIRoutes registers the frontend-compatible routes
func (s *CourseAPIService) RegisterCourseAPIRoutes(router *gin.RouterGroup) {
	log.Printf("Registering CourseAPIRoutes")
	courses := router.Group("/courses")
	{
		log.Printf("Registering GET /courses endpoint")
		courses.GET("", s.GetCoursesAPI)
		log.Printf("Registering GET /courses/:id endpoint")
		courses.GET("/:id", s.GetCourseAPI)
		log.Printf("Registering POST /courses/generate endpoint")
		courses.POST("/generate", s.GenerateCourseAPI)
	}
	log.Printf("Finished registering CourseAPIRoutes")
}