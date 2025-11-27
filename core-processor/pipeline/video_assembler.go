package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/course-creator/core-processor/models"
	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/utils"
)

// VideoQuality represents video quality settings
type VideoQuality struct {
	Width        int
	Height       int
	Bitrate      string
	Framerate    int
	Codec        string
	PixelFormat  string
}

// VideoConfig holds video assembly configuration
type VideoConfig struct {
	Quality      VideoQuality
	OutputDir    string
	FFmpegPath   string
	FFprobePath  string
	TempDir      string
	Timeout      time.Duration
	MaxRetries   int
	FontPath     string
	SubtitleFont string
}

// BackgroundStyle defines background generation styles
type BackgroundStyle string

const (
	BackgroundSolidColor  BackgroundStyle = "solid"
	BackgroundGradient    BackgroundStyle = "gradient"
	BackgroundPattern    BackgroundStyle = "pattern"
	BackgroundAnimated   BackgroundStyle = "animated"
)

// VideoAssembler handles video assembly from audio and visual elements
type VideoAssembler struct {
	Config           VideoConfig
	storage          storage.StorageInterface
	backgroundGen    *BackgroundGenerator
}

// NewVideoAssembler creates a new video assembler
func NewVideoAssembler(storage storage.StorageInterface) *VideoAssembler {
	config := VideoConfig{
		Quality: VideoQuality{
			Width:       1920,
			Height:      1080,
			Bitrate:     "2M",
			Framerate:   30,
			Codec:       "libx264",
			PixelFormat: "yuv420p",
		},
		OutputDir:    "/tmp/course_videos",
		FFmpegPath:   "ffmpeg",
		FFprobePath:  "ffprobe",
		TempDir:      "/tmp/video_temp",
		Timeout:      300 * time.Second,
		MaxRetries:   3,
		FontPath:     "/System/Library/Fonts/Helvetica.ttc", // macOS default
		SubtitleFont: "Arial",
	}

	// Ensure directories exist
	utils.EnsureDir(config.OutputDir)
	utils.EnsureDir(config.TempDir)

	return &VideoAssembler{
		Config:        config,
		storage:       storage,
		backgroundGen: NewBackgroundGenerator(storage),
	}
}

// NewVideoAssemblerWithConfig creates a new video assembler with custom config
func NewVideoAssemblerWithConfig(config VideoConfig, storage storage.StorageInterface) *VideoAssembler {
	utils.EnsureDir(config.OutputDir)
	utils.EnsureDir(config.TempDir)
	
	return &VideoAssembler{
		Config:        config,
		storage:       storage,
		backgroundGen: NewBackgroundGenerator(storage),
	}
}

// CreateVideo creates video from audio and text content
func (va *VideoAssembler) CreateVideo(
	audioPath string,
	textContent string,
	courseID string,
	lessonID string,
	options models.ProcessingOptions,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), va.Config.Timeout)
	defer cancel()

	fmt.Printf("Creating video from audio: %s\n", audioPath)

	// Generate unique output path using storage
	videoName := fmt.Sprintf("video_%d.mp4", utils.HashString(textContent))
	storagePath := storage.GetVideoStoragePath(courseID, lessonID, videoName)

	// Get audio duration
	duration, err := va.getAudioDuration(ctx, audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to get audio duration: %w", err)
	}

	fmt.Printf("Audio duration: %.2f seconds\n", duration)

	// Generate background
	backgroundPath, err := va.generateBackground(ctx, textContent, duration, options)
	if err != nil {
		return "", fmt.Errorf("failed to generate background: %w", err)
	}
	defer os.Remove(backgroundPath)

	// Parse text content into segments
	segments := va.parseTextSegments(textContent, duration)

	// Create video with text overlays
	tempOutputPath := filepath.Join(va.Config.TempDir, fmt.Sprintf("temp_video_%d.mp4", utils.HashString(textContent)))
	if err := va.assembleVideo(ctx, audioPath, backgroundPath, segments, tempOutputPath, options); err != nil {
		return "", fmt.Errorf("failed to assemble video: %w", err)
	}
	
	// Read the video file and save to storage
	videoData, err := os.ReadFile(tempOutputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read video file: %w", err)
	}
	
	// Save to storage
	err = va.storage.Save(storagePath, videoData)
	if err != nil {
		return "", fmt.Errorf("failed to save video to storage: %w", err)
	}

	// Clean up temporary files
	va.cleanup()
	os.Remove(tempOutputPath) // Remove temp video file

	fmt.Printf("Created video: %s\n", storagePath)
	return storagePath, nil
}

