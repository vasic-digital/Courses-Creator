package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Clear environment variables to test defaults
	clearEnvVars()

	config, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test server defaults
	assert.Equal(t, "localhost", config.Server.Host)
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, config.Server.IdleTimeout)

	// Test database defaults
	assert.Equal(t, "sqlite", config.Database.Type)
	assert.Equal(t, "./data/course_creator.db", config.Database.Path)
	assert.Equal(t, 10, config.Database.MaxConnections)
	assert.Equal(t, time.Hour, config.Database.ConnMaxLifetime)
	assert.False(t, config.Database.Debug)

	// Test storage defaults
	assert.Contains(t, config.Storage, "default")
	storage := config.Storage["default"]
	assert.Equal(t, "local", storage.Type)
	assert.Equal(t, "./storage", storage.BasePath)
	assert.Equal(t, "http://localhost:8080/storage", storage.PublicURL)

	// Test LLM defaults
	assert.Equal(t, "ollama", config.LLM.DefaultProvider)
	assert.Equal(t, 1.0, config.LLM.MaxCostPerRequest)
	assert.True(t, config.LLM.PrioritizeQuality)
	assert.True(t, config.LLM.AllowPaid)

	// Test OpenAI defaults
	assert.Equal(t, "", config.LLM.OpenAI.APIKey)
	assert.Equal(t, "https://api.openai.com/v1", config.LLM.OpenAI.BaseURL)
	assert.Equal(t, "gpt-4", config.LLM.OpenAI.DefaultModel)
	assert.Equal(t, 4096, config.LLM.OpenAI.MaxTokens)
	assert.Equal(t, 0.7, config.LLM.OpenAI.Temperature)
	assert.Equal(t, 30*time.Second, config.LLM.OpenAI.Timeout)

	// Test Anthropic defaults
	assert.Equal(t, "", config.LLM.Anthropic.APIKey)
	assert.Equal(t, "https://api.anthropic.com", config.LLM.Anthropic.BaseURL)
	assert.Equal(t, "claude-3-sonnet-20240229", config.LLM.Anthropic.DefaultModel)
	assert.Equal(t, 4096, config.LLM.Anthropic.MaxTokens)
	assert.Equal(t, 30*time.Second, config.LLM.Anthropic.Timeout)

	// Test Ollama defaults
	assert.Equal(t, "http://localhost:11434", config.LLM.Ollama.BaseURL)
	assert.Equal(t, "llama2", config.LLM.Ollama.DefaultModel)
	assert.Equal(t, 60*time.Second, config.LLM.Ollama.Timeout)

	// Test TTS defaults
	assert.Equal(t, "bark", config.TTS.Provider)
	assert.Equal(t, "http://localhost:8000", config.TTS.BarkURL)
	assert.Equal(t, "http://localhost:8001", config.TTS.SpeechT5URL)
	assert.Equal(t, 60*time.Second, config.TTS.Timeout)

	// Test security defaults
	assert.Equal(t, "change-me-in-production", config.Security.JWTSecret)
	assert.Equal(t, 24*time.Hour, config.Security.JWTExpiration)
	assert.False(t, config.Security.EnableAuth)
	assert.True(t, config.Security.EnableRateLimit)
	assert.Equal(t, 60, config.Security.RateLimitRPM)
}

func TestLoadConfig_WithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("PORT", "3000")
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("ENABLE_AUTH", "true")
	defer clearEnvVars()

	config, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test overridden values
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, "3000", config.Server.Port)
	assert.Equal(t, "postgres", config.Database.Type)
	assert.Equal(t, "test-key", config.LLM.OpenAI.APIKey)
	assert.Equal(t, "test-secret", config.Security.JWTSecret)
	assert.True(t, config.Security.EnableAuth)
}

func TestGetEnv(t *testing.T) {
	// Test with value set
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")
	assert.Equal(t, "test-value", getEnv("TEST_VAR", "default"))

	// Test with empty value (should return default)
	os.Setenv("TEST_VAR", "")
	assert.Equal(t, "default", getEnv("TEST_VAR", "default"))

	// Test with unset variable
	os.Unsetenv("TEST_VAR")
	assert.Equal(t, "default", getEnv("TEST_VAR", "default"))
}

func TestGetIntEnv(t *testing.T) {
	// Test with valid integer
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")
	assert.Equal(t, 42, getIntEnv("TEST_INT", 10))

	// Test with invalid integer (should return default)
	os.Setenv("TEST_INT", "invalid")
	assert.Equal(t, 10, getIntEnv("TEST_INT", 10))

	// Test with empty value (should return default)
	os.Setenv("TEST_INT", "")
	assert.Equal(t, 10, getIntEnv("TEST_INT", 10))

	// Test with unset variable
	os.Unsetenv("TEST_INT")
	assert.Equal(t, 10, getIntEnv("TEST_INT", 10))
}

