package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/course-creator/core-processor/models"
)

// MockProvider is a mock LLM provider for testing
type MockProvider struct {
	*BaseProvider
	responses map[string]string // prompt -> response mapping
	failures  map[string]error  // prompt -> error mapping
	callCount map[string]int    // prompt -> call count
}

// NewMockProvider creates a new mock LLM provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		BaseProvider: NewBaseProvider(name, ProviderTypeFree),
		responses:    make(map[string]string),
		failures:     make(map[string]error),
		callCount:    make(map[string]int),
	}
}

// GenerateText generates text using the mock provider
func (m *MockProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	m.callCount[prompt]++

	// Check if this prompt should fail
	if err, ok := m.failures[prompt]; ok {
		return "", err
	}

	// Check for predefined response
	if response, ok := m.responses[prompt]; ok {
		return response, nil
	}

	// Default response based on prompt type
	if strings.Contains(strings.ToLower(prompt), "title") {
		return "Mock Course Title", nil
	}
	if strings.Contains(strings.ToLower(prompt), "description") {
		return "Mock course description for testing purposes.", nil
	}
	if strings.Contains(strings.ToLower(prompt), "lesson") {
		return "Mock lesson content with key concepts and examples.", nil
	}

	// Generic default response
	length := 50
	if len(prompt) < length {
		length = len(prompt)
	}
	return fmt.Sprintf("Mock response for: %s", prompt[:length]), nil
}

// GetCostEstimate returns a mock cost estimate
func (m *MockProvider) GetCostEstimate(textLength int) float64 {
	return 0.0 // Free for testing
}

// SetResponse sets a predefined response for a prompt
func (m *MockProvider) SetResponse(prompt, response string) {
	m.responses[prompt] = response
}

// SetFailure sets a failure for a prompt
func (m *MockProvider) SetFailure(prompt string, err error) {
	m.failures[prompt] = err
}

// GetCallCount returns how many times a prompt was called
func (m *MockProvider) GetCallCount(prompt string) int {
	return m.callCount[prompt]
}

// Reset clears all mock data
func (m *MockProvider) Reset() {
	m.responses = make(map[string]string)
	m.failures = make(map[string]error)
	m.callCount = make(map[string]int)
	m.available = true
}
