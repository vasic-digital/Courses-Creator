package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/course-creator/core-processor/mcp_servers"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/utils"
)

// TTSProvider represents a TTS provider
type TTSProvider string

const (
	TTSProviderBark    TTSProvider = "bark"
	TTSProviderSpeechT5 TTSProvider = "speecht5"
	TTSProviderAzure   TTSProvider = "azure"
	TTSProviderGoogle  TTSProvider = "google"
)

// TTSConfig holds TTS configuration
type TTSConfig struct {
	DefaultProvider    TTSProvider
	OutputDir         string
	SampleRate        int
	BitRate           int
	Format            string // "wav", "mp3", "ogg"
	Timeout           time.Duration
	MaxRetries        int
	ChunkSize         int // Maximum text length per chunk
	Parallelism       int // Number of parallel TTS generations
}

// TTSProcessor handles text-to-speech generation
type TTSProcessor struct {
	Config        TTSConfig
	BarkServer    *mcp_servers.BarkTTSServer
	mu            sync.RWMutex
	Running       bool
}

// AudioSegment represents a processed audio segment
type AudioSegment struct {
	Path      string
	Text      string
	Duration  float64
	StartTime float64
	EndTime   float64
}

// NewTTSProcessor creates a new TTS processor
func NewTTSProcessor() *TTSProcessor {
	config := TTSConfig{
		DefaultProvider: TTSProviderBark,
		OutputDir:      "/tmp/course_audio",
		SampleRate:     24000,
		BitRate:        128000,
		Format:         "wav",
		Timeout:        60 * time.Second,
		MaxRetries:     3,
		ChunkSize:      200,
		Parallelism:    2,
	}

	// Ensure output directory exists
	utils.EnsureDir(config.OutputDir)

	return &TTSProcessor{
		Config:         config,
		BarkServer:     mcp_servers.NewBarkTTSServer(),
		Running:        true,
	}
}

// NewTTSProcessorWithConfig creates a new TTS processor with custom config
func NewTTSProcessorWithConfig(config TTSConfig) *TTSProcessor {
	// Ensure output directory exists
	utils.EnsureDir(config.OutputDir)

	return &TTSProcessor{
		Config:         config,
		BarkServer:     mcp_servers.NewBarkTTSServer(),
		Running:        true,
	}
}

// GenerateAudio generates audio from text using configured TTS
func (tp *TTSProcessor) GenerateAudio(text string, options models.ProcessingOptions) (string, error) {
	tp.mu.RLock()
	if !tp.Running {
		tp.mu.RUnlock()
		return "", fmt.Errorf("TTS processor is not running")
	}
	tp.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), tp.Config.Timeout)
	defer cancel()

	fmt.Printf("Generating audio for text length: %d\n", len(text))

	// Choose TTS provider based on options or default
	provider := tp.Config.DefaultProvider
	if options.Voice != nil {
		switch *options.Voice {
		case "speecht5":
			provider = TTSProviderSpeechT5
		case "bark":
			provider = TTSProviderBark
		// Add other providers as needed
		}
	}

	// Split text into chunks if too long
	chunks := tp.splitText(text)
	fmt.Printf("Text split into %d chunks\n", len(chunks))

	// Generate audio segments
	segments, err := tp.generateAudioSegments(ctx, chunks, provider, options)
	if err != nil {
		return "", fmt.Errorf("failed to generate audio segments: %w", err)
	}

	// Combine audio segments
	finalPath, err := tp.combineAudioSegments(segments, text)
	if err != nil {
		return "", fmt.Errorf("failed to combine audio segments: %w", err)
	}

	fmt.Printf("Audio generation completed: %s\n", finalPath)
	return finalPath, nil
}

// generateAudioSegments generates audio for multiple text chunks
func (tp *TTSProcessor) generateAudioSegments(ctx context.Context, chunks []string, provider TTSProvider, options models.ProcessingOptions) ([]AudioSegment, error) {
	segments := make([]AudioSegment, len(chunks))
	sem := make(chan struct{}, tp.Config.Parallelism)
	errCh := make(chan error, len(chunks))
	wg := sync.WaitGroup{}

	for i, chunk := range chunks {
		wg.Add(1)
		go func(index int, text string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				segment, err := tp.generateSingleSegment(ctx, text, provider, options, index)
				if err != nil {
					errCh <- fmt.Errorf("segment %d failed: %w", index, err)
					return
				}
				segments[index] = *segment
			}
		}(i, chunk)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Check for errors
	if err := <-errCh; err != nil {
		return nil, err
	}

	return segments, nil
}

