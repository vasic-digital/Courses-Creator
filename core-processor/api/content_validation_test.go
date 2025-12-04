package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentSecurityValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup test database
	dbConfig := database.DefaultConfig()
	db, err := database.NewDatabase(dbConfig)
	require.NoError(t, err)
	defer db.Close()
	
	// Setup services and handlers
	auth := middleware.NewAuthMiddleware()
	authService := services.NewAuthService(db.GetGormDB(), auth)
	authHandler := NewAuthHandler(authService, auth)
	
	courseHandler := NewCourseHandler(db)
	courseAPIService := NewCourseAPIService(courseHandler)
	
	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	})
	
	v1 := router.Group("/api/v1")
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
	}
	
	courseGroup := v1.Group("/public")
	courseAPIService.RegisterCourseAPIRoutes(courseGroup)
	
	t.Run("Valid content should be accepted", func(t *testing.T) {
		// Valid markdown content
		validMarkdown := `# Test Course

This is a valid course content with no XSS or injection attempts.

## Introduction

Welcome to the course! This is just normal text content.

## Main Content

Here's some more content with *italics* and **bold** text.`
		
		reqBody := map[string]interface{}{
			"markdown": validMarkdown,
			"options": map[string]string{
				"quality": "standard",
			},
		}
		
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/public/courses/generate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Should not be rejected for security reasons (might fail for other reasons)
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
		if w.Code == http.StatusBadRequest {
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.NotContains(t, response["error"], "Invalid content detected")
		}
	})
	
	t.Run("XSS content should be rejected", func(t *testing.T) {
		// Malicious content with XSS
		xssMarkdown := `# Test Course

<script>alert('XSS');</script>

This content has malicious scripts.

javascript:alert('XSS')

<img src=x onerror=alert('XSS')>`
		
		reqBody := map[string]interface{}{
			"markdown": xssMarkdown,
			"options": map[string]string{
				"quality": "standard",
			},
		}
		
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/public/courses/generate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Should be rejected for security reasons
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid content detected")
	})
}