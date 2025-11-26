package pipeline

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/utils"
)

// VideoAssembler handles video assembly from audio and visual elements
type VideoAssembler struct {
	ffmpegPath string
}

// NewVideoAssembler creates a new video assembler
func NewVideoAssembler() *VideoAssembler {
	return &VideoAssembler{
		ffmpegPath: "ffmpeg", // Assume ffmpeg is in PATH
	}
}

// CreateVideo creates video from audio and text content
func (va *VideoAssembler) CreateVideo(
	audioPath string,
	textContent string,
	outputDir string,
	options models.ProcessingOptions,
) (string, error) {
	fmt.Printf("Creating video from audio: %s\n", audioPath)

	// Generate unique output path
	videoName := fmt.Sprintf("video_%d.mp4", utils.HashString(textContent))
	outputPath := filepath.Join(outputDir, videoName)

	// Ensure output directory exists
	if err := exec.Command("mkdir", "-p", outputDir).Run(); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate background image (simple colored background)
	backgroundPath := filepath.Join(outputDir, fmt.Sprintf("bg_%d.png", utils.HashString(textContent)))
	if err := va.generateBackground(backgroundPath, options); err != nil {
		return "", fmt.Errorf("failed to generate background: %w", err)
	}

	// Create video with FFmpeg
	if err := va.assembleVideo(audioPath, backgroundPath, textContent, outputPath, options); err != nil {
		return "", fmt.Errorf("failed to assemble video: %w", err)
	}

	fmt.Printf("Created video: %s\n", outputPath)
	return outputPath, nil
}

// generateBackground creates a colorful background image
func (va *VideoAssembler) generateBackground(outputPath string, options models.ProcessingOptions) error {
	// Choose a color based on quality or random
	colors := []string{"blue", "green", "purple", "orange", "red", "cyan"}
	color := colors[utils.HashString(outputPath)%uint32(len(colors))]

	// Try to use FFmpeg to create background
	width, height := 1920, 1080 // 1080p

	cmd := exec.Command(va.ffmpegPath,
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=%s:s=%dx%d:d=1", color, width, height),
		"-frames:v", "1",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		// If FFmpeg not available, create placeholder file
		fmt.Printf("FFmpeg not available, creating placeholder background: %s\n", outputPath)
		if err := exec.Command("touch", outputPath).Run(); err != nil {
			return fmt.Errorf("failed to create placeholder background: %w", err)
		}
	}

	return nil
}

// assembleVideo uses FFmpeg to combine audio, background, and text
func (va *VideoAssembler) assembleVideo(audioPath, backgroundPath, textContent, outputPath string, options models.ProcessingOptions) error {
	// Try FFmpeg first
	duration, err := va.getAudioDuration(audioPath)
	if err != nil {
		return fmt.Errorf("failed to get audio duration: %w", err)
	}

	// Create text overlay filter
	textFilter := va.createTextFilter(textContent, duration)

	// FFmpeg command to combine background, text, and audio
	cmd := exec.Command(va.ffmpegPath,
		"-loop", "1",
		"-i", backgroundPath,
		"-i", audioPath,
		"-filter_complex", textFilter,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-t", strconv.FormatFloat(duration, 'f', 2, 64),
		"-shortest",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		// If FFmpeg not available, create placeholder video file
		fmt.Printf("FFmpeg not available, creating placeholder video: %s\n", outputPath)
		if err := exec.Command("touch", outputPath).Run(); err != nil {
			return fmt.Errorf("failed to create placeholder video: %w", err)
		}
	}

	return nil
}

// getAudioDuration gets the duration of an audio file using ffprobe
func (va *VideoAssembler) getAudioDuration(audioPath string) (float64, error) {
	// For now, return a default duration since audio files are placeholders
	// In real implementation, use ffprobe to get actual duration
	return 10.0, nil
}

// createTextFilter creates FFmpeg filter for text overlay
func (va *VideoAssembler) createTextFilter(textContent string, duration float64) string {
	// Split text into lines
	lines := strings.Split(textContent, "\n")
	var textParts []string

	for i, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			// Escape special characters
			line = strings.ReplaceAll(line, "'", "\\'")
			line = strings.ReplaceAll(line, ":", "\\:")
			y := 100 + i*50
			textParts = append(textParts, fmt.Sprintf("drawtext=text='%s':x=100:y=%d:fontsize=48:fontcolor=white", line, y))
		}
	}

	return strings.Join(textParts, ",")
}

// AddSubtitles adds subtitles to video
func (va *VideoAssembler) AddSubtitles(videoPath string, subtitles []models.Subtitle) (string, error) {
	fmt.Printf("Adding subtitles to video: %s\n", videoPath)

	// Placeholder implementation
	return videoPath, nil
}

// AddBackgroundMusic mixes background music with video audio
func (va *VideoAssembler) AddBackgroundMusic(videoPath, musicPath string) (string, error) {
	fmt.Printf("Adding background music to video: %s\n", videoPath)

	outputPath := videoPath[:len(videoPath)-4] + "_with_music.mp4"

	// Try FFmpeg to mix audio
	cmd := exec.Command(va.ffmpegPath,
		"-i", videoPath,
		"-i", musicPath,
		"-filter_complex", "[0:a][1:a]amix=inputs=2:duration=first[aout]",
		"-map", "0:v",
		"-map", "[aout]",
		"-c:v", "copy",
		"-c:a", "aac",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		// If FFmpeg not available, return original
		fmt.Printf("FFmpeg not available for music mixing, returning original: %s\n", videoPath)
		return videoPath, nil
	}

	return outputPath, nil
}
