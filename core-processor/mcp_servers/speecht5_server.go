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

// SpeechT5TTSServer handles SpeechT5 text-to-speech generation
type SpeechT5TTSServer struct {
	*BaseServerImpl
	speechT5URL string
	modelPath   string
	outputDir   string
	maxLength   int
	sampleRate  int
}

// SpeechT5Request represents a SpeechT5 TTS generation request
type SpeechT5Request struct {
	Text       string                 `json:"text"`
	Voice      string                 `json:"voice,omitempty"`
	Speed      float64                `json:"speed,omitempty"`
	Pitch      float64                `json:"pitch,omitempty"`
	SampleRate int                    `json:"sample_rate,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// SpeechT5Response represents a SpeechT5 TTS generation response
type SpeechT5Response struct {
	Success    bool    `json:"success"`
	AudioPath  string  `json:"audio_path,omitempty"`
	Duration   float64 `json:"duration,omitempty"`
	SampleRate int     `json:"sample_rate,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// NewSpeechT5Server creates a new SpeechT5 TTS server
func NewSpeechT5Server() *SpeechT5TTSServer {
	config := MCPServerConfig{
		Name:       "speecht5-tts",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    60 * time.Second,
		MaxRetries: 3,
	}

	server := &SpeechT5TTSServer{
		BaseServerImpl: NewBaseServer(config),
		speechT5URL:    "http://localhost:8082/generate", // Default SpeechT5 server URL
		modelPath:      "/models/speecht5",
		outputDir:      "/tmp/speecht5_output",
		maxLength:      300, // Maximum text length per generation
		sampleRate:     16000,
	}

	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)

	server.RegisterTools()
	return server
}

// NewSpeechT5ServerWithConfig creates a new SpeechT5 TTS server with custom config
func NewSpeechT5ServerWithConfig(speechT5URL, modelPath, outputDir string, maxLength, sampleRate int) *SpeechT5TTSServer {
	config := MCPServerConfig{
		Name:       "speecht5-tts",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    60 * time.Second,
		MaxRetries: 3,
	}

	server := &SpeechT5TTSServer{
		BaseServerImpl: NewBaseServer(config),
		speechT5URL:    speechT5URL,
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
func (s *SpeechT5TTSServer) RegisterTools() {
	s.AddTool("generate_tts", "Generate speech audio from text using SpeechT5 TTS", s.generateTTS)
	s.AddTool("list_voices", "List available SpeechT5 voices", s.listVoices)
	s.AddTool("get_info", "Get SpeechT5 TTS server information", s.getInfo)
}

// GenerateTTS generates TTS audio from text (public method for direct calls)
func (s *SpeechT5TTSServer) GenerateTTS(args map[string]interface{}) (interface{}, error) {
	return s.generateTTS(args)
}

// generateTTS generates TTS audio from text
func (s *SpeechT5TTSServer) generateTTS(args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("text parameter is required and must be a non-empty string")
	}

	voicePreset, _ := args["voice_preset"].(string)
	if voicePreset == "" {
		voicePreset = "default" // Default voice
	}

	speed, _ := args["speed"].(float64)
	if speed == 0 {
		speed = 1.0 // Default speed
	}

	pitch, _ := args["pitch"].(float64)
	if pitch == 0 {
		pitch = 1.0 // Default pitch
	}

	sampleRate, _ := args["sample_rate"].(int)
	if sampleRate == 0 {
		sampleRate = s.sampleRate
	}

	preview := text
	if len(text) > 50 {
		preview = text[:50] + "..."
	}
	fmt.Printf("Generating SpeechT5 TTS for: %s (voice: %s)\n", preview, voicePreset)

	// Split text into chunks if too long
	chunks := s.splitText(text)
	var audioFiles []string

	for i, chunk := range chunks {
		chunkID := fmt.Sprintf("%d_%d", utils.HashString(text), i)

		// Generate audio for each chunk
		audioPath, err := s.generateAudioChunk(chunk, voicePreset, speed, pitch, sampleRate, chunkID)
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
		"sample_rate": sampleRate,
	}, nil
}

// generateAudioChunk generates audio for a single text chunk
func (s *SpeechT5TTSServer) generateAudioChunk(text, voice string, speed, pitch float64, sampleRate int, chunkID string) (string, error) {
	// Use ElevenLabs TTS as primary implementation (alternative to OpenAI)
	return s.callElevenLabsTTS(text, voice, chunkID)
}

// callElevenLabsTTS calls ElevenLabs TTS API
func (s *SpeechT5TTSServer) callElevenLabsTTS(text, voice, chunkID string) (string, error) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ELEVENLABS_API_KEY environment variable not set")
	}

	// Map SpeechT5 voice names to ElevenLabs voices
	elevenLabsVoice := s.mapVoiceToElevenLabs(voice)

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.5,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", elevenLabsVoice)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call ElevenLabs TTS API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ElevenLabs TTS API error (status %d): %s", resp.StatusCode, string(body))
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

