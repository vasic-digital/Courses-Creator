package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Mock course data
type Course struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	CreatedAt   string `json:"created_at"`
}

var courses = []Course{
	{
		ID:          "test-course-1",
		Title:       "Test Course 1",
		Description: "This is a test course for demonstration",
		Duration:    30,
		CreatedAt:   "2025-01-01T00:00:00Z",
	},
	{
		ID:          "test-course-2",
		Title:       "Test Course 2",
		Description: "Another test course",
		Duration:    45,
		CreatedAt:   "2025-01-02T00:00:00Z",
	},
}

func main() {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.Default()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	v1 := r.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		// Public courses
		public := v1.Group("/public")
		{
			public.GET("/courses", func(c *gin.Context) {
				page := 1
				pageSize := 20
				search := ""

				if p := c.Query("page"); p != "" {
					fmt.Sscanf(p, "%d", &page)
				}
				if ps := c.Query("pageSize"); ps != "" {
					fmt.Sscanf(ps, "%d", &pageSize)
				}
				if s := c.Query("search"); s != "" {
					search = s
				}

				// Simple filter for search
				filteredCourses := courses
				if search != "" {
					var filtered []Course
					for _, course := range courses {
						if contains(course.Title, search) || contains(course.Description, search) {
							filtered = append(filtered, course)
						}
					}
					filteredCourses = filtered
				}

				// Simple pagination
				start := (page - 1) * pageSize
				end := start + pageSize
				if start > len(filteredCourses) {
					start = len(filteredCourses)
				}
				if end > len(filteredCourses) {
					end = len(filteredCourses)
				}

				paginated := filteredCourses[start:end]

				c.JSON(http.StatusOK, gin.H{
					"courses": paginated,
					"total":   int64(len(filteredCourses)),
					"page":    page,
					"limit":   pageSize,
				})
			})

			public.GET("/courses/:id", func(c *gin.Context) {
				id := c.Param("id")
				for _, course := range courses {
					if course.ID == id {
						c.JSON(http.StatusOK, course)
						return
					}
				}
				c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			})
		}
	}

	// Start server
	port := "8081"
	log.Printf("Starting test API server on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}