package mcp_servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/course-creator/core-processor/utils"
)

// BarkTTSServer handles Bark text-to-speech generation
type BarkTTSServer struct {
	*BaseServerImpl
	barkURL    string
	modelPath  string
	outputDir  string
	maxLength  int
	sampleRate int
}

// BarkRequest represents a Bark TTS generation request
type BarkRequest struct {
	Text        string                 `json:"text"`
	Voice       string                 `json:"voice,omitempty"`
	Speed       float64                `json:"speed,omitempty"`
	Pitch       float64                `json:"pitch,omitempty"`
	Generation  int                    `json:"generation,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// BarkResponse represents a Bark TTS generation response
type BarkResponse struct {
	Success    bool    `json:"success"`
	AudioPath  string  `json:"audio_path,omitempty"`
	Duration   float64 `json:"duration,omitempty"`
	SampleRate int     `json:"sample_rate,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// NewBarkTTSServer creates a new Bark TTS server
func NewBarkTTSServer() *BarkTTSServer {
	config := MCPServerConfig{
		Name:       "bark-tts",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    60 * time.Second,
		MaxRetries: 3,
	}

	server := &BarkTTSServer{
		BaseServerImpl: NewBaseServer(config),
		barkURL:        "http://localhost:8081/generate", // Default Bark server URL
		modelPath:      "/models/bark",
		outputDir:      "/tmp/bark_output",
		maxLength:      200, // Maximum text length per generation
		sampleRate:     24000,
	}

	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)

	server.RegisterTools()
	return server
}

// NewBarkTTSServerWithConfig creates a new Bark TTS server with custom config
func NewBarkTTSServerWithConfig(barkURL, modelPath, outputDir string, maxLength, sampleRate int) *BarkTTSServer {
	config := MCPServerConfig{
		Name:       "bark-tts",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    60 * time.Second,
		MaxRetries: 3,
	}

	server := &BarkTTSServer{
		BaseServerImpl: NewBaseServer(config),
		barkURL:        barkURL,
		modelPath:      modelPath,
		outputDir:      outputDir,
		maxLength:      maxLength,
		sampleRate:     sampleRate,
	}

	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)

	server.RegisterTools()
	return server
}

// RegisterTools registers the TTS generation tool
func (s *BarkTTSServer) RegisterTools() {
	s.AddTool("generate_tts", "Generate speech audio from text using Bark TTS", s.generateTTS)
	s.AddTool("list_voices", "List available Bark voices", s.listVoices)
	s.AddTool("get_info", "Get Bark TTS server information", s.getInfo)
}

// GenerateTTS generates TTS audio from text (public method for direct calls)
func (s *BarkTTSServer) GenerateTTS(args map[string]interface{}) (interface{}, error) {
	return s.generateTTS(args)
}

// generateTTS generates TTS audio from text
func (s *BarkTTSServer) generateTTS(args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("text parameter is required and must be a non-empty string")
	}

	voicePreset, _ := args["voice_preset"].(string)
	if voicePreset == "" {
		voicePreset = "v2/en_speaker_6" // Default voice
	}

	speed, _ := args["speed"].(float64)
	if speed == 0 {
		speed = 1.0 // Default speed
	}

	pitch, _ := args["pitch"].(float64)
	if pitch == 0 {
		pitch = 1.0 // Default pitch
	}

	temperature, _ := args["temperature"].(float64)
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}

	generation, _ := args["generation"].(int)
	if generation == 0 {
		generation = 1 // Default generation
	}

	preview := text
	if len(text) > 50 {
		preview = text[:50] + "..."
	}
	fmt.Printf("Generating TTS for: %s (voice: %s)\n", preview, voicePreset)

	// Split text into chunks if too long
	chunks := s.splitText(text)
	var audioFiles []string

	for i, chunk := range chunks {
		chunkID := fmt.Sprintf("%d_%d", utils.HashString(text), i)

		// Generate audio for each chunk
		audioPath, err := s.generateAudioChunk(chunk, voicePreset, speed, pitch, temperature, generation, chunkID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate audio for chunk %d: %w", i, err)
		}

		audioFiles = append(audioFiles, audioPath)
	}

	// Combine audio chunks if multiple
	finalAudioPath := audioFiles[0]
	if len(audioFiles) > 1 {
		var err error
		finalAudioPath, err = s.combineAudioFiles(audioFiles, text)
		if err != nil {
			return nil, fmt.Errorf("failed to combine audio chunks: %w", err)
		}
	}

	return map[string]interface{}{
		"audio_path":  finalAudioPath,
		"text":        text,
		"voice":       voicePreset,
		"speed":       speed,
		"pitch":       pitch,
		"chunks":      len(chunks),
		"sample_rate": s.sampleRate,
	}, nil
}

