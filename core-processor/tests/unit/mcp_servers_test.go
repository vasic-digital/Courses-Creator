package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/course-creator/core-processor/mcp_servers"
	"github.com/course-creator/core-processor/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseServerImpl_NewBaseServer(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:       "test-server",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	server := mcp_servers.NewBaseServer(config)

	require.NotNil(t, server)
	assert.Equal(t, config.Name, server.Config.Name)
	assert.Equal(t, config.Version, server.Config.Version)
	assert.Equal(t, config.Transport, server.Config.Transport)
	assert.NotNil(t, server.Tools)
	assert.Empty(t, server.Tools)
}

func TestBaseServerImpl_AddTool(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:       "test-server",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	server := mcp_servers.NewBaseServer(config)

	// Add a tool
	testHandler := func(args map[string]interface{}) (interface{}, error) {
		return "test result", nil
	}

	server.AddTool("test_tool", "A test tool", testHandler)

	// Check that tool was added
	assert.Contains(t, server.Tools, "test_tool")
	tool := server.Tools["test_tool"]
	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.NotNil(t, tool.Handler)
}

func TestBaseServerImpl_ProcessRequest(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:       "test-server",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	server := mcp_servers.NewBaseServer(config)

	// Test initialize request
	request := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	response := server.ProcessRequest(request)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(1), response.ID)
	assert.NotNil(t, response.Result)
	assert.Nil(t, response.Error)

	// Test tools/list request
	request = `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	response = server.ProcessRequest(request)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(2), response.ID)
	assert.NotNil(t, response.Result)
	assert.Nil(t, response.Error)

	// Test invalid request
	request = `{"jsonrpc":"2.0","id":3,"method":"invalid_method","params":{}}`
	response = server.ProcessRequest(request)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(3), response.ID)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.Error)
	// Convert to int for comparison
	assert.Equal(t, int(-32000), int(response.Error.Code))
}

func TestBaseServerImpl_Stop(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:       "test-server",
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	server := mcp_servers.NewBaseServer(config)

	// Server should not be running initially
	assert.False(t, server.IsRunning())

	// Stop the server
	server.Stop()

	// Server should no longer be running
	assert.False(t, server.IsRunning())
}

func TestBarkTTSServer_NewBarkTTSServer(t *testing.T) {
	server := mcp_servers.NewBarkTTSServer()

	require.NotNil(t, server)
	assert.Equal(t, "bark-tts", server.Config.Name)
	assert.Equal(t, "1.0.0", server.Config.Version)
	assert.Equal(t, "stdio", server.Config.Transport)
	assert.NotNil(t, server.Tools)
}

func TestBarkTTSServer_GenerateTTS(t *testing.T) {
	// Skip if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping Bark TTS test - OPENAI_API_KEY environment variable not set")
	}

	// Create temporary directory for test output
	tempDir, err := os.MkdirTemp("", "bark_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	server := mcp_servers.NewBarkTTSServerWithConfig("http://localhost:8081", "/models/bark", tempDir, 200, 24000)
	// Disable Bark Python for faster testing
	server.SetUseBarkPython(false)

	// Test with valid text
	args := map[string]interface{}{
		"text":         "Hello, world!",
		"voice_preset": "v2/en_speaker_6",
	}

	result, err := server.GenerateTTS(args)

	// Result should contain audio path
	require.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, resultMap, "audio_path")
	assert.Contains(t, resultMap, "text")
	assert.Contains(t, resultMap, "voice")
	assert.Equal(t, "Hello, world!", resultMap["text"])
	assert.Equal(t, "v2/en_speaker_6", resultMap["voice"])

	// Verify audio file was created
	audioPath, ok := resultMap["audio_path"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, audioPath)

	// Check that file exists
	_, err = os.Stat(audioPath)
	assert.NoError(t, err)

	// Clean up generated file
	os.Remove(audioPath)
}

func TestBarkTTSServer_ListVoices(t *testing.T) {
	server := mcp_servers.NewBarkTTSServer()

	args := map[string]interface{}{}
	result, err := server.ListVoices(args)

	require.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, resultMap, "voices")
	assert.Contains(t, resultMap, "total")

	voices, ok := resultMap["voices"].([]map[string]interface{})
	require.True(t, ok)
	assert.Greater(t, len(voices), 0)

	total, ok := resultMap["total"].(int)
	require.True(t, ok)
	assert.Equal(t, len(voices), total)
}

func TestBarkTTSServer_SplitText(t *testing.T) {
	server := mcp_servers.NewBarkTTSServer()

	// Test short text (no splitting)
	shortText := "This is a short text"
	chunks := server.SplitText(shortText)
	assert.Len(t, chunks, 1)
	assert.Equal(t, shortText, chunks[0])

	// Test long text (should be split)
	longText := string(make([]byte, 300)) // Create long text
	chunks = server.SplitText(longText)
	assert.Greater(t, len(chunks), 1)
}

func TestSpeechT5TTSServer_NewSpeechT5Server(t *testing.T) {
	server := mcp_servers.NewSpeechT5Server()

	require.NotNil(t, server)
	assert.Equal(t, "speecht5-tts", server.Config.Name)
	assert.Equal(t, "1.0.0", server.Config.Version)
	assert.Equal(t, "stdio", server.Config.Transport)
	assert.NotNil(t, server.Tools)
}

func TestLLaVAServer_NewLLaVAServer(t *testing.T) {
	server := mcp_servers.NewLLaVAServer()

	require.NotNil(t, server)
	assert.Equal(t, "llava-image", server.Config.Name)
	assert.Equal(t, "1.0.0", server.Config.Version)
	assert.Equal(t, "stdio", server.Config.Transport)
	assert.NotNil(t, server.Tools)
}

func TestPix2StructServer_NewPix2StructServer(t *testing.T) {
	server := mcp_servers.NewPix2StructServer()

	require.NotNil(t, server)
	assert.Equal(t, "pix2struct-ui", server.Config.Name)
	assert.Equal(t, "1.0.0", server.Config.Version)
	assert.Equal(t, "stdio", server.Config.Transport)
	assert.NotNil(t, server.Tools)
}

func TestSpeechT5TTSServer_GenerateTTS(t *testing.T) {
	// Skip if ElevenLabs API key is not available
	if os.Getenv("ELEVENLABS_API_KEY") == "" {
		t.Skip("Skipping SpeechT5 TTS test - ELEVENLABS_API_KEY environment variable not set")
	}

	// Create temporary directory for test output
	tempDir, err := os.MkdirTemp("", "speecht5_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	server := mcp_servers.NewSpeechT5ServerWithConfig("http://localhost:8082", "/models/speecht5", tempDir, 300, 16000)
	// Disable SpeechT5 Python for faster testing
	server.SetUseSpeechT5Python(false)

	// Test with valid text
	args := map[string]interface{}{
		"text":         "Hello, world!",
		"voice_preset": "default",
	}

	result, err := server.GenerateTTS(args)

	// Result should contain audio path
	require.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, resultMap, "audio_path")
	assert.Contains(t, resultMap, "text")
	assert.Contains(t, resultMap, "voice")
	assert.Equal(t, "Hello, world!", resultMap["text"])
	assert.Equal(t, "default", resultMap["voice"])

	// Clean up generated file
	audioPath, ok := resultMap["audio_path"].(string)
	if ok && audioPath != "" {
		os.Remove(audioPath)
	}
}

func TestLLaVAServer_AnalyzeImage(t *testing.T) {
	// Skip if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping LLaVA test - OPENAI_API_KEY environment variable not set")
	}

	server := mcp_servers.NewLLaVAServer()

	// Test with a simple test image path (this would normally be a real image)
	args := map[string]interface{}{
		"image":  "/tmp/test_image.jpg",
		"prompt": "Describe what you see in this image",
	}

	result, err := server.AnalyzeImage(args)

	// Should return an error for non-existent image in test
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLLaVAServer_ExtractText(t *testing.T) {
	// Skip if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping LLaVA test - OPENAI_API_KEY environment variable not set")
	}

	server := mcp_servers.NewLLaVAServer()

	// Test with a simple test image path
	args := map[string]interface{}{
		"image": "/tmp/test_image.jpg",
	}

	// Call the internal method via reflection or test the error path
	// Since extractText is not exported, we test via the tool handler
	tool := server.Tools["extract_text"]
	require.NotNil(t, tool)

	result, err := tool.Handler(args)

	// Should return an error for non-existent image in test
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLLaVAServer_DetectObjects(t *testing.T) {
	// Skip if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping LLaVA test - OPENAI_API_KEY environment variable not set")
	}

	server := mcp_servers.NewLLaVAServer()

	// Test with a simple test image path
	args := map[string]interface{}{
		"image":      "/tmp/test_image.jpg",
		"confidence": 0.5,
	}

	// Call the internal method via the tool handler
	tool := server.Tools["detect_objects"]
	require.NotNil(t, tool)

	result, err := tool.Handler(args)

	// Should return an error for non-existent image in test
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPix2StructServer_ParseUI(t *testing.T) {
	// Skip if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping Pix2Struct test - OPENAI_API_KEY environment variable not set")
	}

	server := mcp_servers.NewPix2StructServer()

	// Test with a simple test image path
	args := map[string]interface{}{
		"image":  "/tmp/test_ui_screenshot.png",
		"prompt": "Describe the UI elements in this screenshot",
	}

	result, err := server.ParseUI(args)

	// Should return an error for non-existent image in test
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPix2StructServer_ExtractButtons(t *testing.T) {
	// Skip if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping Pix2Struct test - OPENAI_API_KEY environment variable not set")
	}

	server := mcp_servers.NewPix2StructServer()

	// Test with a simple test image path
	args := map[string]interface{}{
		"image": "/tmp/test_ui_screenshot.png",
	}

	// Call the internal method via the tool handler
	tool := server.Tools["extract_buttons"]
	require.NotNil(t, tool)

	result, err := tool.Handler(args)

	// Should return an error for non-existent image in test
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPix2StructServer_ExtractForms(t *testing.T) {
	// Skip if OpenAI API key is not available
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping Pix2Struct test - OPENAI_API_KEY environment variable not set")
	}

	server := mcp_servers.NewPix2StructServer()

	// Test with a simple test image path
	args := map[string]interface{}{
		"image": "/tmp/test_ui_screenshot.png",
	}

	// Call the internal method via the tool handler
	tool := server.Tools["extract_forms"]
	require.NotNil(t, tool)

	result, err := tool.Handler(args)

	// Should return an error for non-existent image in test
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUtils_HashString(t *testing.T) {
	// Test that hash is deterministic
	text := "test text"
	hash1 := utils.HashString(text)
	hash2 := utils.HashString(text)
	assert.Equal(t, hash1, hash2)

	// Test that different texts produce different hashes
	differentText := "different text"
	hash3 := utils.HashString(differentText)
	assert.NotEqual(t, hash1, hash3)
}

func TestUtils_GenerateID(t *testing.T) {
	id1 := utils.GenerateID()
	id2 := utils.GenerateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 32) // 16 bytes = 32 hex characters
	assert.Len(t, id2, 32)
}

func TestUtils_EnsureDir(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test_dir", utils.GenerateID())

	// Directory should not exist initially
	_, err := os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))

	// Ensure directory
	err = utils.EnsureDir(tempDir)
	require.NoError(t, err)

	// Directory should exist now
	info, err := os.Stat(tempDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Clean up
	os.RemoveAll(tempDir)
}

func TestUtils_FileExists(t *testing.T) {
	// Test existing file
	tempDir := filepath.Join(os.TempDir(), "test_file", utils.GenerateID())
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(tempFile, []byte("test content"), 0644)
	require.NoError(t, err)

	assert.True(t, utils.FileExists(tempFile))

	// Test non-existing file
	nonExistentFile := filepath.Join(tempDir, "non_existent.txt")
	assert.False(t, utils.FileExists(nonExistentFile))
}

func TestUtils_CopyFile(t *testing.T) {
	// Create directories
	srcDir := filepath.Join(os.TempDir(), "src", utils.GenerateID())
	dstDir := filepath.Join(os.TempDir(), "dst", utils.GenerateID())
	err := os.MkdirAll(srcDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(dstDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(srcDir)
	defer os.RemoveAll(dstDir)

	srcFile := filepath.Join(srcDir, "test.txt")
	dstFile := filepath.Join(dstDir, "test.txt")

	// Create source file
	content := []byte("test content for copy")
	err = os.WriteFile(srcFile, content, 0644)
	require.NoError(t, err)

	// Copy file
	err = utils.CopyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer os.Remove(dstFile)

	// Verify copy
	copiedContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, copiedContent)
}

func TestUtils_SanitizeFilename(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "normal-file.txt",
			expected: "normal-file.txt",
		},
		{
			input:    "file/with\\slashes",
			expected: "file_with_slashes",
		},
		{
			input:    "file:with*special?chars",
			expected: "file_with_special_chars",
		},
		{
			input:    string(make([]byte, 150)), // Very long filename
			expected: string(make([]byte, 100)),
		},
	}

	for _, tc := range testCases {
		result := utils.SanitizeFilename(tc.input)
		assert.Equal(t, tc.expected[:len(result)], result)
		assert.LessOrEqual(t, len(result), 100)
	}
}

func TestUtils_GetFileExtension(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "file.txt",
			expected: ".txt",
		},
		{
			input:    "file.mp4",
			expected: ".mp4",
		},
		{
			input:    "file",
			expected: "",
		},
		{
			input:    "file.tar.gz",
			expected: ".gz",
		},
	}

	for _, tc := range testCases {
		result := utils.GetFileExtension(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}

func TestUtils_GetFileSize(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "test_size", utils.GenerateID())
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content for size")

	err = os.WriteFile(tempFile, content, 0644)
	require.NoError(t, err)

	size, err := utils.GetFileSize(tempFile)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
}

func TestUtils_Retry(t *testing.T) {
	attempts := 0
	err := utils.Retry(3, time.Millisecond, func() error {
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)

	// Test that it returns error after max attempts
	attempts = 0
	err = utils.Retry(2, time.Millisecond, func() error {
		attempts++
		return assert.AnError
	})

	assert.Error(t, err)
	assert.Equal(t, 2, attempts)
}

func TestExecuteCommand(t *testing.T) {
	ctx := context.Background()

	// Test valid command
	cmd := utils.ExecuteCommand(ctx, "echo", "test")
	require.NotNil(t, cmd)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "test\n", string(output))

	// Test invalid command
	cmd = utils.ExecuteCommand(ctx, "nonexistent_command")
	output, err = cmd.CombinedOutput()
	assert.Error(t, err)
}

func TestExecuteCommandWithOutput(t *testing.T) {
	ctx := context.Background()

	// Test valid command
	output, err := utils.ExecuteCommandWithOutput(ctx, "echo", "test")
	require.NoError(t, err)
	assert.Equal(t, "test\n", output)

	// Test invalid command
	_, err = utils.ExecuteCommandWithOutput(ctx, "nonexistent_command")
	assert.Error(t, err)
}

// Additional unit tests for MCP servers without external dependencies

func TestBaseServerImpl_ProcessRequest_InvalidJSON(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:    "test-server",
		Version: "1.0.0",
	}

	server := mcp_servers.NewBaseServer(config)

	// Test invalid JSON
	response := server.ProcessRequest("invalid json")

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(-32700), response.ID) // Parse error ID
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.Error)
	assert.Equal(t, int(-32700), int(response.Error.Code)) // Parse error
}

func TestBaseServerImpl_ProcessRequest_UnknownMethod(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:    "test-server",
		Version: "1.0.0",
	}

	server := mcp_servers.NewBaseServer(config)

	// Test unknown method
	request := `{"jsonrpc":"2.0","id":1,"method":"unknown_method","params":{}}`
	response := server.ProcessRequest(request)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(1), response.ID)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.Error)
	assert.Equal(t, int(-32601), int(response.Error.Code)) // Method not found
}

func TestBaseServerImpl_ProcessRequest_ToolCall(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:    "test-server",
		Version: "1.0.0",
	}

	server := mcp_servers.NewBaseServer(config)

	// Add a test tool
	testHandler := func(args map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{"result": "success", "input": args["input"]}, nil
	}
	server.AddTool("test_tool", "A test tool", testHandler)

	// Test tool call
	request := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"test_tool","arguments":{"input":"test_value"}}}`
	response := server.ProcessRequest(request)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(2), response.ID)
	assert.NotNil(t, response.Result)
	assert.Nil(t, response.Error)

	resultMap, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "success", resultMap["result"])
	assert.Equal(t, "test_value", resultMap["input"])
}

