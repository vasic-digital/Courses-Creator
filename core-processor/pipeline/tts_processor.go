package pipeline

import (
	"fmt"
	"path/filepath"

	"github.com/course-creator/core-processor/mcp_servers"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/utils"
)

// TTSProcessor handles text-to-speech generation
type TTSProcessor struct {
	barkServer *mcp_servers.BarkTTSServer
}

// NewTTSProcessor creates a new TTS processor
func NewTTSProcessor() *TTSProcessor {
	return &TTSProcessor{
		barkServer: mcp_servers.NewBarkTTSServer(),
	}
}

// GenerateAudio generates audio from text using configured TTS
func (tp *TTSProcessor) GenerateAudio(text string, options models.ProcessingOptions) (string, error) {
	fmt.Printf("Generating audio for text length: %d\n", len(text))

	// Choose TTS based on options or preferences
	ttsType := "bark" // Default, could be configurable
	if options.Voice != nil && *options.Voice == "speecht5" {
		ttsType = "speecht5"
	}

	var audioPath string
	var err error

	switch ttsType {
	case "bark":
		audioPath, err = tp.generateBarkTTS(text, options)
	case "speecht5":
		audioPath, err = tp.generateSpeechT5TTS(text, options)
	default:
		return "", fmt.Errorf("unknown TTS type: %s", ttsType)
	}

	if err != nil {
		return "", fmt.Errorf("TTS generation failed: %w", err)
	}

	return audioPath, nil
}

// generateBarkTTS generates TTS using Bark
func (tp *TTSProcessor) generateBarkTTS(text string, options models.ProcessingOptions) (string, error) {
	preview := text
	if len(text) > 50 {
		preview = text[:50]
	}
	fmt.Printf("Generating Bark TTS for: %s...\n", preview)

	// Call the MCP server directly
	args := map[string]interface{}{
		"text": text,
	}
	if options.Voice != nil {
		args["voice_preset"] = *options.Voice
	}

	result, err := tp.barkServer.GenerateTTS(args)
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
func (tp *TTSProcessor) generateSpeechT5TTS(text string, options models.ProcessingOptions) (string, error) {
	// Placeholder for MCP call to SpeechT5 server
	preview := text
	if len(text) > 50 {
		preview = text[:50]
	}
	fmt.Printf("Generating SpeechT5 TTS for: %s...\n", preview)

	outputPath := filepath.Join("/tmp", fmt.Sprintf("speecht5_audio_%d.wav", utils.HashString(text)))

	if err := writePlaceholderFile(outputPath, "SpeechT5 audio data"); err != nil {
		return "", err
	}

	return outputPath, nil
}

// Helper function to create placeholder files
func writePlaceholderFile(path, content string) error {
	// In real implementation, use os.WriteFile
	// For now, just simulate
	fmt.Printf("Would write file: %s\n", path)
	return nil
}
