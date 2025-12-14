package pipeline

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/course-creator/core-processor/config"
	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
)

// MockCourseContentGenerator implements llm.CourseContentGenerator for testing
type MockCourseContentGenerator struct{}

func (m *MockCourseContentGenerator) GenerateCourseTitle(ctx context.Context, content string) (string, error) {
	return "Test Course", nil
}

func (m *MockCourseContentGenerator) GenerateCourseDescription(ctx context.Context, title, content string) (string, error) {
	return "Test course description", nil
}

func (m *MockCourseContentGenerator) GenerateLessonContent(ctx context.Context, title, rawContent string) (string, error) {
	return "Test lesson content", nil
}

func (m *MockCourseContentGenerator) GenerateInteractiveElements(ctx context.Context, lessonContent string) ([]string, error) {
	return []string{"quiz", "exercise"}, nil
}

func (m *MockCourseContentGenerator) GenerateMetadata(ctx context.Context, title, description string) (map[string]interface{}, error) {
	return map[string]interface{}{"difficulty": "beginner"}, nil
}

func (m *MockCourseContentGenerator) GetAvailableProviders() []interface{} {
	return []interface{}{"ollama"}
}

func (m *MockCourseContentGenerator) TestProviders(ctx context.Context) map[string]error {
	return map[string]error{}
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

// MockStorage implements storage.StorageInterface for testing
type MockStorage struct{}

func (m *MockStorage) Save(path string, data []byte) error {
	return nil
}

func (m *MockStorage) SaveReader(path string, reader io.Reader) error {
	return nil
}

func (m *MockStorage) Load(path string) ([]byte, error) {
	return []byte("mock data"), nil
}

func (m *MockStorage) Delete(path string) error {
	return nil
}

func (m *MockStorage) Exists(path string) bool {
	return true
}

func (m *MockStorage) List(dir string) ([]string, error) {
	return []string{"file1.jpg", "file2.jpg"}, nil
}

func (m *MockStorage) CreateDir(path string) error {
	return nil
}

func (m *MockStorage) GetURL(path string) string {
	return "http://localhost:8080/storage/" + path
}

func (m *MockStorage) GetSize(path string) (int64, error) {
	return 1024, nil
}

// BackgroundGenerator tests

func TestNewBackgroundGenerator(t *testing.T) {
	storage := &MockStorage{}
	generator := NewBackgroundGenerator(storage)

	assert.NotNil(t, generator)
	assert.NotNil(t, generator.storage)
	assert.NotEmpty(t, generator.colorPalettes)
	assert.NotEmpty(t, generator.patterns)
}

func TestNewBackgroundGeneratorWithConfig(t *testing.T) {
	config := BackgroundConfig{
		Width:      1920,
		Height:     1080,
		Quality:    90,
		OutputDir:  "/tmp/output",
		CacheDir:   "/tmp/cache",
		TempDir:    "/tmp/temp",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	storage := &MockStorage{}
	generator := NewBackgroundGeneratorWithConfig(config, storage)

	assert.NotNil(t, generator)
	assert.Equal(t, config, generator.config)
	assert.Equal(t, storage, generator.storage)
	assert.NotEmpty(t, generator.colorPalettes)
	assert.NotEmpty(t, generator.patterns)
}

// DiagramProcessor tests

func TestNewDiagramProcessor(t *testing.T) {
	storage := &MockStorage{}
	processor := NewDiagramProcessor(storage)

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.storage)
	assert.NotNil(t, processor.llava)
}

func TestNewDiagramProcessorWithConfig(t *testing.T) {
	config := DiagramConfig{
		Width:      800,
		Height:     600,
		Quality:    90,
		OutputDir:  "/tmp/output",
		CacheDir:   "/tmp/cache",
		TempDir:    "/tmp/temp",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	storage := &MockStorage{}
	processor := NewDiagramProcessorWithConfig(config, storage)

	assert.NotNil(t, processor)
	assert.Equal(t, config, processor.config)
	assert.Equal(t, storage, processor.storage)
	assert.NotNil(t, processor.llava)
}

func TestDiagramProcessor_ProcessDiagrams_EmptyContent(t *testing.T) {
	storage := &MockStorage{}
	processor := NewDiagramProcessor(storage)

	ctx := context.Background()
	options := models.ProcessingOptions{}

	diagrams, err := processor.ProcessDiagrams(ctx, "", options)

	assert.NoError(t, err)
	assert.Empty(t, diagrams)
}

// TTSProcessor tests

func TestNewTTSProcessor(t *testing.T) {
	processor := NewTTSProcessor()

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.BarkServer)
	assert.True(t, processor.Running)
}

func TestNewTTSProcessorWithConfig(t *testing.T) {
	config := TTSConfig{
		DefaultProvider: TTSProviderBark,
		OutputDir:       "/tmp/output",
		SampleRate:      24000,
		BitRate:         128000,
		Format:          "wav",
		Timeout:         60 * time.Second,
		MaxRetries:      3,
		ChunkSize:       200,
		Parallelism:     2,
	}

	processor := NewTTSProcessorWithConfig(config)

	assert.NotNil(t, processor)
	assert.Equal(t, config, processor.Config)
	assert.NotNil(t, processor.BarkServer)
	assert.True(t, processor.Running)
}

func TestTTSProcessor_GenerateAudio_EmptyText(t *testing.T) {
	processor := NewTTSProcessor()

	options := models.ProcessingOptions{}

	_, err := processor.GenerateAudio("", options)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text cannot be empty")
}

// VideoAssembler tests

func TestNewVideoAssembler(t *testing.T) {
	storage := &MockStorage{}
	assembler := NewVideoAssembler(storage)

	assert.NotNil(t, assembler)
	assert.NotNil(t, assembler.storage)
	assert.NotNil(t, assembler.backgroundGen)
}

func TestNewVideoAssemblerWithConfig(t *testing.T) {
	config := VideoConfig{
		Quality: VideoQuality{
			Width:       1920,
			Height:      1080,
			Bitrate:     "2000k",
			Framerate:   30,
			Codec:       "libx264",
			PixelFormat: "yuv420p",
		},
		OutputDir:   "/tmp/output",
		FFmpegPath:  "/usr/bin/ffmpeg",
		FFprobePath: "/usr/bin/ffprobe",
		TempDir:     "/tmp/temp",
		Timeout:     300 * time.Second,
		MaxRetries:  3,
	}

	storage := &MockStorage{}
	assembler := NewVideoAssemblerWithConfig(config, storage)

	assert.NotNil(t, assembler)
	assert.Equal(t, config, assembler.Config)
	assert.Equal(t, storage, assembler.storage)
	assert.NotNil(t, assembler.backgroundGen)
}