// generateSingleSegment generates audio for a single text segment
func (tp *TTSProcessor) generateSingleSegment(ctx context.Context, text string, provider TTSProvider, options models.ProcessingOptions, index int) (*AudioSegment, error) {
	var audioPath string
	var err error

	// Retry logic
	for attempt := 0; attempt < tp.Config.MaxRetries; attempt++ {
		switch provider {
		case TTSProviderBark:
			audioPath, err = tp.generateBarkTTS(ctx, text, options, index)
		case TTSProviderSpeechT5:
			audioPath, err = tp.generateSpeechT5TTS(ctx, text, options, index)
		default:
			return nil, fmt.Errorf("unsupported TTS provider: %s", provider)
		}

		if err == nil {
			break
		}

		if attempt < tp.Config.MaxRetries-1 {
			fmt.Printf("TTS attempt %d failed, retrying: %v\n", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("all TTS attempts failed: %w", err)
	}

	// Get audio duration
	duration, err := tp.getAudioDuration(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	return &AudioSegment{
		Path:      audioPath,
		Text:      text,
		Duration:  duration,
		StartTime: 0, // Will be calculated when combining
		EndTime:   duration,
	}, nil
}

// generateBarkTTS generates TTS using Bark
func (tp *TTSProcessor) generateBarkTTS(ctx context.Context, text string, options models.ProcessingOptions, index int) (string, error) {
	preview := text
	if len(text) > 50 {
		preview = text[:50] + "..."
	}
	fmt.Printf("Generating Bark TTS segment %d: %s\n", index, preview)

	// Prepare TTS arguments
	args := map[string]interface{}{
		"text": text,
	}

	if options.Voice != nil && *options.Voice != "" {
		args["voice_preset"] = *options.Voice
	}

	// Add quality settings if available
	if options.Quality == "high" {
		args["temperature"] = 0.5
		args["generation"] = 2
	} else {
		args["temperature"] = 0.7
		args["generation"] = 1
	}

	// Call Bark server
	result, err := tp.BarkServer.GenerateTTS(args)
	if err != nil {
		return "", fmt.Errorf("Bark TTS failed: %w", err)
	}

	// Extract audio path from result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid result from Bark TTS")
	}

	audioPath, ok := resultMap["audio_path"].(string)
	if !ok {
		return "", fmt.Errorf("audio_path not found in result")
	}

	return audioPath, nil
}

// generateSpeechT5TTS generates TTS using SpeechT5
func (tp *TTSProcessor) generateSpeechT5TTS(ctx context.Context, text string, options models.ProcessingOptions, index int) (string, error) {
	preview := text
	if len(text) > 50 {
		preview = text[:50] + "..."
	}
	fmt.Printf("Generating SpeechT5 TTS segment %d: %s\n", index, preview)

	// Prepare TTS arguments
	args := map[string]interface{}{
		"text": text,
	}

	if options.Voice != nil && *options.Voice != "" {
		args["voice_preset"] = *options.Voice
	}

	// Add quality settings
	if options.Quality == "high" {
		args["sample_rate"] = 22050
		args["bitrate"] = 192000
	} else {
		args["sample_rate"] = 16000
		args["bitrate"] = 128000
	}

	// Call Bark server (using Bark as fallback since SpeechT5 server was removed)
	result, err := tp.BarkServer.GenerateTTS(args)
	if err != nil {
		return "", fmt.Errorf("Bark TTS failed: %w", err)
	}

	// Extract audio path from result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid result from SpeechT5 TTS")
	}

	audioPath, ok := resultMap["audio_path"].(string)
	if !ok {
		return "", fmt.Errorf("audio_path not found in result")
	}

	return audioPath, nil
}

// splitText splits text into manageable chunks
func (tp *TTSProcessor) splitText(text string) []string {
	if len(text) <= tp.Config.ChunkSize {
		return []string{text}
	}

	// Split by sentences first, then by words if needed
	sentences := tp.splitBySentences(text)
	var chunks []string
	var currentChunk []string
	currentLength := 0

	for _, sentence := range sentences {
		sentenceLength := len(sentence)
		if currentLength+sentenceLength <= tp.Config.ChunkSize {
			currentChunk = append(currentChunk, sentence)
			currentLength += sentenceLength
		} else {
			if len(currentChunk) > 0 {
				chunks = append(chunks, joinSentences(currentChunk))
				currentChunk = []string{sentence}
				currentLength = sentenceLength
			} else {
				// Single sentence too long, split by words
				words := tp.splitByWords(sentence)
				var currentWords []string
				wordLength := 0

				for _, word := range words {
					if wordLength+len(word)+1 > tp.Config.ChunkSize {
						if len(currentWords) > 0 {
							chunks = append(chunks, joinWords(currentWords))
							currentWords = []string{word}
							wordLength = len(word)
						} else {
							// Single word too long, split it
							for i := 0; i < len(word); i += tp.Config.ChunkSize {
								end := i + tp.Config.ChunkSize
								if end > len(word) {
									end = len(word)
								}
								chunks = append(chunks, word[i:end])
							}
							currentWords = []string{}
							wordLength = 0
						}
					} else {
						currentWords = append(currentWords, word)
						wordLength += len(word) + 1 // +1 for space
					}
				}

				if len(currentWords) > 0 {
					currentChunk = append(currentChunk, joinWords(currentWords))
					currentLength = len(joinWords(currentWords))
				}
			}
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, joinSentences(currentChunk))
	}

	return chunks
}

// splitBySentences splits text into sentences
func (tp *TTSProcessor) splitBySentences(text string) []string {
	// Simple sentence splitting - in production, use more sophisticated NLP
	var sentences []string
	start := 0
	for i, char := range text {
		if char == '.' || char == '!' || char == '?' {
			sentences = append(sentences, text[start:i+1])
			start = i + 1
		}
	}
	if start < len(text) {
		sentences = append(sentences, text[start:])
	}
	return sentences
}

// splitByWords splits text into words
func (tp *TTSProcessor) splitByWords(text string) []string {
	// Simple word splitting
	var words []string
	start := 0
	for i, char := range text {
		if char == ' ' || char == '\t' || char == '\n' {
			if start < i {
				words = append(words, text[start:i])
			}
			start = i + 1
		}
	}
	if start < len(text) {
		words = append(words, text[start:])
	}
	return words
}

// joinSentences joins sentences with proper spacing
func joinSentences(sentences []string) string {
	result := ""
	for i, sentence := range sentences {
		if i > 0 && !strings.HasPrefix(sentence, " ") {
			result += " "
		}
		result += strings.TrimSpace(sentence)
	}
	return result
}

// joinWords joins words with spaces
func joinWords(words []string) string {
	return strings.Join(words, " ")
}

// getAudioDuration gets the duration of an audio file
func (tp *TTSProcessor) getAudioDuration(path string) (float64, error) {
	// Use FFmpeg to get duration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := utils.ExecuteCommandWithOutput(
		ctx,
		"ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		path,
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

// combineAudioSegments combines multiple audio segments into one
func (tp *TTSProcessor) combineAudioSegments(segments []AudioSegment, originalText string) (string, error) {
	if len(segments) == 0 {
		return "", fmt.Errorf("no audio segments to combine")
	}

	if len(segments) == 1 {
		return segments[0].Path, nil
	}

	// Calculate start times
	var totalTime float64
	for i := range segments {
		segments[i].StartTime = totalTime
		segments[i].EndTime = totalTime + segments[i].Duration
		totalTime += segments[i].Duration
	}

	// Create a list file for FFmpeg concat
	listPath := filepath.Join(tp.Config.OutputDir, fmt.Sprintf("concat_list_%d.txt", utils.HashString(originalText)))
	var listContent string
	for _, segment := range segments {
		listContent += fmt.Sprintf("file '%s'\n", segment.Path)
	}

	if err := utils.WriteFile(listPath, listContent); err != nil {
		return "", fmt.Errorf("failed to write concat list: %w", err)
	}
	defer os.Remove(listPath)

	// Combine audio using FFmpeg
	outputPath := filepath.Join(tp.Config.OutputDir, fmt.Sprintf("combined_%d.%s", utils.HashString(originalText), tp.Config.Format))
	
	ctx, cancel := context.WithTimeout(context.Background(), tp.Config.Timeout)
	defer cancel()

	_, err := utils.ExecuteCommandWithOutput(
		ctx,
		"ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy",
		outputPath,
	)
	if err != nil {
		return "", fmt.Errorf("failed to combine audio segments: %w", err)
	}

	// Clean up individual segment files
	for _, segment := range segments {
		os.Remove(segment.Path)
	}

	return outputPath, nil
}

// Stop stops the TTS processor
func (tp *TTSProcessor) Stop() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.Running = false
}

// IsRunning checks if the TTS processor is running
func (tp *TTSProcessor) IsRunning() bool {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.Running
}
