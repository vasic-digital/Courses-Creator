package security_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/course-creator/core-processor/cmd"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/services"
	"github.com/gorilla/mux"
)

func TestSQLInjection(t *testing.T) {
	// Test SQL injection attempts on various endpoints
	router := mux.NewRouter()
	authService := services.NewAuthService()
	
	// Mock database for testing
	// In real implementation, this would be connected to test DB
	
	testCases := []struct {
		name        string
		input       string
		shouldBlock bool
	}{
		{
			name:        "Valid email",
			input:       "user@example.com",
			shouldBlock: false,
		},
		{
			name:        "SQL injection in email",
			input:       "user@example.com'; DROP TABLE users; --",
			shouldBlock: true,
		},
		{
			name:        "SQL injection with quotes",
			input:       "' OR '1'='1",
			shouldBlock: true,
		},
		{
			name:        "SQL injection with UNION",
			input:       "user@example.com' UNION SELECT * FROM users --",
			shouldBlock: true,
		},
		{
			name:        "XSS attempt in email",
			input:       "<script>alert('xss')</script>@example.com",
			shouldBlock: false, // Email validation should handle this
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test login endpoint with malicious input
			loginData := map[string]string{
				"email":    tc.input,
				"password": "password123",
			}
			
			jsonData, _ := json.Marshal(loginData)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// This would be the actual handler in implementation
			// For now, we'll test the validation logic directly
			err := authService.ValidateLoginInput(tc.input, "password123")
			
			if tc.shouldBlock && err == nil {
				t.Errorf("Expected SQL injection to be blocked but was allowed: %s", tc.input)
			}
			
			if !tc.shouldBlock && err != nil {
				t.Errorf("Expected valid input to be allowed but was blocked: %s", tc.input)
			}
		})
	}
}

func TestXSSPrevention(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		shouldBlock bool
	}{
		{
			name:        "Valid text",
			input:       "This is valid course content",
			shouldBlock: false,
		},
		{
			name:        "Script tag injection",
			input:       "<script>alert('xss')</script>",
			shouldBlock: true,
		},
		{
			name:        "JavaScript URI",
			input:       "javascript:alert('xss')",
			shouldBlock: true,
		},
		{
			name:        "On event handler",
			input:       "<img src=x onerror=alert('xss')>",
			shouldBlock: true,
		},
		{
			name:        "IFrame injection",
			input:       "<iframe src=\"javascript:alert('xss')\"></iframe>",
			shouldBlock: true,
		},
		{
			name:        "SVG injection",
			input:       "<svg onload=alert('xss')>",
			shouldBlock: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test content validation
			isValid := services.ValidateContent(tc.input)
			
			if tc.shouldBlock && isValid {
				t.Errorf("Expected XSS attempt to be blocked but was allowed: %s", tc.input)
			}
			
			if !tc.shouldBlock && !isValid {
				t.Errorf("Expected valid content to be allowed but was blocked: %s", tc.input)
			}
		})
	}
}

func TestAuthenticationSecurity(t *testing.T) {
	authService := services.NewAuthService()

	testCases := []struct {
		name        string
		email       string
		password    string
		expectError bool
	}{
		{
			name:        "Valid credentials",
			email:       "user@example.com",
			password:    "validPassword123!",
			expectError:  false,
		},
		{
			name:        "Weak password",
			email:       "user@example.com",
			password:    "123",
			expectError:  true,
		},
		{
			name:        "Common password",
			email:       "user@example.com",
			password:    "password",
			expectError:  true,
		},
		{
			name:        "Password without numbers",
			email:       "user@example.com",
			password:    "passwordOnly",
			expectError:  true,
		},
		{
			name:        "Password without uppercase",
			email:       "user@example.com",
			password:    "password123!",
			expectError:  true,
		},
		{
			name:        "Password without special chars",
			email:       "user@example.com",
			password:    "Password123",
			expectError:  true,
		},
		{
			name:        "Valid strong password",
			email:       "user@example.com",
			password:    "StrongP@ssw0rd123!",
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := authService.ValidatePasswordStrength(tc.password)
			
			if tc.expectError && err == nil {
				t.Errorf("Expected weak password to be rejected but was accepted: %s", tc.password)
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Expected strong password to be accepted but was rejected: %s", tc.password)
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	// Test rate limiting on sensitive endpoints
	router := mux.NewRouter()
	
	// Mock rate limiter
	rateLimiter := services.NewRateLimiter(5, time.Minute) // 5 requests per minute
	
	testCases := []struct {
		name        string
		numRequests int
		shouldLimit bool
	}{
		{
			name:        "Within limit",
			numRequests: 3,
			shouldLimit: false,
		},
		{
			name:        "At limit",
			numRequests: 5,
			shouldLimit: false,
		},
		{
			name:        "Exceeds limit",
			numRequests: 7,
			shouldLimit: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipAddress := "192.168.1.1"
			
			for i := 0; i < tc.numRequests; i++ {
				allowed := rateLimiter.Allow(ipAddress)
				
				if i < tc.numRequests-1 {
					// Should allow all but potentially the last request
					if !allowed {
						t.Errorf("Rate limiter blocked request %d unexpectedly", i+1)
					}
				} else {
					// Last request might be blocked
					if tc.shouldLimit && allowed {
						t.Errorf("Rate limiter should have blocked request %d but allowed it", i+1)
					}
					if !tc.shouldLimit && !allowed {
						t.Errorf("Rate limiter should have allowed request %d but blocked it", i+1)
					}
				}
			}
		})
	}
}

func TestCSRFProtection(t *testing.T) {
	// Test CSRF token generation and validation
	csrfService := services.NewCSRFService()

	testCases := []struct {
		name        string
		token       string
		sessionID   string
		shouldPass  bool
	}{
		{
			name:       "Valid CSRF token",
			token:      "",
			sessionID:  "session123",
			shouldPass: true,
		},
		{
			name:        "Invalid CSRF token",
			token:       "invalid_token",
			sessionID:   "session123",
			shouldPass:  false,
		},
		{
			name:        "Empty token",
			token:       "",
			sessionID:   "session123",
			shouldPass:  false,
		},
		{
			name:        "Token from different session",
			token:       "",
			sessionID:   "different_session",
			shouldPass:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.token == "" && tc.shouldPass {
				// Generate valid token
				tc.token = csrfService.GenerateToken(tc.sessionID)
			}

			valid := csrfService.ValidateToken(tc.token, tc.sessionID)
			
			if tc.shouldPass && !valid {
				t.Errorf("Expected valid CSRF token to pass validation")
			}
			
			if !tc.shouldPass && valid {
				t.Errorf("Expected invalid CSRF token to fail validation")
			}
		})
	}
}

func TestFileUploadSecurity(t *testing.T) {
	testCases := []struct {
		name        string
		filename    string
		content     []byte
		contentType string
		shouldAllow bool
	}{
		{
			name:        "Valid image",
			filename:    "image.jpg",
			content:     []byte{0xFF, 0xD8, 0xFF}, // JPEG header
			contentType: "image/jpeg",
			shouldAllow: true,
		},
		{
			name:        "Executable file",
			filename:    "malware.exe",
			content:     []byte{0x4D, 0x5A}, // MZ header
			contentType: "application/x-msdownload",
			shouldAllow: false,
		},
		{
			name:        "Script file",
			filename:    "script.js",
			content:     []byte("<script>alert('xss')</script>"),
			contentType: "application/javascript",
			shouldAllow: false,
		},
		{
			name:        "PHP file",
			filename:    "shell.php",
			content:     []byte("<?php echo 'shell'; ?>"),
			contentType: "application/x-php",
			shouldAllow: false,
		},
		{
			name:        "Large file",
			filename:    "large.jpg",
			content:     make([]byte, 100*1024*1024), // 100MB
			contentType: "image/jpeg",
			shouldAllow: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := models.UploadFile{
				Filename:    tc.filename,
				Content:     tc.content,
				ContentType: tc.contentType,
			}

			allowed := services.ValidateFileUpload(file)
			
			if tc.shouldAllow && !allowed {
				t.Errorf("Expected file upload to be allowed but was blocked: %s", tc.filename)
			}
			
			if !tc.shouldAllow && allowed {
				t.Errorf("Expected file upload to be blocked but was allowed: %s", tc.filename)
			}
		})
	}
}

