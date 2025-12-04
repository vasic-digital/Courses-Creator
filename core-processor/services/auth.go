package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/course-creator/core-processor/middleware"
	"github.com/course-creator/core-processor/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication and user management
type AuthService struct {
	db          *gorm.DB
	auth        *middleware.AuthMiddleware
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB, auth *middleware.AuthMiddleware) *AuthService {
	return &AuthService{
		db:   db,
		auth: auth,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Password  string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Validate login input for security issues
	if err := ValidateLoginInput(req.Email, req.Password); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	
	// Validate password strength
	if err := ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("weak password: %w", err)
	}
	
	// Check if user already exists
	var existingUser models.UserDB
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.UserDB{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "creator", // Default role for new registrations
		Active:    true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default preferences
	preferences := &models.UserPreferencesDB{
		ID:              uuid.New().String(),
		UserID:          user.ID,
		Voice:           "default",
		BackgroundStyle: "gradient",
		Quality:         "standard",
		Language:        "en",
		Preferences:     "{}", // Empty JSON object
	}

	if err := s.db.Create(preferences).Error; err != nil {
		// Log error but don't fail registration
		fmt.Printf("Warning: failed to create default preferences for user %s: %v\n", user.ID, err)
	}

	// Generate tokens
	userModel := s.convertToUserModel(user)
	accessToken, refreshToken, err := s.auth.GenerateToken(userModel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	if err := s.createSession(ctx, userModel, accessToken, refreshToken); err != nil {
		fmt.Printf("Warning: failed to create session for user %s: %v\n", user.ID, err)
	}

	return &AuthResponse{
		User:         userModel,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(24 * time.Hour.Seconds()),
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Validate login input for security issues
	if err := ValidateLoginInput(req.Email, req.Password); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	
	// Find user
	var user models.UserDB
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.Active {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate tokens
	userModel := s.convertToUserModel(&user)
	accessToken, refreshToken, err := s.auth.GenerateToken(userModel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create or update session
	if err := s.createSession(ctx, userModel, accessToken, refreshToken); err != nil {
		fmt.Printf("Warning: failed to create session for user %s: %v\n", user.ID, err)
	}

	return &AuthResponse{
		User:         userModel,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(24 * time.Hour.Seconds()),
	}, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	newAccessToken, err := s.auth.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user from token to create response
	claims, err := s.auth.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Find user
	var user models.UserDB
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	userModel := s.convertToUserModel(&user)

	return &AuthResponse{
		User:         userModel,
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(24 * time.Hour.Seconds()),
	}, nil
}

// Logout invalidates user session
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	// Delete all user sessions
	if err := s.db.Where("user_id = ?", userID).Delete(&models.UserSessionDB{}).Error; err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.UserDB
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return s.convertToUserModel(&user), nil
}

// UpdateUser updates user information
func (s *AuthService) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) (*models.User, error) {
	var user models.UserDB
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Don't allow password updates through this method
	if _, exists := updates["password"]; exists {
		return nil, fmt.Errorf("password updates should use the dedicated password update method")
	}

	// Update allowed fields
	allowedFields := []string{"first_name", "last_name", "active", "role"}
	for _, field := range allowedFields {
		if value, exists := updates[field]; exists {
			s.db.Model(&user).Update(field, value)
		}
	}

	// Refresh user data
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to refresh user data: %w", err)
	}

	return s.convertToUserModel(&user), nil
}

// UpdatePassword updates user password
func (s *AuthService) UpdatePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Validate new password strength
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("weak password: %w", err)
	}
	
	var user models.UserDB
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions (force re-login)
	if err := s.Logout(ctx, userID); err != nil {
		fmt.Printf("Warning: failed to invalidate sessions after password update: %v\n", err)
	}

	return nil
}

// createSession creates a new user session
func (s *AuthService) createSession(ctx context.Context, user *models.User, accessToken, refreshToken string) error {
	// Get context information
	var ipAddress, userAgent string
	if ginCtx, ok := ctx.(*gin.Context); ok {
		ipAddress = ginCtx.ClientIP()
		userAgent = ginCtx.GetHeader("User-Agent")
	}

	// Hash access token for storage
	tokenHash := sha256.Sum256([]byte(accessToken))
	tokenHashStr := fmt.Sprintf("%x", tokenHash)

	// Create session
	session := &models.UserSessionDB{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		TokenHash:    tokenHashStr,
		RefreshToken: &refreshToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		LastActivity: time.Now(),
	}

	// Delete existing sessions for this user (single session per user for now)
	s.db.Where("user_id = ?", user.ID).Delete(&models.UserSessionDB{})

	// Create new session
	if err := s.db.Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// convertToUserModel converts database model to service model
func (s *AuthService) convertToUserModel(user *models.UserDB) *models.User {
	return &models.User{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// CreateUserByAdmin creates a user with admin privileges
func (s *AuthService) CreateUserByAdmin(ctx context.Context, req *RegisterRequest, role string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.UserDB
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Validate role
	validRoles := map[string]bool{"admin": true, "creator": true, "viewer": true}
	if !validRoles[role] {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	// Create user
	user := &models.UserDB{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
		Active:    true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default preferences
	preferences := &models.UserPreferencesDB{
		ID:              uuid.New().String(),
		UserID:          user.ID,
		Voice:           "default",
		BackgroundStyle: "gradient",
		Quality:         "standard",
		Language:        "en",
		Preferences:     "{}",
	}

	s.db.Create(preferences)

	return s.convertToUserModel(user), nil
}