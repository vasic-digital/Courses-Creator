package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create LLM manager
	llmManager := llm.NewProviderManager(&cfg.LLM)

	// Register a mock provider for testing
	mockProvider := &MockProvider{
		BaseProvider: llm.NewBaseProvider("Mock", llm.ProviderTypeFree),
	}
	llmManager.RegisterProvider(mockProvider)

	// Get available providers
	providers := llmManager.GetProviderInfo()
	fmt.Println("Available LLM Providers:")
	for _, provider := range providers {
		fmt.Printf("- %s (Type: %s, Available: %v)\n",
			provider.Name, provider.Type, provider.Available)
	}

	// Test with a simple prompt
	ctx := context.Background()
	testPrompt := "Explain machine learning in one sentence."
	
	fmt.Println("\nTesting providers with prompt:", testPrompt)
	fmt.Println("----------------------------------------")

	response, err := llmManager.GenerateWithFallback(ctx, testPrompt, models.ProcessingOptions{
		Quality: "standard",
	})
	if err != nil {
		fmt.Printf("All providers failed: %v\n", err)
	} else {
		fmt.Printf("Response from mock provider: %s\n", response)
	}

	// Test content generator with the same manager
	fmt.Println("\nTesting Course Content Generator:")
	fmt.Println("----------------------------------")
	
	// Create content generator with existing manager
	contentGen := &llm.CourseContentGenerator{}
	
	// Use reflection or modify the generator to use our manager
	// For now, let's create a new one with mock provider pre-registered
	mockCfg := &config.LLMConfig{
		DefaultProvider:   "mock",
		MaxCostPerRequest: 1.0,
		PrioritizeQuality:  false,
		AllowPaid:         true,
	}
	contentGen = llm.NewCourseContentGenerator(mockCfg)
	
	// Register mock provider manually
	mockProvider := &MockProvider{
		BaseProvider: llm.NewBaseProvider("Mock", llm.ProviderTypeFree),
	}
	// We need access to the internal provider manager - let's create our own for testing
	
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

	// Test interactive elements
	interactive, err := contentGen.GenerateInteractiveElements(ctx, lessonContent)
	if err != nil {
		fmt.Printf("Failed to generate interactive elements: %v\n", err)
	} else {
		fmt.Printf("Interactive elements: %v\n", interactive)
	}

	// Test metadata
	metadata, err := contentGen.GenerateMetadata(ctx, "Python Basics", "A course about Python programming")
	if err != nil {
		fmt.Printf("Failed to generate metadata: %v\n", err)
	} else {
		fmt.Printf("Metadata: %+v\n", metadata)
	}
}

// MockProvider is a mock LLM provider for testing
type MockProvider struct {
	*llm.BaseProvider
}

// GenerateText generates mock text
func (p *MockProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	if len(prompt) == 0 {
		return "", fmt.Errorf("empty prompt")
	}
	
	// Generate different responses based on prompt content
	if contains(prompt, "title") {
		return "Machine Learning Fundamentals", nil
	}
	if contains(prompt, "description") {
		return "This course provides a comprehensive introduction to machine learning, covering supervised and unsupervised learning techniques, neural networks, and deep learning fundamentals. Perfect for beginners looking to understand the core concepts of ML.",
		nil
	}
	if contains(prompt, "enhance") {
		return "# Python Basics\n\n## Learning Objectives\nBy the end of this lesson, you will understand:\n- What Python is and why it's popular\n- Basic Python syntax and data types\n- How to write simple Python programs\n\n## What is Python?\nPython is a high-level, interpreted programming language known for its simplicity and readability. Created by Guido van Rossum and first released in 1991, Python has become one of the most popular programming languages in the world.\n\n## Why Python?\n- Easy to learn syntax\n- Extensive standard library\n- Large, active community\n- Versatile - used for web development, data science, AI, and more\n\n## Basic Syntax\nPython code is known for its clean, readable syntax. For example:\n```python\nprint(\"Hello, World!\")\n```\n\n## Summary\nPython is a powerful yet beginner-friendly programming language that serves as an excellent starting point for your programming journey.",
		nil
	}
	if contains(prompt, "interactive") {
		return `[{"type":"quiz","title":"Python Basics Quiz","content":"What makes Python a good language for beginners?"},{"type":"exercise","title":"Your First Program","content":"Write a Python program that prints your name"},{"type":"code","title":"Fix the Code","content":"Identify and fix the error in: print 'Hello, World'"}]`,
		nil
	}
	if contains(prompt, "metadata") {
		return `{"difficulty":"beginner","duration_hours":3.0,"prerequisites":["Basic computer literacy"],"learning_outcomes":["Understand Python basics","Write simple Python programs","Know Python's main applications"],"target_audience":"Absolute beginners to programming","tags":["python","programming","basics","beginner"]}`,
		nil
	}
	
	// Default response
	return "This is a mock response for testing purposes.", nil
}

// GetCostEstimate returns 0 for mock provider
func (p *MockProvider) GetCostEstimate(textLength int) float64 {
	return 0.0
}

// contains checks if a string contains another string (case insensitive)
func contains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}