package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockProvider_NewMockProvider(t *testing.T) {
	provider := NewMockProvider("test-mock")
	assert.Equal(t, "test-mock", provider.GetName())
	assert.Equal(t, ProviderTypeFree, provider.GetType())
	assert.True(t, provider.IsAvailable())
}

func TestMockProvider_GenerateText_DefaultResponses(t *testing.T) {
	provider := NewMockProvider("test-mock")
	ctx := context.Background()
	options := models.ProcessingOptions{}

	// Test title prompt
	response, err := provider.GenerateText(ctx, "Generate a course title about Python", options)
	require.NoError(t, err)
	assert.Equal(t, "Mock Course Title", response)

	// Test description prompt
	response, err = provider.GenerateText(ctx, "Write a course description", options)
	require.NoError(t, err)
	assert.Equal(t, "Mock course description for testing purposes.", response)

	// Test lesson prompt
	response, err = provider.GenerateText(ctx, "Create lesson content", options)
	require.NoError(t, err)
	assert.Equal(t, "Mock lesson content with key concepts and examples.", response)

	// Test generic prompt
	response, err = provider.GenerateText(ctx, "Some random prompt text", options)
	require.NoError(t, err)
	assert.Contains(t, response, "Mock response for:")
}

func TestMockProvider_GenerateText_PredefinedResponse(t *testing.T) {
	provider := NewMockProvider("test-mock")
	provider.SetResponse("specific prompt", "predefined response")

	ctx := context.Background()
	options := models.ProcessingOptions{}

	response, err := provider.GenerateText(ctx, "specific prompt", options)
	require.NoError(t, err)
	assert.Equal(t, "predefined response", response)
}

func TestMockProvider_GenerateText_Failure(t *testing.T) {
	provider := NewMockProvider("test-mock")
	mockError := errors.New("mock error")
	provider.SetFailure("failing prompt", mockError)

	ctx := context.Background()
	options := models.ProcessingOptions{}

	response, err := provider.GenerateText(ctx, "failing prompt", options)
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
	assert.Empty(t, response)
}

func TestMockProvider_GetCostEstimate(t *testing.T) {
	provider := NewMockProvider("test-mock")
	cost := provider.GetCostEstimate(1000)
	assert.Equal(t, 0.0, cost)
}

func TestMockProvider_CallCount(t *testing.T) {
	provider := NewMockProvider("test-mock")
	ctx := context.Background()
	options := models.ProcessingOptions{}

	// Call multiple times
	provider.GenerateText(ctx, "prompt1", options)
	provider.GenerateText(ctx, "prompt1", options)
	provider.GenerateText(ctx, "prompt2", options)

	assert.Equal(t, 2, provider.GetCallCount("prompt1"))
	assert.Equal(t, 1, provider.GetCallCount("prompt2"))
	assert.Equal(t, 0, provider.GetCallCount("nonexistent"))
}

func TestMockProvider_Reset(t *testing.T) {
	provider := NewMockProvider("test-mock")
	ctx := context.Background()
	options := models.ProcessingOptions{}

	// Set up some state
	provider.SetResponse("test", "response")
	provider.SetFailure("fail", errors.New("error"))
	provider.GenerateText(ctx, "test", options)

	// Reset
	provider.Reset()

	// Verify reset - check that predefined responses are cleared
	response, err := provider.GenerateText(ctx, "test", options)
	require.NoError(t, err)
	assert.NotEqual(t, "response", response) // Should use default response now

	response, err = provider.GenerateText(ctx, "fail", options)
	require.NoError(t, err) // Should not fail after reset

	// Call count for "test" after reset should be 1 (the call we just made)
	assert.Equal(t, 1, provider.GetCallCount("test"))
}