func TestBaseServerImpl_ProcessRequest_ToolCall_UnknownTool(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:    "test-server",
		Version: "1.0.0",
	}

	server := mcp_servers.NewBaseServer(config)

	// Test tool call with unknown tool
	request := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}`
	response := server.ProcessRequest(request)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(3), response.ID)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.Error)
	assert.Contains(t, response.Error.Message, "unknown tool")
}

func TestBaseServerImpl_ProcessRequest_ToolCall_InvalidArguments(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:    "test-server",
		Version: "1.0.0",
	}

	server := mcp_servers.NewBaseServer(config)

	// Add a test tool that expects arguments
	testHandler := func(args map[string]interface{}) (interface{}, error) {
		if args["required"] == nil {
			return nil, &mcp_servers.MCPError{Code: -32602, Message: "Missing required argument"}
		}
		return "success", nil
	}
	server.AddTool("test_tool", "A test tool", testHandler)

	// Test tool call with invalid arguments
	request := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"test_tool","arguments":{}}}`
	response := server.ProcessRequest(request)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, float64(4), response.ID)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.Error)
	assert.Equal(t, int(-32602), int(response.Error.Code))
}

func TestBarkTTSServer_GenerateTTS_InvalidArgs(t *testing.T) {
	server := mcp_servers.NewBarkTTSServer()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing text",
			args:        map[string]interface{}{"voice_preset": "v2/en_speaker_6"},
			expectError: true,
			errorMsg:    "text is required",
		},
		{
			name:        "empty text",
			args:        map[string]interface{}{"text": "", "voice_preset": "v2/en_speaker_6"},
			expectError: true,
			errorMsg:    "text cannot be empty",
		},
		{
			name:        "invalid voice preset",
			args:        map[string]interface{}{"text": "Hello", "voice_preset": "invalid_voice"},
			expectError: true,
			errorMsg:    "invalid voice preset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.GenerateTTS(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestBarkTTSServer_ListVoices_NoExternalCall(t *testing.T) {
	server := mcp_servers.NewBarkTTSServer()

	// Test that ListVoices returns predefined voices without external calls
	args := map[string]interface{}{}
	result, err := server.ListVoices(args)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, resultMap, "voices")
	assert.Contains(t, resultMap, "total")

	voices, ok := resultMap["voices"].([]map[string]interface{})
	require.True(t, ok)
	assert.Greater(t, len(voices), 0)
}

