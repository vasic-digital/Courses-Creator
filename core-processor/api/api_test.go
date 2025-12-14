package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/jobs"
	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper function to create a test database
func setupAPITestDB(t *testing.T) *database.DB {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	db := &database.DB{DB: gormDB}

	// Auto-migrate the models
	err = db.AutoMigrate(
		&models.UserDB{},
		&models.UserPreferencesDB{},
		&models.UserSessionDB{},
		&models.JobDB{},
		&models.CourseDB{},
		&models.CourseMetadataDB{},
		&models.LessonDB{},
		&models.SubtitleDB{},
		&models.InteractiveElementDB{},
		&models.ProcessingJobDB{},
	)
	require.NoError(t, err)

	return db
}

// Helper function to create test services
func setupTestServices(db *database.DB) (*services.AuthService, *middleware.AuthMiddleware, *jobs.JobQueue) {
	authMiddleware := middleware.NewAuthMiddleware()
	authService := services.NewAuthService(db.GetGormDB(), authMiddleware)

	// Create a simple job queue for testing
	jobQueue := jobs.NewJobQueue(db.GetGormDB(), 1)

	return authService, authMiddleware, jobQueue
}

// Helper function to create test handlers
func setupTestHandlers(db *database.DB) (*AuthHandler, *CourseHandler, *JobHandler, *CourseAPIService) {
	authService, authMiddleware, jobQueue := setupTestServices(db)

	authHandler := NewAuthHandler(authService, authMiddleware)
	courseHandler := NewCourseHandler(db)
	jobHandler := NewJobHandler(jobQueue)
	courseAPIService := NewCourseAPIService(courseHandler)

	return authHandler, courseHandler, jobHandler, courseAPIService
}

// Test AuthHandler Register
func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	router := gin.New()
	router.POST("/register", authHandler.Register)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful registration",
			requestBody: map[string]interface{}{
				"email":      "test@example.com",
				"first_name": "Test",
				"last_name":  "User",
				"password":   "ValidPass123!",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "invalid email",
			requestBody: map[string]interface{}{
				"email":      "invalid-email",
				"first_name": "Test",
				"last_name":  "User",
				"password":   "ValidPass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "weak password",
			requestBody: map[string]interface{}{
				"email":      "weak@example.com",
				"first_name": "Weak",
				"last_name":  "Pass",
				"password":   "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"email": "incomplete@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "user")
				assert.Contains(t, response, "access_token")
				assert.Contains(t, response, "refresh_token")
			}
		})
	}
}

// Test AuthHandler Login
func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	// Pre-create a user for testing
	authService, _, _ := setupTestServices(db)
	_, err := authService.Register(context.Background(), &services.RegisterRequest{
		Email:     "login@example.com",
		FirstName: "Login",
		LastName:  "Test",
		Password:  "ValidPass123!",
	})
	require.NoError(t, err)

	router := gin.New()
	router.POST("/login", authHandler.Login)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"email":    "login@example.com",
				"password": "ValidPass123!",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "invalid credentials",
			requestBody: map[string]interface{}{
				"email":    "login@example.com",
				"password": "WrongPass123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "nonexistent user",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "ValidPass123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "missing fields",
			requestBody: map[string]interface{}{
				"email": "login@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "user")
				assert.Contains(t, response, "access_token")
				assert.Contains(t, response, "refresh_token")
			}
		})
	}
}

// Test AuthHandler RefreshToken
func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	// Create a user and get tokens
	authService, _, _ := setupTestServices(db)
	resp, err := authService.Register(context.Background(), &services.RegisterRequest{
		Email:     "refresh@example.com",
		FirstName: "Refresh",
		LastName:  "Test",
		Password:  "ValidPass123!",
	})
	require.NoError(t, err)

	router := gin.New()
	router.POST("/refresh", authHandler.RefreshToken)

	t.Run("successful token refresh", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"refresh_token": resp.RefreshToken,
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "user")
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"refresh_token": "invalid-token",
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

// Test AuthHandler Logout
func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	})
	router.POST("/logout", authHandler.Logout)

	t.Run("successful logout", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "message")
	})
}

