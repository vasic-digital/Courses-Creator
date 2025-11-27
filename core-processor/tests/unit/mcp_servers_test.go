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
		Name:      "test-server",
		Version:   "1.0.0",
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
		Name:      "test-server",
		Version:   "1.0.0",
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
		Name:      "test-server",
		Version:   "1.0.0",
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
	assert.Equal(t, int32(-32000), response.Error.Code)
}

func TestBaseServerImpl_Stop(t *testing.T) {
	config := mcp_servers.MCPServerConfig{
		Name:      "test-server",
		Version:   "1.0.0",
		Transport:  "stdio",
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	server := mcp_servers.NewBaseServer(config)

	// Server should be running initially
	assert.True(t, server.IsRunning())

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
	server := mcp_servers.NewBarkTTSServer()

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

	// Clean up generated file
	audioPath, ok := resultMap["audio_path"].(string)
	if ok && audioPath != "" {
		os.Remove(audioPath)
	}
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

func TestSpeechT5TTSServer_GenerateTTS(t *testing.T) {
	server := mcp_servers.NewSpeechT5Server()

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
	tempFile := filepath.Join(os.TempDir(), "test_file", utils.GenerateID())
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	assert.True(t, utils.FileExists(tempFile))

	// Test non-existing file
	nonExistentFile := filepath.Join(os.TempDir(), "non_existent", utils.GenerateID())
	assert.False(t, utils.FileExists(nonExistentFile))
}

func TestUtils_CopyFile(t *testing.T) {
	srcFile := filepath.Join(os.TempDir(), "src", utils.GenerateID())
	dstFile := filepath.Join(os.TempDir(), "dst", utils.GenerateID())
	
	// Create source file
	content := []byte("test content for copy")
	err := os.WriteFile(srcFile, content, 0644)
	require.NoError(t, err)
	defer os.Remove(srcFile)

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
	tempFile := filepath.Join(os.TempDir(), "test_size", utils.GenerateID())
	content := []byte("test content for size")
	
	err := os.WriteFile(tempFile, content, 0644)
	require.NoError(t, err)
	defer os.Remove(tempFile)

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