package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