// Test AuthHandler GetProfile
func TestAuthHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	// Create a user
	authService, _, _ := setupTestServices(db)
	resp, err := authService.Register(context.Background(), &services.RegisterRequest{
		Email:     "profile@example.com",
		FirstName: "Profile",
		LastName:  "Test",
		Password:  "ValidPass123!",
	})
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", resp.User.ID)
		c.Next()
	})
	router.GET("/profile", authHandler.GetProfile)

	t.Run("successful profile retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/profile", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "email")
		assert.Equal(t, "profile@example.com", response["email"])
	})
}

// Test JobHandler CreateJob
func TestJobHandler_CreateJob(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, _, jobHandler, _ := setupTestHandlers(db)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	})
	router.POST("/jobs", jobHandler.CreateJob)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful job creation",
			requestBody: map[string]interface{}{
				"type":     "course_generation",
				"payload":  map[string]interface{}{"course_id": "test-course"},
				"priority": 1,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"type": "course_generation",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "unauthenticated request",
			requestBody: map[string]interface{}{
				"type":    "course_generation",
				"payload": map[string]interface{}{"course_id": "test-course"},
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/jobs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// For the unauthenticated test, don't set user_id
			if tt.name == "unauthenticated request" {
				routerWithoutAuth := gin.New()
				routerWithoutAuth.POST("/jobs", jobHandler.CreateJob)
				w := httptest.NewRecorder()
				routerWithoutAuth.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedStatus, w.Code)
				return
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "id")
				assert.Contains(t, response, "type")
			}
		})
	}
}

// Test JobHandler GetJob
func TestJobHandler_GetJob(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, _, jobHandler, _ := setupTestHandlers(db)

	// Create a test job
	jobQueue := jobs.NewJobQueue(db.GetGormDB(), 1)
	job, err := jobQueue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "test-user", map[string]interface{}{"test": "data"}, jobs.JobPriorityNormal)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("user_role", "creator")
		c.Next()
	})
	router.GET("/jobs/:id", jobHandler.GetJob)

	t.Run("successful job retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/jobs/"+job.ID, nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "type")
	})

	t.Run("job not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/jobs/nonexistent-id", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

// Test JobHandler ListJobs
func TestJobHandler_ListJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, _, jobHandler, _ := setupTestHandlers(db)

	// Create test jobs
	jobQueue := jobs.NewJobQueue(db.GetGormDB(), 1)
	_, err := jobQueue.Enqueue(context.Background(), jobs.JobTypeCourseGeneration, "test-user", map[string]interface{}{"test": "data1"}, jobs.JobPriorityNormal)
	require.NoError(t, err)
	_, err = jobQueue.Enqueue(context.Background(), jobs.JobTypeVideoProcessing, "test-user", map[string]interface{}{"test": "data2"}, jobs.JobPriorityNormal)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	router.GET("/jobs", jobHandler.GetUserJobs)

	t.Run("successful job listing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/jobs", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var jobs []interface{}
		err := json.Unmarshal(w.Body.Bytes(), &jobs)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(jobs), 2)
	})
}

// Test CourseAPIService GetCoursesAPI
func TestCourseAPIService_GetCoursesAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, _, _, courseAPIService := setupTestHandlers(db)

	router := gin.New()
	router.GET("/courses", courseAPIService.GetCoursesAPI)

	t.Run("successful courses retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/courses?page=1&pageSize=10", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return OK even if no courses exist
		assert.Equal(t, http.StatusOK, w.Code)

		var courses []interface{}
		err := json.Unmarshal(w.Body.Bytes(), &courses)
		require.NoError(t, err)
		// Should be empty array if no courses exist
		assert.IsType(t, []interface{}{}, courses)
	})

	t.Run("courses with search", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/courses?search=test&page=1&pageSize=10", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var courses []interface{}
		err := json.Unmarshal(w.Body.Bytes(), &courses)
		require.NoError(t, err)
		assert.IsType(t, []interface{}{}, courses)
	})
}

