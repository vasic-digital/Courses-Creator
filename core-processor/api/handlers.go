package api

import (
	"net/http"
	"path/filepath"

	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/gin-gonic/gin"
)

// CourseHandler handles course-related API endpoints
type CourseHandler struct {
	courseGenerator *pipeline.CourseGenerator
}

// NewCourseHandler creates a new course handler
func NewCourseHandler() *CourseHandler {
	return &CourseHandler{
		courseGenerator: pipeline.NewCourseGenerator(),
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
		outputDir = filepath.Dir(req.MarkdownPath) // Default to same directory as input
	}

	// Set default options
	options := req.Options
	if options.Quality == "" {
		options.Quality = "standard"
	}
	if options.Languages == nil {
		options.Languages = []string{"en"}
	}

	// Generate course
	course, err := h.courseGenerator.GenerateCourse(req.MarkdownPath, outputDir, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := GenerateCourseResponse{
		Course:     *course,
		Status:     "success",
		OutputPath: outputDir,
	}

	c.JSON(http.StatusOK, response)
}

// GetCourse handles GET /api/courses/:id
func (h *CourseHandler) GetCourse(c *gin.Context) {
	courseID := c.Param("id")

	// Placeholder - in real implementation, fetch from database
	c.JSON(http.StatusNotFound, gin.H{"error": "Course not found", "id": courseID})
}

// ListCourses handles GET /api/courses
func (h *CourseHandler) ListCourses(c *gin.Context) {
	// Placeholder - in real implementation, fetch from database
	courses := []models.Course{}
	c.JSON(http.StatusOK, gin.H{"courses": courses})
}

// HealthCheck handles GET /api/health
func (h *CourseHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// API request/response types

type GenerateCourseRequest struct {
	MarkdownPath string                   `json:"markdown_path" binding:"required"`
	OutputDir    string                   `json:"output_dir,omitempty"`
	Options      models.ProcessingOptions `json:"options,omitempty"`
}

type GenerateCourseResponse struct {
	Course     models.Course `json:"course"`
	Status     string        `json:"status"`
	OutputPath string        `json:"output_path"`
}