func TestGetFloatEnv(t *testing.T) {
	// Test with valid float
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")
	assert.Equal(t, 3.14, getFloatEnv("TEST_FLOAT", 1.0))

	// Test with invalid float (should return default)
	os.Setenv("TEST_FLOAT", "invalid")
	assert.Equal(t, 1.0, getFloatEnv("TEST_FLOAT", 1.0))

	// Test with empty value (should return default)
	os.Setenv("TEST_FLOAT", "")
	assert.Equal(t, 1.0, getFloatEnv("TEST_FLOAT", 1.0))

	// Test with unset variable
	os.Unsetenv("TEST_FLOAT")
	assert.Equal(t, 1.0, getFloatEnv("TEST_FLOAT", 1.0))
}

func TestGetBoolEnv(t *testing.T) {
	// Test with true values (only those supported by strconv.ParseBool)
	for _, val := range []string{"true", "TRUE", "1"} {
		os.Setenv("TEST_BOOL", val)
		assert.True(t, getBoolEnv("TEST_BOOL", false), "Should parse %s as true", val)
	}

	// Test with false values (only those supported by strconv.ParseBool)
	for _, val := range []string{"false", "FALSE", "0"} {
		os.Setenv("TEST_BOOL", val)
		assert.False(t, getBoolEnv("TEST_BOOL", true), "Should parse %s as false", val)
	}

	// Test with invalid boolean (should return default)
	os.Setenv("TEST_BOOL", "invalid")
	assert.True(t, getBoolEnv("TEST_BOOL", true))

	// Test with empty value (should return default)
	os.Setenv("TEST_BOOL", "")
	assert.True(t, getBoolEnv("TEST_BOOL", true))

	// Test with unset variable
	os.Unsetenv("TEST_BOOL")
	assert.True(t, getBoolEnv("TEST_BOOL", true))
}

func TestGetDurationEnv(t *testing.T) {
	// Test with valid duration
	os.Setenv("TEST_DURATION", "5m30s")
	defer os.Unsetenv("TEST_DURATION")
	assert.Equal(t, 5*time.Minute+30*time.Second, getDurationEnv("TEST_DURATION", time.Minute))

	// Test with invalid duration (should return default)
	os.Setenv("TEST_DURATION", "invalid")
	assert.Equal(t, time.Minute, getDurationEnv("TEST_DURATION", time.Minute))

	// Test with empty value (should return default)
	os.Setenv("TEST_DURATION", "")
	assert.Equal(t, time.Minute, getDurationEnv("TEST_DURATION", time.Minute))

	// Test with unset variable
	os.Unsetenv("TEST_DURATION")
	assert.Equal(t, time.Minute, getDurationEnv("TEST_DURATION", time.Minute))
}

func TestLoadConfig_ReadTimeout(t *testing.T) {
	os.Setenv("READ_TIMEOUT", "45s")
	defer os.Unsetenv("READ_TIMEOUT")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 45*time.Second, config.Server.ReadTimeout)
}

func TestLoadConfig_WriteTimeout(t *testing.T) {
	os.Setenv("WRITE_TIMEOUT", "60s")
	defer os.Unsetenv("WRITE_TIMEOUT")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, config.Server.WriteTimeout)
}

func TestLoadConfig_IdleTimeout(t *testing.T) {
	os.Setenv("IDLE_TIMEOUT", "180s")
	defer os.Unsetenv("IDLE_TIMEOUT")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 180*time.Second, config.Server.IdleTimeout)
}

func TestLoadConfig_DBMaxConnections(t *testing.T) {
	os.Setenv("DB_MAX_CONNECTIONS", "20")
	defer os.Unsetenv("DB_MAX_CONNECTIONS")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 20, config.Database.MaxConnections)
}

func TestLoadConfig_DBConnMaxLifetime(t *testing.T) {
	os.Setenv("DB_CONN_MAX_LIFETIME", "2h30m")
	defer os.Unsetenv("DB_CONN_MAX_LIFETIME")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 2*time.Hour+30*time.Minute, config.Database.ConnMaxLifetime)
}

func TestLoadConfig_DBDebug(t *testing.T) {
	os.Setenv("DB_DEBUG", "true")
	defer os.Unsetenv("DB_DEBUG")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.True(t, config.Database.Debug)
}

func TestLoadConfig_OpenAIMaxTokens(t *testing.T) {
	os.Setenv("OPENAI_MAX_TOKENS", "8192")
	defer os.Unsetenv("OPENAI_MAX_TOKENS")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 8192, config.LLM.OpenAI.MaxTokens)
}

