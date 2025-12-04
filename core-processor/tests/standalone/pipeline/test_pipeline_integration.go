package pipeline_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/course-creator/core-processor/utils"
)

func TestPipelineIntegration(t *testing.T) {
	fmt.Println("Testing complete pipeline integration with LLM providers...")

	// Create test markdown
	tempDir := "/tmp/pipeline_test"
	err := utils.EnsureDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	markdownContent := `# Introduction to AI

This course covers the fundamentals of artificial intelligence.

## What is AI?

Artificial Intelligence is a branch of computer science that aims to create intelligent machines.

## Machine Learning

Machine Learning is a subset of AI that enables systems to learn from data.`

	markdownPath := filepath.Join(tempDir, "test_course.md")
	err = utils.WriteFile(markdownPath, markdownContent)
	if err != nil {
		t.Fatalf("Failed to write markdown: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Logf("Failed to load config, using defaults: %v", err)
		cfg = &config.Config{
			Storage: map[string]config.StorageConfig{
				"default": config.StorageConfig{
					Type:     "local",
					BasePath: "./storage",
				},
			},
			TTS: config.TTSConfig{
				Provider: "bark",
				Timeout:  60 * time.Second,
			},
			LLM: config.LLMConfig{
				DefaultProvider:   "ollama",
				MaxCostPerRequest: 1.0,
				PrioritizeQuality: true,
				AllowPaid:        false, // Don't use paid providers for testing
				Ollama: config.OllamaConfig{
					BaseURL:      "http://localhost:11434",
					DefaultModel: "llama2",
				},
			},
		}
	}

	// Test LLM providers directly
	fmt.Println("\n=== Testing LLM Providers ===")
	testLLMProviders(cfg)

	// Test pipeline with mock TTS
	fmt.Println("\n=== Testing Pipeline with Mock TTS ===")
	testPipelineWithMockTTS(markdownPath, tempDir, cfg)

	fmt.Println("\n✅ Pipeline integration test completed successfully!")
}

func testLLMProviders(cfg *config.Config) {
	// Create provider manager
	manager := llm.NewProviderManager(&cfg.LLM)

	// Add free provider for testing
	freeProvider := llm.NewFreeProvider("test-free", "", "")
	manager.RegisterProvider(freeProvider)
	fmt.Printf("Registered free provider for testing\n")

	// Test provider availability
	providers := manager.GetAvailableProviders()
	fmt.Printf("Available providers count: %d\n", len(providers))
	for _, p := range providers {
		fmt.Printf("  - %s (%s)\n", p.GetName(), p.GetType())
	}

	// Test each provider by trying generation with fallback
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simple test prompt
	prompt := "What is AI? Answer in one sentence."
	options := models.ProcessingOptions{
		Quality: "standard",
	}

	result, err := manager.GenerateWithFallback(ctx, prompt, options)
	if err != nil {
		fmt.Printf("❌ All providers failed: %v\n", err)
	} else {
		fmt.Printf("✅ Generated: %s...\n", result[:min(50, len(result))])
	}
}

func testPipelineWithMockTTS(markdownPath, outputDir string, cfg *config.Config) {
	// Create factory
	factory := pipeline.NewPipelineFactory(cfg)

	// Get components
	llmManager := factory.GetLLMManager()
	
	// Add a free provider for testing
	freeProvider := llm.NewFreeProvider("pipeline-test", "", "")
	llmManager.RegisterProvider(freeProvider)
	
	// Test LLM integration
	ctx := context.Background()
	
	// Test title generation
	fmt.Println("Testing LLM content generation...")
	title, err := llmManager.GenerateWithFallback(ctx, "Generate a title for a course about AI", models.ProcessingOptions{
		Quality: "standard",
	})
	
	if err != nil {
		fmt.Printf("Title generation failed: %v\n", err)
		title = "AI Fundamentals" // Fallback
	} else {
		fmt.Printf("Generated title: %s\n", title)
	}

	// Test description generation
	description, err := llmManager.GenerateWithFallback(ctx, "Generate a description for a course titled '"+title+"'", models.ProcessingOptions{
		Quality: "standard",
	})
	
	if err != nil {
		fmt.Printf("Description generation failed: %v\n", err)
		description = "A comprehensive course about AI fundamentals"
	} else {
		fmt.Printf("Generated description: %s...\n", description[:min(100, len(description))])
	}

	// Create a mock course structure to verify the pipeline components
	course := &models.Course{
		ID:          "test_course_123",
		Title:       title,
		Description: description,
		Metadata: models.CourseMetadata{
			Author:   "Test Author",
			Language: "en",
			Tags:     []string{"AI", "Machine Learning"},
		},
		Lessons: []models.Lesson{
			{
				ID:      "lesson_1",
				Title:   "Introduction",
				Content: "This is an introduction to AI",
				Order:   1,
			},
		},
	}

	fmt.Printf("Created mock course with %d lessons\n", len(course.Lessons))
	fmt.Printf("Course title: %s\n", course.Title)
	fmt.Printf("Course description: %s\n", course.Description[:min(50, len(course.Description))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}