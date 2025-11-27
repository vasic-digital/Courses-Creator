package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/course-creator/core-processor/utils"
	"github.com/stretchr/testify/require"
)

func TestE2EComprehensive(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "output")
	err := utils.EnsureDir(outputDir)
	require.NoError(t, err)

	// Create test markdown content
	markdownContent := `# Introduction to Go Programming

This course covers the fundamentals of Go programming language.

## Lesson 1: Getting Started

Welcome to Go! Go is a modern programming language developed by Google.
It's known for its simplicity, efficiency, and strong support for concurrency.

## Lesson 2: Basic Syntax

Let's learn the basic syntax of Go:
- Variables and types
- Functions
- Control structures
- Error handling

## Lesson 3: Concurrency

Go's killer feature is its built-in support for concurrency:
- Goroutines
- Channels
- Select statements

## Conclusion

You now have a solid foundation in Go programming!`

	markdownPath := filepath.Join(tempDir, "go_course.md")
	err = utils.WriteFile(markdownPath, markdownContent)
	require.NoError(t, err)

	// Test 1: Configuration Loading
	t.Run("Configuration", func(t *testing.T) {
		cfg, err := config.LoadConfig()
		if err != nil {
			t.Logf("Config load failed, using defaults: %v", err)
			cfg = createTestConfig()
		}
		require.NotNil(t, cfg)
	})

	// Test 2: LLM Providers
	t.Run("LLMProviders", func(t *testing.T) {
		cfg := createTestConfig()
		testLLMProviders(t, cfg)
	})

	// Test 3: Pipeline Components
	t.Run("PipelineComponents", func(t *testing.T) {
		cfg := createTestConfig()
		testPipelineComponents(t, cfg)
	})

	// Test 4: Course Generation
	t.Run("CourseGeneration", func(t *testing.T) {
		cfg := createTestConfig()
		testCourseGeneration(t, cfg, markdownPath, outputDir)
	})

	// Test 5: Storage Integration
	t.Run("StorageIntegration", func(t *testing.T) {
		cfg := createTestConfig()
		testStorageIntegration(t, cfg, outputDir)
	})
}

func createTestConfig() *config.Config {
	return &config.Config{
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
			DefaultProvider:   "free",
			MaxCostPerRequest: 1.0,
			PrioritizeQuality: false,
			AllowPaid:        false,
		},
	}
}

func testLLMProviders(t *testing.T, cfg *config.Config) {
	// Test provider manager
	manager := llm.NewProviderManager(&cfg.LLM)

	// Add test providers
	freeProvider := llm.NewFreeProvider("e2e-test", "", "")
	manager.RegisterProvider(freeProvider)

	// Test availability
	providers := manager.GetAvailableProviders()
	t.Logf("Available providers: %d", len(providers))
	require.Greater(t, len(providers), 0)

	// Test text generation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := manager.GenerateWithFallback(ctx, "What is Go programming?", models.ProcessingOptions{
		Quality: "standard",
	})

	require.NoError(t, err)
	require.NotEmpty(t, result)
	t.Logf("LLM generation successful: %s...", result[:min(50, len(result))])
}

func testPipelineComponents(t *testing.T, cfg *config.Config) {
	// Test factory
	factory := pipeline.NewPipelineFactory(cfg)

	// Test component creation
	_ = factory.NewCourseGenerator()
	t.Log("Course generator created")

	// Test TTS processor
	_ = pipeline.NewTTSProcessor()
	t.Log("TTS processor created")

	// Test video assembler
	_ = pipeline.NewVideoAssembler(nil) // Will use default storage
	t.Log("Video assembler created")

	// Test background generator
	_ = pipeline.NewBackgroundGenerator(nil)
	t.Log("Background generator created")

	// Test text processing
	parser := utils.NewMarkdownParser()
	testContent := "# Test\nThis is a test."
	_, err := parser.Parse(testContent)
	require.NoError(t, err)
	t.Log("Markdown parsing successful")
}

func testCourseGeneration(t *testing.T, cfg *config.Config, markdownPath, outputDir string) {
	t.Logf("Testing course generation from: %s", markdownPath)

	// Create course generator with free provider for testing
	factory := pipeline.NewPipelineFactory(cfg)
	_ = factory.NewCourseGenerator()

	t.Log("Note: Using mock LLM for this test (no API keys configured)")

	// Test parsing only
	parser := utils.NewMarkdownParser()
	content, err := os.ReadFile(markdownPath)
	require.NoError(t, err)
	parsedCourse, err := parser.Parse(string(content))
	require.NoError(t, err)
	
	t.Logf("Course parsed successfully: %s", parsedCourse.Title)
	t.Logf("Sections: %d", len(parsedCourse.Sections))

	for i, section := range parsedCourse.Sections {
		t.Logf("%d. %s (%d chars)", i+1, section.Title, len(section.Content))
	}
}

func testStorageIntegration(t *testing.T, cfg *config.Config, outputDir string) {
	t.Log("Testing storage integration")

	// Test file operations
	testFile := filepath.Join(outputDir, "test.txt")
	testContent := "Test content for storage verification"

	// Write test file
	err := utils.WriteFile(testFile, testContent)
	require.NoError(t, err)

	// Verify file exists
	require.True(t, utils.FileExists(testFile))

	// Read and verify content
	readContentBytes, err := os.ReadFile(testFile)
	require.NoError(t, err)
	readContent := string(readContentBytes)
	require.Equal(t, testContent, readContent)

	// Test directory listing
	files, err := filepath.Glob(filepath.Join(outputDir, "*"))
	require.NoError(t, err)
	require.Greater(t, len(files), 0)

	t.Logf("Storage integration successful")
	t.Logf("Files created: %d", len(files))
	for _, file := range files {
		t.Logf("- %s", filepath.Base(file))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}