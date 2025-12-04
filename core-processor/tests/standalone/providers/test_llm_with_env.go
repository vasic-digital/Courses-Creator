package providers_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
)

func TestLLMWithEnv(t *testing.T) {
	// Set test environment variables (these would normally be set externally)
	os.Setenv("OPENAI_API_KEY", "sk-test-key-placeholder")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-placeholder")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("LLM Default Provider: %s\n", cfg.LLM.DefaultProvider)
	fmt.Printf("OpenAI API Key: %s\n", maskAPIKey(cfg.LLM.OpenAI.APIKey))
	fmt.Printf("Anthropic API Key: %s\n", maskAPIKey(cfg.LLM.Anthropic.APIKey))

	// Create LLM manager
	llmManager := llm.NewProviderManager(&cfg.LLM)

	// Get available providers
	providers := llmManager.GetProviderInfo()
	fmt.Println("\nAvailable LLM Providers:")
	for _, provider := range providers {
		fmt.Printf("- %s (Type: %s, Available: %v)\n",
			provider.Name, provider.Type, provider.Available)
	}

	// Test with a simple prompt
	ctx := context.Background()
	testPrompt := "Explain machine learning in one sentence."

	fmt.Println("\nTesting providers with prompt:", testPrompt)
	fmt.Println("----------------------------------------")

	// Since we don't have real API keys, the providers will fail when trying to make API calls
	// But we can verify they are registered correctly
	response, err := llmManager.GenerateWithFallback(ctx, testPrompt, models.ProcessingOptions{
		Quality: "standard",
	})
	if err != nil {
		fmt.Printf("Expected error (fake API keys): %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	// Test direct provider methods
	fmt.Println("\nTesting individual providers:")
	fmt.Println("---------------------------------")

	providerList := llmManager.GetProviders()
	for _, provider := range providerList {
		fmt.Printf("Testing %s (Available: %v)...\n", provider.GetName(), provider.IsAvailable())

		if !provider.IsAvailable() {
			continue
		}

		resp, err := provider.GenerateText(ctx, "Test", models.ProcessingOptions{})
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Response: %s\n", resp)
		}
	}
}

func maskAPIKey(key string) string {
	if key == "" {
		return "[not set]"
	}
	if len(key) <= 8 {
		return "[masked]"
	}
	return key[:4] + "[...]" + key[len(key)-4:]
}