func TestLoadConfig_OpenAITemperature(t *testing.T) {
	os.Setenv("OPENAI_TEMPERATURE", "0.5")
	defer os.Unsetenv("OPENAI_TEMPERATURE")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 0.5, config.LLM.OpenAI.Temperature)
}

func TestLoadConfig_OpenAITimeout(t *testing.T) {
	os.Setenv("OPENAI_TIMEOUT", "45s")
	defer os.Unsetenv("OPENAI_TIMEOUT")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 45*time.Second, config.LLM.OpenAI.Timeout)
}

func TestLoadConfig_AnthropicMaxTokens(t *testing.T) {
	os.Setenv("ANTHROPIC_MAX_TOKENS", "8192")
	defer os.Unsetenv("ANTHROPIC_MAX_TOKENS")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 8192, config.LLM.Anthropic.MaxTokens)
}

func TestLoadConfig_AnthropicTimeout(t *testing.T) {
	os.Setenv("ANTHROPIC_TIMEOUT", "45s")
	defer os.Unsetenv("ANTHROPIC_TIMEOUT")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 45*time.Second, config.LLM.Anthropic.Timeout)
}

func TestLoadConfig_OllamaTimeout(t *testing.T) {
	os.Setenv("OLLAMA_TIMEOUT", "90s")
	defer os.Unsetenv("OLLAMA_TIMEOUT")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 90*time.Second, config.LLM.Ollama.Timeout)
}

func TestLoadConfig_LLMMaxCostPerRequest(t *testing.T) {
	os.Setenv("LLM_MAX_COST_PER_REQUEST", "2.50")
	defer os.Unsetenv("LLM_MAX_COST_PER_REQUEST")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 2.50, config.LLM.MaxCostPerRequest)
}

func TestLoadConfig_LLMPrioritizeQuality(t *testing.T) {
	os.Setenv("LLM_PRIORITIZE_QUALITY", "false")
	defer os.Unsetenv("LLM_PRIORITIZE_QUALITY")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.False(t, config.LLM.PrioritizeQuality)
}

func TestLoadConfig_LLMAllowPaid(t *testing.T) {
	os.Setenv("LLM_ALLOW_PAID", "false")
	defer os.Unsetenv("LLM_ALLOW_PAID")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.False(t, config.LLM.AllowPaid)
}

func TestLoadConfig_TTSTimeout(t *testing.T) {
	os.Setenv("TTS_TIMEOUT", "90s")
	defer os.Unsetenv("TTS_TIMEOUT")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 90*time.Second, config.TTS.Timeout)
}

func TestLoadConfig_JWTExpiration(t *testing.T) {
	os.Setenv("JWT_EXPIRATION", "48h")
	defer os.Unsetenv("JWT_EXPIRATION")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 48*time.Hour, config.Security.JWTExpiration)
}

func TestLoadConfig_RateLimitRPM(t *testing.T) {
	os.Setenv("RATE_LIMIT_RPM", "100")
	defer os.Unsetenv("RATE_LIMIT_RPM")

	config, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 100, config.Security.RateLimitRPM)
}

// Helper function to clear environment variables used in tests
func clearEnvVars() {
	envVars := []string{
		"HOST", "PORT", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT",
		"DB_TYPE", "DB_PATH", "DB_MAX_CONNECTIONS", "DB_CONN_MAX_LIFETIME", "DB_DEBUG",
		"STORAGE_TYPE", "STORAGE_BASE_PATH", "STORAGE_PUBLIC_URL",
		"OPENAI_API_KEY", "OPENAI_BASE_URL", "OPENAI_DEFAULT_MODEL", "OPENAI_MAX_TOKENS", "OPENAI_TEMPERATURE", "OPENAI_TIMEOUT",
		"ANTHROPIC_API_KEY", "ANTHROPIC_BASE_URL", "ANTHROPIC_DEFAULT_MODEL", "ANTHROPIC_MAX_TOKENS", "ANTHROPIC_TIMEOUT",
		"OLLAMA_BASE_URL", "OLLAMA_DEFAULT_MODEL", "OLLAMA_TIMEOUT",
		"LLM_DEFAULT_PROVIDER", "LLM_MAX_COST_PER_REQUEST", "LLM_PRIORITIZE_QUALITY", "LLM_ALLOW_PAID",
		"TTS_PROVIDER", "TTS_BARK_URL", "TTS_SPEECHT5_URL", "TTS_TIMEOUT",
		"JWT_SECRET", "JWT_EXPIRATION", "ENABLE_AUTH", "ENABLE_RATE_LIMIT", "RATE_LIMIT_RPM",
		"TEST_VAR", "TEST_INT", "TEST_FLOAT", "TEST_BOOL", "TEST_DURATION",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
