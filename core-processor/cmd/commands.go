package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/gin-gonic/gin"
)

// StartServer starts the API server
func StartServer() {
	// Initialize database
	dbConfig := database.DefaultConfig()
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return
	}
	defer db.Close()

	// Create router with all routes configured
	r := SetupServerRouter()

	// Note: In the current implementation, we use SetupServerRouter which has placeholder handlers
	// In a real implementation, we'd initialize services and handlers here

	// Replace placeholder routes with real handlers
	// Note: In a real implementation, we'd need to modify the router to replace routes
	// For now, we'll keep the placeholder routes for testing

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

// SetupServerRouter creates and configures the full server router (extracted for testing)
func SetupServerRouter() *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.Default()

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

	// Rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute

	// Public API routes (no auth required)
	v1 := r.Group("/api/v1")
	v1.Use(rateLimiter.Middleware())
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		// Authentication routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "register endpoint"})
			})
			authGroup.POST("/login", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "login endpoint"})
			})
			authGroup.POST("/refresh", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "refresh endpoint"})
			})
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
	{
		publicCourses.GET("/courses", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"courses": []string{}})
		})
	}

	// Protected API routes (auth required)
	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Original endpoints
		protected.POST("/courses/generate", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "generate course"})
		})
		protected.GET("/courses/original", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"courses": []string{}})
		})
		protected.GET("/courses/original/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"course_id": id})
		})

		// Job endpoints
		protected.GET("/jobs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"jobs": []string{}})
		})
		protected.GET("/jobs/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"job_id": id})
		})
	}

	return r
}