// generateBackground creates a dynamic background using BackgroundGenerator
func (va *VideoAssembler) generateBackground(ctx context.Context, textContent string, duration float64, options models.ProcessingOptions) (string, error) {
	// Use the BackgroundGenerator to create background
	return va.backgroundGen.GenerateBackground(ctx, textContent, options)
}

// ParseTextSegments parses text content into timed segments (exported for testing)
func (va *VideoAssembler) ParseTextSegments(textContent string, duration float64) []TextSegment {
	return va.parseTextSegments(textContent, duration)
}

// parseTextSegments parses text content into timed segments
func (va *VideoAssembler) parseTextSegments(textContent string, duration float64) []TextSegment {
	lines := strings.Split(textContent, "\n")
	var segments []TextSegment

	// Distribute time evenly across non-empty lines
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	timePerLine := duration / float64(nonEmptyLines)
	currentTime := 0.0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			segments = append(segments, TextSegment{
				Text:      line,
				StartTime: currentTime,
				EndTime:   currentTime + timePerLine,
			})
			currentTime += timePerLine
		}
	}

	return segments
}

// TextSegment represents a text segment with timing
type TextSegment struct {
	Text      string
	StartTime float64
	EndTime   float64
}

// assembleVideo uses FFmpeg to combine audio, background, and text
func (va *VideoAssembler) assembleVideo(ctx context.Context, audioPath, backgroundPath string, segments []TextSegment, outputPath string, options models.ProcessingOptions) error {
	// Create text overlay filter
	textFilter := va.createTextOverlayFilter(segments, options)

	// Base FFmpeg command
	args := []string{
		"-loop", "1",
		"-i", backgroundPath,
		"-i", audioPath,
		"-filter_complex", textFilter,
		"-c:v", va.Config.Quality.Codec,
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-pix_fmt", va.Config.Quality.PixelFormat,
		"-r", strconv.Itoa(va.Config.Quality.Framerate),
		"-t", fmt.Sprintf("%.2f", segments[len(segments)-1].EndTime),
		"-shortest",
		outputPath,
	}

	cmd := utils.ExecuteCommand(ctx, va.Config.FFmpegPath, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assemble video: %w", err)
	}

	return nil
}

// createTextOverlayFilter creates FFmpeg filter for text overlays
func (va *VideoAssembler) createTextOverlayFilter(segments []TextSegment, options models.ProcessingOptions) string {
	var filterParts []string

	// Create background input
	filterParts = append(filterParts, "[0:v]base")

	for _, segment := range segments {
		// Escape special characters in text
		escapedText := va.escapeFFmpegText(segment.Text)
		
		// Calculate text properties based on quality
		fontSize := 48
		if options.Quality == "high" {
			fontSize = 56
		}

		// Create text overlay for this segment
		textFilter := fmt.Sprintf(
			"drawtext=text='%s':x=(w-tw)/2:y=h-150:fontsize=%d:fontcolor=white:fontfile='%s':enable='between(t,%.2f,%.2f)'",
			escapedText,
			fontSize,
			va.Config.FontPath,
			segment.StartTime,
			segment.EndTime,
		)

		filterParts = append(filterParts, fmt.Sprintf("base%s", textFilter))
	}

	// Combine all filters
	return strings.Join(filterParts, ",")
}

