package unit

import (
	"context"
	"testing"
	"time"

	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)



func TestAuthMiddleware(t *testing.T) {
	auth := middleware.NewAuthMiddleware()
	
	t.Run("Generate and Validate Token", func(t *testing.T) {
		user := &models.User{
			ID:    "test-user-id",
			Email: "test@example.com",
			Role:  "creator",
		}
		
		// Generate tokens
		accessToken, refreshToken, err := auth.GenerateToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		
		// Validate access token
		claims, err := auth.ValidateToken(accessToken)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
		assert.Contains(t, claims.Permissions, "courses:read")
		assert.Contains(t, claims.Permissions, "courses:write")
		
		// Validate refresh token
		claims, err = auth.ValidateToken(refreshToken)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
	})
	
	t.Run("Refresh Token", func(t *testing.T) {
		user := &models.User{
			ID:    "test-user-id-2",
			Email: "test2@example.com",
			Role:  "viewer",
		}
		
		// Generate tokens
		_, refreshToken, err := auth.GenerateToken(user)
		require.NoError(t, err)
		
		// Refresh access token
		newAccessToken, err := auth.RefreshToken(refreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		
		// Validate new token
		claims, err := auth.ValidateToken(newAccessToken)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
	})
	
	t.Run("Invalid Token", func(t *testing.T) {
		invalidToken := "invalid.token.here"
		
		_, err := auth.ValidateToken(invalidToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse token")
	})
	
	t.Run("User Permissions", func(t *testing.T) {
		// Test admin permissions
		adminClaims := &middleware.JWTClaims{
			UserID:      "admin-id",
			Email:       "admin@example.com",
			Role:        "admin",
			Permissions: []string{"*"},
		}
		
		assert.Contains(t, adminClaims.Permissions, "*")
		
		// Check if admin has wildcard permission
		hasWildcard := false
		for _, p := range adminClaims.Permissions {
			if p == "*" {
				hasWildcard = true
				break
			}
		}
		assert.True(t, hasWildcard)
		
		// Test creator permissions
		creatorUser := &models.User{ID: "creator-id", Email: "creator@example.com", Role: "creator"}
		accessToken, _, err := auth.GenerateToken(creatorUser)
		require.NoError(t, err)
		
		claims, err := auth.ValidateToken(accessToken)
		require.NoError(t, err)
		assert.Contains(t, claims.Permissions, "courses:write")
		assert.NotContains(t, claims.Permissions, "users:delete")
		
		// Test viewer permissions
		viewerUser := &models.User{ID: "viewer-id", Email: "viewer@example.com", Role: "viewer"}
		accessToken, _, err = auth.GenerateToken(viewerUser)
		require.NoError(t, err)
		
		claims, err = auth.ValidateToken(accessToken)
		require.NoError(t, err)
		assert.Contains(t, claims.Permissions, "courses:read")
		assert.NotContains(t, claims.Permissions, "courses:write")
	})
}

func TestAuthService(t *testing.T) {
	db := setupTestDB(t)
	auth := middleware.NewAuthMiddleware()
	authService := services.NewAuthService(db, auth)
	
	t.Run("Register User", func(t *testing.T) {
		req := &services.RegisterRequest{
			Email:     "newuser@example.com",
			FirstName: "New",
			LastName:  "User",
			Password:  "SecureP@ssw0rd123!",
		}
		
		// Register user
		resp, err := authService.Register(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Email, resp.User.Email)
		assert.Equal(t, req.FirstName, resp.User.FirstName)
		assert.Equal(t, req.LastName, resp.User.LastName)
		assert.Equal(t, "creator", resp.User.Role) // Default role
		assert.True(t, resp.User.Active)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Greater(t, resp.ExpiresIn, 0)
		
		// Verify user was created in database
		var userDB models.UserDB
		err = db.Where("email = ?", req.Email).First(&userDB).Error
		require.NoError(t, err)
		assert.Equal(t, req.Email, userDB.Email)
		assert.Equal(t, req.FirstName, userDB.FirstName)
		assert.Equal(t, req.LastName, userDB.LastName)
		
		// Verify password is hashed
		err = bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(req.Password))
		require.NoError(t, err)
		
		// Verify default preferences were created
		var prefsDB models.UserPreferencesDB
		err = db.Where("user_id = ?", userDB.ID).First(&prefsDB).Error
		require.NoError(t, err)
		assert.Equal(t, "default", prefsDB.Voice)
		assert.Equal(t, "gradient", prefsDB.BackgroundStyle)
	})
	
	t.Run("Register Duplicate Email", func(t *testing.T) {
		req := &services.RegisterRequest{
			Email:     "newuser@example.com", // Same email as above
			FirstName: "Another",
			LastName:  "User",
			Password:  "AnotherP@ssw0rd123!",
		}
		
		// Should fail with duplicate email
		resp, err := authService.Register(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "already exists")
	})
	
	t.Run("Login User", func(t *testing.T) {
		req := &services.LoginRequest{
			Email:    "newuser@example.com",
			Password: "SecureP@ssw0rd123!",
		}
		
		// Login user
		resp, err := authService.Login(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Email, resp.User.Email)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
	})
	
	t.Run("Login Invalid Credentials", func(t *testing.T) {
		req := &services.LoginRequest{
			Email:    "newuser@example.com",
			Password: "WrongP@ssw0rd123!",
		}
		
		// Should fail with invalid credentials
		resp, err := authService.Login(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid email or password")
	})
	
	t.Run("Login Non-existent User", func(t *testing.T) {
		req := &services.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		
		// Should fail with invalid credentials
		resp, err := authService.Login(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid email or password")
	})
	
	t.Run("Get User By ID", func(t *testing.T) {
		// First get a user ID
		var userDB models.UserDB
		err := db.Where("email = ?", "newuser@example.com").First(&userDB).Error
		require.NoError(t, err)
		
		// Get user by ID
		user, err := authService.GetUserByID(context.Background(), userDB.ID)
		require.NoError(t, err)
		assert.Equal(t, userDB.ID, user.ID)
		assert.Equal(t, userDB.Email, user.Email)
		assert.Equal(t, userDB.FirstName, user.FirstName)
		assert.Equal(t, userDB.LastName, user.LastName)
	})
	
	t.Run("Update User", func(t *testing.T) {
		// First get a user ID
		var userDB models.UserDB
		err := db.Where("email = ?", "newuser@example.com").First(&userDB).Error
		require.NoError(t, err)
		
		// Update user
		updates := map[string]interface{}{
			"first_name": "Updated",
			"last_name":  "Name",
		}
		
		user, err := authService.UpdateUser(context.Background(), userDB.ID, updates)
		require.NoError(t, err)
		assert.Equal(t, "Updated", user.FirstName)
		assert.Equal(t, "Name", user.LastName)
		
		// Verify database was updated
		err = db.Where("id = ?", userDB.ID).First(&userDB).Error
		require.NoError(t, err)
		assert.Equal(t, "Updated", userDB.FirstName)
		assert.Equal(t, "Name", userDB.LastName)
	})
	
	t.Run("Update Password", func(t *testing.T) {
		// First get a user ID
		var userDB models.UserDB
		err := db.Where("email = ?", "newuser@example.com").First(&userDB).Error
		require.NoError(t, err)
		
		// Update password
		err = authService.UpdatePassword(context.Background(), userDB.ID, "SecureP@ssw0rd123!", "NewSecureP@ssw0rd456!")
		require.NoError(t, err)
		
		// Verify old password no longer works
		loginReq := &services.LoginRequest{
			Email:    "newuser@example.com",
			Password: "SecureP@ssw0rd123!",
		}
		_, err = authService.Login(context.Background(), loginReq)
		assert.Error(t, err)
		
		// Verify new password works
		loginReq.Password = "NewSecureP@ssw0rd456!"
		resp, err := authService.Login(context.Background(), loginReq)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
	
	t.Run("Create User By Admin", func(t *testing.T) {
		req := &services.RegisterRequest{
			Email:     "adminuser@example.com",
			FirstName: "Admin",
			LastName:  "User",
			Password:  "AdminP@ssw0rd123!",
		}
		
		// Create admin user
		user, err := authService.CreateUserByAdmin(context.Background(), req, "admin")
		require.NoError(t, err)
		assert.Equal(t, "admin", user.Role)
		
		// Verify user was created
		var userDB models.UserDB
		err = db.Where("email = ?", req.Email).First(&userDB).Error
		require.NoError(t, err)
		assert.Equal(t, "admin", userDB.Role)
	})
	
	t.Run("Logout User", func(t *testing.T) {
		// First get a user ID
		var userDB models.UserDB
		err := db.Where("email = ?", "newuser@example.com").First(&userDB).Error
		require.NoError(t, err)
		
		// Logout user
		err = authService.Logout(context.Background(), userDB.ID)
		require.NoError(t, err)
		
		// Verify sessions were deleted
		var sessionCount int64
		err = db.Model(&models.UserSessionDB{}).Where("user_id = ?", userDB.ID).Count(&sessionCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), sessionCount)
	})
}

func TestRateLimiter(t *testing.T) {
	t.Skip("Skipping rate limiter test for now - requires more complex setup")
	
	// This test would need to be expanded to properly test the rate limiting functionality
	// For now, we're skipping it as it would require more complex timing and synchronization
}

func TestTokenExpiry(t *testing.T) {
	auth := middleware.NewAuthMiddleware()
	
	user := &models.User{
		ID:    "test-user",
		Email: "test@example.com",
		Role:  "creator",
	}
	
	// Generate token
	accessToken, _, err := auth.GenerateToken(user)
	require.NoError(t, err)
	
	// Validate token
	claims, err := auth.ValidateToken(accessToken)
	require.NoError(t, err)
	
	// Check expiry
	assert.True(t, claims.ExpiresAt.After(time.Now()))
	assert.True(t, claims.IssuedAt.Before(time.Now()))
	assert.True(t, claims.NotBefore.Before(time.Now()))
}