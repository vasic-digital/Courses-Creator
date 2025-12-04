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

// SunoServer handles music generation using Suno AI
type SunoServer struct {
	*BaseServerImpl
	sunoURL     string
	apiKey      string
	outputDir   string
	maxDuration int
	sampleRate  int
}

// SunoRequest represents a Suno music generation request
type SunoRequest struct {
	Prompt     string                 `json:"prompt"`
	Duration   int                    `json:"duration,omitempty"`
	Style      string                 `json:"style,omitempty"`
	Mood       string                 `json:"mood,omitempty"`
	Tempo      int                    `json:"tempo,omitempty"`
	Instrument string                 `json:"instrument,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// SunoResponse represents a Suno music generation response
type SunoResponse struct {
	Success    bool    `json:"success"`
	AudioPath  string  `json:"audio_path,omitempty"`
	Duration   float64 `json:"duration,omitempty"`
	SampleRate int     `json:"sample_rate,omitempty"`
	Style      string  `json:"style,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// NewSunoServer creates a new Suno music generation server
func NewSunoServer() *SunoServer {
	config := MCPServerConfig{
		Name:       "suno-music",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    120 * time.Second, // Music generation takes longer
		MaxRetries: 2,
	}

	server := &SunoServer{
		BaseServerImpl: NewBaseServer(config),
		sunoURL:        "http://localhost:8766/generate", // Default Suno server URL
		apiKey:         os.Getenv("SUNO_API_KEY"),
		outputDir:      "/tmp/suno_output",
		maxDuration:    30, // Maximum 30 seconds per generation
		sampleRate:     44100,
	}

	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)

	server.RegisterTools()
	return server
}

// NewSunoServerWithConfig creates a new Suno server with custom config
func NewSunoServerWithConfig(sunoURL, apiKey, outputDir string, maxDuration, sampleRate int) *SunoServer {
	config := MCPServerConfig{
		Name:       "suno-music",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    120 * time.Second,
		MaxRetries: 2,
	}

	server := &SunoServer{
		BaseServerImpl: NewBaseServer(config),
		sunoURL:        sunoURL,
		apiKey:         apiKey,
		outputDir:      outputDir,
		maxDuration:    maxDuration,
		sampleRate:     sampleRate,
	}

	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)

	server.RegisterTools()
	return server
}

// RegisterTools registers the music generation tools
func (s *SunoServer) RegisterTools() {
	s.AddTool("generate_music", "Generate background music using Suno AI", s.generateMusic)
	s.AddTool("list_styles", "List available music styles", s.listStyles)
	s.AddTool("get_info", "Get Suno server information", s.getInfo)
}

// GenerateMusic generates music from prompt (public method for direct calls)
func (s *SunoServer) GenerateMusic(args map[string]interface{}) (interface{}, error) {
	return s.generateMusic(args)
}

// generateMusic generates background music
func (s *SunoServer) generateMusic(args map[string]interface{}) (interface{}, error) {
	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return nil, fmt.Errorf("prompt parameter is required and must be a non-empty string")
	}

	duration, _ := args["duration"].(int)
	if duration == 0 || duration > s.maxDuration {
		duration = 10 // Default 10 seconds
	}

	style, _ := args["style"].(string)
	if style == "" {
		style = "ambient" // Default style
	}

	mood, _ := args["mood"].(string)
	if mood == "" {
		mood = "neutral" // Default mood
	}

	tempo, _ := args["tempo"].(int)
	if tempo == 0 {
		tempo = 120 // Default tempo
	}

	instrument, _ := args["instrument"].(string)
	if instrument == "" {
		instrument = "piano" // Default instrument
	}

	preview := prompt
	if len(preview) > 50 {
		preview = prompt[:50] + "..."
	}
	fmt.Printf("Generating music for: %s (style: %s, mood: %s)\n", preview, style, mood)

	request := SunoRequest{
		Prompt:     prompt,
		Duration:   duration,
		Style:      style,
		Mood:       mood,
		Tempo:      tempo,
		Instrument: instrument,
		Settings: map[string]interface{}{
			"sample_rate": s.sampleRate,
			"format":      "wav",
		},
	}

	// Use cloud music generation API (placeholder for now)
	// In a real implementation, this would call a service like Mubert, AIVA, or similar
	return s.generateMusicWithAPI(request)
}

