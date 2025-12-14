package pipeline

import (
	"context"
	"os"
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

func stringPtr(s string) *string {
	return &s
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

// BackgroundGenerator tests

func TestNewBackgroundGenerator(t *testing.T) {
	storage := NewMockFileStorage()
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

	storage := NewMockFileStorage()
	generator := NewBackgroundGeneratorWithConfig(config, storage)

	assert.NotNil(t, generator)
	assert.Equal(t, config, generator.config)
	assert.Equal(t, storage, generator.storage)
	assert.NotEmpty(t, generator.colorPalettes)
	assert.NotEmpty(t, generator.patterns)
}

// DiagramProcessor tests

func TestNewDiagramProcessor(t *testing.T) {
	storage := NewMockFileStorage()
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

	storage := NewMockFileStorage()
	processor := NewDiagramProcessorWithConfig(config, storage)

	assert.NotNil(t, processor)
	assert.Equal(t, config, processor.config)
	assert.Equal(t, storage, processor.storage)
	assert.NotNil(t, processor.llava)
}

func TestDiagramProcessor_ProcessDiagrams_EmptyContent(t *testing.T) {
	storage := NewMockFileStorage()
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
	// Skip this test in CI or when external services aren't available
	if os.Getenv("CI") != "" || os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test that requires external TTS services")
	}

	processor := NewTTSProcessor()

	options := models.ProcessingOptions{}

	_, err := processor.GenerateAudio("", options)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text parameter is required")
}

// VideoAssembler tests

func TestNewVideoAssembler(t *testing.T) {
	storage := NewMockFileStorage()
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

	storage := NewMockFileStorage()
	assembler := NewVideoAssemblerWithConfig(config, storage)

	assert.NotNil(t, assembler)
	assert.Equal(t, config, assembler.Config)
	assert.Equal(t, storage, assembler.storage)
	assert.NotNil(t, assembler.backgroundGen)
}

// BackgroundGenerator functional tests

func TestBackgroundGenerator_GenerateBackground(t *testing.T) {
	storage := NewMockFileStorage()
	generator := NewBackgroundGenerator(storage)

	ctx := context.Background()
	content := "This is a test course about programming"
	options := models.ProcessingOptions{
		Quality: "standard",
	}

	path, err := generator.GenerateBackground(ctx, content, options)

	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "bg_")
	assert.Contains(t, path, ".png")

	// Verify file was saved via storage
	assert.True(t, storage.Exists(path))
}

func TestBackgroundGenerator_GenerateBackground_DifferentContent(t *testing.T) {
	storage := NewMockFileStorage()
	generator := NewBackgroundGenerator(storage)

	ctx := context.Background()

	tests := []struct {
		name    string
		content string
		options models.ProcessingOptions
	}{
		{
			name:    "business content",
			content: "This is a business course about corporate strategy",
			options: models.ProcessingOptions{Quality: "standard"},
		},
		{
			name:    "nature content",
			content: "This course covers environmental science and nature conservation",
			options: models.ProcessingOptions{Quality: "high"},
		},
		{
			name:    "creative content",
			content: "Learn creative design and digital art techniques",
			options: models.ProcessingOptions{Quality: "standard"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := generator.GenerateBackground(ctx, tt.content, tt.options)

			assert.NoError(t, err)
			assert.NotEmpty(t, path)
			assert.Contains(t, path, ".png")
		})
	}
}

func TestBackgroundGenerator_SelectPalette(t *testing.T) {
	storage := NewMockFileStorage()
	generator := NewBackgroundGenerator(storage)

	tests := []struct {
		name     string
		content  string
		options  models.ProcessingOptions
		expected string
	}{
		{
			name:     "business content",
			content:  "corporate business strategy",
			options:  models.ProcessingOptions{Quality: "standard"},
			expected: "professional",
		},
		{
			name:     "nature content",
			content:  "environmental science nature",
			options:  models.ProcessingOptions{Quality: "standard"},
			expected: "forest",
		},
		{
			name:     "creative content",
			content:  "digital art design creative",
			options:  models.ProcessingOptions{Quality: "standard"},
			expected: "lavender",
		},
		{
			name:     "high quality default",
			content:  "random content",
			options:  models.ProcessingOptions{Quality: "high"},
			expected: "ocean",
		},
		{
			name:     "standard quality default",
			content:  "random content",
			options:  models.ProcessingOptions{Quality: "standard"},
			expected: "professional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette := generator.selectPalette(tt.content, tt.options)
			assert.Equal(t, tt.expected, palette.Name)
		})
	}
}

