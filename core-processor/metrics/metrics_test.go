package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	// Test that Init doesn't panic
	assert.NotPanics(t, func() {
		Init()
	})
}

func TestMiddleware(t *testing.T) {
	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Apply middleware
	middleware := Middleware()
	assert.NotNil(t, middleware)

	// Test that middleware doesn't panic
	assert.NotPanics(t, func() {
		middleware(c)
	})
}

func TestHandler(t *testing.T) {
	handler := Handler()
	assert.NotNil(t, handler)

	// Test that handler responds
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecordJobCompletion(t *testing.T) {
	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		duration := 5 * time.Second
		RecordJobCompletion("completed", "course_generation", duration)
	})
}

func TestRecordCourseGeneration(t *testing.T) {
	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		duration := 120 * time.Second
		RecordCourseGeneration("success", "high", duration)
	})
}

func TestUpdateActiveConnections(t *testing.T) {
	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		UpdateActiveConnections(42)
	})
}

func TestUpdateStorageUsage(t *testing.T) {
	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		bytes := int64(1024 * 1024 * 100) // 100MB
		UpdateStorageUsage("local", bytes)
	})
}

func TestMiddleware_WithDifferentStatusCodes(t *testing.T) {
	// Test different status codes
	statusCodes := []int{200, 404, 500}

	for _, status := range statusCodes {
		t.Run(string(rune(status)), func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Set status code
			w.WriteHeader(status)

			middleware := Middleware()
			assert.NotPanics(t, func() {
				middleware(c)
			})
		})
	}
}

func TestMiddleware_WithDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(method, "/test", nil)

			middleware := Middleware()
			assert.NotPanics(t, func() {
				middleware(c)
			})
		})
	}
}

func TestRecordJobCompletion_WithDifferentStatuses(t *testing.T) {
	statuses := []string{"completed", "failed", "cancelled"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			assert.NotPanics(t, func() {
				RecordJobCompletion(status, "test_type", time.Second)
			})
		})
	}
}

func TestRecordCourseGeneration_WithDifferentQualities(t *testing.T) {
	qualities := []string{"standard", "high"}

	for _, quality := range qualities {
		t.Run(quality, func(t *testing.T) {
			assert.NotPanics(t, func() {
				RecordCourseGeneration("success", quality, time.Minute)
			})
		})
	}
}

func TestUpdateStorageUsage_WithDifferentTypes(t *testing.T) {
	types := []string{"local", "s3", "azure"}

	for _, storageType := range types {
		t.Run(storageType, func(t *testing.T) {
			assert.NotPanics(t, func() {
				bytes := int64(1024)
				UpdateStorageUsage(storageType, bytes)
			})
		})
	}
}
