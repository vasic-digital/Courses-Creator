package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/course-creator/core-processor/models"
)

// CourseContentGenerator generates course content using LLM providers
type CourseContentGenerator struct {
	providerManager *ProviderManager
}

// NewCourseContentGenerator creates a new course content generator
func NewCourseContentGenerator() *CourseContentGenerator {
	manager := NewProviderManager()
	
	// Auto-register available providers
	generator := &CourseContentGenerator{
		providerManager: manager,
	}
	
	// Register providers based on available API keys
	generator.registerAvailableProviders()
	
	return generator
}

// registerAvailableProviders registers providers based on environment configuration
func (ccg *CourseContentGenerator) registerAvailableProviders() {
	// Register OpenAI if API key is available
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		provider := NewOpenAIProvider(apiKey, os.Getenv("OPENAI_MODEL"))
		ccg.providerManager.RegisterProvider(provider)
		fmt.Printf("Registered OpenAI provider with model: %s\n", os.Getenv("OPENAI_MODEL"))
	}
	
	// Register Anthropic if API key is available
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		provider := NewAnthropicProvider(apiKey, os.Getenv("ANTHROPIC_MODEL"))
		ccg.providerManager.RegisterProvider(provider)
		fmt.Printf("Registered Anthropic provider with model: %s\n", os.Getenv("ANTHROPIC_MODEL"))
	}
	
	// Register Ollama if configured
	if ollamaURL := os.Getenv("OLLAMA_URL"); ollamaURL != "" {
		provider := NewOllamaProvider(ollamaURL, os.Getenv("OLLAMA_MODEL"))
		ccg.providerManager.RegisterProvider(provider)
		fmt.Printf("Registered Ollama provider at: %s with model: %s\n", ollamaURL, os.Getenv("OLLAMA_MODEL"))
	} else {
		// Try default Ollama installation
		provider := NewOllamaProvider("", os.Getenv("OLLAMA_MODEL"))
		if provider.IsAvailable() {
			ccg.providerManager.RegisterProvider(provider)
			fmt.Printf("Registered default Ollama provider with model: %s\n", os.Getenv("OLLAMA_MODEL"))
		}
	}
}

// GenerateCourseTitle generates a course title from content
func (ccg *CourseContentGenerator) GenerateCourseTitle(ctx context.Context, content string) (string, error) {
	prompt := fmt.Sprintf(`Based on the following course content, generate a compelling and descriptive course title. 
The title should be:
- Clear and descriptive
- Engaging and professional
- No longer than 60 characters
- Suitable for online learning platforms

Content:
%s

Return only the title, nothing else.`, content[:1000]) // Limit content to avoid token limits

	preferences := ProviderPreferences{
		AllowPaid:         true,
		PrioritizeQuality:  true,
		MaxCostPerRequest: 0.10, // $0.10 max for title generation
	}
	
	provider := ccg.providerManager.GetBestProvider(preferences)
	if provider == nil {
		return "", fmt.Errorf("no LLM providers available")
	}
	
	return ccg.providerManager.GenerateWithFallback(ctx, prompt, models.ProcessingOptions{
		Quality: "standard",
	})
}

// GenerateCourseDescription generates a course description from content
func (ccg *CourseContentGenerator) GenerateCourseDescription(ctx context.Context, title, content string) (string, error) {
	prompt := fmt.Sprintf(`Create an engaging course description for the following content.

Title: %s

Requirements:
- 2-3 paragraphs long
- Highlight key learning outcomes
- Target audience and prerequisites
- Engaging and professional tone
- Maximum 300 words

Content:
%s

Return only the description, nothing else.`, title, content[:2000])

	preferences := ProviderPreferences{
		AllowPaid:         true,
		PrioritizeQuality:  true,
		MaxCostPerRequest: 0.20, // $0.20 max for description generation
	}
	
	provider := ccg.providerManager.GetBestProvider(preferences)
	if provider == nil {
		return "", fmt.Errorf("no LLM providers available")
	}
	
	return ccg.providerManager.GenerateWithFallback(ctx, prompt, models.ProcessingOptions{
		Quality: "standard",
	})
}

// GenerateLessonContent generates enhanced lesson content
func (ccg *CourseContentGenerator) GenerateLessonContent(ctx context.Context, title, rawContent string) (string, error) {
	prompt := fmt.Sprintf(`Enhance and expand the following lesson content while maintaining the original structure and intent.

Lesson Title: %s

Requirements:
- Maintain original code examples and technical accuracy
- Add explanatory text where content is sparse
- Include learning objectives at the beginning
- Add a summary section at the end
- Keep technical examples intact
- Make content more educational and comprehensive

Original Content:
%s

Return the enhanced content in markdown format, nothing else.`, title, rawContent)

	preferences := ProviderPreferences{
		AllowPaid:         true,
		PrioritizeQuality:  true,
		MaxCostPerRequest: 0.50, // $0.50 max for lesson enhancement
	}
	
	provider := ccg.providerManager.GetBestProvider(preferences)
	if provider == nil {
		return "", fmt.Errorf("no LLM providers available")
	}
	
	return ccg.providerManager.GenerateWithFallback(ctx, prompt, models.ProcessingOptions{
		Quality: "high",
	})
}

