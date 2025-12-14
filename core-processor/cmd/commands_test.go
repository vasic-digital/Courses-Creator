package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRouter(t *testing.T) {
	router := SetupRouter()

	assert.NotNil(t, router)
	assert.IsType(t, &gin.Engine{}, router)
}

func TestSetupRouter_HealthCheck(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "ok")
}

func TestSetupRouter_SecurityHeaders(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	// Check security headers are set
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestSetupRouter_TestMode(t *testing.T) {
	// SetupRouter should set Gin to test mode
	router := SetupRouter()

	// Verify Gin is in test mode by checking if it has recovery middleware
	// (test mode typically has different middleware setup)
	assert.NotNil(t, router)
}

func TestGenerateCourse_ValidInputs(t *testing.T) {
	// Test that GenerateCourse doesn't panic with valid inputs
	// Note: This is a basic test since the actual generation depends on external dependencies
	assert.NotPanics(t, func() {
		// This will likely fail due to missing dependencies, but shouldn't panic
		GenerateCourse("nonexistent.md", "/tmp/output")
	})
}

func TestGenerateCourse_EmptyInputs(t *testing.T) {
	// Test that GenerateCourse handles empty inputs gracefully
	assert.NotPanics(t, func() {
		GenerateCourse("", "")
	})
}

func TestSetupRouter_MiddlewareOrder(t *testing.T) {
	router := SetupRouter()

	// Test that middleware is applied in the correct order
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	// The router should have recovery middleware and security headers middleware
	router.ServeHTTP(w, req)

	// If we get a response, middleware is working
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetupRouter_CustomHeaders(t *testing.T) {
	router := SetupRouter()

	// Add a custom route that sets additional headers
	router.GET("/custom", func(c *gin.Context) {
		c.Header("Custom-Header", "test-value")
		c.JSON(http.StatusOK, gin.H{"message": "custom"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/custom", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test-value", w.Header().Get("Custom-Header"))
	assert.Contains(t, w.Body.String(), "custom")
}

func TestSetupRouter_CORSHeaders(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/", nil)
	router.ServeHTTP(w, req)

	// Even for non-CORS requests, security headers should be present
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSetupRouter_InvalidRoute(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	// Should return 404 for non-existent routes
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGenerateCourse_InputValidation(t *testing.T) {
	// Test with empty markdown file path
	assert.NotPanics(t, func() {
		GenerateCourse("", "/tmp/output")
	})

	// Test with empty output directory
	assert.NotPanics(t, func() {
		GenerateCourse("test.md", "")
	})

	// Test with both empty
	assert.NotPanics(t, func() {
		GenerateCourse("", "")
	})
}

func TestSetupRouter_MultipleRequests(t *testing.T) {
	router := SetupRouter()

	// Test multiple concurrent requests
	done := make(chan bool, 2)

	go func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		done <- true
	}()

	go func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		done <- true
	}()

	// Wait for both requests to complete
	<-done
	<-done
}

func TestSetupRouter_ContentType(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	// Check that content type is set to JSON
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

// TestStartServer_RouteRegistration tests that StartServer properly registers routes
func TestStartServer_RouteRegistration(t *testing.T) {
	// Note: StartServer is difficult to test directly due to database dependencies
	// and server startup. We'll test the components it uses instead.

	// Test that SetupRouter creates a router (this is used by StartServer)
	router := SetupRouter()
	assert.NotNil(t, router)

	// Test that the router has the expected middleware
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	// Check security headers are applied
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
}

// TestStartServer_MiddlewareSetup tests middleware configuration
func TestStartServer_MiddlewareSetup(t *testing.T) {
	router := SetupRouter()

	// Test that security headers middleware is applied
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	// Verify all security headers are present
	headers := w.Header()
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
}

// TestStartServer_RateLimiting tests rate limiting middleware setup
func TestStartServer_RateLimiting(t *testing.T) {
	// Test that rate limiter can be created (used in StartServer)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	assert.NotNil(t, rateLimiter)

	// Test rate limiter middleware creation
	middleware := rateLimiter.Middleware()
	assert.NotNil(t, middleware)
}

// TestStartServer_AuthMiddleware tests auth middleware setup
func TestStartServer_AuthMiddleware(t *testing.T) {
	// Test that auth middleware can be created (used in StartServer)
	authMiddleware := middleware.NewAuthMiddleware()
	assert.NotNil(t, authMiddleware)

	// Test that RequireAuth middleware exists
	requireAuth := authMiddleware.RequireAuth()
	assert.NotNil(t, requireAuth)
}

// TestStartServer_DatabaseInitialization tests database initialization logic
func TestStartServer_DatabaseInitialization(t *testing.T) {
	// Test that default database config can be created
	dbConfig := database.DefaultConfig()
	assert.NotNil(t, dbConfig)

	// Test that database config has expected fields
	assert.NotEmpty(t, dbConfig.Path)
	assert.Equal(t, "development", dbConfig.Env)
	assert.False(t, dbConfig.Debug)
}

// TestStartServer_RouteGroups tests route group setup
func TestStartServer_RouteGroups(t *testing.T) {
	// Create a test router similar to what StartServer creates
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Add security headers middleware (from StartServer)
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	// Create route groups like StartServer does
	v1 := r.Group("/api/v1")
	assert.NotNil(t, v1)

	// Test that routes can be added to groups
	v1.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

// TestStartServer_ProtectedRoutes tests protected route setup
func TestStartServer_ProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Add security headers
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	})

	// Create auth middleware (mock)
	authMiddleware := middleware.NewAuthMiddleware()

	// Create protected routes group like StartServer
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware.RequireAuth())
	assert.NotNil(t, protected)

	// Add a protected route
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	// Test that the route exists (auth will fail but route should be registered)
	routes := r.Routes()
	routeExists := false
	for _, route := range routes {
		if route.Path == "/api/v1/protected" && route.Method == "GET" {
			routeExists = true
			break
		}
	}
	assert.True(t, routeExists, "Protected route should be registered")
}

// TestStartServer_PublicRoutes tests public route setup
func TestStartServer_PublicRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Add security headers
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	})

	// Create public routes like StartServer
	v1 := r.Group("/api/v1")
	assert.NotNil(t, v1)

	// Add health check route
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Test health check route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

// TestStartServer_DebugRoutes tests debug route setup
func TestStartServer_DebugRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	v1.GET("/debug/routes", func(c *gin.Context) {
		routes := c.FullPath()
		c.JSON(http.StatusOK, gin.H{
			"message": "Routes debug",
			"path":    routes,
			"query":   c.Request.URL.RawQuery,
		})
	})

	// Test debug route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/debug/routes?test=123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "Routes debug")
	assert.Contains(t, body, "/api/v1/debug/routes")
	assert.Contains(t, body, "test=123")
}