func TestInputValidation(t *testing.T) {
	testCases := []struct {
		name        string
		field       string
		value       string
		shouldAllow bool
	}{
		{
			name:        "Valid course title",
			field:       "title",
			value:       "Introduction to Machine Learning",
			shouldAllow: true,
		},
		{
			name:        "Empty course title",
			field:       "title",
			value:       "",
			shouldAllow: false,
		},
		{
			name:        "Too long course title",
			field:       "title",
			value:       strings.Repeat("A", 300),
			shouldAllow: false,
		},
		{
			name:        "Valid user ID",
			field:       "userId",
			value:       "user123",
			shouldAllow: true,
		},
		{
			name:        "Invalid user ID with special chars",
			field:       "userId",
			value:       "user@#$",
			shouldAllow: false,
		},
		{
			name:        "Valid job ID",
			field:       "jobId",
			value:       "job-12345",
			shouldAllow: true,
		},
		{
			name:        "Invalid job ID",
			field:       "jobId",
			value:       "../../etc/passwd",
			shouldAllow: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allowed := services.ValidateInputField(tc.field, tc.value)
			
			if tc.shouldAllow && !allowed {
				t.Errorf("Expected input to be allowed but was blocked: %s = %s", tc.field, tc.value)
			}
			
			if !tc.shouldAllow && allowed {
				t.Errorf("Expected input to be blocked but was allowed: %s = %s", tc.field, tc.value)
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	// Test that security headers are properly set
	router := cmd.SetupRouter()
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()

	// Check for security headers
	requiredHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range requiredHeaders {
		actualValue := resp.Header.Get(header)
		if actualValue == "" {
			t.Errorf("Missing security header: %s", header)
		} else if actualValue != expectedValue {
			t.Errorf("Incorrect security header value for %s: got %s, want %s", 
				header, actualValue, expectedValue)
		}
	}
}

func TestPasswordHashing(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{
			name:     "Strong password",
			password: "StrongP@ssw0rd123!",
		},
		{
			name:     "Medium password",
			password: "Password123",
		},
		{
			name:     "Simple password",
			password: "simple123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Hash password
			hash, err := services.HashPassword(tc.password)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			// Verify hash is not empty
			if hash == "" {
				t.Error("Hashed password is empty")
			}

			// Verify hash is not the same as password
			if hash == tc.password {
				t.Error("Hashed password is the same as original password")
			}

			// Verify password can be verified
			valid := services.VerifyPassword(tc.password, hash)
			if !valid {
				t.Error("Failed to verify correct password")
			}

			// Verify wrong password is rejected
			valid = services.VerifyPassword("wrongpassword", hash)
			if valid {
				t.Error("Verified incorrect password")
			}

			// Verify two hashes of the same password are different (due to salt)
			hash2, err := services.HashPassword(tc.password)
			if err != nil {
				t.Fatalf("Failed to hash second password: %v", err)
			}

			if hash == hash2 {
				t.Error("Two hashes of the same password are identical (missing salt)")
			}
		})
	}
}