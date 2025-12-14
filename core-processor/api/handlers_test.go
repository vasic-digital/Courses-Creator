package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-creator/core-processor/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database
	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	// Create handler
	handler := NewCourseHandler(db)

	// Create test request
	req, err := http.NewRequest("GET", "/api/health", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(rr)
	c.Request = req

	// Call handler
	handler.HealthCheck(c)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
}

func TestGetCourse(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database
	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	// Create handler
	handler := NewCourseHandler(db)

	// Test 1: Course not found
	req, err := http.NewRequest("GET", "/api/courses/nonexistent", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}

	handler.GetCourse(c)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Test 2: Another non-existent course (handler doesn't validate UUID format)
	req, err = http.NewRequest("GET", "/api/courses/invalid-uuid", nil)
	require.NoError(t, err)

	rr = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(rr)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "invalid-uuid"}}

	handler.GetCourse(c)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestListCourses(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database
	db, err := database.NewDatabase(database.DefaultConfig())
	require.NoError(t, err)
	defer db.Close()

	// Create handler
	handler := NewCourseHandler(db)

	// Test: Empty list
	req, err := http.NewRequest("GET", "/api/courses", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = req

	handler.ListCourses(c)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	courses, ok := response["courses"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, courses)
}