// Test CourseAPIService GenerateCourseAPI
func TestCourseAPIService_GenerateCourseAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, _, _, courseAPIService := setupTestHandlers(db)

	router := gin.New()
	router.POST("/courses/generate", courseAPIService.GenerateCourseAPI)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful course generation",
			requestBody: map[string]interface{}{
				"markdown": "# Test Course\n\nThis is a test course.",
				"options": map[string]interface{}{
					"quality": "standard",
				},
			},
			expectedStatus: http.StatusAccepted,
			expectError:    false,
		},
		{
			name: "missing markdown",
			requestBody: map[string]interface{}{
				"options": map[string]interface{}{
					"quality": "standard",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid content",
			requestBody: map[string]interface{}{
				"markdown": "<script>alert('xss')</script>",
				"options": map[string]interface{}{
					"quality": "standard",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/courses/generate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "job_id")
				assert.Contains(t, response, "message")
				assert.Contains(t, response, "status")
			}
		})
	}
}

// Test CourseAPIService GetCourseAPI
func TestCourseAPIService_GetCourseAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, _, _, courseAPIService := setupTestHandlers(db)

	router := gin.New()
	router.GET("/courses/:id", courseAPIService.GetCourseAPI)

	t.Run("course not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/courses/nonexistent-id", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

// Test CourseHandler GenerateCourse
func TestCourseHandler_GenerateCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, courseHandler, _, _ := setupTestHandlers(db)

	router := gin.New()
	router.POST("/generate", courseHandler.GenerateCourse)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful course generation",
			requestBody: map[string]interface{}{
				"markdown_path": "/tmp/test.md",
				"output_dir":    "/tmp/output",
			},
			expectedStatus: http.StatusAccepted,
			expectError:    false,
		},
		{
			name: "missing markdown path",
			requestBody: map[string]interface{}{
				"output_dir": "/tmp/output",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/generate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "job_id")
				assert.Contains(t, response, "status")
				assert.Contains(t, response, "message")
			}
		})
	}
}

// Test AuthHandler UpdateProfile
func TestAuthHandler_UpdateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	// Create a user first
	authService, _, _ := setupTestServices(db)
	resp, err := authService.Register(context.Background(), &services.RegisterRequest{
		Email:     "update@example.com",
		FirstName: "Update",
		LastName:  "Test",
		Password:  "ValidPass123!",
	})
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", resp.User.ID)
		c.Next()
	})
	router.PUT("/profile", authHandler.UpdateProfile)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful profile update",
			requestBody: map[string]interface{}{
				"first_name": "Updated",
				"last_name":  "Name",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid JSON",
			requestBody:    nil, // This will cause JSON parsing error
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "unauthenticated request",
			requestBody: map[string]interface{}{
				"first_name": "ShouldFail",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.requestBody != nil {
				body, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest("PUT", "/profile", bytes.NewBufferString("invalid json"))
				req.Header.Set("Content-Type", "application/json")
			}

			if tt.name == "unauthenticated request" {
				routerWithoutAuth := gin.New()
				routerWithoutAuth.PUT("/profile", authHandler.UpdateProfile)
				w := httptest.NewRecorder()
				routerWithoutAuth.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedStatus, w.Code)
				return
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "id")
				assert.Contains(t, response, "email")
				assert.Equal(t, "Updated", response["first_name"])
			}
		})
	}
}

// Test AuthHandler UpdatePassword
func TestAuthHandler_UpdatePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	// Create a user first
	authService, _, _ := setupTestServices(db)
	resp, err := authService.Register(context.Background(), &services.RegisterRequest{
		Email:     "password@example.com",
		FirstName: "Password",
		LastName:  "Test",
		Password:  "ValidPass123!",
	})
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", resp.User.ID)
		c.Next()
	})
	router.PUT("/password", authHandler.UpdatePassword)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful password update",
			requestBody: map[string]interface{}{
				"current_password": "ValidPass123!",
				"new_password":     "NewValidPass456!",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "weak new password",
			requestBody: map[string]interface{}{
				"current_password": "ValidPass123!",
				"new_password":     "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "wrong current password",
			requestBody: map[string]interface{}{
				"current_password": "WrongPass123!",
				"new_password":     "NewValidPass456!",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing fields",
			requestBody: map[string]interface{}{
				"current_password": "ValidPass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "unauthenticated request",
			requestBody: map[string]interface{}{
				"current_password": "ValidPass123!",
				"new_password":     "NewValidPass456!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.name == "unauthenticated request" {
				routerWithoutAuth := gin.New()
				routerWithoutAuth.PUT("/password", authHandler.UpdatePassword)
				w := httptest.NewRecorder()
				routerWithoutAuth.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedStatus, w.Code)
				return
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
			}
		})
	}
}