// escapeFFmpegText escapes special characters for FFmpeg drawtext
func (va *VideoAssembler) escapeFFmpegText(text string) string {
	// Escape special characters for FFmpeg
	replacements := map[string]string{
		"'":  "\\'",
		":":  "\\:",
		"[":  "\\[",
		"]":  "\\]",
		",":  "\\,",
		";":  "\\;",
		"(":  "\\(",
		")":  "\\)",
		"%":  "\\%",
	}

	escaped := text
	for old, new := range replacements {
		escaped = strings.ReplaceAll(escaped, old, new)
	}

	return escaped
}

// getAudioDuration gets duration of an audio file using ffprobe
func (va *VideoAssembler) getAudioDuration(ctx context.Context, audioPath string) (float64, error) {
	output, err := utils.ExecuteCommandWithOutput(
		ctx,
		va.Config.FFprobePath,
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		audioPath,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	var duration float64
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%f", &duration)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// AddSubtitles adds subtitles to video
func (va *VideoAssembler) AddSubtitles(ctx context.Context, videoPath string, subtitles []models.Subtitle) (string, error) {
	fmt.Printf("Adding subtitles to video: %s\n", videoPath)

	// Create subtitle file
	subtitlePath := filepath.Join(va.Config.TempDir, fmt.Sprintf("subs_%s.srt", utils.GenerateID()))
	defer os.Remove(subtitlePath)

	if err := va.createSRTSubtitleFile(subtitlePath, subtitles); err != nil {
		return "", fmt.Errorf("failed to create subtitle file: %w", err)
	}

	// Output path with subtitles
	outputPath := videoPath[:len(videoPath)-4] + "_subtitled.mp4"

	// Add subtitles using FFmpeg
	args := []string{
		"-i", videoPath,
		"-i", subtitlePath,
		"-c", "copy",
		"-c:s", "mov_text",
		"-metadata:s:s:0", "language=eng",
		outputPath,
	}

	cmd := utils.ExecuteCommand(ctx, va.Config.FFmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to add subtitles: %w", err)
	}

	return outputPath, nil
}

// createSRTSubtitleFile creates an SRT subtitle file
func (va *VideoAssembler) createSRTSubtitleFile(path string, subtitles []models.Subtitle) error {
	var content strings.Builder
	
	for _, subtitle := range subtitles {
		for i, timestampMap := range subtitle.Timestamps {
			startTime := va.formatSRTTime(timestampMap["start"].(float64))
			endTime := va.formatSRTTime(timestampMap["end"].(float64))
			
			content.WriteString(fmt.Sprintf("%d\n", i+1))
			content.WriteString(fmt.Sprintf("%s --> %s\n", startTime, endTime))
			content.WriteString(fmt.Sprintf("%s\n\n", timestampMap["text"].(string)))
		}
	}

	return utils.WriteFile(path, content.String())
}

// formatSRTTime formats time for SRT subtitle format
func (va *VideoAssembler) formatSRTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}

// AddBackgroundMusic mixes background music with video audio
func (va *VideoAssembler) AddBackgroundMusic(ctx context.Context, videoPath, musicPath string, volume float64) (string, error) {
	fmt.Printf("Adding background music to video: %s\n", videoPath)

	outputPath := videoPath[:len(videoPath)-4] + "_with_music.mp4"

	// Mix audio using FFmpeg
	args := []string{
		"-i", videoPath,
		"-i", musicPath,
		"-filter_complex", fmt.Sprintf("[0:a][1:a]amix=inputs=2:weights=1 %.2f[aout]", volume),
		"-map", "0:v",
		"-map", "[aout]",
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "128k",
		outputPath,
	}

	cmd := utils.ExecuteCommand(ctx, va.Config.FFmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to add background music: %w", err)
	}

	return outputPath, nil
}

// cleanup removes temporary files
func (va *VideoAssembler) cleanup() {
	utils.CleanTempFiles(va.Config.TempDir, time.Hour)
}