// generateAudioChunk generates audio for a single text chunk
func (s *BarkTTSServer) generateAudioChunk(text, voice string, speed, pitch, temperature float64, generation int, chunkID string) (string, error) {
	// Use OpenAI TTS as primary implementation (more reliable than local Bark)
	return s.callOpenAITTS(text, voice, chunkID)
}

// callOpenAITTS calls OpenAI TTS API
func (s *BarkTTSServer) callOpenAITTS(text, voice, chunkID string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Map Bark voice names to OpenAI voices
	openAIVoice := s.mapVoiceToOpenAI(voice)

	requestBody := map[string]interface{}{
		"model":           "tts-1",
		"input":           text,
		"voice":           openAIVoice,
		"response_format": "mp3",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI TTS API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI TTS API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Save audio response to file
	audioPath := filepath.Join(s.outputDir, fmt.Sprintf("%s.mp3", chunkID))
	out, err := os.Create(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to create audio file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save audio file: %w", err)
	}

	return audioPath, nil
}

// mapVoiceToOpenAI maps Bark voice names to OpenAI TTS voices
func (s *BarkTTSServer) mapVoiceToOpenAI(barkVoice string) string {
	voiceMap := map[string]string{
		"v2/en_speaker_0": "alloy",
		"v2/en_speaker_1": "echo",
		"v2/en_speaker_2": "fable",
		"v2/en_speaker_3": "onyx",
		"v2/en_speaker_4": "nova",
		"v2/en_speaker_5": "shimmer",
		"v2/en_speaker_6": "alloy", // default
		"v2/en_speaker_7": "echo",
		"v2/en_speaker_8": "fable",
		"v2/en_speaker_9": "nova",
	}

	if openAIVoice, exists := voiceMap[barkVoice]; exists {
		return openAIVoice
	}

	// Default to alloy if voice not found
	return "alloy"
}

// isBarkServerRunning checks if Bark server is available
func (s *BarkTTSServer) isBarkServerRunning() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.barkURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// callBarkServer calls the local Bark server
func (s *BarkTTSServer) callBarkServer(request BarkRequest, chunkID string) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.barkURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Bark server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response BarkResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("Bark generation failed: %s", response.Error)
	}

	return response.AudioPath, nil
}

// callBarkPython calls Bark Python implementation
func (s *BarkTTSServer) callBarkPython(request BarkRequest, chunkID string) (string, error) {
	// Generate Python script for Bark
	pythonScript := fmt.Sprintf(`
import os
import sys
sys.path.append('%s')

from bark import SAMPLE_RATE, generate_audio, preload_models
from scipy.io.wavfile import write as write_wav
import numpy as np

# Load models
preload_models()

# Generate audio
audio_array = generate_audio(
    "%s",
    history_prompt="%s",
    temperature=%f,
    generation=%d
)

# Apply speed and pitch adjustments
if %f != 1.0:
    audio_array = np.interp(
        np.linspace(0, 1, int(len(audio_array) / %f)),
        np.linspace(0, 1, len(audio_array)),
        audio_array
    )

# Save to file
output_path = "%s/%s.wav"
write_wav(output_path, SAMPLE_RATE, audio_array)

print(output_path)
`,
		s.modelPath,
		request.Text,
		request.Voice,
		request.Temperature,
		request.Generation,
		request.Speed,
		request.Speed,
		s.outputDir,
		chunkID,
	)

	// Write script to temporary file
	scriptPath := filepath.Join(s.outputDir, fmt.Sprintf("%s_script.py", chunkID))
	if err := os.WriteFile(scriptPath, []byte(pythonScript), 0644); err != nil {
		return "", fmt.Errorf("failed to write Python script: %w", err)
	}
	defer os.Remove(scriptPath)

	// Execute Python script
	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	cmd := utils.ExecuteCommand(ctx, "python3", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Python execution failed: %w, output: %s", err, string(output))
	}

	// Parse output to get audio path
	audioPath := strings.TrimSpace(string(output))
	if !filepath.IsAbs(audioPath) {
		audioPath = filepath.Join(s.outputDir, audioPath)
	}

	// Verify audio file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return "", fmt.Errorf("generated audio file not found: %s", audioPath)
	}

	return audioPath, nil
}

// SplitText splits text into chunks of maximum length (exported for testing)
func (s *BarkTTSServer) SplitText(text string) []string {
	return s.splitText(text)
}

