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
	"strings"
	"time"
)

// Pix2StructServer handles UI parsing using Pix2Struct (or similar vision models)
type Pix2StructServer struct {
	*BaseServerImpl
	apiURL       string
	apiKey       string
	outputDir    string
	maxImageSize int
	timeout      time.Duration
}

// Pix2StructRequest represents a Pix2Struct UI parsing request
type Pix2StructRequest struct {
	Image    string                 `json:"image"`  // Base64 encoded image or file path
	Prompt   string                 `json:"prompt"` // Parsing prompt
	Detail   string                 `json:"detail"` // Analysis detail level
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// Pix2StructResponse represents a Pix2Struct UI parsing response
type Pix2StructResponse struct {
	Success   bool                   `json:"success"`
	Structure map[string]interface{} `json:"structure,omitempty"`
	Text      []string               `json:"text,omitempty"`
	Elements  []UIElement            `json:"elements,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// UIElement represents a UI element extracted from an image
type UIElement struct {
	Type       string                 `json:"type"`       // button, input, text, etc.
	Text       string                 `json:"text"`       // element text content
	Bounds     map[string]int         `json:"bounds"`     // x, y, width, height
	Properties map[string]interface{} `json:"properties"` // additional properties
	Confidence float64                `json:"confidence"` // detection confidence
}

// NewPix2StructServer creates a new Pix2Struct UI parsing server
func NewPix2StructServer() *Pix2StructServer {
	config := MCPServerConfig{
		Name:       "pix2struct-ui",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    60 * time.Second,
		MaxRetries: 2,
	}

	server := &Pix2StructServer{
		BaseServerImpl: NewBaseServer(config),
		apiURL:         "https://api.openai.com/v1/chat/completions", // Using GPT-4 Vision for now
		apiKey:         os.Getenv("OPENAI_API_KEY"),
		outputDir:      "/tmp/pix2struct_output",
		maxImageSize:   5 * 1024 * 1024, // 5MB
		timeout:        60 * time.Second,
	}

	// Ensure output directory exists
	os.MkdirAll(server.outputDir, 0755)

	server.RegisterTools()
	return server
}

// RegisterTools registers the UI parsing tools
func (s *Pix2StructServer) RegisterTools() {
	s.AddTool("parse_ui", "Parse UI elements from screenshot or image", s.parseUI)
	s.AddTool("extract_buttons", "Extract button elements and their properties", s.extractButtons)
	s.AddTool("extract_forms", "Extract form elements and input fields", s.extractForms)
	s.AddTool("get_info", "Get Pix2Struct server information", s.getInfo)
}

// ParseUI parses UI elements from an image (public method for direct calls)
func (s *Pix2StructServer) ParseUI(args map[string]interface{}) (interface{}, error) {
	return s.parseUI(args)
}

// parseUI performs comprehensive UI parsing
func (s *Pix2StructServer) parseUI(args map[string]interface{}) (interface{}, error) {
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("image parameter is required and must be a non-empty string")
	}

	prompt, _ := args["prompt"].(string)
	if prompt == "" {
		prompt = "Analyze this UI screenshot and describe all visible elements, their types, text content, and layout. Identify buttons, inputs, text fields, and other interactive elements."
	}

	detail, _ := args["detail"].(string)
	if detail == "" {
		detail = "high"
	}

	fmt.Printf("Parsing UI with detail level: %s\n", detail)

	// Use OpenAI GPT-4 Vision for UI parsing
	return s.callOpenAIVisionForUI(image, prompt, detail)
}

// callOpenAIVisionForUI calls OpenAI GPT-4 Vision API for UI parsing
func (s *Pix2StructServer) callOpenAIVisionForUI(imageInput, prompt, detail string) (interface{}, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Prepare image data for OpenAI API
	imageData, err := s.prepareImageForOpenAI(imageInput)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare image: %w", err)
	}

	// Build messages for GPT-4 Vision
	messages := []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": prompt,
				},
				{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": imageData,
					},
				},
			},
		},
	}

	requestBody := map[string]interface{}{
		"model":      "gpt-4-vision-preview",
		"messages":   messages,
		"max_tokens": 1000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI Vision API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI Vision API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI Vision API")
	}

	analysis := response.Choices[0].Message.Content

	// Parse the analysis to extract structured UI elements
	elements := s.parseUIElementsFromAnalysis(analysis)

	return map[string]interface{}{
		"structure": map[string]interface{}{
			"description": analysis,
			"detail":      detail,
		},
		"text":     s.extractTextFromUIAnalysis(analysis),
		"elements": elements,
		"model":    "gpt-4-vision-preview",
	}, nil
}

// prepareImageForOpenAI prepares image data for OpenAI API
func (s *Pix2StructServer) prepareImageForOpenAI(imageInput string) (string, error) {
	// Check if it's already a URL
	if strings.HasPrefix(imageInput, "http") {
		return imageInput, nil
	}

	// Check if it's a file path
	if _, err := os.Stat(imageInput); err == nil {
		// Read file and convert to base64
		data, err := os.ReadFile(imageInput)
		if err != nil {
			return "", fmt.Errorf("failed to read image file: %w", err)
		}

		// Check file size
		if len(data) > s.maxImageSize {
			return "", fmt.Errorf("image file too large: %d bytes (max %d)", len(data), s.maxImageSize)
		}

		// Determine content type
		contentType := "image/jpeg" // default
		if strings.HasSuffix(strings.ToLower(imageInput), ".png") {
			contentType = "image/png"
		}

		base64Data := base64.StdEncoding.EncodeToString(data)
		return fmt.Sprintf("data:%s;base64,%s", contentType, base64Data), nil
	}

	// Assume it's base64 data
	if strings.Contains(imageInput, "base64,") {
		return imageInput, nil
	}

	// Try to decode as raw base64
	if _, err := base64.StdEncoding.DecodeString(imageInput); err == nil {
		return fmt.Sprintf("data:image/jpeg;base64,%s", imageInput), nil
	}

	return "", fmt.Errorf("invalid image input format")
}

// parseUIElementsFromAnalysis extracts structured UI elements from analysis text
func (s *Pix2StructServer) parseUIElementsFromAnalysis(analysis string) []UIElement {
	elements := []UIElement{}

	// Simple parsing logic - look for common UI patterns
	analysisLower := strings.ToLower(analysis)

	// Look for buttons
	if strings.Contains(analysisLower, "button") {
		elements = append(elements, UIElement{
			Type:       "button",
			Text:       s.extractButtonText(analysis),
			Confidence: 0.8,
			Properties: map[string]interface{}{
				"detected": true,
			},
		})
	}

	// Look for input fields
	if strings.Contains(analysisLower, "input") || strings.Contains(analysisLower, "field") || strings.Contains(analysisLower, "textbox") {
		elements = append(elements, UIElement{
			Type:       "input",
			Text:       "",
			Confidence: 0.7,
			Properties: map[string]interface{}{
				"detected": true,
			},
		})
	}

	// Look for text elements
	if strings.Contains(analysisLower, "text") {
		elements = append(elements, UIElement{
			Type:       "text",
			Text:       s.extractVisibleText(analysis),
			Confidence: 0.9,
			Properties: map[string]interface{}{
				"detected": true,
			},
		})
	}

	return elements
}

// extractButtonText extracts button text from analysis
func (s *Pix2StructServer) extractButtonText(analysis string) string {
	// Simple extraction - look for button-related text
	lines := strings.Split(analysis, "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "button") && (strings.Contains(lineLower, "says") || strings.Contains(lineLower, "with") || strings.Contains(lineLower, "labeled")) {
			// Extract text after keywords
			words := strings.Fields(line)
			for i, word := range words {
				if word == "says" || word == "with" || word == "labeled" {
					if i+1 < len(words) {
						return strings.Join(words[i+1:], " ")
					}
				}
			}
		}
	}
	return "Button"
}

// extractVisibleText extracts visible text from analysis
func (s *Pix2StructServer) extractVisibleText(analysis string) string {
	// Simple extraction - look for quoted text or text mentions
	lines := strings.Split(analysis, "\n")
	texts := []string{}

	for _, line := range lines {
		// Look for quoted text
		if strings.Contains(line, "\"") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != end && start >= 0 && end > start {
				text := line[start+1 : end]
				if len(text) > 0 {
					texts = append(texts, text)
				}
			}
		}
	}

	if len(texts) > 0 {
		return strings.Join(texts, " ")
	}

	return "Text content detected"
}

// extractTextFromUIAnalysis extracts all text from UI analysis
func (s *Pix2StructServer) extractTextFromUIAnalysis(analysis string) []string {
	texts := []string{}

	// Simple extraction - split by sentences and look for text content
	sentences := strings.Split(analysis, ".")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 10 && (strings.Contains(strings.ToLower(sentence), "text") || strings.Contains(strings.ToLower(sentence), "says") || strings.Contains(strings.ToLower(sentence), "shows")) {
			texts = append(texts, sentence)
		}
	}

	return texts
}

// extractButtons extracts button elements specifically
func (s *Pix2StructServer) extractButtons(args map[string]interface{}) (interface{}, error) {
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("image parameter is required")
	}

	result, err := s.parseUI(map[string]interface{}{
		"image":  image,
		"prompt": "Identify all buttons in this UI screenshot. For each button, describe its text, position, and appearance.",
		"detail": "high",
	})
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	elements, ok := resultMap["elements"].([]UIElement)
	if !ok {
		return nil, fmt.Errorf("no elements in response")
	}

	// Filter for buttons only
	buttons := []UIElement{}
	for _, element := range elements {
		if element.Type == "button" {
			buttons = append(buttons, element)
		}
	}

	return map[string]interface{}{
		"buttons": buttons,
		"count":   len(buttons),
	}, nil
}

// extractForms extracts form elements
func (s *Pix2StructServer) extractForms(args map[string]interface{}) (interface{}, error) {
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf("image parameter is required")
	}

	result, err := s.parseUI(map[string]interface{}{
		"image":  image,
		"prompt": "Identify all form elements in this UI screenshot, including input fields, dropdowns, checkboxes, and text areas. Describe their labels and types.",
		"detail": "high",
	})
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	elements, ok := resultMap["elements"].([]UIElement)
	if !ok {
		return nil, fmt.Errorf("no elements in response")
	}

	// Filter for form elements
	formElements := []UIElement{}
	for _, element := range elements {
		if element.Type == "input" || element.Type == "select" || element.Type == "textarea" || element.Type == "checkbox" {
			formElements = append(formElements, element)
		}
	}

	return map[string]interface{}{
		"form_elements": formElements,
		"count":         len(formElements),
	}, nil
}

// getInfo returns Pix2Struct server information
func (s *Pix2StructServer) getInfo(args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"name":           "Pix2Struct UI Parser",
		"version":        "1.0.0",
		"api_url":        s.apiURL,
		"output_dir":     s.outputDir,
		"max_image_size": s.maxImageSize,
		"timeout":        s.timeout.String(),
		"model":          "gpt-4-vision-preview",
	}, nil
}