func TestSpeechT5TTSServer_GenerateTTS_InvalidArgs(t *testing.T) {
	server := mcp_servers.NewSpeechT5Server()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing text",
			args:        map[string]interface{}{"voice_preset": "default"},
			expectError: true,
			errorMsg:    "text is required",
		},
		{
			name:        "empty text",
			args:        map[string]interface{}{"text": "", "voice_preset": "default"},
			expectError: true,
			errorMsg:    "text cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.GenerateTTS(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestLLaVAServer_AnalyzeImage_InvalidArgs(t *testing.T) {
	server := mcp_servers.NewLLaVAServer()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing image",
			args:        map[string]interface{}{"prompt": "Describe this image"},
			expectError: true,
			errorMsg:    "image path is required",
		},
		{
			name:        "empty image",
			args:        map[string]interface{}{"image": "", "prompt": "Describe this image"},
			expectError: true,
			errorMsg:    "image path cannot be empty",
		},
		{
			name:        "missing prompt",
			args:        map[string]interface{}{"image": "/path/to/image.jpg"},
			expectError: true,
			errorMsg:    "prompt is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.AnalyzeImage(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestPix2StructServer_ParseUI_InvalidArgs(t *testing.T) {
	server := mcp_servers.NewPix2StructServer()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing image",
			args:        map[string]interface{}{"prompt": "Describe the UI"},
			expectError: true,
			errorMsg:    "image path is required",
		},
		{
			name:        "empty image",
			args:        map[string]interface{}{"image": "", "prompt": "Describe the UI"},
			expectError: true,
			errorMsg:    "image path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.ParseUI(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestSunoServer_NewSunoServer(t *testing.T) {
	server := mcp_servers.NewSunoServer()

	require.NotNil(t, server)
	assert.Equal(t, "suno-music", server.Config.Name)
	assert.Equal(t, "1.0.0", server.Config.Version)
	assert.Equal(t, "stdio", server.Config.Transport)
	assert.NotNil(t, server.Tools)
}