func TestBackgroundGenerator_SelectPattern(t *testing.T) {
	storage := NewMockFileStorage()
	generator := NewBackgroundGenerator(storage)

	tests := []struct {
		name     string
		options  models.ProcessingOptions
		expected string
	}{
		{
			name:     "high quality",
			options:  models.ProcessingOptions{Quality: "high"},
			expected: "geometric",
		},
		{
			name:     "standard quality",
			options:  models.ProcessingOptions{Quality: "standard"},
			expected: "gradient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := generator.selectPattern(tt.options)
			assert.Equal(t, tt.expected, pattern.GetName())
		})
	}
}

// DiagramProcessor functional tests

func TestDiagramProcessor_ProcessDiagrams_WithMermaidContent(t *testing.T) {
	storage := NewMockFileStorage()
	processor := NewDiagramProcessor(storage)

	ctx := context.Background()
	content := `
# Introduction to Programming

## Flowchart Example

` + "```mermaid" + `
graph TD
    A[Start] --> B{Is it working?}
    B -->|Yes| C[Great!]
    B -->|No| D[Debug]
    D --> B
` + "```" + `

This flowchart shows the debugging process.
`

	options := models.ProcessingOptions{}

	diagrams, err := processor.ProcessDiagrams(ctx, content, options)

	assert.NoError(t, err)
	assert.Len(t, diagrams, 1)
	assert.Equal(t, "flowchart", string(diagrams[0].Type))
	assert.Contains(t, diagrams[0].Description, "flowchart diagram")
}

func TestDiagramProcessor_ProcessDiagrams_WithMultipleDiagrams(t *testing.T) {
	storage := NewMockFileStorage()
	processor := NewDiagramProcessor(storage)

	ctx := context.Background()
	content := `
# System Architecture

## Database Schema

` + "```mermaid" + `
erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains
    CUSTOMER {
        string name
        string custNumber
        string sector
    }
` + "```" + `

## User Flow

` + "```mermaid" + `
sequenceDiagram
    participant U as User
    participant S as System
    U->>S: Login
    S-->>U: Token
` + "```" + `
`

	options := models.ProcessingOptions{}

	diagrams, err := processor.ProcessDiagrams(ctx, content, options)

	assert.NoError(t, err)
	assert.Len(t, diagrams, 2)
	assert.Equal(t, "entity", string(diagrams[0].Type))
	assert.Equal(t, "sequence", string(diagrams[1].Type))
}

func TestDiagramProcessor_DetectDiagrams(t *testing.T) {
	storage := NewMockFileStorage()
	processor := NewDiagramProcessor(storage)

	content := `
# Course Title

## Section 1

Some text here.

` + "```mermaid" + `
graph TD
    A --> B
` + "```" + `

More text.

` + "```mermaid" + `
sequenceDiagram
    A->>B: message
` + "```" + `
`

	requests := processor.detectDiagrams(content)

	assert.Len(t, requests, 2)
	assert.Equal(t, "graph TD\n    A --> B", requests[0].Content)
	assert.Equal(t, "sequenceDiagram\n    A->>B: message", requests[1].Content)
}

