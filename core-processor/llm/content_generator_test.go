package llm

import (
	"context"
	"testing"

	"github.com/course-creator/core-processor/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCourseContentGenerator(t *testing.T) {
	cfg := &config.LLMConfig{
		DefaultProvider: "ollama",
	}

	generator := NewCourseContentGenerator(cfg)

	assert.NotNil(t, generator)
	assert.NotNil(t, generator.providerManager)
}

func TestCourseContentGenerator_GenerateCourseTitle(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	// Mock the provider manager
	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GenerateText", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > 0
	}), mock.Anything).Return("Generated Course Title", nil)

	// Replace the provider manager with one containing our mock
	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	title, err := generator.GenerateCourseTitle(context.Background(), "Sample course content about programming")

	assert.NoError(t, err)
	assert.Equal(t, "Generated Course Title", title)
	mockProvider.AssertExpectations(t)
}

func TestCourseContentGenerator_GenerateCourseTitle_NoProvider(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	// Mock provider manager with no available providers
	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{},
	}

	_, err := generator.GenerateCourseTitle(context.Background(), "Sample content")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no LLM providers available")
}

func TestCourseContentGenerator_GenerateCourseDescription(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GenerateText", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > 0
	}), mock.Anything).Return("Generated course description with engaging content.", nil)

	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	description, err := generator.GenerateCourseDescription(context.Background(), "Course Title", "Course content details")

	assert.NoError(t, err)
	assert.Contains(t, description, "Generated course description")
	mockProvider.AssertExpectations(t)
}

func TestCourseContentGenerator_GenerateLessonContent(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GenerateText", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > 0
	}), mock.Anything).Return("Enhanced lesson content with learning objectives.", nil)

	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	content, err := generator.GenerateLessonContent(context.Background(), "Lesson Title", "Original lesson content")

	assert.NoError(t, err)
	assert.Contains(t, content, "Enhanced lesson content")
	mockProvider.AssertExpectations(t)
}

func TestCourseContentGenerator_GenerateInteractiveElements(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GenerateText", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > 0
	}), mock.Anything).Return(`[{"type":"quiz","title":"Test Quiz","content":"What is X?"}]`, nil)

	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	elements, err := generator.GenerateInteractiveElements(context.Background(), "Lesson content about topic X")

	assert.NoError(t, err)
	assert.Len(t, elements, 1)
	assert.Contains(t, elements[0], "quiz")
	mockProvider.AssertExpectations(t)
}

func TestCourseContentGenerator_GenerateMetadata(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GenerateText", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > 0
	}), mock.Anything).Return(`{"difficulty":"intermediate","duration_hours":2.5}`, nil)

	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	metadata, err := generator.GenerateMetadata(context.Background(), "Course Title", "Course description")

	assert.NoError(t, err)
	assert.Equal(t, `{"difficulty":"intermediate","duration_hours":2.5}`, metadata["generated_metadata"])
	assert.Equal(t, "intermediate", metadata["difficulty"])
	assert.Equal(t, 2.0, metadata["duration_hours"])
	mockProvider.AssertExpectations(t)
}

func TestCourseContentGenerator_GenerateMetadata_Fallback(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GenerateText", mock.Anything, mock.Anything, mock.Anything).Return("", assert.AnError)

	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	metadata, err := generator.GenerateMetadata(context.Background(), "Course Title", "Course description")

	// Should return fallback metadata without error
	assert.NoError(t, err)
	assert.Equal(t, "intermediate", metadata["difficulty"])
	assert.Equal(t, 2.0, metadata["duration_hours"])
	mockProvider.AssertExpectations(t)
}

func TestCourseContentGenerator_GetAvailableProviders(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GetCostEstimate", 1000).Return(0.0)

	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	providers := generator.GetAvailableProviders()

	assert.Len(t, providers, 1)
	assert.Equal(t, "mock-provider", providers[0].Name)
	assert.Equal(t, ProviderTypeFree, providers[0].Type)
	assert.True(t, providers[0].Available)
	mockProvider.AssertExpectations(t)
}

func TestCourseContentGenerator_TestProviders(t *testing.T) {
	cfg := &config.LLMConfig{}
	generator := NewCourseContentGenerator(cfg)

	mockProvider := NewMockLLMProvider("mock-provider", ProviderTypeFree, true)
	mockProvider.On("GenerateText", mock.Anything, "Test prompt", mock.Anything).Return("response", nil)

	generator.providerManager = &ProviderManager{
		providers: []LLMProvider{mockProvider},
	}

	results := generator.TestProviders(context.Background())

	assert.Len(t, results, 1)
	assert.NoError(t, results["mock-provider"])
	mockProvider.AssertExpectations(t)
}

func TestMinString(t *testing.T) {
	// Test with string shorter than limit
	result := minString("hello", 10)
	assert.Equal(t, "hello", result)

	// Test with string longer than limit
	result = minString("this is a very long string that should be truncated", 10)
	assert.Equal(t, "this is a ", result)
	assert.Len(t, result, 10)
}