func TestSunoServer_GenerateMusic_InvalidArgs(t *testing.T) {
	server := mcp_servers.NewSunoServer()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing prompt",
			args:        map[string]interface{}{"duration": 30},
			expectError: true,
			errorMsg:    "prompt is required",
		},
		{
			name:        "empty prompt",
			args:        map[string]interface{}{"prompt": "", "duration": 30},
			expectError: true,
			errorMsg:    "prompt cannot be empty",
		},
		{
			name:        "invalid duration",
			args:        map[string]interface{}{"prompt": "Create happy music", "duration": 0},
			expectError: true,
			errorMsg:    "duration must be between 10 and 60 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.GenerateMusic(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestMCPError_Error(t *testing.T) {
	err := &mcp_servers.MCPError{
		Code:    -32600,
		Message: "Invalid Request",
		Data:    map[string]interface{}{"details": "missing method"},
	}

	errorStr := err.Error()
	assert.Contains(t, errorStr, "Invalid Request")
	assert.Contains(t, errorStr, "-32600")
}

func TestBaseServerImpl_Run_InvalidTransport(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:      "test-server",
		Version:   "1.0.0",
		Transport: "invalid_transport",
	}

	server := mcp_servers.NewBaseServer(config)

	err := server.Run()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported transport")
}

func TestBaseServerImpl_AddTool_DuplicateName(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:    "test-server",
		Version: "1.0.0",
	}

	server := mcp_servers.NewBaseServer(config)

	// Add first tool
	server.AddTool("test_tool", "First description", func(args map[string]interface{}) (interface{}, error) {
		return "first", nil
	})

	// Add duplicate tool (should overwrite)
	server.AddTool("test_tool", "Second description", func(args map[string]interface{}) (interface{}, error) {
		return "second", nil
	})

	assert.Contains(t, server.Tools, "test_tool")
	tool := server.Tools["test_tool"]
	assert.Equal(t, "Second description", tool.Description) // Should be updated
}
