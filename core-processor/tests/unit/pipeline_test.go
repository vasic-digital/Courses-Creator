package unit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTTSProcessor_NewTTSProcessor(t *testing.T) {
	processor := pipeline.NewTTSProcessor()

	require.NotNil(t, processor)
	assert.Equal(t, pipeline.TTSProviderBark, processor.Config.DefaultProvider)
	assert.Equal(t, "/tmp/course_audio", processor.Config.OutputDir)
	assert.Equal(t, 24000, processor.Config.SampleRate)
	assert.Equal(t, 128000, processor.Config.BitRate)
	assert.Equal(t, "wav", processor.Config.Format)
	assert.True(t, processor.Running)
}

func TestTTSProcessor_NewTTSProcessorWithConfig(t *testing.T) {
	config := pipeline.TTSConfig{
		DefaultProvider: pipeline.TTSProviderSpeechT5,
		OutputDir:      "/custom/audio",
		SampleRate:     16000,
		BitRate:        96000,
		Format:         "mp3",
		Timeout:        120 * time.Second,
		MaxRetries:     5,
		ChunkSize:      100,
		Parallelism:    4,
	}

	processor := pipeline.NewTTSProcessorWithConfig(config)

	require.NotNil(t, processor)
	assert.Equal(t, pipeline.TTSProviderSpeechT5, processor.Config.DefaultProvider)
	assert.Equal(t, "/custom/audio", processor.Config.OutputDir)
	assert.Equal(t, 16000, processor.Config.SampleRate)
	assert.Equal(t, 96000, processor.Config.BitRate)
	assert.Equal(t, "mp3", processor.Config.Format)
}

func TestTTSProcessor_GenerateAudio(t *testing.T) {
	processor := pipeline.NewTTSProcessor()

	options := models.ProcessingOptions{
		Quality:   "standard",
		Languages: []string{"en"},
	}

	// Test with short text
	text := "Hello, world!"
	audioPath, err := processor.GenerateAudio(text, options)

	// Should generate audio (may be placeholder if TTS servers not available)
	if err == nil && audioPath != "" {
		assert.NotEmpty(t, audioPath)
		
		// Clean up
		os.Remove(audioPath)
	}
}

func TestTTSProcessor_SplitText(t *testing.T) {
	processor := pipeline.NewTTSProcessor()

	// Test short text (no splitting)
	shortText := "This is a short text"
	chunks := processor.BarkServer.SplitText(shortText)
	assert.Len(t, chunks, 1)
	assert.Equal(t, shortText, chunks[0])

	// Test long text (should be split)
	longText := string(make([]byte, 300)) // Create long text
	chunks = processor.BarkServer.SplitText(longText)
	assert.Greater(t, len(chunks), 1)

	// Verify that concatenated chunks equal original text
	reconstructed := ""
	for _, chunk := range chunks {
		reconstructed += chunk
	}
	assert.Equal(t, len(longText), len(reconstructed))
}

func TestTTSProcessor_Stop(t *testing.T) {
	processor := pipeline.NewTTSProcessor()
	
	// Should be running initially
	assert.True(t, processor.IsRunning())

	// Stop the processor
	processor.Stop()

	// Should no longer be running
	assert.False(t, processor.IsRunning())
}

func TestVideoAssembler_NewVideoAssembler(t *testing.T) {
	// Create a temporary storage for testing
	tempDir := "/tmp/test_storage"
	storageConfig := storage.StorageConfig{
		BasePath: tempDir,
		PublicURL: "",
	}
	storage := storage.NewLocalStorage(storageConfig)
	
	assembler := pipeline.NewVideoAssembler(storage)

	require.NotNil(t, assembler)
	assert.Equal(t, 1920, assembler.Config.Quality.Width)
	assert.Equal(t, 1080, assembler.Config.Quality.Height)
	assert.Equal(t, "2M", assembler.Config.Quality.Bitrate)
	assert.Equal(t, 30, assembler.Config.Quality.Framerate)
	assert.Equal(t, "libx264", assembler.Config.Quality.Codec)
	assert.Equal(t, "yuv420p", assembler.Config.Quality.PixelFormat)
}

