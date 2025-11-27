package api

import (
	"log"
	"net/http"
	"strconv"

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

	// Set default options
	if req.Options.Quality == "" {
		req.Options.Quality = "standard"
	}
	if req.Options.Languages == nil {
		req.Options.Languages = []string{"en"}
	}

	// Create a temporary file for the markdown
	// For now, we'll return a mock response
	// TODO: Implement actual markdown processing
	
	response := GenerateCourseResponse{
		JobID:   "mock-job-id",
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