// Test AuthHandler CreateUserByAdmin
func TestAuthHandler_CreateUserByAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-user-id")
		c.Set("user_role", "admin")
		c.Next()
	})
	router.POST("/admin/users", authHandler.CreateUserByAdmin)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful user creation by admin",
			requestBody: map[string]interface{}{
				"email":      "admincreated@example.com",
				"first_name": "Admin",
				"last_name":  "Created",
				"password":   "ValidPass123!",
				"role":       "creator",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "invalid email",
			requestBody: map[string]interface{}{
				"email":      "invalid-email",
				"first_name": "Test",
				"last_name":  "User",
				"password":   "ValidPass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "weak password",
			requestBody: map[string]interface{}{
				"email":      "weakpass@example.com",
				"first_name": "Weak",
				"last_name":  "Pass",
				"password":   "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"email": "incomplete@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "unauthorized - not admin",
			requestBody: map[string]interface{}{
				"email":      "notadmin@example.com",
				"first_name": "Not",
				"last_name":  "Admin",
				"password":   "ValidPass123!",
			},
			expectedStatus: http.StatusForbidden,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/admin/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.name == "unauthorized - not admin" {
				routerWithoutAdmin := gin.New()
				routerWithoutAdmin.Use(func(c *gin.Context) {
					c.Set("user_id", "regular-user-id")
					c.Set("user_role", "creator")
					c.Next()
				})
				routerWithoutAdmin.POST("/admin/users", authHandler.CreateUserByAdmin)
				w := httptest.NewRecorder()
				routerWithoutAdmin.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedStatus, w.Code)
				return
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				// Check that response contains user fields directly (not wrapped in "user" object)
				assert.Contains(t, response, "id")
				assert.Contains(t, response, "email")
				assert.Contains(t, response, "first_name")
			}
		})
	}
}

// Test AuthHandler Logout with different scenarios
func TestAuthHandler_Logout_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	router := gin.New()
	router.POST("/logout", authHandler.Logout)

	t.Run("logout without authentication", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 401 when not authenticated
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

// Test AuthHandler GetProfile with different scenarios
func TestAuthHandler_GetProfile_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	authHandler, _, _, _ := setupTestHandlers(db)

	router := gin.New()
	router.GET("/profile", authHandler.GetProfile)

	t.Run("profile without authentication", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/profile", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})

	t.Run("profile with invalid user ID", func(t *testing.T) {
		routerWithInvalidUser := gin.New()
		routerWithInvalidUser.Use(func(c *gin.Context) {
			c.Set("user_id", "invalid-user-id")
			c.Next()
		})
		routerWithInvalidUser.GET("/profile", authHandler.GetProfile)

		req := httptest.NewRequest("GET", "/profile", nil)

		w := httptest.NewRecorder()
		routerWithInvalidUser.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

// Test CourseAPIService GenerateCourseAPI with more edge cases
func TestCourseAPIService_GenerateCourseAPI_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAPITestDB(t)
	_, _, _, courseAPIService := setupTestHandlers(db)

	router := gin.New()
	router.POST("/courses/generate", courseAPIService.GenerateCourseAPI)

	t.Run("generate course with empty markdown", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"markdown": "",
			"options": map[string]interface{}{
				"quality": "standard",
			},
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/courses/generate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})

	t.Run("generate course with very long markdown", func(t *testing.T) {
		longMarkdown := strings.Repeat("# Test\n\nContent\n\n", 1000)
		requestBody := map[string]interface{}{
			"markdown": longMarkdown,
			"options": map[string]interface{}{
				"quality": "standard",
			},
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/courses/generate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should either succeed or fail based on validation
		assert.True(t, w.Code == http.StatusAccepted || w.Code == http.StatusBadRequest)
	})
}