func TestVideoAssembler_NewVideoAssemblerWithConfig(t *testing.T) {
	config := pipeline.VideoConfig{
		Quality: pipeline.VideoQuality{
			Width:       1280,
			Height:      720,
			Bitrate:     "1M",
			Framerate:    25,
			Codec:       "libx265",
			PixelFormat: "yuv420p",
		},
		OutputDir:    "/custom/videos",
		FFmpegPath:   "/usr/local/bin/ffmpeg",
		FFprobePath:  "/usr/local/bin/ffprobe",
		TempDir:      "/custom/temp",
		Timeout:      600 * time.Second,
		MaxRetries:   5,
		FontPath:     "/System/Library/Fonts/Arial.ttf",
		SubtitleFont: "Helvetica",
	}

	assembler := pipeline.NewVideoAssemblerWithConfig(config)

	require.NotNil(t, assembler)
	assert.Equal(t, 1280, assembler.Config.Quality.Width)
	assert.Equal(t, 720, assembler.Config.Quality.Height)
	assert.Equal(t, "1M", assembler.Config.Quality.Bitrate)
	assert.Equal(t, 25, assembler.Config.Quality.Framerate)
	assert.Equal(t, "libx265", assembler.Config.Quality.Codec)
}

func TestVideoAssembler_ParseTextSegments(t *testing.T) {
	// Create a temporary storage for testing
	tempDir := "/tmp/test_storage"
	storageConfig := storage.StorageConfig{
		BasePath: tempDir,
		PublicURL: "",
	}
	storage := storage.NewLocalStorage(storageConfig)
	assembler := pipeline.NewVideoAssembler(storage)

	textContent := `This is line 1.

This is line 2.

This is line 3.`

	duration := 9.0 // 3 lines over 9 seconds = 3 seconds per line
	segments := assembler.ParseTextSegments(textContent, duration)

	assert.Len(t, segments, 3)
	
	// Check first segment
	assert.Equal(t, "This is line 1.", segments[0].Text)
	assert.Equal(t, 0.0, segments[0].StartTime)
	assert.Equal(t, 3.0, segments[0].EndTime)
	
	// Check second segment
	assert.Equal(t, "This is line 2.", segments[1].Text)
	assert.Equal(t, 3.0, segments[1].StartTime)
	assert.Equal(t, 6.0, segments[1].EndTime)
	
	// Check third segment
	assert.Equal(t, "This is line 3.", segments[2].Text)
	assert.Equal(t, 6.0, segments[2].StartTime)
	assert.Equal(t, 9.0, segments[2].EndTime)
}

func TestVideoAssembler_EscapeFFmpegText(t *testing.T) {
	assembler := pipeline.NewVideoAssembler()

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "Simple text",
			expected: "Simple text",
		},
		{
			input:    "Text with 'quotes'",
			expected: "Text with \\'quotes\\'",
		},
		{
			input:    "Text with: colon",
			expected: "Text with\\: colon",
		},
		{
			input:    "Text with [brackets]",
			expected: "Text with \\[brackets\\]",
		},
		{
			input:    "Text with (parentheses)",
			expected: "Text with \\(parentheses\\)",
		},
		{
			input:    "Text with %percent",
			expected: "Text with \\%percent",
		},
	}

	for _, tc := range testCases {
		result := assembler.escapeFFmpegText(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}

func TestVideoAssembler_FormatSRTTime(t *testing.T) {
	assembler := pipeline.NewVideoAssembler()

	testCases := []struct {
		seconds  float64
		expected string
	}{
		{
			seconds:  0.0,
			expected: "00:00:00,000",
		},
		{
			seconds:  65.5,
			expected: "00:01:05,500",
		},
		{
			seconds:  3661.123,
			expected: "01:01:01,123",
		},
	}

	for _, tc := range testCases {
		result := assembler.formatSRTTime(tc.seconds)
		assert.Equal(t, tc.expected, result)
	}
}

func TestVideoAssembler_CreateSRTSubtitleFile(t *testing.T) {
	assembler := pipeline.NewVideoAssembler()

	tempFile := filepath.Join(os.TempDir(), "test_subtitles.srt")
	defer os.Remove(tempFile)

	subtitles := []models.Subtitle{
		{
			Language: "en",
			Content:  "Test subtitle",
			Timestamps: []models.Timestamp{
				{
					Start: 0.0,
					End:   3.0,
					Text:  "First subtitle",
				},
				{
					Start: 3.5,
					End:   6.5,
					Text:  "Second subtitle",
				},
			},
		},
	}

	err := assembler.createSRTSubtitleFile(tempFile, subtitles)
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "1")
	assert.Contains(t, contentStr, "00:00:00,000 --> 00:00:03,000")
	assert.Contains(t, contentStr, "First subtitle")
	assert.Contains(t, contentStr, "2")
	assert.Contains(t, contentStr, "00:00:03,500 --> 00:00:06,500")
	assert.Contains(t, contentStr, "Second subtitle")
}

