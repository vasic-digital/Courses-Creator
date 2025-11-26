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
	config VideoConfig
}

// NewVideoAssembler creates a new video assembler
func NewVideoAssembler() *VideoAssembler {
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
		config: config,
	}
}

// NewVideoAssemblerWithConfig creates a new video assembler with custom config
func NewVideoAssemblerWithConfig(config VideoConfig) *VideoAssembler {
	utils.EnsureDir(config.OutputDir)
	utils.EnsureDir(config.TempDir)
	
	return &VideoAssembler{
		config: config,
	}
}

// CreateVideo creates video from audio and text content
func (va *VideoAssembler) CreateVideo(
	audioPath string,
	textContent string,
	outputDir string,
	options models.ProcessingOptions,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), va.config.Timeout)
	defer cancel()

	fmt.Printf("Creating video from audio: %s\n", audioPath)

	// Generate unique output path
	videoName := fmt.Sprintf("video_%d.mp4", utils.HashString(textContent))
	outputPath := filepath.Join(outputDir, videoName)

	// Ensure output directory exists
	if err := utils.EnsureDir(outputDir); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

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
	if err := va.assembleVideo(ctx, audioPath, backgroundPath, segments, outputPath, options); err != nil {
		return "", fmt.Errorf("failed to assemble video: %w", err)
	}

	// Clean up temporary files
	va.cleanup()

	fmt.Printf("Created video: %s\n", outputPath)
	return outputPath, nil
}

// generateBackground creates a dynamic background
func (va *VideoAssembler) generateBackground(ctx context.Context, textContent string, duration float64, options models.ProcessingOptions) (string, error) {
	backgroundPath := filepath.Join(va.config.TempDir, fmt.Sprintf("bg_%d.png", utils.HashString(textContent)))

	// Choose background style based on quality and preferences
	var style BackgroundStyle
	switch options.Quality {
	case "high":
		style = BackgroundAnimated
	default:
		style = BackgroundGradient
	}

	switch style {
	case BackgroundSolidColor:
		return va.generateSolidBackground(ctx, backgroundPath, textContent)
	case BackgroundGradient:
		return va.generateGradientBackground(ctx, backgroundPath, textContent, duration)
	case BackgroundPattern:
		return va.generatePatternBackground(ctx, backgroundPath, textContent)
	case BackgroundAnimated:
		return va.generateAnimatedBackground(ctx, backgroundPath, textContent, duration)
	default:
		return va.generateSolidBackground(ctx, backgroundPath, textContent)
	}
}

// generateSolidBackground creates a solid color background
func (va *VideoAssembler) generateSolidBackground(ctx context.Context, outputPath, textContent string) (string, error) {
	colors := []string{
		"4A90E2", // Blue
		"50C878", // Green
		"9B59B6", // Purple
		"F39C12", // Orange
		"E74C3C", // Red
		"1ABC9C", // Turquoise
		"34495E", // Dark Blue
		"E67E22", // Carrot
	}

	// Choose color based on text hash
	colorIndex := int(utils.HashString(textContent)) % len(colors)
	color := colors[colorIndex]

	cmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath,
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=%s:s=%dx%d:d=1", color, va.config.Quality.Width, va.config.Quality.Height),
		"-frames:v", "1",
		"-q:v", "1",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate solid background: %w", err)
	}

	return outputPath, nil
}

// generateGradientBackground creates a gradient background
func (va *VideoAssembler) generateGradientBackground(ctx context.Context, outputPath, textContent string, duration float64) (string, error) {
	// Create gradient filter
	filter := fmt.Sprintf("color=red:s=%dx%d[d1];color=blue:s=%dx%d[d2];[d1][d2]scale2ref[d2][d1];[d2][d1]blend=all_mode=multiply",
		va.config.Quality.Width, va.config.Quality.Height,
		va.config.Quality.Width, va.config.Quality.Height,
	)

	cmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath,
		"-f", "lavfi",
		"-i", filter,
		"-frames:v", "1",
		"-q:v", "1",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		// Fallback to solid background
		return va.generateSolidBackground(ctx, outputPath, textContent)
	}

	return outputPath, nil
}

// generatePatternBackground creates a pattern background
func (va *VideoAssembler) generatePatternBackground(ctx context.Context, outputPath, textContent string) (string, error) {
	// Create a pattern using FFmpeg's geq filter
	filter := fmt.Sprintf("geq=lum='p(X,Y)':cb='p(X,Y)':cr='p(X,Y)':s=%dx%d",
		va.config.Quality.Width, va.config.Quality.Height)

	cmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath,
		"-f", "lavfi",
		"-i", filter,
		"-frames:v", "1",
		"-q:v", "1",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		// Fallback to solid background
		return va.generateSolidBackground(ctx, outputPath, textContent)
	}

	return outputPath, nil
}

// generateAnimatedBackground creates an animated background
func (va *VideoAssembler) generateAnimatedBackground(ctx context.Context, outputPath, textContent string, duration float64) (string, error) {
	// Create animated gradient
	animOutput := filepath.Join(va.config.TempDir, fmt.Sprintf("anim_bg_%d.mp4", utils.HashString(textContent)))
	defer os.Remove(animOutput)

	// Create animated gradient filter
	filter := fmt.Sprintf("color=red:s=%dx%d:d=%.1f[c1];color=blue:s=%dx%d:d=%.1f[c2];[c1][c2]blend=all_mode=multiply",
		va.config.Quality.Width, va.config.Quality.Height, duration,
		va.config.Quality.Width, va.config.Quality.Height, duration,
	)

	cmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath,
		"-f", "lavfi",
		"-i", filter,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		"-t", fmt.Sprintf("%.1f", duration),
		animOutput,
	)

	if err := cmd.Run(); err != nil {
		// Fallback to static background
		return va.generateGradientBackground(ctx, outputPath, textContent, duration)
	}

	// Extract first frame as background
	frameCmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath,
		"-i", animOutput,
		"-ss", "0.1",
		"-frames:v", "1",
		"-q:v", "1",
		outputPath,
	)

	if err := frameCmd.Run(); err != nil {
		return va.generateGradientBackground(ctx, outputPath, textContent, duration)
	}

	return outputPath, nil
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
		"-c:v", va.config.Quality.Codec,
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-pix_fmt", va.config.Quality.PixelFormat,
		"-r", strconv.Itoa(va.config.Quality.Framerate),
		"-t", fmt.Sprintf("%.2f", segments[len(segments)-1].EndTime),
		"-shortest",
		outputPath,
	}

	cmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath, args...)

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
			va.config.FontPath,
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
		va.config.FFprobePath,
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
	subtitlePath := filepath.Join(va.config.TempDir, fmt.Sprintf("subs_%d.srt", utils.GenerateID()))
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

	cmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath, args...)
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

	cmd := utils.ExecuteCommand(ctx, va.config.FFmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to add background music: %w", err)
	}

	return outputPath, nil
}

// cleanup removes temporary files
func (va *VideoAssembler) cleanup() {
	utils.CleanTempFiles(va.config.TempDir, time.Hour)
}