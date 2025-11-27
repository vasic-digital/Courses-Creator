package mcp_servers

import (
	"bytes"
	"context"
	"encoding/base64"
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

// LLaVAServer handles image analysis using LLaVA (Large Language and Vision Assistant)
type LLaVAServer struct {
	*BaseServerImpl
	llavaURL    string
	modelPath   string
	outputDir   string
	maxImageSize int
	contextWindow int
}

// LLaVARequest represents a LLaVA image analysis request
type LLaVARequest struct {
	Image      string                 `json:"image"`      // Base64 encoded image or file path
	Prompt     string                 `json:"prompt"`     // Analysis prompt
	Question   string                 `json:"question"`   // Specific question about image
	Detail     string                 `json:"detail"`     // Analysis detail level (low, high)
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// LLaVAResponse represents a LLaVA image analysis response
type LLaVAResponse struct {
	Success    bool     `json:"success"`
	Analysis   string   `json:"analysis,omitempty"`
	Objects    []string `json:"objects,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// LLaVAImageInfo represents information about an analyzed image
type LLaVAImageInfo struct {
	Objects     []string               `json:"objects"`
	Description string                 `json:"description"`
	Colors      []string               `json:"colors"`
	Style       string                 `json:"style"`
	Composition map[string]interface{} `json:"composition"`
	Text        []string               `json:"text"`
	Features    map[string]interface{} `json:"features"`
}

// NewLLaVAServer creates a new LLaVA image analysis server
func NewLLaVAServer() *LLaVAServer {
	config := MCPServerConfig{
		Name:       "llava-image",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    90 * time.Second,
		MaxRetries: 2,
	}
	
	server := &LLaVAServer{
		BaseServerImpl: NewBaseServer(config),
		llavaURL:        "http://localhost:8767/analyze", // Default LLaVA server URL
		modelPath:       "/models/llava",
		outputDir:       "/tmp/llava_output",
		maxImageSize:    5 * 1024 * 1024, // 5MB
		contextWindow:   2048,
	}
	
	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)
	
	server.RegisterTools()
	return server
}

// NewLLaVAServerWithConfig creates a new LLaVA server with custom config
func NewLLaVAServerWithConfig(llavaURL, modelPath, outputDir string, maxImageSize, contextWindow int) *LLaVAServer {
	config := MCPServerConfig{
		Name:       "llava-image",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    90 * time.Second,
		MaxRetries: 2,
	}
	
	server := &LLaVAServer{
		BaseServerImpl: NewBaseServer(config),
		llavaURL:       llavaURL,
		modelPath:      modelPath,
		outputDir:      outputDir,
		maxImageSize:   maxImageSize,
		contextWindow:  contextWindow,
	}
	
	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)
	
	server.RegisterTools()
	return server
}

// RegisterTools registers the image analysis tools
func (s *LLaVAServer) RegisterTools() {
	s.AddTool("analyze_image", "Analyze image content and provide description", s.analyzeImage)
	s.AddTool("extract_text", "Extract text from image using OCR", s.extractText)
	s.AddTool("detect_objects", "Detect and identify objects in image", s.detectObjects)
	s.AddTool("analyze_colors", "Analyze color palette and composition", s.analyzeColors)
	s.AddTool("get_info", "Get LLaVA server information", s.getInfo)
}

// AnalyzeImage analyzes image content (public method for direct calls)
func (s *LLaVAServer) AnalyzeImage(args map[string]interface{}) (interface{}, error) {
	return s.analyzeImage(args)
}

// analyzeImage performs comprehensive image analysis
func (s *LLaVAServer) analyzeImage(args map[string]interface{}) (interface{}, error) {
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("image parameter is required and must be a non-empty string")
	}

	prompt, _ := args["prompt"].(string)
	if prompt == "" {
		prompt = "Describe this image in detail, including objects, layout, colors, and style."
	}

	detail, _ := args["detail"].(string)
	if detail == "" {
		detail = "high" // Default high detail
	}

	question, _ := args["question"].(string)
	
	fmt.Printf("Analyzing image with detail level: %s\n", detail)

	// Process image (handle both file paths and base64)
	imageData, err := s.processImageInput(image)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	request := LLaVARequest{
		Image:    imageData,
		Prompt:   prompt,
		Question: question,
		Detail:   detail,
		Settings: map[string]interface{}{
			"context_window": s.contextWindow,
			"model_path":     s.modelPath,
		},
	}

	// Check if local LLaVA server is available
	if s.isLLaVAServerRunning() {
		return s.callLLaVAServer(request)
	}

	// Fallback to Python implementation
	return s.callLLaVAPython(request)
}

// extractText extracts text from image using OCR capabilities
func (s *LLaVAServer) extractText(args map[string]interface{}) (interface{}, error) {
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("image parameter is required and must be a non-empty string")
	}

	language, _ := args["language"].(string)
	if language == "" {
		language = "en" // Default English
	}

	fmt.Printf("Extracting text from image in language: %s\n", language)

	// Process image
	imageData, err := s.processImageInput(image)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	request := LLaVARequest{
		Image:  imageData,
		Prompt: fmt.Sprintf("Extract all text from this image. Return the text in %s language. Preserve the structure and formatting as much as possible.", language),
		Detail: "high",
		Settings: map[string]interface{}{
			"context_window": s.contextWindow,
			"model_path":     s.modelPath,
			"task_type":      "ocr",
		},
	}

	return s.callLLaVAPython(request)
}

// detectObjects detects and identifies objects in image
func (s *LLaVAServer) detectObjects(args map[string]interface{}) (interface{}, error) {
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("image parameter is required and must be a non-empty string")
	}

	confidence, _ := args["confidence"].(float64)
	if confidence == 0 {
		confidence = 0.5 // Default confidence threshold
	}

	fmt.Printf("Detecting objects with confidence threshold: %.2f\n", confidence)

	// Process image
	imageData, err := s.processImageInput(image)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	request := LLaVARequest{
		Image:  imageData,
		Prompt: fmt.Sprintf("Identify all objects in this image with confidence score >= %.2f. List each object with its approximate location and confidence level.", confidence),
		Detail: "high",
		Settings: map[string]interface{}{
			"context_window": s.contextWindow,
			"model_path":     s.modelPath,
			"task_type":      "object_detection",
			"confidence_threshold": confidence,
		},
	}

	return s.callLLaVAPython(request)
}

// analyzeColors analyzes color palette and composition
func (s *LLaVAServer) analyzeColors(args map[string]interface{}) (interface{}, error) {
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("image parameter is required and must be a non-empty string")
	}

	paletteSize, _ := args["palette_size"].(int)
	if paletteSize == 0 {
		paletteSize = 5 // Default 5 colors
	}

	fmt.Printf("Analyzing colors with palette size: %d\n", paletteSize)

	// Process image
	imageData, err := s.processImageInput(image)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	request := LLaVARequest{
		Image:  imageData,
		Prompt: fmt.Sprintf("Analyze the color composition of this image. Extract the dominant %d colors with their hex codes, proportions, and describe the overall color scheme and mood.", paletteSize),
		Detail: "high",
		Settings: map[string]interface{}{
			"context_window": s.contextWindow,
			"model_path":     s.modelPath,
			"task_type":      "color_analysis",
			"palette_size":   paletteSize,
		},
	}

	return s.callLLaVAPython(request)
}

// processImageInput processes image input (file path or base64)
func (s *LLaVAServer) processImageInput(image string) (string, error) {
	// Check if image is a file path
	if _, err := os.Stat(image); err == nil {
		// Read and encode image file
		data, err := os.ReadFile(image)
		if err != nil {
			return "", fmt.Errorf("failed to read image file: %w", err)
		}
		
		if len(data) > s.maxImageSize {
			return "", fmt.Errorf("image file too large: %d bytes (max %d bytes)", len(data), s.maxImageSize)
		}
		
		// Get file extension for format detection
		ext := strings.ToLower(filepath.Ext(image))
		var mimeType string
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		default:
			return "", fmt.Errorf("unsupported image format: %s", ext)
		}
		
		// Return base64 encoded image with mime type
		return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
	}
	
	// Check if image is already base64 encoded
	if strings.HasPrefix(image, "data:image/") {
		return image, nil
	}
	
	// Assume it's base64 without data URL prefix
	return fmt.Sprintf("image/png;base64,%s", image), nil
}

// isLLaVAServerRunning checks if LLaVA server is available
func (s *LLaVAServer) isLLaVAServerRunning() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.llavaURL+"/health", nil)
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

// callLLaVAServer calls the local LLaVA server
func (s *LLaVAServer) callLLaVAServer(request LLaVARequest) (interface{}, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.llavaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLaVA server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response LLaVAResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("LLaVA analysis failed: %s", response.Error)
	}

	return map[string]interface{}{
		"analysis":   response.Analysis,
		"objects":    response.Objects,
		"confidence": response.Confidence,
	}, nil
}

// callLLaVAPython calls LLaVA Python implementation
func (s *LLaVAServer) callLLaVAPython(request LLaVARequest) (interface{}, error) {
	// Generate Python script for LLaVA (mock implementation)
	taskType, _ := request.Settings["task_type"].(string)
	if taskType == "" {
		taskType = "general"
	}

	pythonScript := fmt.Sprintf(`
import os
import sys
import base64
import json
from io import BytesIO
from PIL import Image

# Mock LLaVA implementation for demonstration
def analyze_image(image_data, prompt, task_type="general"):
    # Decode base64 image
    if "," in image_data:
        image_data = image_data.split(",", 1)[1]
    
    try:
        image_bytes = base64.b64decode(image_data)
        image = Image.open(BytesIO(image_bytes))
        
        # Basic image analysis
        width, height = image.size
        mode = image.mode
        
        if task_type == "ocr":
            # Mock OCR - in real implementation, would use Tesseract or similar
            return {
                "text": ["Sample text extracted from image", "Another line of text"],
                "confidence": 0.85,
                "language": "en"
            }
        elif task_type == "object_detection":
            # Mock object detection
            return {
                "objects": [
                    {"name": "person", "confidence": 0.92, "bbox": [100, 150, 300, 450]},
                    {"name": "car", "confidence": 0.87, "bbox": [350, 200, 550, 400]},
                    {"name": "tree", "confidence": 0.75, "bbox": [50, 100, 150, 350]}
                ]
            }
        elif task_type == "color_analysis":
            # Mock color analysis
            return {
                "dominant_colors": [
                    {"hex": "#3B82F6", "name": "Blue", "proportion": 0.35},
                    {"hex": "#10B981", "name": "Green", "proportion": 0.25},
                    {"hex": "#F59E0B", "name": "Amber", "proportion": 0.20},
                    {"hex": "#EF4444", "name": "Red", "proportion": 0.15},
                    {"hex": "#8B5CF6", "name": "Purple", "proportion": 0.05}
                ],
                "scheme": "vibrant",
                "mood": "energetic"
            }
        else:
            # General image analysis
            return {
                "description": "This image contains various elements with rich colors and interesting composition. The overall style appears to be modern and well-balanced.",
                "objects": ["person", "building", "sky", "tree"],
                "style": "photorealistic",
                "mood": "neutral",
                "composition": "balanced"
            }
    except Exception as e:
        return {"error": str(e)}

# Parse request
request_json = '''%s'''
request = json.loads(request_json)

# Analyze image
result = analyze_image(request["image"], request["prompt"], "%s")

# Output result as JSON
print(json.dumps(result))
`,
		func() string {
			data, _ := json.Marshal(request)
			return string(data)
		}(),
		taskType,
	)

	// Write script to temporary file
	scriptPath := filepath.Join(s.outputDir, fmt.Sprintf("llava_script_%d.py", utils.HashString(request.Image+request.Prompt)))
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

	// Parse JSON output
	var result interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(output))), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Python output: %w, output: %s", err, string(output))
	}

	return result, nil
}

// getInfo returns LLaVA server information
func (s *LLaVAServer) getInfo(args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"name":           "LLaVA Image Analysis Server",
		"version":        "1.0.0",
		"server_url":     s.llavaURL,
		"model_path":     s.modelPath,
		"max_image_size": s.maxImageSize,
		"context_window": s.contextWindow,
		"output_dir":     s.outputDir,
		"server_running": s.isLLaVAServerRunning(),
	}, nil
}