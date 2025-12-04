package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/course-creator/core-processor/api"
	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/course-creator/core-processor/services"
	"github.com/gin-gonic/gin"
)

// StartServer starts the API server
func StartServer() {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.Default()

	// Initialize database
	dbConfig := database.DefaultConfig()
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return
	}
	defer db.Close()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Add security headers middleware
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	// Initialize authentication
	authMiddleware := middleware.NewAuthMiddleware()

	// Initialize services
	authService := services.NewAuthService(db.GetGormDB(), authMiddleware)

	// Create handlers
	courseHandler := api.NewCourseHandler(db)
	authHandler := api.NewAuthHandler(authService, authMiddleware)
	courseAPIService := api.NewCourseAPIService(courseHandler)

	// Rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute

	// Public API routes (no auth required)
	v1 := r.Group("/api/v1")
	v1.Use(rateLimiter.Middleware())
	{
		v1.GET("/health", courseHandler.HealthCheck)

		// Authentication routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		// Debug route to see all registered routes
		v1.GET("/debug/routes", func(c *gin.Context) {
			routes := c.FullPath()
			c.JSON(200, gin.H{
				"message": "Routes debug",
				"path":    routes,
				"query":   c.Request.URL.RawQuery,
			})
		})
	}

	// Frontend-compatible routes (public for now)
	publicCourses := v1.Group("/public")
	log.Printf("Registering public courses routes under /api/v1/public")
	courseAPIService.RegisterCourseAPIRoutes(publicCourses)

	// Protected API routes (auth required)
	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Original endpoints
		protected.POST("/courses/generate", courseHandler.GenerateCourse)
		protected.GET("/courses/original", courseHandler.ListCourses)
		protected.GET("/courses/original/:id", courseHandler.GetCourse)

		// Job endpoints
		protected.GET("/jobs", courseHandler.ListJobs)
		protected.GET("/jobs/:id", courseHandler.GetJob)
	}

	// Start server
	port := "8080"
	log.Printf("Starting Course Creator API server on port %s", port)

	// Print all registered routes for debugging
	log.Printf("Registered routes:")
	for _, route := range r.Routes() {
		log.Printf("  %s %s", route.Method, route.Path)
	}

	if err := r.Run(":" + port); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}

// GenerateCourse generates a course from markdown file via CLI
func GenerateCourse(markdownFile, outputDir string) {
	fmt.Printf("Generating course from %s to %s\n", markdownFile, outputDir)

	generator := pipeline.NewCourseGenerator()
	options := models.ProcessingOptions{
		Quality:   "standard",
		Languages: []string{"en"},
	}

	course, err := generator.GenerateCourse(markdownFile, outputDir, options)
	if err != nil {
		log.Printf("Failed to generate course: %v", err)
		return
	}

	fmt.Printf("Course generated successfully: %s\n", course.Title)
	fmt.Printf("Lessons: %d\n", len(course.Lessons))
	fmt.Printf("Output directory: %s\n", outputDir)
}

// SetupRouter creates and configures a Gin router for testing
func SetupRouter() *gin.Engine {
	// Set Gin mode to test mode
	gin.SetMode(gin.TestMode)

	// Create Gin router
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())

	// Add security headers middleware
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	// Add a simple health check route for testing
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
