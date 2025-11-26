package main

import (
	"fmt"
	"log"
	"os"

	"github.com/course-creator/core-processor/api"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
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
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.Default()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Create handlers
	courseHandler := api.NewCourseHandler()

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

func generateCourse(markdownFile, outputDir string) {
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
