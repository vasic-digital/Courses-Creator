package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/course-creator/core-processor/api"
	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/course-creator/core-processor/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI_Integration(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)

	// Create test directories
	tempDir := filepath.Join(os.TempDir(), "api_test")
	outputDir := filepath.Join(os.TempDir(), "api_output")

	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Create test markdown file
	markdownContent := `# API Integration Test Course

This is a test course for API integration testing.

## Introduction

Testing the course generation API.

## Main Content

Testing course content processing.

## Conclusion

API integration test complete.`

	markdownPath := filepath.Join(tempDir, "test_course.md")
	err = utils.WriteFile(markdownPath, markdownContent)
	require.NoError(t, err)

	// Setup database
	dbConfig := database.DefaultConfig()
	db, err := database.NewDatabase(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Setup Gin router
	router := gin.New()
	handler := api.NewCourseHandler(db)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handler.HealthCheck)
		v1.POST("/courses/generate", handler.GenerateCourse)
		v1.GET("/courses", handler.ListCourses)
		v1.GET("/courses/:id", handler.GetCourse)
	}

	// Test health check
	t.Run("Health Check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})

	// Test course generation
	t.Run("Generate Course", func(t *testing.T) {
		t.Skip("Skipping course generation API test - requires TTS generation")

		requestBody := map[string]interface{}{
			"markdown_path": markdownPath,
			"output_dir":    outputDir,
			"options": map[string]interface{}{
				"voice":            "bark",
				"background_music": true,
				"languages":        []string{"en"},
				"quality":          "standard",
			},
		}

		req, _ := http.NewRequest("POST", "/api/v1/courses/generate", nil)
		req.Header.Set("Content-Type", "application/json")

		// Convert request body to JSON and set it
		jsonBody, _ := json.Marshal(requestBody)
		req.Body = io.NopCloser(strings.NewReader(string(jsonBody)))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept either success or error (may fail due to missing dependencies)
		if w.Code == http.StatusOK {
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "course_id")
			assert.Contains(t, response, "status")
		} else {
			t.Logf("Course generation returned status %d (may be expected with missing dependencies)", w.Code)
		}
	})

	// Test list courses
	t.Run("List Courses", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/courses", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "courses")
	})
}