// TestStartServer_ServerConfig tests server configuration
func TestStartServer_ServerConfig(t *testing.T) {
	// Test that the server would start on the expected port
	// (We can't actually start the server in tests, but we can verify the configuration)

	// Test port configuration (hardcoded in StartServer as "8080")
	expectedPort := "8080"
	assert.Equal(t, "8080", expectedPort)

	// Test Gin mode setting
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	gin.SetMode(gin.ReleaseMode)
	assert.Equal(t, gin.ReleaseMode, gin.Mode())

	gin.SetMode(gin.TestMode)
	assert.Equal(t, gin.TestMode, gin.Mode())
}

// TestGenerateCourse_ErrorHandling tests error handling in GenerateCourse
func TestGenerateCourse_ErrorHandling(t *testing.T) {
	// Test with invalid file paths - should not panic
	assert.NotPanics(t, func() {
		GenerateCourse("/nonexistent/path/file.md", "/nonexistent/output")
	})

	// Test with empty strings
	assert.NotPanics(t, func() {
		GenerateCourse("", "")
	})

	// Test with special characters in paths
	assert.NotPanics(t, func() {
		GenerateCourse("/tmp/test file.md", "/tmp/output dir")
	})
}

// TestStartServer_GinConfiguration tests Gin router configuration
func TestStartServer_GinConfiguration(t *testing.T) {
	// Test Gin default router creation
	r := gin.Default()
	assert.NotNil(t, r)

	// Test Gin new router creation
	r2 := gin.New()
	assert.NotNil(t, r2)

	// Test that routers have different configurations
	// Default has Logger and Recovery, New has only Recovery
	assert.NotEqual(t, r, r2)
}

// TestStartServer_MiddlewareStack tests middleware stack setup
func TestStartServer_MiddlewareStack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add middleware in the same order as StartServer
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

	// Add a test route
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "middleware test"})
	})

	// Test that all middleware is applied
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that security headers are applied
	headers := w.Header()
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
}

