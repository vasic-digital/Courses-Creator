package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/course-creator/core-processor/models"
)

// OpenAIProvider represents OpenAI API provider
type OpenAIProvider struct {
	*BaseProvider
	apiKey       string
	model        string
	baseURL      string
	costPerToken float64
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &OpenAIProvider{
		BaseProvider: NewBaseProvider("OpenAI", ProviderTypePaid),
		apiKey:       apiKey,
		model:        model,
		baseURL:      "https://api.openai.com/v1",
		costPerToken: getOpenAICostPerToken(model),
	}
}

// GenerateText generates text using OpenAI API
func (p *OpenAIProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	// Prepare request payload
	requestBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":        2000,
		"temperature":       0.7,
		"top_p":             1,
		"frequency_penalty": 0,
		"presence_penalty":  0,
	}

	// Add quality settings
	if options.Quality == "high" {
		requestBody["max_tokens"] = 4000
		requestBody["temperature"] = 0.5
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		p.SetAvailable(false)
		return "", fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		p.SetAvailable(false)
		return "", fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	p.SetAvailable(true)

	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return response.Choices[0].Message.Content, nil
}

// GetCostEstimate returns cost estimate for OpenAI
func (p *OpenAIProvider) GetCostEstimate(textLength int) float64 {
	// Rough estimate: assume 4 chars per token
	tokens := textLength / 4
	return float64(tokens) * p.costPerToken
}

// getOpenAICostPerToken returns cost per 1K tokens for OpenAI models
func getOpenAICostPerToken(model string) float64 {
	costs := map[string]float64{
		"gpt-3.5-turbo": 0.002,   // $0.002 per 1K tokens
		"gpt-4":         0.03,    // $0.03 per 1K tokens
		"gpt-4-turbo":   0.01,    // $0.01 per 1K tokens
		"gpt-4o":        0.005,   // $0.005 per 1K tokens
		"gpt-4o-mini":   0.00015, // $0.00015 per 1K tokens
	}

	if cost, exists := costs[model]; exists {
		return cost / 1000.0 // Convert to per-token cost
	}
	return 0.002 / 1000.0 // Default to GPT-3.5-turbo pricing
}

// AnthropicProvider represents Anthropic Claude API provider
type AnthropicProvider struct {
	*BaseProvider
	apiKey       string
	model        string
	baseURL      string
	costPerToken float64
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if model == "" {
		model = "claude-3-haiku-20240307"
	}

	return &AnthropicProvider{
		BaseProvider: NewBaseProvider("Anthropic", ProviderTypePaid),
		apiKey:       apiKey,
		model:        model,
		baseURL:      "https://api.anthropic.com/v1",
		costPerToken: getAnthropicCostPerToken(model),
	}
}

// GenerateText generates text using Anthropic API
func (p *AnthropicProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("Anthropic API key not configured")
	}

	// Prepare request payload
	requestBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 2000,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	// Add quality settings
	if options.Quality == "high" {
		requestBody["max_tokens"] = 4000
		requestBody["temperature"] = 0.5
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		p.SetAvailable(false)
		return "", fmt.Errorf("Anthropic API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		p.SetAvailable(false)
		return "", fmt.Errorf("Anthropic API error: %d - %s", resp.StatusCode, string(body))
	}

	p.SetAvailable(true)

	// Parse response
	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in Anthropic response")
	}

	return response.Content[0].Text, nil
}

// GetCostEstimate returns cost estimate for Anthropic
func (p *AnthropicProvider) GetCostEstimate(textLength int) float64 {
	tokens := textLength / 4
	return float64(tokens) * p.costPerToken
}

// getAnthropicCostPerToken returns cost per 1K tokens for Anthropic models
func getAnthropicCostPerToken(model string) float64 {
	costs := map[string]float64{
		"claude-3-haiku-20240307":    0.00025, // $0.00025 per 1K tokens
		"claude-3-sonnet-20240229":   0.003,   // $0.003 per 1K tokens
		"claude-3-opus-20240229":     0.015,   // $0.015 per 1K tokens
		"claude-3-5-sonnet-20241022": 0.003,   // $0.003 per 1K tokens
	}

	if cost, exists := costs[model]; exists {
		return cost / 1000.0 // Convert to per-token cost
	}
	return 0.00025 / 1000.0 // Default to Claude 3 Haiku pricing
}

// OllamaProvider represents local Ollama provider
type OllamaProvider struct {
	*BaseProvider
	baseURL string
	model   string
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama2"
	}

	return &OllamaProvider{
		BaseProvider: NewBaseProvider("Ollama", ProviderTypeFree),
		baseURL:      baseURL,
		model:        model,
	}
}

// GenerateText generates text using Ollama API
func (p *OllamaProvider) GenerateText(ctx context.Context, prompt string, options models.ProcessingOptions) (string, error) {
	// Prepare request payload
	requestBody := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}

	// Add quality settings
	if options.Quality == "high" {
		requestBody["options"] = map[string]interface{}{
			"temperature": 0.5,
			"num_predict": 4000,
		}
	} else {
		requestBody["options"] = map[string]interface{}{
			"temperature": 0.7,
			"num_predict": 2000,
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 120 * time.Second} // Ollama can be slower
	resp, err := client.Do(req)
	if err != nil {
		p.SetAvailable(false)
		return "", fmt.Errorf("Ollama API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		p.SetAvailable(false)
		return "", fmt.Errorf("Ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	p.SetAvailable(true)

	// Parse response
	var response struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Done {
		return "", fmt.Errorf("Ollama response incomplete")
	}

	return response.Response, nil
}

// GetCostEstimate returns cost estimate for Ollama (free = 0)
func (p *OllamaProvider) GetCostEstimate(textLength int) float64 {
	return 0.0 // Local models are free
}

// IsAvailable checks if Ollama is running and model is available
func (p *OllamaProvider) IsAvailable() bool {
	if !p.BaseProvider.IsAvailable() {
		return false
	}

	// Quick health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		p.SetAvailable(false)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.SetAvailable(false)
		return false
	}

	// Check if model is available
	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false
	}

	for _, model := range response.Models {
		if strings.HasPrefix(model.Name, p.model) {
			return true
		}
	}

	return false
}

// GetModel returns the model name for OpenAI provider
func (p *OpenAIProvider) GetModel() string {
	return p.model
}

// GetBaseURL returns the base URL for OpenAI provider
func (p *OpenAIProvider) GetBaseURL() string {
	return p.baseURL
}

// GetModel returns the model name for Anthropic provider
func (p *AnthropicProvider) GetModel() string {
	return p.model
}

// GetBaseURL returns the base URL for Anthropic provider
func (p *AnthropicProvider) GetBaseURL() string {
	return p.baseURL
}

// GetModel returns the model name for Ollama provider
func (p *OllamaProvider) GetModel() string {
	return p.model
}

// GetBaseURL returns the base URL for Ollama provider
func (p *OllamaProvider) GetBaseURL() string {
	return p.baseURL
}
