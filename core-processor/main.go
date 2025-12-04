package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/course-creator/core-processor/api"
	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/database"
	filestorage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/jobs"
	"github.com/course-creator/core-processor/metrics"
	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/course-creator/core-processor/services"
	"github.com/course-creator/core-processor/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: course-creator <command>")
		fmt.Println("Commands:")
		fmt.Println("  server    Start the API server")
		fmt.Println("  generate  Generate course from markdown file")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "server":
		startServer()
	case "generate":
		if len(os.Args) < 4 {
			fmt.Println("Usage: course-creator generate <markdown-file> <output-dir>")
			os.Exit(1)
		}
		markdownFile := os.Args[2]
		outputDir := os.Args[3]
		generateCourse(markdownFile, outputDir)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func startServer() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Initialize database
	dbConfig := database.DefaultConfig()
	dbConfig.Debug = gin.Mode() == gin.DebugMode

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Set Gin mode
	gin.SetMode(gin.DebugMode)

	// Initialize metrics
	metrics.Init()

	// Create Gin router
	r := gin.Default()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(metrics.Middleware())

	// Add a middleware to log all requests
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		fmt.Printf("Request: %s %s\n", method, path)
		c.Next()
	})

	// Initialize authentication
	authMiddleware := middleware.NewAuthMiddleware()

	// Initialize services
	authService := services.NewAuthService(db.GetGormDB(), authMiddleware)

	// Initialize job queue
	jobQueue := jobs.NewJobQueue(db.GetGormDB(), 4) // 4 workers

	// Initialize storage manager
	storageManager, err := filestorage.NewStorageManagerWithDefault(filestorage.DefaultStorageConfig())
	if err != nil {
		log.Printf("Failed to initialize storage: %v", err)
		os.Exit(1)
	}

	jobCtx := &jobs.JobContext{
		Queue:           jobQueue,
		Storage:         storageManager.DefaultProvider(),
		MarkdownParser:  utils.NewMarkdownParser(),
		CourseGenerator: pipeline.NewCourseGenerator(),
	}
	jobCtx.RegisterDefaultHandlers()

	// Start job queue
	if err := jobQueue.Start(); err != nil {
		log.Printf("Failed to start job queue: %v", err)
		os.Exit(1)
	}
	defer jobQueue.Stop()

	// Create handlers
	courseHandler := api.NewCourseHandler(db)
	authHandler := api.NewAuthHandler(authService, authMiddleware)
	jobHandler := api.NewJobHandler(jobQueue)
	courseAPIService := api.NewCourseAPIService(courseHandler)

	// Rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute

	// Public API routes (no auth required)
	v1 := r.Group("/api/v1")
	v1.Use(rateLimiter.Middleware())
	{
		v1.GET("/health", courseHandler.HealthCheck)

		// Public course routes
		publicCourses := v1.Group("/public")
		courseAPIService.RegisterCourseAPIRoutes(publicCourses)

		// Authentication routes (public)
		fmt.Printf("About to register auth routes...\n")
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.GET("/types", jobHandler.GetJobTypes)
		}
		fmt.Printf("Auth routes registered\n")
	}

	// Protected API routes (auth required)
	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Course routes
		courseGroup := protected.Group("/courses")
		{
			courseGroup.POST("/generate", courseAPIService.GenerateCourseAPI)
			courseGroup.GET("", courseAPIService.GetCoursesAPI)
			courseGroup.GET("/:id", courseAPIService.GetCourseAPI)
		}

		// Job routes
		jobGroup := protected.Group("/jobs")
		{
			jobGroup.POST("", jobHandler.CreateJob)
			jobGroup.GET("", jobHandler.GetUserJobs)
			jobGroup.GET("/:id", jobHandler.GetJob)
			jobGroup.POST("/:id/cancel", jobHandler.CancelJob)
			jobGroup.PUT("/:id/progress", jobHandler.UpdateJobProgress)
		}

		// User profile routes
		profileGroup := protected.Group("/auth")
		{
			profileGroup.GET("/profile", authHandler.GetProfile)
			profileGroup.PUT("/profile", authHandler.UpdateProfile)
			profileGroup.PUT("/password", authHandler.UpdatePassword)
			profileGroup.POST("/logout", authHandler.Logout)
		}
	}

	// Metrics endpoint
	v1.GET("/metrics", gin.WrapH(metrics.Handler()))

	// Admin routes (admin role required)
	admin := protected.Group("/admin")
	admin.Use(authMiddleware.RequirePermission("system:admin"))
	{
		admin.POST("/users", authHandler.CreateUserByAdmin)
		admin.GET("/jobs", jobHandler.GetSystemJobs)
	}

	// Static file serving from storage
	defaultStorage := cfg.Storage["default"]
	r.Static("/storage", defaultStorage.BasePath)

	// Start server
	port := "8080"
	log.Printf("Starting Course Creator API server on port %s", port)
	log.Printf("Database: %s", dbConfig.Path)
	log.Printf("Storage: %s (type: %s)", defaultStorage.BasePath, defaultStorage.Type)
	log.Printf("Job queue: started with %d workers", 4)
	log.Printf("Authentication: enabled")

	if err := r.Run(":" + port); err != nil {
		log.Printf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

func generateCourse(markdownFile, outputDir string) {
	fmt.Printf("Generating course from %s to %s\n", markdownFile, outputDir)

	generator := pipeline.NewCourseGenerator()
	options := models.ProcessingOptions{
		Quality:   "standard",
		Languages: []string{"en"},
	}

	course, err := generator.GenerateCourse(markdownFile, outputDir, options)
	if err != nil {
		log.Printf("Failed to generate course: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Course generated successfully: %s\n", course.Title)
	fmt.Printf("Lessons: %d\n", len(course.Lessons))
	fmt.Printf("Output directory: %s\n", outputDir)
}
