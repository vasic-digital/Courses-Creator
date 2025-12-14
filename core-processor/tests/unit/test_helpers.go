package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/course-creator/core-processor/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Get underlying SQL DB to configure SQLite
	sqlDB, err := db.DB()
	require.NoError(t, err)
	// Enable foreign keys
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	// Configure SQLite for better concurrent access
	_, err = sqlDB.Exec("PRAGMA busy_timeout = 5000")
	require.NoError(t, err)
	_, err = sqlDB.Exec("PRAGMA journal_mode = WAL")
	require.NoError(t, err)

	// Migrate all tables
	err = db.AutoMigrate(
		&models.UserDB{},
		&models.UserPreferencesDB{},
		&models.UserSessionDB{},
		&models.CourseDB{},
		&models.JobDB{},
		&models.LessonDB{},
		&models.SubtitleDB{},
		&models.InteractiveElementDB{},
		&models.CourseMetadataDB{},
		&models.ProcessingJobDB{},
	)
	require.NoError(t, err)

	// Add BeforeCreate hooks for UUID generation
	db.Callback().Create().Before("gorm:create").Register("generate_user_id", func(db *gorm.DB) {
		if user, ok := db.Statement.Dest.(*models.UserDB); ok {
			if user.ID == "" {
				user.ID = uuid.New().String()
			}
		}
	})

	return db
}

// TestContext returns a context with timeout for tests
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// CreateTempDir creates a temporary directory for tests
func CreateTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "course-creator-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// CreateTestFile creates a test file with content
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

// MockHTTPClient creates a mock HTTP client with canned responses
type MockHTTPClient struct {
	Responses map[string]*http.Response
	Requests  []*http.Request
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Responses: make(map[string]*http.Response),
		Requests:  make([]*http.Request, 0),
	}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.Requests = append(m.Requests, req)

	key := req.Method + " " + req.URL.String()
	if resp, ok := m.Responses[key]; ok {
		return resp, nil
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString(`{"error": "not found"}`)),
		Header:     make(http.Header),
	}, nil
}

func (m *MockHTTPClient) AddResponse(method, url string, status int, body interface{}) {
	var bodyBytes []byte
	switch v := body.(type) {
	case string:
		bodyBytes = []byte(v)
	case []byte:
		bodyBytes = v
	default:
		var err error
		bodyBytes, err = json.Marshal(v)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal response body: %v", err))
		}
	}

	m.Responses[method+" "+url] = &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
		Header:     make(http.Header),
	}
}

// TestServer creates a test HTTP server
func TestServer(t *testing.T, handler http.Handler) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := server.Client()
	return server, client
}

// AssertJSONEqual asserts that two JSON strings are equal
func AssertJSONEqual(t *testing.T, expected, actual string) {
	var expectedObj, actualObj interface{}

	err := json.Unmarshal([]byte(expected), &expectedObj)
	require.NoError(t, err)

	err = json.Unmarshal([]byte(actual), &actualObj)
	require.NoError(t, err)

	assert.Equal(t, expectedObj, actualObj)
}

// Retry executes a function with retry logic
func Retry(t *testing.T, maxAttempts int, delay time.Duration, fn func() error) error {
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err != nil {
			lastErr = err
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return lastErr
}

// LoadTestData loads test data from a file
func LoadTestData(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

// CreateTestUser creates a test user object
type TestUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func CreateTestUser() TestUser {
	return TestUser{
		ID:       "test-user-123",
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "user",
	}
}

// CreateTestCourse creates a test course object
type TestCourse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AuthorID    string    `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CreateTestCourse() TestCourse {
	return TestCourse{
		ID:          "test-course-123",
		Title:       "Test Course",
		Description: "This is a test course",
		AuthorID:    "test-user-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// MockLogger provides a mock logger for tests
type MockLogger struct {
	Messages []string
	Errors   []error
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		Messages: make([]string, 0),
		Errors:   make([]error, 0),
	}
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(msg, args...))
}

func (m *MockLogger) Error(err error, msg string, args ...interface{}) {
	m.Errors = append(m.Errors, err)
	m.Messages = append(m.Messages, fmt.Sprintf(msg, args...))
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(msg, args...))
}

// CleanupTestEnv cleans up test environment variables
func CleanupTestEnv(t *testing.T, vars ...string) {
	original := make(map[string]string)
	for _, v := range vars {
		if val, ok := os.LookupEnv(v); ok {
			original[v] = val
		}
		os.Unsetenv(v)
	}

	t.Cleanup(func() {
		for _, v := range vars {
			if val, ok := original[v]; ok {
				os.Setenv(v, val)
			} else {
				os.Unsetenv(v)
			}
		}
	})
}
