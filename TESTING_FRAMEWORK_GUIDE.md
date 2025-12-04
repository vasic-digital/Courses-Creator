# Course Creator - Testing Framework & Strategy Guide

## Overview

This document outlines the comprehensive testing framework for Course Creator project, covering all 6 supported test types with detailed implementation guidelines, test coverage requirements, and quality assurance procedures.

## Test Types Overview

### 1. Unit Testing
**Purpose**: Verify individual components/functions in isolation
**Coverage Target**: 80% minimum
**Frameworks**: 
- Go: testify/assert, require, mock
- TypeScript: Jest, React Testing Library
- React Native: Jest, React Native Testing Library

### 2. Integration Testing
**Purpose**: Test interaction between components and external services
**Coverage Target**: 70% minimum
**Focus**: Database operations, API integrations, file storage

### 3. Contract Testing
**Purpose**: Ensure API contracts between services are maintained
**Coverage Target**: 100% of public APIs
**Tools**: OpenAPI/Swagger, custom contract validators

### 4. Performance Testing
**Purpose**: Verify system performance under load
**Metrics**: Response times, throughput, resource usage
**Tools**: k6, Apache Bench, Go benchmarks

### 5. Security Testing
**Purpose**: Identify and fix security vulnerabilities
**Coverage**: All authentication, authorization, and data handling
**Tools**: OWASP ZAP, Go security scanners

### 6. Accessibility Testing
**Purpose**: Ensure compliance with accessibility standards
**Standard**: WCAG 2.1 AA
**Tools**: axe-core, Lighthouse, screen reader testing

## Test Implementation Details

### Unit Testing Implementation

#### Go Backend Unit Tests

##### Structure
```
/core-processor/tests/unit/
├── api/
│   ├── handlers_test.go
│   ├── middleware_test.go
│   └── routes_test.go
├── services/
│   ├── auth_test.go
│   ├── course_test.go
│   └── job_test.go
├── models/
│   ├── course_test.go
│   ├── user_test.go
│   └── job_test.go
├── repository/
│   ├── course_repo_test.go
│   ├── user_repo_test.go
│   └── job_repo_test.go
├── utils/
│   ├── validator_test.go
│   ├── markdown_test.go
│   └── crypto_test.go
└── pipeline/
    ├── generator_test.go
    ├── processor_test.go
    └── assembler_test.go
```

##### Example Test Implementation

```go
// /core-processor/tests/unit/services/auth_test.go
package services

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/golang-jwt/jwt/v4"
	"github.com/course-creator/models"
)

// Mock repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// Test suite
type AuthServiceTestSuite struct {
	suite.Suite
	service  *AuthService
	mockRepo *MockUserRepository
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockUserRepository)
	suite.service = NewAuthService(suite.mockRepo, "test-secret")
}

func (suite *AuthServiceTestSuite) TestRegisterUser_ValidData() {
	// Arrange
	user := &models.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	
	suite.mockRepo.On("FindByEmail", user.Email).Return(nil, nil)
	suite.mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	
	// Act
	err := suite.service.RegisterUser(user)
	
	// Assert
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), user.HashedPassword)
	assert.NotZero(suite.T(), user.ID)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLoginUser_ValidCredentials() {
	// Arrange
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	user := &models.User{
		ID:           1,
		Email:        email,
		HashedPassword: string(hashedPassword),
		FirstName:    "John",
		LastName:     "Doe",
	}
	
	suite.mockRepo.On("FindByEmail", email).Return(user, nil)
	
	// Act
	token, err := suite.service.LoginUser(email, password)
	
	// Assert
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)
	
	// Verify token
	claims, err := suite.service.ValidateToken(token)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, claims.UserID)
	assert.Equal(suite.T(), email, claims.Email)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// Benchmarks
func BenchmarkAuthService_LoginUser(b *testing.B) {
	service := NewAuthService(nil, "test-secret")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateToken(1, "test@example.com")
	}
}
```

##### Frontend Unit Tests

