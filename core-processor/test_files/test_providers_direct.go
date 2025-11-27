package main

import (
	"context"
	"fmt"

	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
)

func main() {
	fmt.Println("Testing LLM Provider Implementations")
	fmt.Println("=================================")
	
	ctx := context.Background()
	
	// Test OpenAI Provider (with dummy key)
	fmt.Println("\n1. Testing OpenAI Provider:")
	fmt.Println("---------------------------")
	openaiProvider := llm.NewOpenAIProvider("sk-dummy-key", "gpt-3.5-turbo")
	fmt.Printf("Provider: %s\n", openaiProvider.GetName())
	fmt.Printf("Type: %s\n", openaiProvider.GetType())
	fmt.Printf("Available: %v\n", openaiProvider.IsAvailable())
	fmt.Printf("Cost Estimate (1000 chars): $%.6f\n", openaiProvider.GetCostEstimate(1000))
	
	resp, err := openaiProvider.GenerateText(ctx, "Test prompt", models.ProcessingOptions{})
	if err != nil {
		fmt.Printf("Error (expected): %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp)
	}
	
	// Test Anthropic Provider (with dummy key)
	fmt.Println("\n2. Testing Anthropic Provider:")
	fmt.Println("------------------------------")
	anthropicProvider := llm.NewAnthropicProvider("sk-ant-dummy-key", "claude-3-haiku-20240307")
	fmt.Printf("Provider: %s\n", anthropicProvider.GetName())
	fmt.Printf("Type: %s\n", anthropicProvider.GetType())
	fmt.Printf("Available: %v\n", anthropicProvider.IsAvailable())
	fmt.Printf("Cost Estimate (1000 chars): $%.6f\n", anthropicProvider.GetCostEstimate(1000))
	
	resp, err = anthropicProvider.GenerateText(ctx, "Test prompt", models.ProcessingOptions{})
	if err != nil {
		fmt.Printf("Error (expected): %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp)
	}
	
	// Test Ollama Provider
	fmt.Println("\n3. Testing Ollama Provider:")
	fmt.Println("----------------------------")
	ollamaProvider := llm.NewOllamaProvider("http://localhost:11434", "llama2")
	fmt.Printf("Provider: %s\n", ollamaProvider.GetName())
	fmt.Printf("Type: %s\n", ollamaProvider.GetType())
	fmt.Printf("Available: %v\n", ollamaProvider.IsAvailable())
	fmt.Printf("Cost Estimate (1000 chars): $%.6f\n", ollamaProvider.GetCostEstimate(1000))
	
	resp, err = ollamaProvider.GenerateText(ctx, "Test prompt", models.ProcessingOptions{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp)
	}
	
	// Test Free Provider
	fmt.Println("\n4. Testing Free Provider:")
	fmt.Println("-------------------------")
	freeProvider := llm.NewFreeProvider("FreeLocal", "", "")
	fmt.Printf("Provider: %s\n", freeProvider.GetName())
	fmt.Printf("Type: %s\n", freeProvider.GetType())
	fmt.Printf("Available: %v\n", freeProvider.IsAvailable())
	fmt.Printf("Cost Estimate (1000 chars): $%.6f\n", freeProvider.GetCostEstimate(1000))
	
	resp, err = freeProvider.GenerateText(ctx, "Test prompt", models.ProcessingOptions{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp)
	}
	
	// Test Paid Provider
	fmt.Println("\n5. Testing Paid Provider:")
	fmt.Println("-------------------------")
	paidProvider := llm.NewPaidProvider("PaidTest", "", "", 0.01)
	fmt.Printf("Provider: %s\n", paidProvider.GetName())
	fmt.Printf("Type: %s\n", paidProvider.GetType())
	fmt.Printf("Available: %v\n", paidProvider.IsAvailable())
	fmt.Printf("Cost Estimate (1000 chars): $%.6f\n", paidProvider.GetCostEstimate(1000))
	
	resp, err = paidProvider.GenerateText(ctx, "Test prompt", models.ProcessingOptions{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", resp)
	}
	
	// Test Provider Manager
	fmt.Println("\n6. Testing Provider Manager:")
	fmt.Println("-----------------------------")
	manager := llm.NewProviderManager(nil) // Using nil config for test
	
	// Register our providers
	manager.RegisterProvider(openaiProvider)
	manager.RegisterProvider(anthropicProvider)
	manager.RegisterProvider(ollamaProvider)
	manager.RegisterProvider(freeProvider)
	manager.RegisterProvider(paidProvider)
	
	// Get all providers info
	fmt.Println("\nRegistered Providers:")
	for _, info := range manager.GetProviderInfo() {
		fmt.Printf("- %s: %s, Available: %v, Cost: $%.6f/token\n", 
			info.Name, info.Type, info.Available, info.CostPerToken)
	}
	
	// Test fallback mechanism
	fmt.Println("\nTesting Fallback Mechanism:")
	preferences := llm.ProviderPreferences{
		PreferredType:     llm.ProviderTypeFree,
		MaxCostPerRequest: 0.10,
		PrioritizeQuality:  false,
		AllowPaid:         false,
	}
	
	bestProvider := manager.GetBestProvider(preferences)
	if bestProvider != nil {
		fmt.Printf("Best provider: %s\n", bestProvider.GetName())
		resp, err := manager.GenerateWithFallback(ctx, "Test prompt", models.ProcessingOptions{})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Response: %s\n", resp)
		}
	} else {
		fmt.Println("No suitable provider found")
	}
}