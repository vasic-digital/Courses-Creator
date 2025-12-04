package llm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
)

func TestLLMIntegration(t *testing.T) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create LLM manager
	llmManager := llm.NewProviderManager(&cfg.LLM)

	// Get available providers
	providers := llmManager.GetProviderInfo()
	fmt.Println("Available LLM Providers:")
	for _, provider := range providers {
		fmt.Printf("- %s (Type: %s, Available: %v, Cost per token: $%.6f)\n",
			provider.Name, provider.Type, provider.Available, provider.CostPerToken)
	}

	// Test providers with a simple prompt
	ctx := context.Background()
	testPrompt := "Explain quantum computing in one sentence."
	
	fmt.Println("\nTesting providers with prompt:", testPrompt)
	fmt.Println("----------------------------------------")
	
	// Test using fallback mechanism
	response, err := llmManager.GenerateWithFallback(ctx, testPrompt, models.ProcessingOptions{
		Quality: "standard",
	})
	if err != nil {
		fmt.Printf("All providers failed: %v\n", err)
	} else {
		fmt.Printf("Response from best provider: %s\n", response)
	}

	// Test content generator
	fmt.Println("\nTesting Course Content Generator:")
	fmt.Println("----------------------------------")
	
	contentGen := llm.NewCourseContentGenerator(&cfg.LLM)
	
	// Test title generation
	sampleContent := `
# Introduction to Machine Learning

Machine learning is a subset of artificial intelligence that enables systems to learn and improve from experience without being explicitly programmed.

## Topics Covered
- Supervised Learning
- Unsupervised Learning  
- Neural Networks
- Deep Learning
`

	title, err := contentGen.GenerateCourseTitle(ctx, sampleContent)
	if err != nil {
		fmt.Printf("Failed to generate title: %v\n", err)
	} else {
		fmt.Printf("Generated title: %s\n", title)
	}

	// Test description generation
	description, err := contentGen.GenerateCourseDescription(ctx, "Machine Learning Basics", sampleContent)
	if err != nil {
		fmt.Printf("Failed to generate description: %v\n", err)
	} else {
		fmt.Printf("Generated description: %s\n", description[:min(200, len(description))])
	}

	// Test lesson enhancement
	lessonContent := `## Python Basics
Python is a programming language.`
	
	enhanced, err := contentGen.GenerateLessonContent(ctx, "Python Basics", lessonContent)
	if err != nil {
		fmt.Printf("Failed to enhance lesson: %v\n", err)
	} else {
		fmt.Printf("Enhanced lesson preview: %s\n", enhanced[:min(300, len(enhanced))])
	}
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}