```typescript
// /player-app/src/__tests__/components/CourseCard.test.tsx
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import CourseCard from '../components/CourseCard';
import { mockCourse } from '../__mocks__/course';

const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
});

const renderWithQueryClient = (component: React.ReactElement) => {
  return render(
    <QueryClientProvider client={createTestQueryClient()}>
      {component}
    </QueryClientProvider>
  );
};

describe('CourseCard', () => {
  it('renders course information correctly', () => {
    renderWithQueryClient(<CourseCard course={mockCourse} />);
    
    expect(screen.getByText(mockCourse.title)).toBeInTheDocument();
    expect(screen.getByText(mockCourse.description)).toBeInTheDocument();
    expect(screen.getByText(/duration/i)).toBeInTheDocument();
  });
  
  it('shows loading state while thumbnail loads', () => {
    const courseWithLoadingThumb = {
      ...mockCourse,
      thumbnailUrl: null,
    };
    
    renderWithQueryClient(<CourseCard course={courseWithLoadingThumb} />);
    expect(screen.getByTestId('thumbnail-loader')).toBeInTheDocument();
  });
  
  it('navigates to course detail on click', () => {
    const mockPush = jest.fn();
    jest.mock('next/router', () => ({
      useRouter: () => ({
        push: mockPush,
      }),
    }));
    
    renderWithQueryClient(<CourseCard course={mockCourse} />);
    
    fireEvent.click(screen.getByTestId('course-card'));
    expect(mockPush).toHaveBeenCalledWith(`/courses/${mockCourse.id}`);
  });
});
```

### Integration Testing Implementation

#### Database Integration Tests

```go
// /core-processor/tests/integration/database_test.go
package integration

import (
	"context"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"course-creator/database"
	"course-creator/models"
	"course-processor/repository"
)

type DatabaseIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *database.Connection
	repo      repository.CourseRepository
}

func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	
	// Start test container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	
	var err error
	suite.container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)
	
	// Get connection details
	host, _ := suite.container.Host(ctx)
	port, _ := suite.container.MappedPort(ctx, "5432")
	
	// Connect to test database
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	suite.db = database.NewConnection(dsn)
	suite.repo = repository.NewCourseRepository(suite.db)
	
	// Run migrations
	err = suite.db.Migrate()
	suite.Require().NoError(err)
}

func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	suite.db.Close()
	suite.container.Terminate(context.Background())
}

func (suite *DatabaseIntegrationTestSuite) SetupTest() {
	// Clean database before each test
	suite.db.Exec("TRUNCATE TABLE courses, users, jobs RESTART IDENTITY CASCADE")
}

func (suite *DatabaseIntegrationTestSuite) TestCreateCourse() {
	course := &models.Course{
		Title:       "Integration Test Course",
		Description: "A course created during integration testing",
		Duration:    3600,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	err := suite.repo.Create(course)
	suite.Require().NoError(err)
	
	assert.NotZero(suite.T(), course.ID)
	
	// Verify course exists in database
	retrieved, err := suite.repo.FindByID(course.ID)
	suite.Require().NoError(err)
	
	assert.Equal(suite.T(), course.Title, retrieved.Title)
	assert.Equal(suite.T(), course.Description, retrieved.Description)
}

func (suite *DatabaseIntegrationTestSuite) TestCourseTransactions() {
	// Test transaction rollback on error
	tx := suite.db.Begin()
	
	course := &models.Course{
		Title:       "Transaction Test",
		Description: "Should be rolled back",
	}
	
	err := tx.Create(course).Error
	suite.Require().NoError(err)
	
	// Simulate error
	err = tx.Model(&models.User{}).Create(&models.User{
		Email: "invalid-email",
	}).Error
	suite.Require().Error(err)
	
	tx.Rollback()
	
	// Verify course was not committed
	_, err = suite.repo.FindByID(course.ID)
	suite.Assert().Error(err)
}

func TestDatabaseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationIntegrationTestSuite))
}
```

#### API Integration Tests

