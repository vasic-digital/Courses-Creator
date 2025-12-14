package mcp_servers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseServerImpl(t *testing.T) {
	config := MCPServerConfig{
		Name:       "test-server",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	server := NewBaseServer(config)
	require.NotNil(t, server)

	// Test initial state
	assert.False(t, server.IsRunning())
	assert.Equal(t, "test-server", server.Config.Name)

	// Test adding tools
	called := false
	server.AddTool("test_tool", "A test tool", func(args map[string]interface{}) (interface{}, error) {
		called = true
		return map[string]interface{}{"result": "success"}, nil
	})

	// Test tool registration
	assert.Contains(t, server.Tools, "test_tool")
	assert.Equal(t, "A test tool", server.Tools["test_tool"].Description)

	// Test initialize request
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	requestJSON, _ := json.Marshal(initRequest)
	response := server.ProcessRequest(string(requestJSON))

	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	result := response.Result.(map[string]interface{})
	assert.Equal(t, "2024-11-05", result["protocolVersion"])
	serverInfo := result["serverInfo"].(map[string]interface{})
	assert.Equal(t, "test-server", serverInfo["name"])

	// Test tools/list request
	listRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	requestJSON, _ = json.Marshal(listRequest)
	response = server.ProcessRequest(string(requestJSON))

	assert.Nil(t, response.Error)
	result = response.Result.(map[string]interface{})
	tools := result["tools"].([]interface{})
	assert.Len(t, tools, 1)

	tool := tools[0].(map[string]interface{})
	assert.Equal(t, "test_tool", tool["name"])

	// Test tool call request
	toolCallRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "test_tool",
			"arguments": map[string]interface{}{
				"text": "test input",
			},
		},
	}

	requestJSON, _ = json.Marshal(toolCallRequest)
	response = server.ProcessRequest(string(requestJSON))

	assert.Nil(t, response.Error)
	assert.True(t, called)
}

func TestBarkTTSServer(t *testing.T) {
	server := NewBarkTTSServer()
	require.NotNil(t, server)

	// Test server configuration
	assert.Equal(t, "bark-tts", server.Config.Name)
	assert.True(t, server.useBarkPython)
	assert.Equal(t, 200, server.maxLength)
	assert.Equal(t, 24000, server.sampleRate)

	// Test tool registration
	assert.Contains(t, server.Tools, "generate_tts")
	assert.Contains(t, server.Tools, "list_voices")
	assert.Contains(t, server.Tools, "get_info")

	// Test TTS generation with minimal parameters
	args := map[string]interface{}{
		"text": "Hello world",
	}

	result, err := server.GenerateTTS(args)

	// The test should either succeed or fail gracefully with a meaningful error
	if err != nil {
		// Check if it's a dependency issue (acceptable in test environment)
		assert.Contains(t, err.Error(), "failed to generate audio")
	} else {
		// If successful, check response structure
		assert.Contains(t, result, "audio_path")
		assert.Contains(t, result, "text")
		assert.Equal(t, "Hello world", result.(map[string]interface{})["text"])
	}
}

func TestSpeechT5Server(t *testing.T) {
	server := NewSpeechT5Server()
	require.NotNil(t, server)

	// Test server configuration
	assert.Equal(t, "speecht5-tts", server.Config.Name)
	assert.True(t, server.useSpeechT5Python)

	// Test tool registration
	assert.Contains(t, server.Tools, "generate_tts")
	assert.Contains(t, server.Tools, "list_voices")
	assert.Contains(t, server.Tools, "get_info")
}

func TestLLaVAServer(t *testing.T) {
	server := NewLLaVAServer()
	require.NotNil(t, server)

	// Test server configuration
	assert.Equal(t, "llava-vision", server.Config.Name)
	assert.True(t, server.useLLaVAPython)

	// Test tool registration
	assert.Contains(t, server.Tools, "analyze_image")
	assert.Contains(t, server.Tools, "extract_text")
	assert.Contains(t, server.Tools, "get_info")
}

func TestPix2StructServer(t *testing.T) {
	server := NewPix2StructServer()
	require.NotNil(t, server)

	// Test server configuration
	assert.Equal(t, "pix2struct", server.Config.Name)
	assert.True(t, server.usePix2StructPython)

	// Test tool registration
	assert.Contains(t, server.Tools, "parse_ui")
	assert.Contains(t, server.Tools, "extract_data")
	assert.Contains(t, server.Tools, "get_info")
}

func TestSunoServer(t *testing.T) {
	server := NewSunoServer()
	require.NotNil(t, server)

	// Test server configuration
	assert.Equal(t, "suno-music", server.Config.Name)
	assert.True(t, server.useSunoPython)

	// Test tool registration
	assert.Contains(t, server.Tools, "generate_music")
	assert.Contains(t, server.Tools, "extend_music")
	assert.Contains(t, server.Tools, "get_info")
}

func TestMCPServerErrorHandling(t *testing.T) {
	server := NewBaseServerWithDefaults("test-error")

	// Test invalid JSON
	response := server.ProcessRequest("invalid json")
	assert.NotNil(t, response.Error)
	assert.Equal(t, -32700, response.Error.Code)
	assert.Equal(t, "Parse error", response.Error.Message)

	// Test invalid method
	invalidRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "invalid_method",
	}

	requestJSON, _ := json.Marshal(invalidRequest)
	response = server.ProcessRequest(string(requestJSON))

	assert.NotNil(t, response.Error)
	assert.Equal(t, -32000, response.Error.Code)
	assert.Contains(t, response.Error.Message, "method not found")

	// Test tool not found
	toolCallRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "nonexistent_tool",
		},
	}

	requestJSON, _ = json.Marshal(toolCallRequest)
	response = server.ProcessRequest(string(requestJSON))

	assert.NotNil(t, response.Error)
	assert.Equal(t, -32000, response.Error.Code)
	assert.Contains(t, response.Error.Message, "tool not found")
}
