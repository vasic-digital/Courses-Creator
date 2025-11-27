package main

import (
	"fmt"
	"log"

	"github.com/course-creator/core-processor/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("LLM Default Provider: %s\n", cfg.LLM.DefaultProvider)
	fmt.Printf("OpenAI API Key: %s\n", maskKey(cfg.LLM.OpenAI.APIKey))
	fmt.Printf("Anthropic API Key: %s\n", maskKey(cfg.LLM.Anthropic.APIKey))
	fmt.Printf("Ollama Base URL: %s\n", cfg.LLM.Ollama.BaseURL)
	fmt.Printf("Ollama Model: %s\n", cfg.LLM.Ollama.DefaultModel)
	fmt.Printf("Allow Paid: %v\n", cfg.LLM.AllowPaid)
	fmt.Printf("Max Cost per Request: $%.2f\n", cfg.LLM.MaxCostPerRequest)
}

func maskKey(key string) string {
	if key == "" {
		return "[not set]"
	}
	if len(key) <= 8 {
		return "[masked]"
	}
	return key[:4] + "[...]" + key[len(key)-4:]
}