```typescript
// /core-processor/tests/integration/api_test.ts
import request from 'supertest';
import { app } from '../app';
import { setupTestDatabase, cleanupTestDatabase } from './helpers';

describe('API Integration Tests', () => {
  beforeAll(async () => {
    await setupTestDatabase();
  });
  
  afterAll(async () => {
    await cleanupTestDatabase();
  });
  
  describe('Course API', () => {
    it('should create and retrieve a course', async () => {
      // Create course
      const createResponse = await request(app)
        .post('/api/v1/courses')
        .send({
          title: 'Integration Test Course',
          description: 'Created during integration testing',
          duration: 3600,
        })
        .expect(201);
      
      const courseId = createResponse.body.id;
      expect(courseId).toBeDefined();
      
      // Retrieve course
      const getResponse = await request(app)
        .get(`/api/v1/courses/${courseId}`)
        .expect(200);
      
      expect(getResponse.body.title).toBe('Integration Test Course');
      expect(getResponse.body.description).toBe('Created during integration testing');
    });
    
    it('should handle validation errors', async () => {
      await request(app)
        .post('/api/v1/courses')
        .send({
          title: '', // Invalid: empty title
          description: 'Valid description',
        })
        .expect(400);
    });
  });
});
```

### Contract Testing Implementation

#### API Contract Tests

```go
// /core-processor/tests/contract/api_contract_test.go
package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

// OpenAPI schema loader
func loadAPISchema() *gojsonschema.Schema {
	schemaLoader := gojsonschema.NewStringLoader(`
	{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"id": {
				"type": "integer"
			},
			"title": {
				"type": "string",
				"minLength": 1,
				"maxLength": 255
			},
			"description": {
				"type": "string",
				"maxLength": 1000
			},
			"duration": {
				"type": "integer",
				"minimum": 0
			},
			"created_at": {
				"type": "string",
				"format": "date-time"
			}
		},
		"required": ["id", "title", "description", "created_at"]
	}
	`)
	
	schema, _ := gojsonschema.NewSchema(schemaLoader)
	return schema
}

func TestCourseAPIContract(t *testing.T) {
	schema := loadAPISchema()
	
	// Test POST /courses response
	course := map[string]interface{}{
		"title":       "Contract Test Course",
		"description": "Testing API contract",
		"duration":    3600,
	}
	
	body, _ := json.Marshal(course)
	req := httptest.NewRequest("POST", "/api/v1/courses", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	
	// Assuming we have a router setup
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	// Validate response against schema
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	documentLoader := gojsonschema.NewGoLoader(response)
	result, err := schema.Validate(documentLoader)
	
	assert.NoError(t, err)
	assert.True(t, result.Valid())
	
	// Validate required fields
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "title")
	assert.Contains(t, response, "description")
	assert.Contains(t, response, "created_at")
	
	// Validate field types
	assert.IsType(t, float64(0), response["id"])
	assert.IsType(t, "", response["title"])
	assert.IsType(t, "", response["description"])
}

func TestErrorResponsesContract(t *testing.T) {
	errorSchema := gojsonschema.NewStringLoader(`
	{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"error": {
				"type": "string"
			},
			"code": {
				"type": "integer"
			},
			"details": {
				"type": "string"
			}
		},
		"required": ["error", "code"]
	}
	`)
	
	schema, _ := gojsonschema.NewSchema(errorSchema)
	
	// Test 400 error response
	req := httptest.NewRequest("POST", "/api/v1/courses", bytes.NewBuffer([]byte(`{}`)))
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	documentLoader := gojsonschema.NewGoLoader(response)
	result, err := schema.Validate(documentLoader)
	
	assert.NoError(t, err)
	assert.True(t, result.Valid())
	assert.Equal(t, float64(400), response["code"])
	assert.NotEmpty(t, response["error"])
}
```

### Performance Testing Implementation

#### Load Testing with k6