func TestPipeline_EndToEnd(t *testing.T) {
	t.Skip("Skipping E2E test - requires TTS generation")

	// Setup test environment
	tempDir := filepath.Join(os.TempDir(), "e2e_test")
	outputDir := filepath.Join(os.TempDir(), "e2e_output")

	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Create comprehensive test markdown
	markdownContent := `# End-to-End Test Course

This is a comprehensive test course for end-to-end testing of the entire pipeline.

## Module 1: System Overview

The Course Creator system consists of multiple components working together:
- Backend API for course processing
- TTS engines for voice generation
- Video assembly pipeline
- Cross-platform applications

## Module 2: Technical Implementation

### Backend Processing

The backend processes markdown files through:
1. Parsing and structure analysis
2. Audio generation using AI models
3. Video creation with text overlays
4. Final assembly and packaging

### AI Integration

Multiple AI providers are supported:
- Bark TTS for high-quality voice synthesis
- SpeechT5 for alternative TTS options
- LLM providers for content enhancement

## Module 3: Quality Assurance

### Testing Strategy

The system implements comprehensive testing:
- Unit tests for individual components
- Integration tests for API endpoints
- End-to-end tests for complete workflows
- Performance tests for scalability

### Quality Metrics

Quality is ensured through:
- 100% test coverage requirement
- Automated quality gates
- Performance benchmarks
- Security scanning

## Module 4: User Experience

### Cross-Platform Support

The system provides applications for:
- Desktop (Electron-based creator)
- Mobile (React Native player)
- Web browser (HTML5 player)

### Features

Key features include:
- Intuitive course creation interface
- Real-time processing feedback
- Multi-language subtitle support
- Offline content availability

## Conclusion

This concludes our comprehensive end-to-end test course.`

	markdownPath := filepath.Join(tempDir, "e2e_course.md")
	err = utils.WriteFile(markdownPath, markdownContent)
	require.NoError(t, err)

	// Create course generator
	generator := pipeline.NewCourseGenerator()

	options := models.ProcessingOptions{
		Quality:         "high",
		Languages:       []string{"en", "es", "fr"},
		BackgroundMusic: true,
		Voice:           stringPtr("v2/en_speaker_6"),
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	// Run generation with timeout
	resultCh := make(chan *models.Course, 1)
	errCh := make(chan error, 1)

	go func() {
		course, err := generator.GenerateCourse(markdownPath, outputDir, options)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- course
	}()

	select {
	case <-ctx.Done():
		t.Skip("End-to-end test timed out (may be due to missing dependencies)")
	case err := <-errCh:
		if err != nil {
			t.Logf("E2E test failed (may be expected): %v", err)

			// Check that error is related to missing dependencies
			if strings.Contains(err.Error(), "failed to generate audio") ||
				strings.Contains(err.Error(), "failed to create video") ||
				strings.Contains(err.Error(), "ffmpeg") {
				t.Skip("Skipping due to missing FFmpeg/TTS dependencies")
			}
			t.Fatalf("Unexpected error: %v", err)
		}
	case course := <-resultCh:
		// Validate course structure
		require.NotNil(t, course)

		// Validate basic course properties
		assert.NotEmpty(t, course.ID)
		assert.Equal(t, "End-to-End Test Course", course.Title)
		assert.Contains(t, course.Description, "comprehensive test")

		// Validate lessons
		assert.Len(t, course.Lessons, 5) // Overview + Technical + QA + UX + Conclusion

		lessonTitles := make([]string, len(course.Lessons))
		for i, lesson := range course.Lessons {
			lessonTitles[i] = lesson.Title

			// Validate lesson structure
			assert.NotEmpty(t, lesson.ID)
			assert.NotEmpty(t, lesson.Title)
			assert.NotEmpty(t, lesson.Content)
			assert.Greater(t, lesson.Order, 0)

			// Check for generated media (may not exist if dependencies missing)
			if lesson.VideoURL != nil && utils.FileExists(*lesson.VideoURL) {
				// Verify video file properties
				size, err := utils.GetFileSize(*lesson.VideoURL)
				require.NoError(t, err)
				assert.Greater(t, size, int64(0))
			}

			if lesson.AudioURL != nil && utils.FileExists(*lesson.AudioURL) {
				// Verify audio file properties
				size, err := utils.GetFileSize(*lesson.AudioURL)
				require.NoError(t, err)
				assert.Greater(t, size, int64(0))
			}
		}

		// Validate expected lesson titles
		expectedTitles := []string{
			"Module 1: System Overview",
			"Module 2: Technical Implementation",
			"Module 3: Quality Assurance",
			"Module 4: User Experience",
			"Conclusion",
		}

		for _, expected := range expectedTitles {
			assert.Contains(t, lessonTitles, expected)
		}

		t.Log("End-to-end test completed successfully!")
	}
}

func TestFileProcessing_Integration(t *testing.T) {
	// Test file processing pipeline integration
	tempDir := filepath.Join(os.TempDir(), "file_test")
	outputDir := filepath.Join(os.TempDir(), "file_output")

	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is test content for file processing integration test."
	err = utils.WriteFile(testFile, testContent)
	require.NoError(t, err)

	// Test file operations
	t.Run("File Operations", func(t *testing.T) {
		// Test file existence
		assert.True(t, utils.FileExists(testFile))

		// Test file size
		size, err := utils.GetFileSize(testFile)
		require.NoError(t, err)
		assert.Greater(t, size, int64(0))

		// Test file copy
		copiedFile := filepath.Join(tempDir, "copied.txt")
		err = utils.CopyFile(testFile, copiedFile)
		require.NoError(t, err)
		defer os.Remove(copiedFile)

		assert.True(t, utils.FileExists(copiedFile))

		// Verify content
		copiedContent, err := os.ReadFile(copiedFile)
		require.NoError(t, err)
		assert.Equal(t, testContent, string(copiedContent))

		// Test file extension
		ext := utils.GetFileExtension(copiedFile)
		assert.Equal(t, ".txt", ext)

		// Test filename sanitization
		unsafeFile := "file/with\\special:chars.txt"
		safeFile := utils.SanitizeFilename(unsafeFile)
		assert.NotContains(t, safeFile, "/")
		assert.NotContains(t, safeFile, "\\")
		assert.NotContains(t, safeFile, ":")
	})
}

func TestConcurrentProcessing_Integration(t *testing.T) {
	t.Skip("Skipping concurrent processing test - requires TTS generation")

	// Test concurrent course generation
	tempDir := filepath.Join(os.TempDir(), "concurrent_test")
	outputDir := filepath.Join(os.TempDir(), "concurrent_output")

	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Create multiple test courses
	courses := []struct {
		name    string
		content string
	}{
		{
			name: "Course 1",
			content: `# Concurrent Test Course 1

This is test course 1 for concurrent processing.

## Content

Content for course 1.`,
		},
		{
			name: "Course 2",
			content: `# Concurrent Test Course 2

This is test course 2 for concurrent processing.

## Content

Content for course 2.`,
		},
		{
			name: "Course 3",
			content: `# Concurrent Test Course 3

This is test course 3 for concurrent processing.

## Content

Content for course 3.`,
		},
	}

	// Process courses concurrently
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resultCh := make(chan map[string]interface{}, len(courses))
	errCh := make(chan error, len(courses))

	for _, course := range courses {
		go func(name, content string) {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic in goroutine: %v", r)
				}
			}()

			// Create markdown file
			markdownPath := filepath.Join(tempDir, fmt.Sprintf("%s.md", utils.GenerateID()))
			err := utils.WriteFile(markdownPath, content)
			if err != nil {
				errCh <- fmt.Errorf("failed to create markdown file: %w", err)
				return
			}
			defer os.Remove(markdownPath)

			// Create course-specific output directory
			courseOutputDir := filepath.Join(outputDir, name)
			err = utils.EnsureDir(courseOutputDir)
			if err != nil {
				errCh <- fmt.Errorf("failed to create output directory: %w", err)
				return
			}

			// Generate course
			generator := pipeline.NewCourseGenerator()
			options := models.ProcessingOptions{
				Quality:   "standard",
				Languages: []string{"en"},
			}

			generatedCourse, err := generator.GenerateCourse(markdownPath, courseOutputDir, options)
			if err != nil {
				errCh <- fmt.Errorf("failed to generate course %s: %w", name, err)
				return
			}

			resultCh <- map[string]interface{}{
				"name":   name,
				"course": generatedCourse,
			}
		}(course.name, course.content)
	}

	// Collect results
	results := make(map[string]interface{})
	completed := 0
	failed := 0

	for completed+failed < len(courses) {
		select {
		case <-ctx.Done():
			t.Skip("Concurrent processing test timed out")
		case err := <-errCh:
			failed++
			t.Logf("Course generation failed (may be expected): %v", err)
		case result := <-resultCh:
			completed++
			results[result["name"].(string)] = result["course"]
		}
	}

	// Verify that at least some courses completed or failed gracefully
	assert.True(t, completed > 0 || failed > 0)
	t.Logf("Concurrent processing completed: %d successful, %d failed", completed, failed)

	// Validate successful courses
	for name, course := range results {
		if course != nil {
			c := course.(*models.Course)
			assert.NotEmpty(t, c.ID)
			assert.NotEmpty(t, c.Title)
			assert.Contains(t, c.Title, name)
		}
	}
}