// generateMusicWithAPI generates music using a cloud API (placeholder implementation)
func (s *SunoServer) generateMusicWithAPI(request SunoRequest) (interface{}, error) {
	// For now, create a placeholder response
	// In a real implementation, this would call a music generation API

	// Generate a unique filename
	filename := fmt.Sprintf("music_%d_%s.wav", utils.HashString(request.Prompt), time.Now().Format("20060102_150405"))
	outputPath := filepath.Join(s.outputDir, filename)

	// Create a placeholder music file (silence for now)
	// In a real implementation, this would download generated music from the API
	err := s.createPlaceholderMusicFile(outputPath, request.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create music file: %w", err)
	}

	return map[string]interface{}{
		"audio_path":  outputPath,
		"prompt":      request.Prompt,
		"duration":    request.Duration,
		"style":       request.Style,
		"mood":        request.Mood,
		"tempo":       request.Tempo,
		"instrument":  request.Instrument,
		"sample_rate": s.sampleRate,
		"note":        "This is a placeholder implementation. Replace with real music generation API.",
	}, nil
}

// createPlaceholderMusicFile creates a placeholder music file
func (s *SunoServer) createPlaceholderMusicFile(outputPath string, duration int) error {
	// For now, create an empty WAV file as placeholder
	// In a real implementation, this would be replaced with actual music generation

	// Create a minimal WAV file header (silence)
	wavHeader := []byte{
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x00, 0x00, 0x00, 0x00, // File size (placeholder)
		0x57, 0x41, 0x56, 0x45, // "WAVE"
		0x66, 0x6D, 0x74, 0x20, // "fmt "
		0x10, 0x00, 0x00, 0x00, // Chunk size
		0x01, 0x00, // Audio format (PCM)
		0x01, 0x00, // Num channels
		0x80, 0x3E, 0x00, 0x00, // Sample rate (16000)
		0x80, 0x3E, 0x00, 0x00, // Byte rate
		0x01, 0x00, // Block align
		0x08, 0x00, // Bits per sample
		0x64, 0x61, 0x74, 0x61, // "data"
		0x00, 0x00, 0x00, 0x00, // Data size (placeholder)
	}

	// Calculate actual sizes
	sampleRate := 16000
	bitsPerSample := 8
	numChannels := 1
	dataSize := duration * sampleRate * numChannels * bitsPerSample / 8
	fileSize := len(wavHeader) + dataSize - 8

	// Update header with correct sizes
	wavHeader[4] = byte(fileSize & 0xFF)
	wavHeader[5] = byte((fileSize >> 8) & 0xFF)
	wavHeader[6] = byte((fileSize >> 16) & 0xFF)
	wavHeader[7] = byte((fileSize >> 24) & 0xFF)

	wavHeader[40] = byte(dataSize & 0xFF)
	wavHeader[41] = byte((dataSize >> 8) & 0xFF)
	wavHeader[42] = byte((dataSize >> 16) & 0xFF)
	wavHeader[43] = byte((dataSize >> 24) & 0xFF)

	// Create silence data
	silenceData := make([]byte, dataSize)

	// Write file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(wavHeader); err != nil {
		return err
	}

	if _, err := file.Write(silenceData); err != nil {
		return err
	}

	return nil
}