```javascript
// /core-processor/tests/performance/load_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 200 }, // Ramp up to 200 users
    { duration: '5m', target: 200 }, // Stay at 200 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'], // 95% of requests under 200ms
    http_req_failed: ['rate<0.01'],    // Error rate below 1%
    errors: ['rate<0.01'],            // Custom error rate below 1%
  },
};

const BASE_URL = 'http://localhost:8080';

export function setup() {
  // Setup - create test user and get auth token
  const loginResponse = http.post(`${BASE_URL}/api/v1/auth/login`, {
    email: 'test@example.com',
    password: 'password',
  });
  
  return {
    token: loginResponse.json('token'),
  };
}

export default function(data) {
  const params = {
    headers: {
      Authorization: `Bearer ${data.token}`,
    },
  };
  
  // Test course listing
  let response = http.get(`${BASE_URL}/api/v1/courses`, params);
  let success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
    'response contains courses': (r) => r.json().length > 0,
  });
  
  errorRate.add(!success);
  
  // Test course creation (10% of iterations)
  if (__ITER % 10 === 0) {
    const courseData = {
      title: `Performance Test Course ${__ITER}`,
      description: 'Created during performance testing',
      duration: 3600,
    };
    
    response = http.post(`${BASE_URL}/api/v1/courses`, JSON.stringify(courseData), params);
    success = check(response, {
      'status is 201': (r) => r.status === 201,
      'response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    errorRate.add(!success);
  }
  
  sleep(1);
}

export function teardown(data) {
  // Cleanup - delete test data if needed
  console.log('Load test completed');
}
```

#### Go Benchmarks

```go
// /core-processor/tests/performance/benchmark_test.go
package performance

import (
	"testing"
	"course-creator/services"
)

func BenchmarkCourseService_CreateCourse(b *testing.B) {
	service := services.NewCourseService(nil)
	course := &models.Course{
		Title:       "Benchmark Course",
		Description: "Used for benchmarking",
		Duration:    3600,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateCourse(course)
	}
}

func BenchmarkAuthService_GenerateToken(b *testing.B) {
	service := services.NewAuthService(nil, "test-secret")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GenerateToken(1, "test@example.com")
	}
}

func BenchmarkMarkdownParser_Parse(b *testing.B) {
	parser := utils.NewMarkdownParser()
	content := `# Test Course
	
This is a test course for benchmarking.

## Lesson 1
Content for lesson 1.

## Lesson 2  
Content for lesson 2.
`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(content)
	}
}
```

### Security Testing Implementation

#### Security Test Suite

```go
// /core-processor/tests/security/security_test.go
package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSQLInjectionPrevention(t *testing.T) {
	// Test various SQL injection payloads
	payloads := []string{
		"'; DROP TABLE courses; --",
		"' OR '1'='1",
		"1'; UPDATE users SET password='hacked' WHERE '1'='1'; --",
	}
	
	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			body := bytes.NewBufferString(`{"title": "` + payload + `"}`)
			req := httptest.NewRequest("POST", "/api/v1/courses", body)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			// Should either reject with 400 or process safely
			assert.NotEqual(t, http.StatusInternalServerError, w.Code)
		})
	}
}

func TestXSSPrevention(t *testing.T) {
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"'><script>alert('xss')</script>",
	}
	
	for _, payload := range xssPayloads {
		t.Run(payload, func(t *testing.T) {
			course := map[string]string{
				"title":       payload,
				"description": payload,
			}
			
			body, _ := json.Marshal(course)
			req := httptest.NewRequest("POST", "/api/v1/courses", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			// Response should not contain unescaped scripts
			assert.NotContains(t, w.Body.String(), "<script>")
			assert.NotContains(t, w.Body.String(), "javascript:")
		})
	}
}

func TestAuthenticationSecurity(t *testing.T) {
	// Test without authentication
	req := httptest.NewRequest("GET", "/api/v1/courses", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should require authentication
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	// Test with invalid token
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRateLimiting(t *testing.T) {
	// Send rapid requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/api/v1/courses", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Should eventually rate limit
		if w.Code == http.StatusTooManyRequests {
			return
		}
	}
	
	t.Error("Rate limiting not working")
}
```

### Accessibility Testing Implementation

#### Accessibility Tests

