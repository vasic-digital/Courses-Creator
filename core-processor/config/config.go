package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Server ServerConfig
	
	// Database settings
	Database DatabaseConfig
	
	// Storage settings
	Storage map[string]StorageConfig
	
	// LLM settings
	LLM LLMConfig
	
	// TTS settings
	TTS TTSConfig
	
	// Security settings
	Security SecurityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string        `env:"HOST" default:"localhost"`
	Port         string        `env:"PORT" default:"8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" default:"120s"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type            string        `env:"DB_TYPE" default:"sqlite"`
	Path            string        `env:"DB_PATH" default:"./data/course_creator.db"`
	MaxConnections   int           `env:"DB_MAX_CONNECTIONS" default:"10"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" default:"1h"`
	Debug           bool          `env:"DB_DEBUG" default:"false"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	Type      string                 `env:"STORAGE_TYPE" default:"local"`
	BasePath  string                 `env:"STORAGE_BASE_PATH" default:"./storage"`
	PublicURL string                 `env:"STORAGE_PUBLIC_URL" default:"http://localhost:8080/storage"`
	Settings  map[string]interface{} // Provider-specific settings
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	// OpenAI settings
	OpenAI OpenAIConfig
	
	// Anthropic settings
	Anthropic AnthropicConfig
	
	// Ollama settings
	Ollama OllamaConfig
	
	// General settings
	DefaultProvider  string  `env:"LLM_DEFAULT_PROVIDER" default:"ollama"`
	MaxCostPerRequest float64 `env:"LLM_MAX_COST_PER_REQUEST" default:"1.00"`
	PrioritizeQuality bool    `env:"LLM_PRIORITIZE_QUALITY" default:"true"`
	AllowPaid        bool    `env:"LLM_ALLOW_PAID" default:"true"`
}

// OpenAIConfig holds OpenAI-specific settings
type OpenAIConfig struct {
	APIKey      string  `env:"OPENAI_API_KEY"`
	BaseURL     string  `env:"OPENAI_BASE_URL" default:"https://api.openai.com/v1"`
	DefaultModel string  `env:"OPENAI_DEFAULT_MODEL" default:"gpt-4"`
	MaxTokens   int     `env:"OPENAI_MAX_TOKENS" default:"4096"`
	Temperature  float64 `env:"OPENAI_TEMPERATURE" default:"0.7"`
	Timeout     time.Duration `env:"OPENAI_TIMEOUT" default:"30s"`
}

// AnthropicConfig holds Anthropic-specific settings
type AnthropicConfig struct {
	APIKey      string        `env:"ANTHROPIC_API_KEY"`
	BaseURL     string        `env:"ANTHROPIC_BASE_URL" default:"https://api.anthropic.com"`
	DefaultModel string        `env:"ANTHROPIC_DEFAULT_MODEL" default:"claude-3-sonnet-20240229"`
	MaxTokens   int           `env:"ANTHROPIC_MAX_TOKENS" default:"4096"`
	Timeout     time.Duration `env:"ANTHROPIC_TIMEOUT" default:"30s"`
}

// OllamaConfig holds Ollama-specific settings
type OllamaConfig struct {
	BaseURL     string        `env:"OLLAMA_BASE_URL" default:"http://localhost:11434"`
	DefaultModel string        `env:"OLLAMA_DEFAULT_MODEL" default:"llama2"`
	Timeout     time.Duration `env:"OLLAMA_TIMEOUT" default:"60s"`
}

// TTSConfig holds TTS configuration
type TTSConfig struct {
	Provider     string        `env:"TTS_PROVIDER" default:"bark"`
	BarkURL      string        `env:"TTS_BARK_URL" default:"http://localhost:8000"`
	SpeechT5URL string        `env:"TTS_SPEECHT5_URL" default:"http://localhost:8001"`
	Timeout      time.Duration `env:"TTS_TIMEOUT" default:"60s"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret      string        `env:"JWT_SECRET" default:"your-secret-key"`
	JWTExpiration  time.Duration `env:"JWT_EXPIRATION" default:"24h"`
	EnableAuth     bool          `env:"ENABLE_AUTH" default:"false"`
	EnableRateLimit bool          `env:"ENABLE_RATE_LIMIT" default:"true"`
	RateLimitRPM  int           `env:"RATE_LIMIT_RPM" default:"60"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("HOST", "localhost"),
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Type:            getEnv("DB_TYPE", "sqlite"),
			Path:            getEnv("DB_PATH", "./data/course_creator.db"),
			MaxConnections:   getIntEnv("DB_MAX_CONNECTIONS", 10),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", time.Hour),
			Debug:           getBoolEnv("DB_DEBUG", false),
		},
		Storage: map[string]StorageConfig{
			"default": {
				Type:      getEnv("STORAGE_TYPE", "local"),
				BasePath:  getEnv("STORAGE_BASE_PATH", "./storage"),
				PublicURL: getEnv("STORAGE_PUBLIC_URL", "http://localhost:8080/storage"),
				Settings:  make(map[string]interface{}),
			},
		},
		LLM: LLMConfig{
			OpenAI: OpenAIConfig{
				APIKey:      getEnv("OPENAI_API_KEY", ""),
				BaseURL:     getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
				DefaultModel: getEnv("OPENAI_DEFAULT_MODEL", "gpt-4"),
				MaxTokens:   getIntEnv("OPENAI_MAX_TOKENS", 4096),
				Temperature:  getFloatEnv("OPENAI_TEMPERATURE", 0.7),
				Timeout:     getDurationEnv("OPENAI_TIMEOUT", 30*time.Second),
			},
			Anthropic: AnthropicConfig{
				APIKey:      getEnv("ANTHROPIC_API_KEY", ""),
				BaseURL:     getEnv("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
				DefaultModel: getEnv("ANTHROPIC_DEFAULT_MODEL", "claude-3-sonnet-20240229"),
				MaxTokens:   getIntEnv("ANTHROPIC_MAX_TOKENS", 4096),
				Timeout:     getDurationEnv("ANTHROPIC_TIMEOUT", 30*time.Second),
			},
			Ollama: OllamaConfig{
				BaseURL:     getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
				DefaultModel: getEnv("OLLAMA_DEFAULT_MODEL", "llama2"),
				Timeout:     getDurationEnv("OLLAMA_TIMEOUT", 60*time.Second),
			},
			DefaultProvider:   getEnv("LLM_DEFAULT_PROVIDER", "ollama"),
			MaxCostPerRequest: getFloatEnv("LLM_MAX_COST_PER_REQUEST", 1.0),
			PrioritizeQuality: getBoolEnv("LLM_PRIORITIZE_QUALITY", true),
			AllowPaid:        getBoolEnv("LLM_ALLOW_PAID", true),
		},
		TTS: TTSConfig{
			Provider:     getEnv("TTS_PROVIDER", "bark"),
			BarkURL:      getEnv("TTS_BARK_URL", "http://localhost:8000"),
			SpeechT5URL: getEnv("TTS_SPEECHT5_URL", "http://localhost:8001"),
			Timeout:      getDurationEnv("TTS_TIMEOUT", 60*time.Second),
		},
		Security: SecurityConfig{
			JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
			JWTExpiration:  getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			EnableAuth:     getBoolEnv("ENABLE_AUTH", false),
			EnableRateLimit: getBoolEnv("ENABLE_RATE_LIMIT", true),
			RateLimitRPM:  getIntEnv("RATE_LIMIT_RPM", 60),
		},
	}
	
	return config, nil
}

// Helper functions for reading environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}