// isSunoServerRunning checks if Suno server is available
func (s *SunoServer) isSunoServerRunning() bool {
	if s.apiKey == "" {
		return false // Need API key for server communication
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.sunoURL+"/health", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// callSunoServer calls the local Suno server
func (s *SunoServer) callSunoServer(request SunoRequest) (interface{}, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.sunoURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Create HTTP client with timeout and security settings
	client := &http.Client{
		Timeout: s.Config.Timeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Suno server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response SunoResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Suno generation failed: %s", response.Error)
	}

	return map[string]interface{}{
		"audio_path":  response.AudioPath,
		"prompt":      request.Prompt,
		"duration":    response.Duration,
		"style":       request.Style,
		"mood":        request.Mood,
		"tempo":       request.Tempo,
		"instrument":  request.Instrument,
		"sample_rate": response.SampleRate,
	}, nil
}

// callSunoPython calls Suno Python implementation
func (s *SunoServer) callSunoPython(request SunoRequest) (interface{}, error) {
	// Generate Python script for Suno (mock implementation)
	pythonScript := fmt.Sprintf(`
import os
import sys
import numpy as np
import soundfile as sf

# Mock Suno implementation - generates simple audio
sample_rate = %d
duration = %d
tempo = %d

# Generate simple sine wave music based on style
def generate_music(style, mood, tempo):
    t = np.linspace(0, duration, sample_rate * duration)
    
    if style == "ambient":
        # Slow, atmospheric sounds
        freq = 220.0 * (tempo / 120.0)  # A3 note
        wave = np.sin(2 * np.pi * freq * t)
        # Add some reverb effect (simple delay)
        wave = wave + 0.3 * np.roll(wave, sample_rate // 4)
    elif style == "classical":
        # Classical piano-style
        freqs = [261.63, 293.66, 329.63, 349.23, 392.00, 440.00, 493.88]  # C major scale
        wave = np.zeros_like(t)
        note_duration = duration / len(freqs)
        for i, freq in enumerate(freqs):
            start = int(i * note_duration * sample_rate)
            end = int((i + 1) * note_duration * sample_rate)
            wave[start:end] = np.sin(2 * np.pi * freq * t[start:end]) * np.exp(-0.5 * (t[start:end] - i * note_duration) / note_duration)
    elif style == "electronic":
        # Electronic/synthesizer
        freq = 130.81 * (tempo / 120.0)  # C3 note
        wave = np.sin(2 * np.pi * freq * t)
        # Add harmonics for richness
        wave += 0.5 * np.sin(4 * np.pi * freq * t)
        wave += 0.25 * np.sin(6 * np.pi * freq * t)
        # Add some rhythm
        beat_freq = tempo / 60.0
        wave *= 0.8 + 0.2 * np.sin(2 * np.pi * beat_freq * t)
    else:
        # Default simple music
        freq = 440.0 * (tempo / 120.0)  # A4 note
        wave = np.sin(2 * np.pi * freq * t)
    
    # Apply mood-based modifications
    if mood == "happy":
        wave *= 1.2
    elif mood == "sad":
        wave *= 0.8
    elif mood == "energetic":
        wave *= 1.5
    
    # Normalize
    wave = wave / np.max(np.abs(wave)) * 0.8
    
    return wave

# Generate music
audio = generate_music("%s", "%s", %d)

# Save to file
output_path = "%s/music_%d.wav"
sf.write(output_path, audio, sample_rate)

print(output_path)
`,
		s.sampleRate,
		request.Duration,
		request.Tempo,
		request.Style,
		request.Mood,
		request.Tempo,
		s.outputDir,
		utils.HashString(request.Prompt+request.Style+request.Mood),
	)

	// Write script to temporary file
	scriptPath := filepath.Join(s.outputDir, fmt.Sprintf("suno_script_%d.py", utils.HashString(request.Prompt)))
	if err := os.WriteFile(scriptPath, []byte(pythonScript), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Python script: %w", err)
	}
	defer os.Remove(scriptPath)

	// Execute Python script
	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	cmd := utils.ExecuteCommand(ctx, "python3", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Python execution failed: %w, output: %s", err, string(output))
	}

	// Parse output to get audio path
	audioPath := strings.TrimSpace(string(output))
	if !filepath.IsAbs(audioPath) {
		audioPath = filepath.Join(s.outputDir, audioPath)
	}

	// Verify audio file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("generated audio file not found: %s", audioPath)
	}

	return map[string]interface{}{
		"audio_path":  audioPath,
		"prompt":      request.Prompt,
		"duration":    float64(request.Duration),
		"style":       request.Style,
		"mood":        request.Mood,
		"tempo":       request.Tempo,
		"instrument":  request.Instrument,
		"sample_rate": s.sampleRate,
	}, nil
}

// listStyles lists available music styles
func (s *SunoServer) listStyles(args map[string]interface{}) (interface{}, error) {
	styles := []map[string]interface{}{
		{
			"id":          "ambient",
			"name":        "Ambient",
			"description": "Atmospheric and calming background music",
			"moods":       []string{"neutral", "calm", "peaceful"},
			"instruments": []string{"synthesizer", "pads", "strings"},
		},
		{
			"id":          "classical",
			"name":        "Classical",
			"description": "Traditional classical music style",
			"moods":       []string{"elegant", "sophisticated", "dramatic"},
			"instruments": []string{"piano", "strings", "orchestra"},
		},
		{
			"id":          "electronic",
			"name":        "Electronic",
			"description": "Modern electronic and synthesizer music",
			"moods":       []string{"energetic", "modern", "futuristic"},
			"instruments": []string{"synthesizer", "drums", "bass"},
		},
		{
			"id":          "jazz",
			"name":        "Jazz",
			"description": "Jazz and swing music styles",
			"moods":       []string{"relaxed", "sophisticated", "upbeat"},
			"instruments": []string{"piano", "saxophone", "bass", "drums"},
		},
		{
			"id":          "rock",
			"name":        "Rock",
			"description": "Rock and pop music styles",
			"moods":       []string{"energetic", "powerful", "dynamic"},
			"instruments": []string{"guitar", "bass", "drums"},
		},
		{
			"id":          "cinematic",
			"name":        "Cinematic",
			"description": "Epic film and soundtrack music",
			"moods":       []string{"dramatic", "emotional", "epic"},
			"instruments": []string{"orchestra", "choir", "percussion"},
		},
	}

	return map[string]interface{}{
		"styles": styles,
		"total":  len(styles),
	}, nil
}

// getInfo returns Suno server information
func (s *SunoServer) getInfo(args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"name":           "Suno Music Generation Server",
		"version":        "1.0.0",
		"server_url":     s.sunoURL,
		"max_duration":   s.maxDuration,
		"sample_rate":    s.sampleRate,
		"output_dir":     s.outputDir,
		"api_key_set":    s.apiKey != "",
		"server_running": s.isSunoServerRunning(),
	}, nil
}