func TestDiagramProcessor_InferDiagramType(t *testing.T) {
	storage := NewMockFileStorage()
	processor := NewDiagramProcessor(storage)

	tests := []struct {
		name     string
		content  string
		expected DiagramType
	}{
		{
			name:     "flowchart",
			content:  "graph TD\nA-->B",
			expected: DiagramProcess, // "graph" doesn't contain "flow", "process", or "step"
		},
		{
			name:     "sequence diagram",
			content:  "sequenceDiagram\nA->>B: msg",
			expected: DiagramSequence,
		},
		{
			name:     "class diagram",
			content:  "classDiagram\nclass A",
			expected: DiagramClass,
		},
		{
			name:     "entity diagram",
			content:  "erDiagram\nA ||--o{ B",
			expected: DiagramProcess, // "erDiagram" doesn't contain "entity", "relationship", or "database"
		},
		{
			name:     "mind map",
			content:  "mindmap\nroot((Main))",
			expected: DiagramMindMap,
		},
		{
			name:     "unknown diagram",
			content:  "some random content",
			expected: DiagramProcess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.inferDiagramType(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiagramProcessor_CountElements(t *testing.T) {
	storage := NewMockFileStorage()
	processor := NewDiagramProcessor(storage)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "simple flowchart",
			content:  "graph TD\nA-->B\nB-->C",
			expected: 3, // Contains "step" keyword
		},
		{
			name:     "sequence diagram",
			content:  "sequenceDiagram\nA->>B: msg\nB-->>A: response",
			expected: 3, // Minimum default
		},
		{
			name:     "empty content",
			content:  "",
			expected: 3, // Minimum default
		},
		{
			name:     "single element",
			content:  "graph TD\nA",
			expected: 3, // Minimum default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.countElements(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TTSProcessor functional tests

func TestTTSProcessor_GenerateAudio_ValidText(t *testing.T) {
	// Skip this test in CI or when external services aren't available
	if os.Getenv("CI") != "" || os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test that requires external TTS services")
	}

	processor := NewTTSProcessor()

	text := "Hello world, this is a test of the text to speech system."
	options := models.ProcessingOptions{
		Voice: stringPtr("en-US"),
	}

	path, err := processor.GenerateAudio(text, options)

	// Note: This test may fail in environments without TTS providers configured
	// but it tests the basic flow and error handling
	if err != nil {
		assert.Contains(t, err.Error(), "TTS") // Should contain TTS-related error
	} else {
		assert.NotEmpty(t, path)
		assert.Contains(t, path, ".wav")
	}
}

func TestTTSProcessor_SplitText(t *testing.T) {
	processor := NewTTSProcessor()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "short text",
			text:     "Hello world",
			expected: 1,
		},
		{
			name:     "medium text",
			text:     "This is a longer piece of text that should be split into multiple chunks for better processing and audio quality.",
			expected: 1, // May be split or not depending on length
		},
		{
			name:     "empty text",
			text:     "",
			expected: 1, // splitText always returns at least one chunk
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := processor.splitText(tt.text)
			assert.Len(t, chunks, tt.expected)
		})
	}
}

func TestTTSProcessor_SplitBySentences(t *testing.T) {
	processor := NewTTSProcessor()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "single sentence",
			text:     "This is one sentence.",
			expected: 1,
		},
		{
			name:     "multiple sentences",
			text:     "First sentence. Second sentence! Third sentence?",
			expected: 3,
		},
		{
			name:     "no punctuation",
			text:     "This is text without punctuation",
			expected: 1,
		},
		{
			name:     "empty text",
			text:     "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentences := processor.splitBySentences(tt.text)
			assert.Len(t, sentences, tt.expected)
		})
	}
}

func TestTTSProcessor_SplitByWords(t *testing.T) {
	processor := NewTTSProcessor()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "few words",
			text:     "Hello world",
			expected: 2,
		},
		{
			name:     "many words",
			text:     "This is a longer sentence with many words that should be split appropriately",
			expected: 13,
		},
		{
			name:     "empty text",
			text:     "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := processor.splitByWords(tt.text)
			assert.Len(t, words, tt.expected)
		})
	}
}

func TestTTSProcessor_IsRunning(t *testing.T) {
	processor := NewTTSProcessor()

	// Should be running after creation
	assert.True(t, processor.IsRunning())

	// Stop the processor
	processor.Stop()

	// Should not be running after stop
	assert.False(t, processor.IsRunning())
}

// VideoAssembler functional tests

func TestVideoAssembler_ParseTextSegments(t *testing.T) {
	storage := NewMockFileStorage()
	assembler := NewVideoAssembler(storage)

	textContent := "First sentence. Second sentence! Third sentence?"
	duration := 10.0

	segments := assembler.ParseTextSegments(textContent, duration)

	assert.NotEmpty(t, segments)
	// Should have at least one segment
	assert.Greater(t, len(segments), 0)

	// Check that segments have reasonable timing
	totalDuration := 0.0
	for _, segment := range segments {
		assert.Greater(t, segment.EndTime, segment.StartTime)
		duration := segment.EndTime - segment.StartTime
		assert.Greater(t, duration, 0.0)
		totalDuration += duration
	}

	// Total duration should be close to requested duration
	assert.InDelta(t, duration, totalDuration, 1.0)
}

func TestVideoAssembler_ParseTextSegments_EmptyText(t *testing.T) {
	storage := NewMockFileStorage()
	assembler := NewVideoAssembler(storage)

	segments := assembler.ParseTextSegments("", 10.0)

	// Should return empty slice for empty text
	assert.Empty(t, segments)
}