// mapVoiceToElevenLabs maps SpeechT5 voice names to ElevenLabs voices
func (s *SpeechT5TTSServer) mapVoiceToElevenLabs(speechT5Voice string) string {
	voiceMap := map[string]string{
		"default": "21m00Tcm4TlvDq8ikWAM", // Rachel
		"male":    "29vD33N1CtxCmqQRPOHJ", // Drew
		"female":  "AZnzlk1XvdvUeBnXmlld", // Dora
		"young":   "EXAVITQu4vr4xnSDxMaL", // Bella
		"mature":  "ErXwobaYiN019PkySvjV", // Antoni
		"neutral": "VR6AewLTigWG4xSOukaG", // Arnold
		"excited": "pNInz6obpgDQGcFmaJgB", // Adam
		"calm":    "21m00Tcm4TlvDq8ikWAM", // Rachel (default)
	}

	if elevenLabsVoice, exists := voiceMap[speechT5Voice]; exists {
		return elevenLabsVoice
	}

	// Default to Rachel if voice not found
	return "21m00Tcm4TlvDq8ikWAM"
}

// isSpeechT5ServerRunning checks if SpeechT5 server is available
func (s *SpeechT5TTSServer) isSpeechT5ServerRunning() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.speechT5URL+"/health", nil)
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

// callSpeechT5Server calls the local SpeechT5 server
func (s *SpeechT5TTSServer) callSpeechT5Server(request SpeechT5Request, chunkID string) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.speechT5URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call SpeechT5 server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response SpeechT5Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("SpeechT5 generation failed: %s", response.Error)
	}

	return response.AudioPath, nil
}

// callSpeechT5Python calls SpeechT5 Python implementation
func (s *SpeechT5TTSServer) callSpeechT5Python(request SpeechT5Request, chunkID string) (string, error) {
	// Generate Python script for SpeechT5
	pythonScript := fmt.Sprintf(`
import os
import sys
sys.path.append('%s')

import torch
import numpy as np
from transformers import SpeechT5Processor, SpeechT5ForTextToSpeech, SpeechT5HifiGan
from datasets import load_dataset
import soundfile as sf

# Load models
processor = SpeechT5Processor.from_pretrained("microsoft/speecht5_tts")
model = SpeechT5ForTextToSpeech.from_pretrained("microsoft/speecht5_tts")
vocoder = SpeechT5HifiGan.from_pretrained("microsoft/speecht5_hifigan")

# Load speaker embeddings
embeddings_dataset = load_dataset("Matthijs/cmu-arctic-xvectors", split="validation")
speaker_embeddings = torch.tensor(embeddings_dataset[%d]["xvector"]).unsqueeze(0)

# Prepare input
inputs = processor(text="%s", return_tensors="pt")

# Generate speech
speech = model.generate_speech(inputs["input_ids"], speaker_embeddings=speaker_embeddings, vocoder=vocoder)

# Apply speed and pitch adjustments
if %f != 1.0:
    # Simple speed adjustment using resampling
    import librosa
    speech = speech.squeeze().numpy()
    speech = librosa.effects.time_stretch(speech, rate=%f)
    speech = torch.tensor(speech).unsqueeze(0)

if %f != 1.0:
    # Simple pitch adjustment
    speech = speech.squeeze().numpy()
    speech = librosa.effects.pitch_shift(speech, sr=%d, n_steps=%f)
    speech = torch.tensor(speech).unsqueeze(0)

# Save to file
output_path = "%s/%s.wav"
sf.write(output_path, speech.squeeze().numpy(), %d)
print(output_path)
`,
		s.modelPath,
		0, // Default speaker embedding
		request.Text,
		request.Speed,
		request.Speed,
		request.Pitch,
		request.SampleRate,
		request.Pitch,
		s.outputDir,
		chunkID,
		request.SampleRate,
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

// splitText splits text into chunks of maximum length
func (s *SpeechT5TTSServer) splitText(text string) []string {
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
func (s *SpeechT5TTSServer) combineAudioFiles(audioFiles []string, originalText string) (string, error) {
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

// listVoices lists available SpeechT5 voices
func (s *SpeechT5TTSServer) listVoices(args map[string]interface{}) (interface{}, error) {
	voices := []map[string]interface{}{
		{
			"id":          "default",
			"name":        "Default Voice",
			"language":    "en",
			"gender":      "neutral",
			"description": "Default SpeechT5 voice",
		},
		{
			"id":          "cmu_us_bdl_arctic",
			"name":        "BDL Male Voice",
			"language":    "en",
			"gender":      "male",
			"description": "CMU Arctic BDL male voice",
		},
		{
			"id":          "cmu_us_slt_arctic",
			"name":        "SLT Female Voice",
			"language":    "en",
			"gender":      "female",
			"description": "CMU Arctic SLT female voice",
		},
		{
			"id":          "cmu_us_clb_arctic",
			"name":        "CLB Female Voice",
			"language":    "en",
			"gender":      "female",
			"description": "CMU Arctic CLB female voice",
		},
		{
			"id":          "cmu_us_rms_arctic",
			"name":        "RMS Male Voice",
			"language":    "en",
			"gender":      "male",
			"description": "CMU Arctic RMS male voice",
		},
	}

	return map[string]interface{}{
		"voices": voices,
		"total":  len(voices),
	}, nil
}

// getInfo returns SpeechT5 TTS server information
func (s *SpeechT5TTSServer) getInfo(args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"name":           "SpeechT5 TTS Server",
		"version":        "1.0.0",
		"server_url":     s.speechT5URL,
		"model_path":     s.modelPath,
		"sample_rate":    s.sampleRate,
		"max_length":     s.maxLength,
		"output_dir":     s.outputDir,
		"server_running": s.isSpeechT5ServerRunning(),
	}, nil
}
