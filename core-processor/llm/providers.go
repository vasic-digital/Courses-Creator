package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/course-creator/core-processor/models"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderTypeFree ProviderType = "free"
	ProviderTypePaid ProviderType = "paid"
)

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error)
	GetType() ProviderType
	GetName() string
	IsAvailable() bool
	GetCostEstimate(textLength int) float64 // Estimated cost in USD
}

// Note: ProviderManager and related types are now in manager.go

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	name         string
	providerType ProviderType
	available    bool
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name string, providerType ProviderType) *BaseProvider {
	return &BaseProvider{
		name:         name,
		providerType: providerType,
		available:    true,
	}
}

// GetType returns the provider type
func (p *BaseProvider) GetType() ProviderType {
	return p.providerType
}

// GetName returns the provider name
func (p *BaseProvider) GetName() string {
	return p.name
}

// IsAvailable returns whether the provider is available
func (p *BaseProvider) IsAvailable() bool {
	return p.available
}

// SetAvailable sets the availability status
func (p *BaseProvider) SetAvailable(available bool) {
	p.available = available
}

// FreeProvider represents a free LLM provider
type FreeProvider struct {
	*BaseProvider
	apiEndpoint string
	apiKey      string
}

// NewFreeProvider creates a new free provider
func NewFreeProvider(name, apiEndpoint, apiKey string) *FreeProvider {
	return &FreeProvider{
		BaseProvider: NewBaseProvider(name, ProviderTypeFree),
		apiEndpoint:  apiEndpoint,
		apiKey:       apiKey,
	}
}

// GenerateText generates text using the free provider
func (p *FreeProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	// Try to use local models first (Ollama)
	if ollama := NewOllamaProvider("", ""); ollama.IsAvailable() {
		return ollama.GenerateText(ctx, prompt, options)
	}
	
	// Fallback to placeholder response
	fmt.Printf("Generating text with free provider %s (placeholder)\n", p.name)

	// Simulate API call
	select {
	case <-time.After(100 * time.Millisecond):
		return fmt.Sprintf("Generated text from %s: %s", p.name, prompt[:min(50, len(prompt))]), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// GetCostEstimate returns cost estimate (free = 0)
func (p *FreeProvider) GetCostEstimate(textLength int) float64 {
	return 0.0
}

// PaidProvider represents a paid LLM provider
type PaidProvider struct {
	*BaseProvider
	apiEndpoint  string
	apiKey       string
	costPerToken float64
}

// NewPaidProvider creates a new paid provider
func NewPaidProvider(name, apiEndpoint, apiKey string, costPerToken float64) *PaidProvider {
	return &PaidProvider{
		BaseProvider: NewBaseProvider(name, ProviderTypePaid),
		apiEndpoint:  apiEndpoint,
		apiKey:       apiKey,
		costPerToken: costPerToken,
	}
}

// GenerateText generates text using the paid provider
func (p *PaidProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	// Placeholder implementation
	// In real implementation, make API call to paid service
	fmt.Printf("Generating text with paid provider %s\n", p.name)

	// Simulate API call
	select {
	case <-time.After(200 * time.Millisecond):
		return fmt.Sprintf("High-quality text from %s: %s", p.name, minString(prompt, 50)), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// GetCostEstimate returns cost estimate
func (p *PaidProvider) GetCostEstimate(textLength int) float64 {
	// Rough estimate: assume 4 chars per token
	tokens := textLength / 4
	return float64(tokens) * p.costPerToken
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
