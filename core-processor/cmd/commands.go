package cmd

import (
	"fmt"
	"log"

	"github.com/course-creator/core-processor/api"
	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
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
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Create handlers
	courseHandler := api.NewCourseHandler(db)

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", courseHandler.HealthCheck)
		v1.POST("/courses/generate", courseHandler.GenerateCourse)
		v1.GET("/courses", courseHandler.ListCourses)
		v1.GET("/courses/:id", courseHandler.GetCourse)
	}

	// Start server
	port := "8080"
	log.Printf("Starting Course Creator API server on port %s", port)
	log.Fatal(r.Run(":" + port))
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
		log.Fatalf("Failed to generate course: %v", err)
	}

	fmt.Printf("Course generated successfully: %s\n", course.Title)
	fmt.Printf("Lessons: %d\n", len(course.Lessons))
	fmt.Printf("Output directory: %s\n", outputDir)
}