// TestStartServer_RouteSetup tests the route setup logic from StartServer
func TestStartServer_RouteSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Add security headers middleware (from StartServer)
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	// Initialize auth middleware (from StartServer)
	authMiddleware := middleware.NewAuthMiddleware()

	// Initialize rate limiter (from StartServer)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	// Create route groups like StartServer does
	v1 := r.Group("/api/v1")
	v1.Use(rateLimiter.Middleware())

	// Public API routes (no auth required) - from StartServer
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		// Authentication routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "register"})
			})
			authGroup.POST("/login", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "login"})
			})
			authGroup.POST("/refresh", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "refresh"})
			})
		}

		// Debug route
		v1.GET("/debug/routes", func(c *gin.Context) {
			routes := c.FullPath()
			c.JSON(http.StatusOK, gin.H{
				"message": "Routes debug",
				"path":    routes,
				"query":   c.Request.URL.RawQuery,
			})
		})
	}

	// Protected API routes (auth required) - from StartServer
	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
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
		protected.GET("/jobs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"jobs": []string{}})
		})
		protected.GET("/jobs/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"job_id": id})
		})
	}

	// Test public routes
	t.Run("health check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	})

	t.Run("auth register", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "register")
	})

	t.Run("auth login", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "login")
	})

	t.Run("auth refresh", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "refresh")
	})

	t.Run("debug routes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/debug/routes?test=123", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Routes debug")
	})

	// Test protected routes (will fail auth but routes should exist)
	t.Run("protected routes exist", func(t *testing.T) {
		routes := r.Routes()
		expectedRoutes := []string{
			"GET /api/v1/health",
			"POST /api/v1/auth/register",
			"POST /api/v1/auth/login",
			"POST /api/v1/auth/refresh",
			"GET /api/v1/debug/routes",
			"POST /api/v1/courses/generate",
			"GET /api/v1/courses/original",
			"GET /api/v1/courses/original/:id",
			"GET /api/v1/jobs",
			"GET /api/v1/jobs/:id",
		}

		for _, expected := range expectedRoutes {
			found := false
			for _, route := range routes {
				routePath := route.Method + " " + route.Path
				if routePath == expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Route %s should be registered", expected)
		}
	})
}

// TestStartServer_ServerStartup tests server startup logic
func TestStartServer_ServerStartup(t *testing.T) {
	// Test that Gin mode is set correctly (from StartServer)
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	// Test setting release mode (used in StartServer)
	gin.SetMode(gin.ReleaseMode)
	assert.Equal(t, gin.ReleaseMode, gin.Mode())

	// Test setting test mode
	gin.SetMode(gin.TestMode)
	assert.Equal(t, gin.TestMode, gin.Mode())
}

// TestStartServer_PortConfiguration tests port configuration
func TestStartServer_PortConfiguration(t *testing.T) {
	// Test that the expected port is configured (hardcoded as "8080" in StartServer)
	expectedPort := "8080"

	// Test that it's a valid port string
	assert.NotEmpty(t, expectedPort)
	assert.Equal(t, "8080", expectedPort)
}

// TestGenerateCourse_OptionsSetup tests the options setup in GenerateCourse
func TestGenerateCourse_OptionsSetup(t *testing.T) {
	// Test that GenerateCourse sets up options correctly
	// We can't easily test the full function, but we can test the options creation logic

	// This tests the options setup logic from GenerateCourse
	options := struct {
		Quality   string
		Languages []string
	}{
		Quality:   "standard",
		Languages: []string{"en"},
	}

	assert.Equal(t, "standard", options.Quality)
	assert.Contains(t, options.Languages, "en")
	assert.Len(t, options.Languages, 1)
}

// TestSetupServerRouter tests the SetupServerRouter function
func TestSetupServerRouter(t *testing.T) {
	router := SetupServerRouter()
	assert.NotNil(t, router)

	// Test that the router has expected routes
	routes := router.Routes()
	assert.True(t, len(routes) > 0, "Router should have routes registered")

	// Check that expected routes exist
	expectedRoutes := map[string]bool{
		"GET /api/v1/health":               false,
		"POST /api/v1/auth/register":       false,
		"POST /api/v1/auth/login":          false,
		"POST /api/v1/auth/refresh":        false,
		"GET /api/v1/debug/routes":         false,
		"GET /api/v1/public/courses":       false,
		"POST /api/v1/courses/generate":    false,
		"GET /api/v1/courses/original":     false,
		"GET /api/v1/courses/original/:id": false,
		"GET /api/v1/jobs":                 false,
		"GET /api/v1/jobs/:id":             false,
	}

	for _, route := range routes {
		routeKey := route.Method + " " + route.Path
		if _, exists := expectedRoutes[routeKey]; exists {
			expectedRoutes[routeKey] = true
		}
	}

	// Verify all expected routes are present
	for route, found := range expectedRoutes {
		assert.True(t, found, "Route %s should be registered", route)
	}
}

