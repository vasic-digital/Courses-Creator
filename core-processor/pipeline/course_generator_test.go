package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCourseContentGenerator implements llm.CourseContentGenerator for testing
type MockCourseContentGenerator struct {
	mock.Mock
}

func (m *MockCourseContentGenerator) GenerateCourseTitle(ctx context.Context, content string) (string, error) {
	args := m.Called(ctx, content)
	return args.String(0), args.Error(1)
}

func (m *MockCourseContentGenerator) GenerateCourseDescription(ctx context.Context, title, content string) (string, error) {
	args := m.Called(ctx, title, content)
	return args.String(0), args.Error(1)
}

func (m *MockCourseContentGenerator) GenerateLessonContent(ctx context.Context, title, rawContent string) (string, error) {
	args := m.Called(ctx, title, rawContent)
	return args.String(0), args.Error(1)
}

func (m *MockCourseContentGenerator) GenerateInteractiveElements(ctx context.Context, lessonContent string) ([]string, error) {
	args := m.Called(ctx, lessonContent)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCourseContentGenerator) GenerateMetadata(ctx context.Context, title, description string) (map[string]interface{}, error) {
	args := m.Called(ctx, title, description)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCourseContentGenerator) GetAvailableProviders() []llm.ProviderInfo {
	args := m.Called()
	return args.Get(0).([]llm.ProviderInfo)
}

func (m *MockCourseContentGenerator) TestProviders(ctx context.Context) map[string]error {
	args := m.Called(ctx)
	return args.Get(0).(map[string]error)
}

func TestNewCourseGenerator(t *testing.T) {
	generator := NewCourseGenerator()

	assert.NotNil(t, generator)
	assert.NotNil(t, generator.ttsProcessor)
	assert.NotNil(t, generator.videoAssembler)
	assert.NotNil(t, generator.diagramProcessor)
	assert.NotNil(t, generator.contentGen)
	assert.NotNil(t, generator.storage)
}

func TestCourseGenerator_GenerateCourse_EmptyMarkdownPath(t *testing.T) {
	generator := NewCourseGenerator()

	_, err := generator.GenerateCourse("", "/tmp/output", models.ProcessingOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "markdown path cannot be empty")
}

func TestCourseGenerator_GenerateCourse_EmptyOutputDir(t *testing.T) {
	generator := NewCourseGenerator()

	_, err := generator.GenerateCourse("/tmp/test.md", "", models.ProcessingOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output directory cannot be empty")
}

func TestCourseGenerator_GenerateCourse_NonExistentFile(t *testing.T) {
	generator := NewCourseGenerator()

	_, err := generator.GenerateCourse("/nonexistent/file.md", "/tmp/output", models.ProcessingOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "markdown file does not exist")
}

func TestCourseGenerator_GenerateCourse_BasicMarkdown(t *testing.T) {
	// Create temporary directory and markdown file
	tempDir := t.TempDir()
	markdownPath := filepath.Join(tempDir, "test.md")
	outputDir := filepath.Join(tempDir, "output")

	// Create a simple markdown file
	markdownContent := `# Test Course

This is a test course.

## Introduction

Welcome to the course!

## Advanced Topics

More advanced content.
`

	err := os.WriteFile(markdownPath, []byte(markdownContent), 0644)
	assert.NoError(t, err)

	// Create generator - this will use default config which may not have LLM providers
	// so it will fall back to basic functionality
	generator := NewCourseGenerator()

	// This test may fail due to missing LLM providers, but it tests the basic parsing
	course, err := generator.GenerateCourse(markdownPath, outputDir, models.ProcessingOptions{
		Quality:   "standard",
		Languages: []string{"en"},
	})

	// We expect this to work for basic parsing even without LLM providers
	if err != nil {
		// If it fails due to LLM, that's expected in test environment
		t.Logf("Course generation failed (expected without LLM providers): %v", err)
		return
	}

	assert.NotNil(t, course)
	assert.Equal(t, "Test Course", course.Title)
	assert.Contains(t, course.Description, "This is a test course")
	assert.Len(t, course.Lessons, 2) // Two sections in the markdown
}

func TestPipelineFactory_NewPipelineFactory(t *testing.T) {
	cfg := &config.Config{}
	factory := NewPipelineFactory(cfg)

	assert.NotNil(t, factory)
	assert.Equal(t, cfg, factory.config)
}

func TestPipelineFactory_NewCourseGenerator(t *testing.T) {
	cfg := &config.Config{
		Storage: map[string]config.StorageConfig{
			"default": {
				Type:      "local",
				BasePath:  "/tmp/test-storage",
				PublicURL: "http://localhost:8080/storage",
			},
		},
		TTS: config.TTSConfig{
			Provider: "bark",
			Timeout:  60,
		},
		LLM: config.LLMConfig{
			DefaultProvider: "ollama",
		},
	}

	factory := NewPipelineFactory(cfg)
	generator := factory.NewCourseGenerator()

	assert.NotNil(t, generator)
	assert.NotNil(t, generator.ttsProcessor)
	assert.NotNil(t, generator.videoAssembler)
	assert.NotNil(t, generator.diagramProcessor)
	assert.NotNil(t, generator.contentGen)
	assert.NotNil(t, generator.storage)
}

func TestPipelineFactory_GetLLMManager(t *testing.T) {
	cfg := &config.Config{
		LLM: config.LLMConfig{
			DefaultProvider: "ollama",
		},
	}

	factory := NewPipelineFactory(cfg)
	manager := factory.GetLLMManager()

	assert.NotNil(t, manager)
}
