package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid request with markdown path",
			requestBody: map[string]interface{}{
				"markdown_path": "/tmp/test.md",
				"output_dir":    "/tmp/output",
				"options": map[string]interface{}{
					"quality":   "standard",
					"languages": []string{"en"},
				},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "Missing markdown path",
			requestBody: map[string]interface{}{
				"output_dir": "/tmp/output",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MarkdownPath",
		},
		{
			name: "Empty markdown path",
			requestBody: map[string]interface{}{
				"markdown_path": "",
				"output_dir":    "/tmp/output",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MarkdownPath",
		},
		{
			name: "Valid request with default output directory",
			requestBody: map[string]interface{}{
				"markdown_path": "/tmp/test.md",
				"options": map[string]interface{}{
					"quality": "standard",
				},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "Valid request with custom options",
			requestBody: map[string]interface{}{
				"markdown_path": "/tmp/test.md",
				"output_dir":    "/tmp/output",
				"options": map[string]interface{}{
					"quality":          "high",
					"voice":            "en-US-Wavenet-A",
					"background_music": "calm",
					"languages":        []string{"en", "es"},
				},
			},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/courses/generate", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			handler.GenerateCourse(c)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedError != "" {
				var response map[string]string
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tc.expectedError)
			} else if tc.expectedStatus == http.StatusAccepted {
				var response GenerateCourseResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.JobID)
				assert.Equal(t, "pending", response.Status)
				assert.Equal(t, "Course generation started", response.Message)
			}
		})
	}
}

func TestGetCourse_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	req, err := http.NewRequest("GET", "/api/courses/nonexistent-id", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent-id"}}

	handler.GetCourse(c)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Course not found", response["error"])
	assert.Equal(t, "nonexistent-id", response["id"])
}

func TestListCourses_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "Default pagination",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Custom page and limit",
			query:          "?page=2&limit=5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid page (negative)",
			query:          "?page=-1&limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid limit (too high)",
			query:          "?page=1&limit=200",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Search query",
			query:          "?search=test",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/courses"+tc.query, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			handler.ListCourses(c)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var response ListCoursesResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.NotNil(t, response.Courses)
			assert.GreaterOrEqual(t, response.Total, int64(0))
			assert.GreaterOrEqual(t, response.Page, 1)
			assert.GreaterOrEqual(t, response.Limit, 1)
			assert.LessOrEqual(t, response.Limit, 100)
		})
	}
}

func TestGetJob(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	req, err := http.NewRequest("GET", "/api/jobs/nonexistent-job", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent-job"}}

	handler.GetJob(c)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Job not found", response["error"])
	assert.Equal(t, "nonexistent-job", response["id"])
}

func TestListJobs_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "Default pagination",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With status filter",
			query:          "?status=pending",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With course_id filter",
			query:          "?course_id=test-course",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With pagination and filters",
			query:          "?page=2&limit=5&status=completed&course_id=test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid page (negative)",
			query:          "?page=-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid limit (too high)",
			query:          "?limit=200",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/jobs"+tc.query, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)
			c.Request = req

			handler.ListJobs(c)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var response ListJobsResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.NotNil(t, response.Jobs)
			assert.GreaterOrEqual(t, response.Total, int64(0))
			assert.GreaterOrEqual(t, response.Page, 1)
			assert.GreaterOrEqual(t, response.Limit, 1)
			assert.LessOrEqual(t, response.Limit, 100)
		})
	}
}

func TestHealthCheck_DatabaseFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)

	db.Close()

	handler := NewCourseHandler(db)

	req, err := http.NewRequest("GET", "/api/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = req

	handler.HealthCheck(c)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", response["status"])
	assert.Equal(t, "Database connection failed", response["error"])
}

