package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/models"
)

// ProviderManager manages multiple LLM providers
type ProviderManager struct {
	providers []LLMProvider
	config    *config.LLMConfig
	mu        sync.RWMutex
}

// NewProviderManager creates a new provider manager
func NewProviderManager(cfg *config.LLMConfig) *ProviderManager {
	pm := &ProviderManager{
		providers: []LLMProvider{},
		config:    cfg,
	}

	// Auto-register providers based on config
	pm.registerProvidersFromConfig()

	return pm
}

// registerProvidersFromConfig registers providers based on configuration
func (pm *ProviderManager) registerProvidersFromConfig() {
	// Return early if config is nil
	if pm.config == nil {
		return
	}

	// Register OpenAI if API key is available
	if pm.config.OpenAI.APIKey != "" {
		provider := NewOpenAIProvider(pm.config.OpenAI.APIKey, pm.config.OpenAI.DefaultModel)
		pm.RegisterProvider(provider)
		fmt.Printf("Registered OpenAI provider with model: %s\n", pm.config.OpenAI.DefaultModel)
	}

	// Register Anthropic if API key is available
	if pm.config.Anthropic.APIKey != "" {
		provider := NewAnthropicProvider(pm.config.Anthropic.APIKey, pm.config.Anthropic.DefaultModel)
		pm.RegisterProvider(provider)
		fmt.Printf("Registered Anthropic provider with model: %s\n", pm.config.Anthropic.DefaultModel)
	}

	// Always try to register Ollama (with defaults if config is nil)
	ollamaURL := "http://localhost:11434"
	ollamaModel := "llama2"
	if pm.config != nil {
		ollamaURL = pm.config.Ollama.BaseURL
		ollamaModel = pm.config.Ollama.DefaultModel
	}
	ollamaProvider := NewOllamaProvider(ollamaURL, ollamaModel)
	if ollamaProvider.IsAvailable() {
		pm.RegisterProvider(ollamaProvider)
		fmt.Printf("Registered Ollama provider at: %s with model: %s\n", ollamaURL, ollamaModel)
	}
}

// RegisterProvider adds a provider to the manager
func (pm *ProviderManager) RegisterProvider(provider LLMProvider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.providers = append(pm.providers, provider)
}

// GetBestProvider returns the best available provider based on preferences
func (pm *ProviderManager) GetBestProvider(preferences ProviderPreferences) LLMProvider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var bestProvider LLMProvider
	var bestScore float64 = -1

	for _, provider := range pm.providers {
		if !provider.IsAvailable() {
			continue
		}

		score := pm.calculateProviderScore(provider, preferences)
		if score > bestScore {
			bestScore = score
			bestProvider = provider
		}
	}

	return bestProvider
}

// GenerateWithFallback tries providers in order until one succeeds
func (pm *ProviderManager) GenerateWithFallback(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	pm.mu.RLock()
	providers := make([]LLMProvider, len(pm.providers))
	copy(providers, pm.providers)
	pm.mu.RUnlock()

	for _, provider := range providers {
		if !provider.IsAvailable() {
			continue
		}

		result, err := provider.GenerateText(ctx, prompt, options)
		if err == nil {
			return result, nil
		}

		// Log the error and try next provider
		fmt.Printf("Provider %s failed: %v\n", provider.GetName(), err)
	}

	return "", fmt.Errorf("all providers failed")
}

// ProviderPreferences defines user preferences for provider selection
type ProviderPreferences struct {
	PreferredType     ProviderType
	MaxCostPerRequest float64
	PrioritizeQuality bool
	AllowPaid         bool
}

// calculateProviderScore calculates how well a provider matches preferences
func (pm *ProviderManager) calculateProviderScore(provider LLMProvider, prefs ProviderPreferences) float64 {
	score := 0.0

	// Type preference
	if provider.GetType() == prefs.PreferredType {
		score += 10
	}

	// Cost consideration
	if provider.GetType() == ProviderTypePaid && !prefs.AllowPaid {
		return -1 // Disqualify paid providers if not allowed
	}

	// Quality preference (paid providers generally higher quality)
	if prefs.PrioritizeQuality && provider.GetType() == ProviderTypePaid {
		score += 5
	}

	// Availability bonus
	if provider.IsAvailable() {
		score += 2
	}

	return score
}

// GetProviders returns all registered providers
func (pm *ProviderManager) GetProviders() []LLMProvider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers := make([]LLMProvider, len(pm.providers))
	copy(providers, pm.providers)
	return providers
}

// GetAvailableProviders returns only available providers
func (pm *ProviderManager) GetAvailableProviders() []LLMProvider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var available []LLMProvider
	for _, provider := range pm.providers {
		if provider.IsAvailable() {
			available = append(available, provider)
		}
	}
	return available
}

// UpdateConfig updates the configuration and re-registers providers
func (pm *ProviderManager) UpdateConfig(cfg *config.LLMConfig) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.config = cfg
	pm.providers = []LLMProvider{} // Clear existing providers
	pm.registerProvidersFromConfig()
}

// GetDefaultPreferences returns default preferences based on config
func (pm *ProviderManager) GetDefaultPreferences() ProviderPreferences {
	if pm.config == nil {
		return ProviderPreferences{
			PreferredType:     ProviderTypeFree,
			MaxCostPerRequest: 1.0,
			PrioritizeQuality: false,
			AllowPaid:         true,
		}
	}

	return ProviderPreferences{
		PreferredType:     ProviderType(pm.config.DefaultProvider),
		MaxCostPerRequest: pm.config.MaxCostPerRequest,
		PrioritizeQuality: pm.config.PrioritizeQuality,
		AllowPaid:         pm.config.AllowPaid,
	}
}

// GetProviderInfo returns detailed info about all providers
func (pm *ProviderManager) GetProviderInfo() []ProviderInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var info []ProviderInfo
	for _, provider := range pm.providers {
		costEstimate := provider.GetCostEstimate(1000) // Cost for 1000 characters
		info = append(info, ProviderInfo{
			Name:         provider.GetName(),
			Type:         provider.GetType(),
			CostPerToken: costEstimate * 4, // Convert back to per-token estimate
			Available:    provider.IsAvailable(),
		})
	}
	return info
}

// TestAllProviders tests all registered providers
func (pm *ProviderManager) TestAllProviders(ctx context.Context) map[string]error {
	pm.mu.RLock()
	providers := make([]LLMProvider, len(pm.providers))
	copy(providers, pm.providers)
	pm.mu.RUnlock()

	results := make(map[string]error)

	for _, provider := range providers {
		_, err := provider.GenerateText(ctx, "Test prompt", models.ProcessingOptions{})
		results[provider.GetName()] = err
	}

	return results
}
