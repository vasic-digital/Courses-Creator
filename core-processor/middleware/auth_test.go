package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/course-creator/core-processor/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthMiddleware(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	assert.NotNil(t, am)
	assert.Equal(t, "test-secret-key", am.secretKey)
	assert.Equal(t, 24*time.Hour, am.tokenExpiry)
	assert.Equal(t, 7*24*time.Hour, am.refreshExpiry)
	assert.Equal(t, "course-creator-api", am.jwtIssuer)
	assert.Equal(t, "course-creator-users", am.jwtAudience)
}

func TestGenerateToken(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Role:  "creator",
	}

	accessToken, refreshToken, err := am.GenerateToken(user)

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// Verify the access token can be validated
	claims, err := am.ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "creator", claims.Role)
	assert.Contains(t, claims.Permissions, "courses:read")
	assert.Contains(t, claims.Permissions, "courses:write")
}

func TestValidateToken(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Role:  "creator",
	}

	// Generate a token
	accessToken, _, err := am.GenerateToken(user)
	require.NoError(t, err)

	// Validate the token
	claims, err := am.ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "creator", claims.Role)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	// Try to validate an invalid token
	_, err := am.ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	// Call the middleware without Authorization header
	am.RequireAuth()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat token123")

	// Call the middleware
	am.RequireAuth()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_ValidToken(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Role:  "creator",
	}

	// Generate a token
	accessToken, _, err := am.GenerateToken(user)
	require.NoError(t, err)

	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+accessToken)

	// Call the middleware
	am.RequireAuth()(c)

	assert.False(t, c.IsAborted())

	// Check that user info was set in context
	userID, exists := c.Get("user_id")
	assert.True(t, exists)
	assert.Equal(t, "user123", userID)

	userEmail, exists := c.Get("user_email")
	assert.True(t, exists)
	assert.Equal(t, "test@example.com", userEmail)

	userRole, exists := c.Get("user_role")
	assert.True(t, exists)
	assert.Equal(t, "creator", userRole)
}

func TestRequireRole_SufficientRole(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Role:  "admin",
	}

	// Generate a token
	accessToken, _, err := am.GenerateToken(user)
	require.NoError(t, err)

	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+accessToken)

	// Set user info in context (simulating RequireAuth)
	c.Set("user_role", "admin")

	// Call the role middleware
	am.RequireRole("admin")(c)

	assert.False(t, c.IsAborted())
}

func TestRequireRole_InsufficientRole(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user info in context with insufficient role
	c.Set("user_role", "viewer")

	// Call the role middleware
	am.RequireRole("admin")(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_SufficientPermission(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Role:  "creator",
	}

	// Generate a token
	accessToken, _, err := am.GenerateToken(user)
	require.NoError(t, err)

	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+accessToken)

	// Set user info in context (simulating RequireAuth)
	permissions := []string{"courses:read", "courses:write"}
	c.Set("user_permissions", permissions)

	// Call the permission middleware
	am.RequirePermission("courses:read")(c)

	assert.False(t, c.IsAborted())
}

func TestRequirePermission_InsufficientPermission(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	// Create a test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user info in context with insufficient permissions
	permissions := []string{"courses:read"}
	c.Set("user_permissions", permissions)

	// Call the permission middleware
	am.RequirePermission("system:admin")(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRefreshToken(t *testing.T) {
	// Set a test secret key
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	am := NewAuthMiddleware()

	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Role:  "creator",
	}

	// Generate tokens
	_, refreshToken, err := am.GenerateToken(user)
	require.NoError(t, err)

	// Refresh the token
	newAccessToken, err := am.RefreshToken(refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// Verify the new token
	claims, err := am.ValidateToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "creator", claims.Role)
}

func TestGetUserPermissions(t *testing.T) {
	tests := []struct {
		role          string
		expectedLen   int
		hasPermission string
	}{
		{"admin", 13, "system:admin"},
		{"creator", 6, "courses:write"},
		{"viewer", 2, "courses:read"},
		{"unknown", 2, "courses:read"}, // defaults to viewer
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			permissions := getUserPermissions(tt.role)
			assert.Len(t, permissions, tt.expectedLen)
			assert.Contains(t, permissions, tt.hasPermission)
		})
	}
}

func TestGetSecretKey(t *testing.T) {
	// Test with environment variable set
	os.Setenv("JWT_SECRET_KEY", "custom-secret")
	defer os.Unsetenv("JWT_SECRET_KEY")

	secret := getSecretKey()
	assert.Equal(t, "custom-secret", secret)

	// Test with environment variable not set
	os.Unsetenv("JWT_SECRET_KEY")
	secret = getSecretKey()
	assert.Equal(t, "change-this-default-secret-key-in-production", secret)
}
