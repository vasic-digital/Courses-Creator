package llm

import (
	"context"
	"testing"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLMProvider implements LLMProvider for testing
type MockLLMProvider struct {
	mock.Mock
	name         string
	providerType ProviderType
	available    bool
}

func NewMockLLMProvider(name string, providerType ProviderType, available bool) *MockLLMProvider {
	return &MockLLMProvider{
		name:         name,
		providerType: providerType,
		available:    available,
	}
}

func (m *MockLLMProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	args := m.Called(ctx, prompt, options)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) GetType() ProviderType {
	return m.providerType
}

func (m *MockLLMProvider) GetName() string {
	return m.name
}

func (m *MockLLMProvider) IsAvailable() bool {
	return m.available
}

func (m *MockLLMProvider) GetCostEstimate(textLength int) float64 {
	args := m.Called(textLength)
	return args.Get(0).(float64)
}

func TestNewProviderManager(t *testing.T) {
	cfg := &config.LLMConfig{
		DefaultProvider: "ollama",
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:11434",
			DefaultModel: "llama2",
		},
	}

	pm := NewProviderManager(cfg)

	assert.NotNil(t, pm)
	assert.NotNil(t, pm.config)
	assert.Equal(t, cfg, pm.config)
}

func TestProviderManager_RegisterProvider(t *testing.T) {
	pm := NewProviderManager(nil)

	mockProvider := NewMockLLMProvider("test-provider", ProviderTypeFree, true)

	pm.RegisterProvider(mockProvider)

	providers := pm.GetProviders()
	assert.Len(t, providers, 1)
	assert.Equal(t, "test-provider", providers[0].GetName())
}

func TestProviderManager_GetBestProvider(t *testing.T) {
	pm := NewProviderManager(nil)

	// Register providers
	freeProvider := NewMockLLMProvider("free-provider", ProviderTypeFree, true)
	paidProvider := NewMockLLMProvider("paid-provider", ProviderTypePaid, true)
	unavailableProvider := NewMockLLMProvider("unavailable-provider", ProviderTypeFree, false)

	pm.RegisterProvider(freeProvider)
	pm.RegisterProvider(paidProvider)
	pm.RegisterProvider(unavailableProvider)

	// Test with free preference
	prefs := ProviderPreferences{
		PreferredType: ProviderTypeFree,
		AllowPaid:     true,
	}

	best := pm.GetBestProvider(prefs)
	assert.NotNil(t, best)
	assert.Equal(t, "free-provider", best.GetName())

	// Test with paid preference
	prefs.PreferredType = ProviderTypePaid
	best = pm.GetBestProvider(prefs)
	assert.NotNil(t, best)
	assert.Equal(t, "paid-provider", best.GetName())
}

func TestProviderManager_GenerateWithFallback(t *testing.T) {
	pm := NewProviderManager(nil)

	// Mock providers
	failingProvider := NewMockLLMProvider("failing", ProviderTypeFree, true)
	failingProvider.On("GenerateText", mock.Anything, "test prompt", mock.Anything).Return("", assert.AnError)

	successProvider := NewMockLLMProvider("success", ProviderTypeFree, true)
	successProvider.On("GenerateText", mock.Anything, "test prompt", mock.Anything).Return("success response", nil)

	pm.RegisterProvider(failingProvider)
	pm.RegisterProvider(successProvider)

	result, err := pm.GenerateWithFallback(context.Background(), "test prompt", models.ProcessingOptions{})

	assert.NoError(t, err)
	assert.Equal(t, "success response", result)

	failingProvider.AssertExpectations(t)
	successProvider.AssertExpectations(t)
}

func TestProviderManager_GetProviders(t *testing.T) {
	pm := NewProviderManager(nil)

	provider1 := NewMockLLMProvider("provider1", ProviderTypeFree, true)
	provider2 := NewMockLLMProvider("provider2", ProviderTypePaid, true)

	pm.RegisterProvider(provider1)
	pm.RegisterProvider(provider2)

	providers := pm.GetProviders()
	assert.Len(t, providers, 2)

	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.GetName()
	}
	assert.Contains(t, names, "provider1")
	assert.Contains(t, names, "provider2")
}

func TestProviderManager_GetAvailableProviders(t *testing.T) {
	pm := NewProviderManager(nil)

	availableProvider := NewMockLLMProvider("available", ProviderTypeFree, true)
	unavailableProvider := NewMockLLMProvider("unavailable", ProviderTypeFree, false)

	pm.RegisterProvider(availableProvider)
	pm.RegisterProvider(unavailableProvider)

	available := pm.GetAvailableProviders()
	assert.Len(t, available, 1)
	assert.Equal(t, "available", available[0].GetName())
}

func TestProviderManager_UpdateConfig(t *testing.T) {
	pm := NewProviderManager(nil)

	// Initially no providers
	assert.Len(t, pm.GetProviders(), 0)

	// Update with config
	cfg := &config.LLMConfig{
		DefaultProvider: "ollama",
		Ollama: config.OllamaConfig{
			BaseURL:      "http://localhost:11434",
			DefaultModel: "llama2",
		},
	}

	pm.UpdateConfig(cfg)

	// Should have registered providers based on config
	assert.NotNil(t, pm.config)
	assert.Equal(t, cfg, pm.config)
}

func TestProviderManager_GetDefaultPreferences(t *testing.T) {
	// Test with nil config
	pm := NewProviderManager(nil)
	prefs := pm.GetDefaultPreferences()

	assert.Equal(t, ProviderTypeFree, prefs.PreferredType)
	assert.Equal(t, 1.0, prefs.MaxCostPerRequest)
	assert.False(t, prefs.PrioritizeQuality)
	assert.True(t, prefs.AllowPaid)

	// Test with config
	cfg := &config.LLMConfig{
		DefaultProvider:   "openai",
		MaxCostPerRequest: 0.5,
		PrioritizeQuality: true,
		AllowPaid:         false,
	}

	pm = NewProviderManager(cfg)
	prefs = pm.GetDefaultPreferences()

	assert.Equal(t, ProviderType("openai"), prefs.PreferredType)
	assert.Equal(t, 0.5, prefs.MaxCostPerRequest)
	assert.True(t, prefs.PrioritizeQuality)
	assert.False(t, prefs.AllowPaid)
}

func TestProviderManager_TestAllProviders(t *testing.T) {
	pm := NewProviderManager(nil)

	provider1 := NewMockLLMProvider("provider1", ProviderTypeFree, true)
	provider1.On("GenerateText", mock.Anything, "Test prompt", mock.Anything).Return("response", nil)

	provider2 := NewMockLLMProvider("provider2", ProviderTypeFree, true)
	provider2.On("GenerateText", mock.Anything, "Test prompt", mock.Anything).Return("", assert.AnError)

	pm.RegisterProvider(provider1)
	pm.RegisterProvider(provider2)

	results := pm.TestAllProviders(context.Background())

	assert.Len(t, results, 2)
	assert.NoError(t, results["provider1"])
	assert.Error(t, results["provider2"])

	provider1.AssertExpectations(t)
	provider2.AssertExpectations(t)
}