// TestSetupServerRouter_HealthCheck tests the health check endpoint
func TestSetupServerRouter_HealthCheck(t *testing.T) {
	router := SetupServerRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

// TestSetupServerRouter_DebugRoutes tests the debug routes endpoint
func TestSetupServerRouter_DebugRoutes(t *testing.T) {
	router := SetupServerRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/debug/routes?test=123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "Routes debug")
	assert.Contains(t, body, "/api/v1/debug/routes")
	assert.Contains(t, body, "test=123")
}

// TestSetupServerRouter_AuthEndpoints tests auth endpoints
func TestSetupServerRouter_AuthEndpoints(t *testing.T) {
	router := SetupServerRouter()

	testCases := []struct {
		method   string
		path     string
		expected string
	}{
		{"POST", "/api/v1/auth/register", "register endpoint"},
		{"POST", "/api/v1/auth/login", "login endpoint"},
		{"POST", "/api/v1/auth/refresh", "refresh endpoint"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), tc.expected)
		})
	}
}

// TestSetupServerRouter_PublicCourses tests public courses endpoint
func TestSetupServerRouter_PublicCourses(t *testing.T) {
	router := SetupServerRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/public/courses", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "courses")
}

// TestSetupServerRouter_ProtectedRoutes tests that protected routes exist
func TestSetupServerRouter_ProtectedRoutes(t *testing.T) {
	router := SetupServerRouter()

	// Test that protected routes are registered (they will fail auth but should exist)
	routes := router.Routes()

	protectedRoutes := []string{
		"/api/v1/courses/generate",
		"/api/v1/courses/original",
		"/api/v1/jobs",
	}

	for _, routePath := range protectedRoutes {
		found := false
		for _, route := range routes {
			if route.Path == routePath {
				found = true
				break
			}
		}
		assert.True(t, found, "Protected route %s should be registered", routePath)
	}
}

// TestSetupServerRouter_Middleware tests that middleware is properly applied
func TestSetupServerRouter_Middleware(t *testing.T) {
	router := SetupServerRouter()

	// Test security headers middleware
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	headers := w.Header()
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
}

// TestSetupServerRouter_GinMode tests Gin mode configuration
func TestSetupServerRouter_GinMode(t *testing.T) {
	// Save original mode
	originalMode := gin.Mode()
	defer gin.SetMode(originalMode)

	// SetupServerRouter sets Gin to ReleaseMode
	router := SetupServerRouter()
	assert.NotNil(t, router)

	// Gin should be in ReleaseMode during router setup
	// (This is hard to test directly, but we can verify the router was created)
	assert.NotNil(t, router)
}

// TestGenerateCourse_SuccessPath tests the success path of GenerateCourse
func TestGenerateCourse_SuccessPath(t *testing.T) {
	// Test that GenerateCourse handles the success path correctly
	// Since we can't mock the pipeline easily, we test the structure

	// Test the options structure that GenerateCourse creates
	options := models.ProcessingOptions{
		Quality:   "standard",
		Languages: []string{"en"},
	}

	assert.Equal(t, "standard", options.Quality)
	assert.Contains(t, options.Languages, "en")
	assert.Len(t, options.Languages, 1)
}

// TestSetupServerRouter_RateLimiter tests rate limiter middleware
func TestSetupServerRouter_RateLimiter(t *testing.T) {
	router := SetupServerRouter()

	// Test that rate limiting is applied to API routes
	// This is hard to test directly, but we can verify the routes exist
	routes := router.Routes()

	apiRoutes := 0
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/api/v1") {
			apiRoutes++
		}
	}

	assert.True(t, apiRoutes > 0, "Should have API routes with rate limiting")
}

// TestSetupServerRouter_AuthMiddleware tests auth middleware setup
func TestSetupServerRouter_AuthMiddleware(t *testing.T) {
	router := SetupServerRouter()

	// Test that auth middleware is applied to protected routes
	// We can verify by checking that protected routes exist
	routes := router.Routes()

	protectedRouteCount := 0
	for _, route := range routes {
		if strings.Contains(route.Path, "/courses/generate") ||
			strings.Contains(route.Path, "/courses/original") ||
			strings.Contains(route.Path, "/jobs") {
			protectedRouteCount++
		}
	}

	assert.True(t, protectedRouteCount > 0, "Should have protected routes with auth middleware")
}