func TestGenerateCourse_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	req, err := http.NewRequest("POST", "/api/courses/generate", bytes.NewBufferString("invalid json"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = req

	handler.GenerateCourse(c)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")
}

func TestGenerateCourse_OptionsSerializationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	requestBody := map[string]interface{}{
		"markdown_path": "/tmp/test.md",
		"options": map[string]interface{}{
			"quality": func() {}, // Function cannot be serialized to JSON
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/courses/generate", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = req

	handler.GenerateCourse(c)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProcessCourseAsync_Integration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "1" {
		t.Skip("Skipping integration test")
	}

	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()

	db, err := database.NewDatabase(&database.Config{
		Path:  filepath.Join(tempDir, "test.db"),
		Debug: false,
		Env:   "test",
	})
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	markdownPath := filepath.Join(tempDir, "test.md")
	outputDir := filepath.Join(tempDir, "output")

	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	testContent := `# Test Course
## Introduction
This is a test course for async processing.

## Lesson 1
Test lesson content.`

	err = os.WriteFile(markdownPath, []byte(testContent), 0644)
	require.NoError(t, err)

	options := models.ProcessingOptions{
		Quality:   "standard",
		Languages: []string{"en"},
	}

	jobOptions := &repository.JobOptions{
		Voice:           options.Voice,
		BackgroundMusic: options.BackgroundMusic,
		Languages:       options.Languages,
		Quality:         options.Quality,
	}

	optionsJSON, err := repository.SerializeJobOptions(jobOptions)
	require.NoError(t, err)

	job := &models.ProcessingJobDB{
		InputPath:  markdownPath,
		OutputPath: &outputDir,
		Options:    optionsJSON,
		Status:     "pending",
		Progress:   0,
	}

	jobRepo := repository.NewProcessingJobRepository(db)
	job, err = jobRepo.CreateJob(job)
	require.NoError(t, err)

	go handler.processCourseAsync(job.ID, markdownPath, outputDir, options)

	time.Sleep(2 * time.Second)

	updatedJob, err := jobRepo.GetJobByID(job.ID)
	require.NoError(t, err)

	assert.Contains(t, []string{"running", "completed"}, updatedJob.Status)
	assert.GreaterOrEqual(t, updatedJob.Progress, 0)
}

func TestCourseHandler_GetCourses_API(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	tests := []struct {
		name     string
		params   map[string]interface{}
		expected int
	}{
		{
			name:     "Default params",
			params:   map[string]interface{}{},
			expected: 0,
		},
		{
			name: "With pagination",
			params: map[string]interface{}{
				"page":     2,
				"pageSize": 10,
			},
			expected: 0,
		},
		{
			name: "With search",
			params: map[string]interface{}{
				"search": "test",
			},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			courses, err := handler.GetCourses(tc.params)
			require.NoError(t, err)
			assert.Len(t, courses, tc.expected)
		})
	}
}

func TestCourseHandler_GetCourseByID_API(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	handler := NewCourseHandler(db)

	course, err := handler.GetCourseByID("nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, course)
}

func TestGenerateCourseRequestAPI_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     GenerateCourseRequestAPI
		shouldError bool
	}{
		{
			name: "Valid request",
			request: GenerateCourseRequestAPI{
				Markdown: "# Test Course\nContent",
				Options: models.ProcessingOptions{
					Quality: "standard",
				},
			},
			shouldError: false,
		},
		{
			name: "Missing markdown",
			request: GenerateCourseRequestAPI{
				Markdown: "",
				Options: models.ProcessingOptions{
					Quality: "standard",
				},
			},
			shouldError: true,
		},
		{
			name: "Empty markdown",
			request: GenerateCourseRequestAPI{
				Markdown: "   ",
				Options: models.ProcessingOptions{
					Quality: "standard",
				},
			},
			shouldError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldError {
				assert.Error(t, tc.request.Validate())
			} else {
				assert.NoError(t, tc.request.Validate())
			}
		})
	}
}

func (r *GenerateCourseRequestAPI) Validate() error {
	if r.Markdown == "" {
		return fmt.Errorf("markdown is required")
	}
	if len(strings.TrimSpace(r.Markdown)) == 0 {
		return fmt.Errorf("markdown cannot be empty")
	}
	return nil
}
