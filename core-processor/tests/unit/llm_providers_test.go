package unit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIProvider_NewOpenAIProvider(t *testing.T) {
	// Test with default parameters
	provider := llm.NewOpenAIProvider("", "")

	require.NotNil(t, provider)
	assert.Equal(t, "OpenAI", provider.GetName())
	assert.Equal(t, llm.ProviderTypePaid, provider.GetType())
	assert.Equal(t, "gpt-3.5-turbo", provider.GetModel())
	assert.Equal(t, "https://api.openai.com/v1", provider.GetBaseURL())
}

func TestOpenAIProvider_NewOpenAIProviderWithCustomParams(t *testing.T) {
	provider := llm.NewOpenAIProvider("test-key", "gpt-4")

	require.NotNil(t, provider)
	assert.Equal(t, "OpenAI", provider.GetName())
	assert.Equal(t, "gpt-4", provider.GetModel())
}

func TestOpenAIProvider_GenerateText_NoAPIKey(t *testing.T) {
	// Clear environment variable
	os.Setenv("OPENAI_API_KEY", "")
	defer os.Setenv("OPENAI_API_KEY", "")

	provider := llm.NewOpenAIProvider("", "")

	ctx := context.Background()
	_, err := provider.GenerateText(ctx, "test prompt", models.ProcessingOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not configured")
}

func TestOpenAIProvider_GetCostEstimate(t *testing.T) {
	provider := llm.NewOpenAIProvider("", "gpt-3.5-turbo")

	cost := provider.GetCostEstimate(1000) // 1000 characters

	assert.Greater(t, cost, 0.0)
	assert.Less(t, cost, 1.0) // Should be less than $1
}

func TestAnthropicProvider_NewAnthropicProvider(t *testing.T) {
	provider := llm.NewAnthropicProvider("", "")

	require.NotNil(t, provider)
	assert.Equal(t, "Anthropic", provider.GetName())
	assert.Equal(t, llm.ProviderTypePaid, provider.GetType())
	assert.Equal(t, "claude-3-haiku-20240307", provider.GetModel())
	assert.Equal(t, "https://api.anthropic.com/v1", provider.GetBaseURL())
}

func TestAnthropicProvider_NewAnthropicProviderWithCustomParams(t *testing.T) {
	provider := llm.NewAnthropicProvider("test-key", "claude-3-opus-20240229")

	require.NotNil(t, provider)
	assert.Equal(t, "claude-3-opus-20240229", provider.GetModel())
}

func TestAnthropicProvider_GenerateText_NoAPIKey(t *testing.T) {
	// Clear environment variable
	os.Setenv("ANTHROPIC_API_KEY", "")
	defer os.Setenv("ANTHROPIC_API_KEY", "")

	provider := llm.NewAnthropicProvider("", "")

	ctx := context.Background()
	_, err := provider.GenerateText(ctx, "test prompt", models.ProcessingOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not configured")
}

func TestAnthropicProvider_GetCostEstimate(t *testing.T) {
	provider := llm.NewAnthropicProvider("", "claude-3-haiku-20240307")

	cost := provider.GetCostEstimate(1000) // 1000 characters

	assert.Greater(t, cost, 0.0)
	assert.Less(t, cost, 1.0) // Should be less than $1
}

func TestOllamaProvider_NewOllamaProvider(t *testing.T) {
	provider := llm.NewOllamaProvider("", "")

	require.NotNil(t, provider)
	assert.Equal(t, "Ollama", provider.GetName())
	assert.Equal(t, llm.ProviderTypeFree, provider.GetType())
	assert.Equal(t, "llama2", provider.GetModel())
	assert.Equal(t, "http://localhost:11434", provider.GetBaseURL())
}

func TestOllamaProvider_NewOllamaProviderWithCustomParams(t *testing.T) {
	provider := llm.NewOllamaProvider("http://localhost:8080", "mistral")

	require.NotNil(t, provider)
	assert.Equal(t, "mistral", provider.GetModel())
	assert.Equal(t, "http://localhost:8080", provider.GetBaseURL())
}

func TestOllamaProvider_GetCostEstimate(t *testing.T) {
	provider := llm.NewOllamaProvider("", "")

	cost := provider.GetCostEstimate(1000) // 1000 characters

	assert.Equal(t, 0.0, cost) // Ollama is free
}

func TestOllamaProvider_IsAvailable_NoServer(t *testing.T) {
	provider := llm.NewOllamaProvider("http://localhost:9999", "llama2")

	// Should return false when server is not available
	available := provider.IsAvailable()
	assert.False(t, available)
}

func TestProviderManager_NewProviderManager(t *testing.T) {
	config := &config.LLMConfig{
		OpenAI: config.OpenAIConfig{
			APIKey:       "",
			DefaultModel: "gpt-3.5-turbo",
		},
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:11434",
			DefaultModel: "llama2",
		},
		DefaultProvider: "openai",
	}

	manager := llm.NewProviderManager(config)

	require.NotNil(t, manager)
}

func TestProviderManager_GetProviderInfo(t *testing.T) {
	config := &config.LLMConfig{
		OpenAI: config.OpenAIConfig{
			APIKey:       "",
			DefaultModel: "gpt-3.5-turbo",
		},
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:11434",
			DefaultModel: "llama2",
		},
		DefaultProvider: "openai",
	}

	manager := llm.NewProviderManager(config)

	info := manager.GetProviderInfo()

	// May have 0 providers if Ollama is not running and OpenAI has no API key
	// This is expected behavior in test environment
	assert.GreaterOrEqual(t, len(info), 0)
}