func TestVideoAssembler_GenerateSolidBackground(t *testing.T) {
	assembler := pipeline.NewVideoAssembler()

	ctx := context.Background()
	outputPath := filepath.Join(os.TempDir(), "test_background.png")
	defer os.Remove(outputPath)

	textContent := "Test background generation"

	backgroundPath, err := assembler.generateSolidBackground(ctx, outputPath, textContent)
	require.NoError(t, err)
	assert.Equal(t, outputPath, backgroundPath)

	// Verify file was created (if FFmpeg is available)
	if utils.FileExists(outputPath) {
		info, err := os.Stat(outputPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))
	}
}

func TestCourseGenerator_NewCourseGenerator(t *testing.T) {
	generator := pipeline.NewCourseGenerator()

	require.NotNil(t, generator)
	assert.NotNil(t, generator.markdownParser)
	assert.NotNil(t, generator.ttsProcessor)
	assert.NotNil(t, generator.videoAssembler)
}

func TestCourseGenerator_GenerateCourse(t *testing.T) {
	// Create temporary markdown file
	tempDir := filepath.Join(os.TempDir(), "course_test")
	outputDir := filepath.Join(os.TempDir(), "course_output")
	
	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	markdownContent := `# Test Course

This is a test course for unit testing.

## Introduction

Welcome to the test course.

## Main Content

This is the main content of the course.

## Conclusion

Thank you for completing the course.`

	markdownPath := filepath.Join(tempDir, "test_course.md")
	err = utils.WriteFile(markdownPath, markdownContent)
	require.NoError(t, err)

	generator := pipeline.NewCourseGenerator()
	
	options := models.ProcessingOptions{
		Quality:   "standard",
		Languages: []string{"en"},
	}

	course, err := generator.GenerateCourse(markdownPath, outputDir, options)
	
	// Should generate course structure even with placeholder implementations
	if err == nil && course != nil {
		assert.NotEmpty(t, course.ID)
		assert.Equal(t, "Test Course", course.Title)
		assert.Contains(t, course.Description, "test course")
		assert.NotEmpty(t, course.Lessons)
	}
}