```typescript
// /player-app/src/__tests__/accessibility/CourseCard.a11y.test.tsx
import { render, axe, toHaveNoViolations } from 'jest-axe';
import CourseCard from '../../components/CourseCard';
import { mockCourse } from '../../__mocks__/course';

expect.extend(toHaveNoViolations);

describe('CourseCard Accessibility', () => {
  it('should not have accessibility violations', async () => {
    const { container } = render(<CourseCard course={mockCourse} />);
    const results = await axe(container);
    
    expect(results).toHaveNoViolations();
  });
  
  it('should be keyboard navigable', () => {
    const { getByRole } = render(<CourseCard course={mockCourse} />);
    const card = getByRole('button'); // Assuming card is focusable
    
    expect(card).toHaveAttribute('tabIndex', '0');
    
    // Test keyboard navigation
    card.focus();
    expect(card).toHaveFocus();
  });
  
  it('should have appropriate ARIA labels', () => {
    const { getByLabelText } = render(<CourseCard course={mockCourse} />);
    
    expect(getByLabelText(/course title/i)).toBeInTheDocument();
    expect(getByLabelText(/course duration/i)).toBeInTheDocument();
  });
  
  it('should have sufficient color contrast', () => {
    // This would typically be tested with a contrast checker
    const { getByText } = render(<CourseCard course={mockCourse} />);
    const title = getByText(mockCourse.title);
    
    // Check computed styles
    const styles = window.getComputedStyle(title);
    expect(styles.color).not.toBe('rgba(0, 0, 0, 0)'); // Ensure not transparent
  });
});
```

## Test Coverage Requirements

### Minimum Coverage Standards

1. **Unit Tests**: 80% line coverage, 70% branch coverage
2. **Integration Tests**: 70% coverage of critical paths
3. **Contract Tests**: 100% of public APIs
4. **E2E Tests**: All user journeys covered
5. **Performance Tests**: All critical endpoints
6. **Security Tests**: All authentication and data handling

### Coverage Measurement

#### Go Coverage
```bash
# Run tests with coverage
go test ./... -v -coverprofile=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Check specific package coverage
go test ./services -cover
```

#### TypeScript Coverage
```json
// package.json
{
  "scripts": {
    "test": "jest --coverage",
    "test:watch": "jest --watch",
    "coverage:report": "jest --coverage --coverageReporters=text-lcov | coveralls"
  }
}
```

#### Coverage Configuration
```javascript
// jest.config.js
module.exports = {
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/index.ts',
    '!src/types/**',
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
};
```

## Test Data Management

### Test Fixtures

```typescript
// /player-app/src/__mocks__/courses.ts
export const mockCourses = [
  {
    id: '1',
    title: 'Introduction to React',
    description: 'Learn React from scratch',
    duration: 3600,
    thumbnailUrl: 'https://example.com/thumb1.jpg',
    createdAt: '2023-01-01T00:00:00Z',
    updatedAt: '2023-01-01T00:00:00Z',
  },
  {
    id: '2',
    title: 'Advanced TypeScript',
    description: 'Master TypeScript concepts',
    duration: 5400,
    thumbnailUrl: 'https://example.com/thumb2.jpg',
    createdAt: '2023-01-02T00:00:00Z',
    updatedAt: '2023-01-02T00:00:00Z',
  },
];

export const mockCourse = mockCourses[0];
```

```go
// /core-processor/tests/fixtures/courses.go
package fixtures

import "time"

func GetTestCourse() *models.Course {
	return &models.Course{
		ID:          1,
		Title:       "Test Course",
		Description: "A course for testing",
		Duration:    3600,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func GetTestUser() *models.User {
	return &models.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
```

### Database Test Helpers

```go
// /core-processor/tests/helpers/database.go
package helpers

import (
	"course-creator/database"
	"course-creator/models"
	"github.com/stretchr/testify/require"
)

func SetupTestDatabase(t *testing.T) *database.Connection {
	// Create in-memory database for testing
	db := database.NewInMemoryConnection()
	
	// Run migrations
	err := db.Migrate()
	require.NoError(t, err)
	
	return db
}

func CleanupTestDatabase(db *database.Connection) {
	db.Close()
}

func CreateTestUser(t *testing.T, db *database.Connection) *models.User {
	user := fixtures.GetTestUser()
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func CreateTestCourse(t *testing.T, db *database.Connection, userID uint) *models.Course {
	course := fixtures.GetTestCourse()
	course.UserID = userID
	err := db.Create(course).Error
	require.NoError(t, err)
	return course
}
```

## Continuous Integration

