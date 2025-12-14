package pipeline

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/course-creator/core-processor/models"
)

// MockTTSProcessor is a mock TTS processor for testing
type MockTTSProcessor struct {
	Config          TTSConfig
	GeneratedAudios map[string]string // text -> audio path
	ShouldFail      map[string]bool   // text -> should fail
	CallCount       int
}

// NewMockTTSProcessor creates a new mock TTS processor
func NewMockTTSProcessor() *MockTTSProcessor {
	return &MockTTSProcessor{
		Config: TTSConfig{
			DefaultProvider: TTSProviderBark,
			OutputDir:       "/tmp/mock_tts",
			SampleRate:      24000,
			BitRate:         128000,
			Format:          "wav",
			Timeout:         30 * time.Second,
			MaxRetries:      3,
			ChunkSize:       500,
			Parallelism:     2,
		},
		GeneratedAudios: make(map[string]string),
		ShouldFail:      make(map[string]bool),
		CallCount:       0,
	}
}

// GenerateAudio generates mock audio for testing
func (m *MockTTSProcessor) GenerateAudio(text string, options models.ProcessingOptions) (string, error) {
	m.CallCount++

	// Check if this text should fail
	if m.ShouldFail[text] {
		return "", fmt.Errorf("mock TTS failure for text: %s", text)
	}

	// Return cached path if already generated
	if path, ok := m.GeneratedAudios[text]; ok {
		return path, nil
	}

	// Generate mock audio path
	audioPath := filepath.Join(m.Config.OutputDir, fmt.Sprintf("audio_%d.wav", m.CallCount))
	m.GeneratedAudios[text] = audioPath

	return audioPath, nil
}

// GenerateAudioWithContext generates mock audio with context
func (m *MockTTSProcessor) GenerateAudioWithContext(ctx interface{}, text string, options models.ProcessingOptions) (string, error) {
	return m.GenerateAudio(text, options)
}

// SplitText splits text into chunks (mock implementation)
func (m *MockTTSProcessor) SplitText(text string) []string {
	// Simple mock implementation - split by sentences
	var chunks []string
	currentChunk := ""
	sentenceEnd := 0

	for i, char := range text {
		currentChunk += string(char)

		// Check for sentence end
		if char == '.' || char == '!' || char == '?' {
			sentenceEnd = i
			if len(currentChunk) > m.Config.ChunkSize/2 {
				chunks = append(chunks, currentChunk[:sentenceEnd+1])
				currentChunk = currentChunk[sentenceEnd+1:]
			}
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	if len(chunks) == 0 {
		chunks = []string{text}
	}

	return chunks
}

// SplitBySentences splits text by sentences (mock implementation)
func (m *MockTTSProcessor) SplitBySentences(text string) []string {
	// Simple mock - split by period
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

// SplitByWords splits text by word count (mock implementation)
func (m *MockTTSProcessor) SplitByWords(text string, maxWords int) []string {
	// Simple mock - split by spaces
	words := []string{}
	wordStart := 0
	inWord := false

	for i, char := range text {
		if char == ' ' || char == '\n' || char == '\t' {
			if inWord {
				words = append(words, text[wordStart:i])
				inWord = false
			}
		} else if !inWord {
			wordStart = i
			inWord = true
		}
	}

	if inWord {
		words = append(words, text[wordStart:])
	}

	// Group into chunks
	var chunks []string
	currentChunk := ""
	wordCount := 0

	for _, word := range words {
		if wordCount >= maxWords && currentChunk != "" {
			chunks = append(chunks, currentChunk)
			currentChunk = word
			wordCount = 1
		} else {
			if currentChunk != "" {
				currentChunk += " "
			}
			currentChunk += word
			wordCount++
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// IsRunning returns mock running status
func (m *MockTTSProcessor) IsRunning() bool {
	return m.CallCount > 0
}

// SetFailure sets a text to fail
func (m *MockTTSProcessor) SetFailure(text string) {
	m.ShouldFail[text] = true
}

// ClearFailures clears all failure settings
func (m *MockTTSProcessor) ClearFailures() {
	m.ShouldFail = make(map[string]bool)
}

// GetCallCount returns the number of times GenerateAudio was called
func (m *MockTTSProcessor) GetCallCount() int {
	return m.CallCount
}

// Reset resets the mock state
func (m *MockTTSProcessor) Reset() {
	m.GeneratedAudios = make(map[string]string)
	m.ShouldFail = make(map[string]bool)
	m.CallCount = 0
}