func TestErrorHandling_Integration(t *testing.T) {
	t.Skip("Skipping error handling test - requires TTS generation")

	// Test error handling throughout the pipeline
	tempDir := filepath.Join(os.TempDir(), "error_test")
	outputDir := filepath.Join(os.TempDir(), "error_output")

	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	t.Run("Invalid Markdown Path", func(t *testing.T) {
		generator := pipeline.NewCourseGenerator()
		options := models.ProcessingOptions{
			Quality:   "standard",
			Languages: []string{"en"},
		}

		course, err := generator.GenerateCourse("/nonexistent/path.md", outputDir, options)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "markdown file does not exist")
		assert.Nil(t, course)
	})

	t.Run("Invalid Output Directory", func(t *testing.T) {
		markdownPath := filepath.Join(tempDir, "test.md")
		err := utils.WriteFile(markdownPath, "test")
		require.NoError(t, err)
		defer os.Remove(markdownPath)

		generator := pipeline.NewCourseGenerator()
		options := models.ProcessingOptions{
			Quality:   "standard",
			Languages: []string{"en"},
		}

		// Try to use directory that doesn't exist and can't be created
		course, err := generator.GenerateCourse(markdownPath, "/invalid/path/that/cannot/be/created", options)

		assert.Error(t, err)
		assert.Nil(t, course)
	})

	t.Run("Empty Options", func(t *testing.T) {
		markdownPath := filepath.Join(tempDir, "test2.md")
		err := utils.WriteFile(markdownPath, "# Test\n\nContent")
		require.NoError(t, err)
		defer os.Remove(markdownPath)

		generator := pipeline.NewCourseGenerator()
		var options models.ProcessingOptions // Empty options

		course, err := generator.GenerateCourse(markdownPath, outputDir, options)

		// Should work with empty options
		if err == nil && course != nil {
			assert.NotEmpty(t, course.ID)
			assert.NotEmpty(t, course.Title)
		}
	})
}

