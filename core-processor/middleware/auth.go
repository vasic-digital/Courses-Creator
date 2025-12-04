package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/course-creator/core-processor/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	secretKey     string
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
	jwtIssuer     string
	jwtAudience   string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		secretKey:     getSecretKey(),
		tokenExpiry:   24 * time.Hour,     // Access token expires in 24 hours
		refreshExpiry: 7 * 24 * time.Hour, // Refresh token expires in 7 days
		jwtIssuer:     "course-creator-api",
		jwtAudience:   "course-creator-users",
	}
}

// GenerateToken generates a new JWT token for a user
func (am *AuthMiddleware) GenerateToken(user *models.User) (string, string, error) {
	// Create claims
	claims := JWTClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: getUserPermissions(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    am.jwtIssuer,
			Audience:  []string{am.jwtAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(am.tokenExpiry)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(am.secretKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (longer expiry)
	refreshClaims := claims
	refreshClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(am.refreshExpiry))
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(am.secretKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshTokenString, nil
}

// ValidateToken validates a JWT token and returns claims
func (am *AuthMiddleware) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Additional validation
		now := time.Now()
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(now) {
			return nil, fmt.Errorf("token has expired")
		}
		if claims.NotBefore != nil && claims.NotBefore.Time.After(now) {
			return nil, fmt.Errorf("token not yet valid")
		}
		if claims.IssuedAt != nil && claims.IssuedAt.Time.After(now) {
			return nil, fmt.Errorf("token issued in the future")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RequireAuth middleware requires authentication
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format, expected 'Bearer <token>'",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := am.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"code":    "INVALID_TOKEN",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_permissions", claims.Permissions)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RequireRole middleware requires specific role
func (am *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		if userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Requires %s role, current role: %s", requiredRole, userRole),
				"code":  "INSUFFICIENT_ROLE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware requires specific permission
func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("user_permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		userPermissions, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid permissions format",
				"code":  "INVALID_PERMISSIONS",
			})
			c.Abort()
			return
		}

		// Check if user has required permission
		hasPermission := false
		for _, p := range userPermissions {
			if p == permission || p == "*" { // Wildcard permission
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Permission '%s' required", permission),
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RefreshToken generates a new access token from a refresh token
func (am *AuthMiddleware) RefreshToken(refreshToken string) (string, error) {
	claims, err := am.ValidateToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Create new claims with updated expiry
	newClaims := JWTClaims{
		UserID:      claims.UserID,
		Email:       claims.Email,
		Role:        claims.Role,
		Permissions: claims.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    am.jwtIssuer,
			Audience:  []string{am.jwtAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(am.tokenExpiry)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate new access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	accessToken, err := token.SignedString([]byte(am.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessToken, nil
}

// GetUserFromContext extracts user info from gin context
func GetUserFromContext(c *gin.Context) (*models.User, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")
	permissions, _ := c.Get("user_permissions")

	user := &models.User{
		ID:    userID.(string),
		Email: email.(string),
		Role:  role.(string),
	}

	// Store permissions in a custom field (not part of User model normally)
	if perms, ok := permissions.([]string); ok {
		user.CustomData = map[string]interface{}{
			"permissions": perms,
		}
	}

	return user, nil
}

// getUserPermissions returns permissions based on user role
func getUserPermissions(role string) []string {
	permissionMap := map[string][]string{
		"admin": {
			"courses:read", "courses:write", "courses:delete",
			"users:read", "users:write", "users:delete",
			"system:read", "system:write", "system:admin",
			"jobs:read", "jobs:write", "jobs:delete",
			"*", // Admin has all permissions
		},
		"creator": {
			"courses:read", "courses:write", "courses:delete",
			"jobs:read", "jobs:write",
			"system:read",
		},
		"viewer": {
			"courses:read",
			"jobs:read",
		},
	}

	if permissions, exists := permissionMap[role]; exists {
		return permissions
	}

	// Default to viewer permissions
	return permissionMap["viewer"]
}

// getSecretKey retrieves JWT secret key from environment
func getSecretKey() string {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		secret = "change-this-default-secret-key-in-production"
	}
	return secret
}

// RateLimiter middleware for API rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check current requests
		now := time.Now()
		requests, exists := rl.requests[clientIP]

		if !exists {
			rl.requests[clientIP] = []time.Time{now}
			c.Next()
			return
		}

		// Filter requests within the window
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Check if limit exceeded
		if len(validRequests) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":  "Rate limit exceeded",
				"code":   "RATE_LIMIT_EXCEEDED",
				"limit":  rl.limit,
				"window": rl.window.String(),
			})
			c.Abort()
			return
		}

		// Add current request
		rl.requests[clientIP] = append(validRequests, now)
		c.Next()
	}
}

// cleanup removes old requests from memory
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for clientIP, requests := range rl.requests {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			if len(validRequests) == 0 {
				delete(rl.requests, clientIP)
			} else {
				rl.requests[clientIP] = validRequests
			}
		}
	}
}