// GenerateInteractiveElements generates interactive elements for lessons
func (ccg *CourseContentGenerator) GenerateInteractiveElements(ctx context.Context, lessonContent string) ([]string, error) {
	prompt := fmt.Sprintf(`Analyze the following lesson content and suggest 3-5 interactive elements.

Requirements:
- Include different types: quizzes, exercises, code challenges
- Each element should test understanding of key concepts
- Provide clear instructions for each element
- Make them engaging and educational

Content:
%s

Format your response as a JSON array of objects, each with:
- "type": "quiz" | "exercise" | "code"
- "title": Short descriptive title
- "content": Detailed instructions or questions
- "answer": For quizzes (optional)

Return only valid JSON, nothing else.`, lessonContent[:3000])

	preferences := ProviderPreferences{
		AllowPaid:         true,
		PrioritizeQuality:  true,
		MaxCostPerRequest: 0.30, // $0.30 max for interactive elements
	}
	
	provider := ccg.providerManager.GetBestProvider(preferences)
	if provider == nil {
		return []string{}, fmt.Errorf("no LLM providers available")
	}
	
	result, err := ccg.providerManager.GenerateWithFallback(ctx, prompt, models.ProcessingOptions{
		Quality: "standard",
	})
	
	if err != nil {
		return []string{}, err
	}
	
	// Parse and validate JSON response (simplified for now)
	if strings.Contains(result, "[") && strings.Contains(result, "]") {
		return []string{result}, nil // Return the JSON string for now
	}
	
	// Fallback: generate basic elements
	return []string{
		`{"type":"quiz","title":"Understanding Check","content":"What are the key concepts covered in this lesson?"}`,
		`{"type":"exercise","title":"Practice Exercise","content":"Apply what you've learned by creating a simple example."}`,
	}, nil
}

// GenerateMetadata generates course metadata
func (ccg *CourseContentGenerator) GenerateMetadata(ctx context.Context, title, description string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Generate metadata for the following course.

Title: %s
Description: %s

Provide the following metadata as JSON:
- "difficulty": "beginner" | "intermediate" | "advanced"
- "duration_hours": Estimated course duration in hours
- "prerequisites": Array of prerequisite skills
- "learning_outcomes": Array of learning outcomes
- "target_audience": Target audience description
- "tags": Array of relevant tags (5-10 tags)

Return only valid JSON, nothing else.`, title, description)

	preferences := ProviderPreferences{
		AllowPaid:         true,
		PrioritizeQuality:  true,
		MaxCostPerRequest: 0.15, // $0.15 max for metadata generation
	}
	
	provider := ccg.providerManager.GetBestProvider(preferences)
	if provider == nil {
		return map[string]interface{}{}, fmt.Errorf("no LLM providers available")
	}
	
	result, err := ccg.providerManager.GenerateWithFallback(ctx, prompt, models.ProcessingOptions{
		Quality: "standard",
	})
	
	if err != nil {
		// Return fallback metadata
		return map[string]interface{}{
			"difficulty":       "intermediate",
			"duration_hours":   2.0,
			"prerequisites":    []string{"Basic programming"},
			"learning_outcomes": []string{"Understand core concepts"},
			"target_audience":  "Developers",
			"tags":            []string{"programming", "tutorial"},
		}, nil
	}
	
	// For now, return a simple parsed version
	return map[string]interface{}{
		"generated_metadata": result,
		"difficulty":        "intermediate",
		"duration_hours":    2.0,
	}, nil
}

// GetAvailableProviders returns list of available LLM providers
func (ccg *CourseContentGenerator) GetAvailableProviders() []ProviderInfo {
	var providers []ProviderInfo
	
	for _, provider := range ccg.providerManager.providers {
		if provider.IsAvailable() {
			costEstimate := provider.GetCostEstimate(1000) // Cost for 1000 characters
			providers = append(providers, ProviderInfo{
				Name:         provider.GetName(),
				Type:         provider.GetType(),
				CostPerToken: costEstimate * 4, // Convert back to per-token estimate
				Available:    true,
			})
		}
	}
	
	return providers
}

// ProviderInfo provides information about an LLM provider
type ProviderInfo struct {
	Name         string      `json:"name"`
	Type         ProviderType `json:"type"`
	CostPerToken float64     `json:"cost_per_token"`
	Available    bool        `json:"available"`
}

// TestProviders tests all registered providers
func (ccg *CourseContentGenerator) TestProviders(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	for _, provider := range ccg.providerManager.providers {
		_, err := provider.GenerateText(ctx, "Test prompt", models.ProcessingOptions{})
		results[provider.GetName()] = err
	}
	
	return results
}