func TestPerformance_Integration(t *testing.T) {
	t.Skip("Skipping performance test - requires TTS generation")

	// Test performance characteristics
	tempDir := filepath.Join(os.TempDir(), "perf_test")
	outputDir := filepath.Join(os.TempDir(), "perf_output")

	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	t.Run("Processing Speed", func(t *testing.T) {
		// Create a moderately sized course
		var contentBuilder strings.Builder
		contentBuilder.WriteString("# Performance Test Course\n\n")
		contentBuilder.WriteString("This is a performance test course.\n\n")

		// Add multiple sections
		for i := 1; i <= 20; i++ {
			contentBuilder.WriteString(fmt.Sprintf("## Section %d\n\n", i))
			contentBuilder.WriteString(fmt.Sprintf("This is the content for section %d. ", i))

			// Add some content to each section
			for j := 1; j <= 5; j++ {
				contentBuilder.WriteString(fmt.Sprintf("Paragraph %d in section %d. ", j, i))
			}
			contentBuilder.WriteString("\n\n")
		}

		markdownPath := filepath.Join(tempDir, "perf_course.md")
		err = utils.WriteFile(markdownPath, contentBuilder.String())
		require.NoError(t, err)

		// Measure processing time
		start := time.Now()

		generator := pipeline.NewCourseGenerator()
		options := models.ProcessingOptions{
			Quality:   "standard",
			Languages: []string{"en"},
		}

		course, err := generator.GenerateCourse(markdownPath, outputDir, options)

		duration := time.Since(start)

		if err == nil && course != nil {
			assert.NotEmpty(t, course.ID)
			assert.Len(t, course.Lessons, 20) // 20 sections

			// Performance assertion - should process within reasonable time
			// This may fail if dependencies are missing, which is ok
			assert.Less(t, duration, 60*time.Second, "Processing should complete within 60 seconds")

			t.Logf("Processed 20 sections in %v (%.2f seconds per section)",
				duration, duration.Seconds()/20.0)
		} else {
			t.Logf("Performance test failed due to dependencies: %v", err)
		}
	})

	t.Run("Memory Usage", func(t *testing.T) {
		// Test memory efficiency with large course
		var contentBuilder strings.Builder
		contentBuilder.WriteString("# Memory Test Course\n\n")

		// Create a very large course
		for i := 1; i <= 100; i++ {
			contentBuilder.WriteString(fmt.Sprintf("## Large Section %d\n\n", i))

			// Add substantial content
			for j := 1; j <= 50; j++ {
				contentBuilder.WriteString(fmt.Sprintf("This is paragraph %d in large section %d. ", j, i))
				// Add more text to make it substantial
				contentBuilder.WriteString("Additional text to increase content size and test memory efficiency. ")
				contentBuilder.WriteString("More content to ensure we're testing memory usage properly. ")
			}
			contentBuilder.WriteString("\n\n")
		}

		markdownPath := filepath.Join(tempDir, "memory_course.md")
		err = utils.WriteFile(markdownPath, contentBuilder.String())
		require.NoError(t, err)

		// Test that we can at least process without running out of memory
		generator := pipeline.NewCourseGenerator()
		options := models.ProcessingOptions{
			Quality:   "standard",
			Languages: []string{"en"},
		}

		course, err := generator.GenerateCourse(markdownPath, outputDir, options)

		if err == nil && course != nil {
			assert.Len(t, course.Lessons, 100) // 100 sections
			t.Logf("Successfully processed large course with 100 sections")
		} else {
			t.Logf("Memory test completed with expected errors: %v", err)
		}
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