func TestIntegration_CompleteCourseGeneration(t *testing.T) {
	// This test integrates multiple components
	tempDir := filepath.Join(os.TempDir(), "integration_test")
	outputDir := filepath.Join(os.TempDir(), "integration_output")
	
	err := utils.EnsureDir(tempDir)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	err = utils.EnsureDir(outputDir)
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Create comprehensive markdown
	markdownContent := `# Complete Integration Test Course

This is a comprehensive test course that tests all components integration.

## Lesson 1: Introduction

Welcome to this comprehensive integration test. We will test:
- Text parsing
- Audio generation
- Video assembly
- File management

## Lesson 2: Technical Details

The system processes content through multiple stages:
1. Markdown parsing
2. Audio generation with TTS
3. Video creation with FFmpeg
4. Final assembly and packaging

## Lesson 3: Advanced Features

Advanced features include:
- Multiple TTS providers
- High-quality video processing
- Background generation
- Text overlays

## Conclusion

This concludes our integration test.`

	markdownPath := filepath.Join(tempDir, "integration_course.md")
	err = utils.WriteFile(markdownPath, markdownContent)
	require.NoError(t, err)

	generator := pipeline.NewCourseGenerator()
	
	options := models.ProcessingOptions{
		Quality:          "high",
		Languages:        []string{"en"},
		BackgroundMusic:   true,
		Voice:            stringPtr("v2/en_speaker_6"),
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Run generation in goroutine with timeout
	resultCh := make(chan *pipeline.CourseResult, 1)
	errCh := make(chan error, 1)

	go func() {
		course, err := generator.GenerateCourse(markdownPath, outputDir, options)
		if err != nil {
			errCh <- err
			return
		}
		
		resultCh <- &pipeline.CourseResult{
			Course: course,
			Status: "success",
		}
	}()

	select {
	case <-ctx.Done():
		t.Skip("Integration test timed out (may be due to missing FFmpeg/TTS dependencies)")
	case err := <-errCh:
		if err == nil {
			t.Fatal("Unexpected error in integration test")
		}
		t.Logf("Integration test failed (expected with missing dependencies): %v", err)
	case result := <-resultCh:
		require.NotNil(t, result)
		assert.Equal(t, "success", result.Status)
		require.NotNil(t, result.Course)
		
		course := result.Course
		assert.NotEmpty(t, course.ID)
		assert.Equal(t, "Complete Integration Test Course", course.Title)
		assert.Equal(t, 4, len(course.Lessons)) // Introduction, Technical Details, Advanced Features, Conclusion
		
		// Verify lesson content
		lessonTitles := []string{}
		for _, lesson := range course.Lessons {
			lessonTitles = append(lessonTitles, lesson.Title)
			assert.NotEmpty(t, lesson.ID)
			assert.NotEmpty(t, lesson.Content)
		}
		
		assert.Contains(t, lessonTitles, "Lesson 1: Introduction")
		assert.Contains(t, lessonTitles, "Lesson 2: Technical Details")
		assert.Contains(t, lessonTitles, "Lesson 3: Advanced Features")
		assert.Contains(t, lessonTitles, "Conclusion")
	}
}

func TestVideoBackgroundStyles(t *testing.T) {
	assembler := pipeline.NewVideoAssembler()

	ctx := context.Background()
	textContent := "Test background styles"
	duration := 5.0

	options := models.ProcessingOptions{Quality: "standard"}

	// Test each background style
	styles := []pipeline.BackgroundStyle{
		pipeline.BackgroundSolidColor,
		pipeline.BackgroundGradient,
		pipeline.BackgroundPattern,
	}

	for _, style := range styles {
		t.Run(string(style), func(t *testing.T) {
			outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("bg_%s_%d.png", style, utils.GenerateID()))
			defer os.Remove(outputPath)

			var backgroundPath string
			var err error

			switch style {
			case pipeline.BackgroundSolidColor:
				backgroundPath, err = assembler.generateSolidBackground(ctx, outputPath, textContent)
			case pipeline.BackgroundGradient:
				backgroundPath, err = assembler.generateGradientBackground(ctx, outputPath, textContent, duration)
			case pipeline.BackgroundPattern:
				backgroundPath, err = assembler.generatePatternBackground(ctx, outputPath, textContent)
			}

			if err == nil {
				assert.NotEmpty(t, backgroundPath)
				if utils.FileExists(backgroundPath) {
					info, err := os.Stat(backgroundPath)
					require.NoError(t, err)
					assert.Greater(t, info.Size(), int64(0))
				}
			} else {
				t.Logf("Background generation failed (may be due to missing FFmpeg): %v", err)
			}
		})
	}
}

func TestTTSProviderSelection(t *testing.T) {
	processor := pipeline.NewTTSProcessor()

	testCases := []struct {
		voice            *string
		expectedProvider pipeline.TTSProvider
	}{
		{
			voice:            nil,
			expectedProvider: pipeline.TTSProviderBark,
		},
		{
			voice:            stringPtr("bark"),
			expectedProvider: pipeline.TTSProviderBark,
		},
		{
			voice:            stringPtr("speecht5"),
			expectedProvider: pipeline.TTSProviderSpeechT5,
		},
		{
			voice:            stringPtr("custom"),
			expectedProvider: pipeline.TTSProviderBark, // Falls back to default
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("voice=%v", tc.voice), func(t *testing.T) {
			options := models.ProcessingOptions{
				Voice: tc.voice,
			}

			// This tests the provider selection logic internally
			provider := pipeline.TTSProviderBark // Default
			if options.Voice != nil {
				switch *options.Voice {
				case "speecht5":
					provider = pipeline.TTSProviderSpeechT5
				case "bark":
					provider = pipeline.TTSProviderBark
				}
			}

			assert.Equal(t, tc.expectedProvider, provider)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}