### GitHub Actions Configuration

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
            
      - name: Run tests
        run: |
          cd core-processor
          go test ./... -v -race -coverprofile=coverage.out
          
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./core-processor/coverage.out
          
  test-frontend:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        app: [player-app, creator-app, mobile-player]
        
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          
      - name: Cache node modules
        uses: actions/cache@v3
        with:
          path: |
            ~/.npm
            ${{ github.workspace }}/${{ matrix.app }}/node_modules
          key: ${{ runner.os }}-node-${{ hashFiles('${{ matrix.app }}/package-lock.json') }}
          
      - name: Install dependencies
        run: |
          cd ${{ matrix.app }}
          npm ci
          
      - name: Run tests
        run: |
          cd ${{ matrix.app }}
          npm test -- --coverage
          
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          directory: ${{ matrix.app }}/coverage
          
  e2e-tests:
    runs-on: ubuntu-latest
    needs: [test-go, test-frontend]
    
    steps:
      - uses: actions/checkout@v3
      - uses: cypress-io/github-action@v5
        with:
          wait-on: 'http://localhost:8080'
          wait-on-timeout: 120
          
      - name: Start services
        run: |
          docker-compose -f docker-compose.test.yml up -d
          
      - name: Run E2E tests
        run: |
          cd core-processor
          npx cypress run --spec 'cypress/e2e/**/*.cy.js'
          
      - name: Stop services
        run: |
          docker-compose -f docker-compose.test.yml down
```

## Test Reporting & Metrics

### Coverage Reports

```bash
# Generate combined coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Generate coverage badge
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -n1 | awk '{print $3}'
```

### Test Metrics Dashboard

```yaml
# /monitoring/grafana/dashboards/test-metrics.yml
dashboard:
  title: Test Metrics Dashboard
  panels:
    - title: Test Coverage
      type: stat
      targets:
        - expr: go_test_coverage_percent
          
    - title: Test Pass Rate
      type: graph
      targets:
        - expr: go_test_pass_rate
          
    - title: Test Duration
      type: graph
      targets:
        - expr: histogram_quantile(0.95, go_test_duration_seconds)
```

## Best Practices

### Test Writing Guidelines

1. **AAA Pattern**: Arrange, Act, Assert
2. **Independent Tests**: Tests should not depend on each other
3. **Descriptive Names**: Test names should describe what they test
4. **Mock External Dependencies**: Isolate code from external services
5. **Test Edge Cases**: Don't just test happy path
6. **Keep Tests Simple**: Complex tests are hard to maintain

### Code Examples

```go
// Good test example
func TestCourseService_CreateCourse_ValidInput_ReturnsSuccess(t *testing.T) {
	// Arrange
	service := NewCourseService(mockRepo)
	course := &models.Course{
		Title:       "Valid Course",
		Description: "Valid Description",
		Duration:    3600,
	}
	
	mockRepo.On("Create", mock.AnythingOfType("*models.Course")).Return(nil)
	
	// Act
	err := service.CreateCourse(course)
	
	// Assert
	assert.NoError(t, err)
	assert.NotZero(t, course.ID)
	mockRepo.AssertExpectations(t)
}
```

### Common Pitfalls to Avoid

1. **Testing Implementation Details**: Test behavior, not implementation
2. **Brittle Tests**: Tests that break with refactoring
3. **Mock Overuse**: Don't mock everything
4. **Slow Tests**: Keep unit tests fast
5. **Complex Setup**: Keep test setup simple

## Troubleshooting

### Common Issues

1. **Test Environment**
   - Ensure test database is isolated
   - Clean up test data after each test
   - Use test configuration

2. **Flaky Tests**
   - Add proper synchronization
   - Use deterministic test data
   - Avoid time-based assertions

3. **Performance Issues**
   - Profile slow tests
   - Use test doubles for slow dependencies
   - Parallelize independent tests

### Debugging Tips

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestCreateCourse ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Run tests with specific timeout
go test -timeout=30s ./...
```

## Conclusion

This comprehensive testing framework ensures the Course Creator project maintains high quality standards through:

1. **Multi-layer Testing**: Unit, integration, contract, performance, security, and accessibility
2. **High Coverage**: 80% minimum coverage requirements
3. **Automation**: CI/CD pipeline integration
4. **Quality Gates**: Automated checks prevent merging low-quality code
5. **Continuous Monitoring**: Track test metrics over time

Following this framework will result in a robust, reliable, and maintainable codebase that meets enterprise quality standards.