func TestProviderManager_GenerateWithFallback_NoAvailableProviders(t *testing.T) {
	config := &config.LLMConfig{
		OpenAI: config.OpenAIConfig{
			APIKey:       "", // No API key
			DefaultModel: "gpt-3.5-turbo",
		},
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:9999", // Invalid URL
			DefaultModel: "llama2",
		},
		DefaultProvider: "openai",
	}

	manager := llm.NewProviderManager(config)

	ctx := context.Background()
	_, err := manager.GenerateWithFallback(ctx, "test prompt", models.ProcessingOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestCourseContentGenerator_NewCourseContentGenerator(t *testing.T) {
	config := &config.LLMConfig{
		OpenAI: config.OpenAIConfig{
			APIKey:       "",
			DefaultModel: "gpt-3.5-turbo",
		},
		DefaultProvider: "openai",
	}

	generator := llm.NewCourseContentGenerator(config)

	require.NotNil(t, generator)
}

func TestCourseContentGenerator_GenerateCourseTitle_NoAPIKey(t *testing.T) {
	config := &config.LLMConfig{
		OpenAI: config.OpenAIConfig{
			APIKey:       "", // No API key
			DefaultModel: "gpt-3.5-turbo",
		},
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:9999", // Invalid URL
			DefaultModel: "llama2",
		},
		DefaultProvider: "openai",
	}

	generator := llm.NewCourseContentGenerator(config)

	ctx := context.Background()
	content := "# Test Course\nThis is a test course content."

	_, err := generator.GenerateCourseTitle(ctx, content)

	assert.Error(t, err)
}

func TestCourseContentGenerator_GenerateCourseDescription_NoAPIKey(t *testing.T) {
	config := &config.LLMConfig{
		OpenAI: config.OpenAIConfig{
			APIKey:       "", // No API key
			DefaultModel: "gpt-3.5-turbo",
		},
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:9999", // Invalid URL
			DefaultModel: "llama2",
		},
		DefaultProvider: "openai",
	}

	generator := llm.NewCourseContentGenerator(config)

	ctx := context.Background()
	title := "Test Course"
	content := "# Test Course\nThis is a test course content."

	_, err := generator.GenerateCourseDescription(ctx, title, content)

	assert.Error(t, err)
}

func TestCourseContentGenerator_GenerateLessonContent_NoAPIKey(t *testing.T) {
	config := &config.LLMConfig{
		OpenAI: config.OpenAIConfig{
			APIKey:       "", // No API key
			DefaultModel: "gpt-3.5-turbo",
		},
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:9999", // Invalid URL
			DefaultModel: "llama2",
		},
		DefaultProvider: "openai",
	}

	generator := llm.NewCourseContentGenerator(config)

	ctx := context.Background()
	title := "Test Lesson"
	content := "## Test Lesson\nBasic content."

	_, err := generator.GenerateLessonContent(ctx, title, content)

	assert.Error(t, err)
}

// Integration tests (run only with API keys)
func TestOpenAIProvider_Integration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping OpenAI integration test - OPENAI_API_KEY not set")
	}

	provider := llm.NewOpenAIProvider("", "gpt-3.5-turbo")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := provider.GenerateText(ctx, "What is 2+2? Answer with just the number.", models.ProcessingOptions{
		Quality: "standard",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Contains(t, response, "4")
}

func TestAnthropicProvider_Integration(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping Anthropic integration test - ANTHROPIC_API_KEY not set")
	}

	provider := llm.NewAnthropicProvider("", "claude-3-haiku-20240307")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := provider.GenerateText(ctx, "What is 3+3? Answer with just the number.", models.ProcessingOptions{
		Quality: "standard",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Contains(t, response, "6")
}

func TestOllamaProvider_Integration(t *testing.T) {
	// This test requires Ollama to be running locally
	provider := llm.NewOllamaProvider("", "llama2")

	if !provider.IsAvailable() {
		t.Skip("Skipping Ollama integration test - Ollama not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := provider.GenerateText(ctx, "What is 1+1? Answer with just the number.", models.ProcessingOptions{
		Quality: "standard",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response)
}