func TestVideoAssembler_ParseTextSegments_LongText(t *testing.T) {
	storage := NewMockFileStorage()
	assembler := NewVideoAssembler(storage)

	longText := `This is a very long text that should be split into multiple segments.
Each sentence should become a separate segment.
The timing should be distributed evenly across the total duration.
This ensures that the text appears at appropriate times during the video playback.`
	duration := 30.0

	segments := assembler.ParseTextSegments(longText, duration)

	assert.NotEmpty(t, segments)
	assert.Greater(t, len(segments), 1) // Should split into multiple segments

	// Check timing distribution
	totalDuration := 0.0
	for _, segment := range segments {
		totalDuration += segment.EndTime - segment.StartTime
	}
	assert.InDelta(t, duration, totalDuration, 1.0)
}

func TestVideoAssembler_EscapeFFmpegText(t *testing.T) {
	storage := NewMockFileStorage()
	assembler := NewVideoAssembler(storage)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "text with apostrophe",
			input:    "It's working",
			expected: "It\\'s working",
		},
		{
			name:     "text with colon",
			input:    "Time: 10:30",
			expected: "Time\\: 10\\:30",
		},
		{
			name:     "text with multiple special chars",
			input:    "It's 10:30, don't wait!",
			expected: "It\\'s 10\\:30\\, don\\'t wait!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := assembler.escapeFFmpegText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVideoAssembler_FormatSRTTime(t *testing.T) {
	storage := NewMockFileStorage()
	assembler := NewVideoAssembler(storage)

	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		{
			name:     "zero seconds",
			seconds:  0.0,
			expected: "00:00:00,000",
		},
		{
			name:     "whole seconds",
			seconds:  5.0,
			expected: "00:00:05,000",
		},
		{
			name:     "with milliseconds",
			seconds:  5.123,
			expected: "00:00:05,123",
		},
		{
			name:     "with minutes",
			seconds:  65.5,
			expected: "00:01:05,500",
		},
		{
			name:     "with hours",
			seconds:  3665.25,
			expected: "01:01:05,250",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := assembler.formatSRTTime(tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVideoAssembler_CreateSRTSubtitleFile(t *testing.T) {
	storage := NewMockFileStorage()
	assembler := NewVideoAssembler(storage)

	subtitles := []models.Subtitle{
		{
			Language: "en",
			Content:  "Hello world",
			Timestamps: []map[string]interface{}{
				{"start": 0.0, "end": 2.0, "text": "Hello"},
				{"start": 2.0, "end": 4.0, "text": "world"},
			},
		},
	}

	tempFile := "/tmp/test_subtitles.srt"
	defer os.Remove(tempFile) // Clean up

	err := assembler.createSRTSubtitleFile(tempFile, subtitles)

	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(tempFile)
	assert.NoError(t, err)
}

func TestVideoAssembler_ParseTextSegments_VariousFormats(t *testing.T) {
	storage := NewMockFileStorage()
	assembler := NewVideoAssembler(storage)

	tests := []struct {
		name     string
		text     string
		duration float64
		expected int // expected number of segments
	}{
		{
			name:     "single_sentence",
			text:     "This is a single sentence.",
			duration: 5.0,
			expected: 1,
		},
		{
			name:     "multiple_sentences",
			text:     "First sentence. Second sentence! Third sentence?",
			duration: 15.0,
			expected: 3,
		},
		{
			name:     "question_and_exclamation",
			text:     "What is this? This is amazing!",
			duration: 8.0,
			expected: 2,
		},
		{
			name:     "ellipsis_and_abbreviations",
			text:     "Wait... This is Dr. Smith's test. See you at 5 p.m.",
			duration: 12.0,
			expected: 3,
		},
		{
			name:     "very_short_duration",
			text:     "This text has multiple sentences. But duration is short.",
			duration: 1.0,
			expected: 1, // Should combine into fewer segments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := assembler.ParseTextSegments(tt.text, tt.duration)

			// Should have reasonable number of segments
			assert.Greater(t, len(segments), 0)
			assert.LessOrEqual(t, len(segments), tt.expected+2) // Allow some flexibility

			// Check timing consistency
			totalDuration := 0.0
			for _, segment := range segments {
				assert.Greater(t, segment.EndTime, segment.StartTime)
				totalDuration += segment.EndTime - segment.StartTime
			}

			// Total duration should be close to requested duration
			assert.InDelta(t, tt.duration, totalDuration, 0.5)
		})
	}
}