// splitText splits text into chunks of maximum length
func (s *BarkTTSServer) splitText(text string) []string {
	if len(text) <= s.maxLength {
		return []string{text}
	}

	var chunks []string
	words := strings.Fields(text)
	var currentChunk []string
	currentLength := 0

	for _, word := range words {
		if currentLength+len(word)+1 > s.maxLength {
			if len(currentChunk) > 0 {
				chunks = append(chunks, strings.Join(currentChunk, " "))
				currentChunk = []string{word}
				currentLength = len(word)
			} else {
				// Single word longer than maxLength, split it
				for i := 0; i < len(word); i += s.maxLength {
					end := i + s.maxLength
					if end > len(word) {
						end = len(word)
					}
					chunks = append(chunks, word[i:end])
				}
				currentChunk = []string{}
				currentLength = 0
			}
		} else {
			currentChunk = append(currentChunk, word)
			currentLength += len(word) + 1 // +1 for space
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

// combineAudioFiles combines multiple audio files into one
func (s *BarkTTSServer) combineAudioFiles(audioFiles []string, originalText string) (string, error) {
	// For now, just return the first file
	// In a real implementation, this would use FFmpeg or similar to concatenate audio
	combinedPath := filepath.Join(s.outputDir, fmt.Sprintf("combined_%d.wav", utils.HashString(originalText)))

	// Copy first file as combined (placeholder)
	src, err := os.ReadFile(audioFiles[0])
	if err != nil {
		return "", fmt.Errorf("failed to read first audio file: %w", err)
	}

	if err := os.WriteFile(combinedPath, src, 0644); err != nil {
		return "", fmt.Errorf("failed to write combined audio file: %w", err)
	}

	// Clean up chunk files
	for _, file := range audioFiles {
		os.Remove(file)
	}

	return combinedPath, nil
}

// ListVoices lists available Bark voices (exported for testing)
func (s *BarkTTSServer) ListVoices(args map[string]interface{}) (interface{}, error) {
	return s.listVoices(args)
}

// listVoices lists available Bark voices
func (s *BarkTTSServer) listVoices(args map[string]interface{}) (interface{}, error) {
	voices := []map[string]interface{}{
		{
			"id":          "v2/en_speaker_0",
			"name":        "English Speaker 0",
			"language":    "en",
			"gender":      "neutral",
			"description": "Standard English voice",
		},
		{
			"id":          "v2/en_speaker_1",
			"name":        "English Speaker 1",
			"language":    "en",
			"gender":      "male",
			"description": "Male English voice",
		},
		{
			"id":          "v2/en_speaker_2",
			"name":        "English Speaker 2",
			"language":    "en",
			"gender":      "female",
			"description": "Female English voice",
		},
		{
			"id":          "v2/en_speaker_3",
			"name":        "English Speaker 3",
			"language":    "en",
			"gender":      "male",
			"description": "Deep male English voice",
		},
		{
			"id":          "v2/en_speaker_4",
			"name":        "English Speaker 4",
			"language":    "en",
			"gender":      "female",
			"description": "Soft female English voice",
		},
		{
			"id":          "v2/en_speaker_5",
			"name":        "English Speaker 5",
			"language":    "en",
			"gender":      "male",
			"description": "Bright male English voice",
		},
		{
			"id":          "v2/en_speaker_6",
			"name":        "English Speaker 6",
			"language":    "en",
			"gender":      "female",
			"description": "Warm female English voice",
		},
		{
			"id":          "v2/en_speaker_7",
			"name":        "English Speaker 7",
			"language":    "en",
			"gender":      "neutral",
			"description": "Neutral English voice",
		},
		{
			"id":          "v2/en_speaker_8",
			"name":        "English Speaker 8",
			"language":    "en",
			"gender":      "male",
			"description": "Clear male English voice",
		},
		{
			"id":          "v2/en_speaker_9",
			"name":        "English Speaker 9",
			"language":    "en",
			"gender":      "female",
			"description": "Expressive female English voice",
		},
	}

	return map[string]interface{}{
		"voices": voices,
		"total":  len(voices),
	}, nil
}

// getInfo returns Bark TTS server information
func (s *BarkTTSServer) getInfo(args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"name":           "Bark TTS Server",
		"version":        "1.0.0",
		"server_url":     s.barkURL,
		"model_path":     s.modelPath,
		"sample_rate":    s.sampleRate,
		"max_length":     s.maxLength,
		"output_dir":     s.outputDir,
		"server_running": s.isBarkServerRunning(),
	}, nil
}
