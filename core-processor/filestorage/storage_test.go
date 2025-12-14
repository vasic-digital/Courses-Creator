package storage

import (
	"bytes"
	"strings"
	"testing"
)

func TestLocalStorage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create local storage
	config := StorageConfig{
		Type:      "local",
		BasePath:  tempDir,
		PublicURL: "http://localhost:8080/storage",
	}
	storage := NewLocalStorage(config)

	// Test Save
	testContent := "Hello, World!"
	err := storage.Save("test.txt", []byte(testContent))
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Test Exists
	exists := storage.Exists("test.txt")
	if !exists {
		t.Error("File should exist")
	}

	// Test Load
	data, err := storage.Load("test.txt")
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}
	if string(data) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(data))
	}

	// Test GetSize
	size, err := storage.GetSize("test.txt")
	if err != nil {
		t.Fatalf("Failed to get file size: %v", err)
	}
	if size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), size)
	}

	// Test List
	files, err := storage.List("")
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Test GetURL
	url := storage.GetURL("test.txt")
	expectedURL := "http://localhost:8080/storage/test.txt"
	if url != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, url)
	}

	// Test SaveReader with strings.Reader
	readerContent := "Reader test content"
	reader := strings.NewReader(readerContent)
	err = storage.SaveReader("reader_test.txt", reader)
	if err != nil {
		t.Fatalf("Failed to save reader: %v", err)
	}

	// Verify SaveReader worked
	data, err = storage.Load("reader_test.txt")
	if err != nil {
		t.Fatalf("Failed to load reader file: %v", err)
	}
	if string(data) != readerContent {
		t.Errorf("Expected reader content '%s', got '%s'", readerContent, string(data))
	}

	// Test SaveReader with bytes.Buffer
	bufferContent := "Buffer test content"
	buffer := bytes.NewBufferString(bufferContent)
	err = storage.SaveReader("buffer_test.txt", buffer)
	if err != nil {
		t.Fatalf("Failed to save buffer: %v", err)
	}

	// Verify SaveReader with buffer worked
	data, err = storage.Load("buffer_test.txt")
	if err != nil {
		t.Fatalf("Failed to load buffer file: %v", err)
	}
	if string(data) != bufferContent {
		t.Errorf("Expected buffer content '%s', got '%s'", bufferContent, string(data))
	}

	// Test CreateDir
	err = storage.CreateDir("test_dir")
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Test saving file in created directory
	err = storage.Save("test_dir/nested_file.txt", []byte("nested content"))
	if err != nil {
		t.Fatalf("Failed to save in created directory: %v", err)
	}

	// Test GetFile
	file, err := storage.GetFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	if file.Name != "test.txt" {
		t.Errorf("Expected file name 'test.txt', got '%s'", file.Name)
	}
	if file.Size != int64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), file.Size)
	}

	// Test List with subdirectory
	files, err = storage.List("test_dir")
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file in test_dir, got %d", len(files))
	}
	if files[0] != "nested_file.txt" {
		t.Errorf("Expected file 'nested_file.txt', got '%s'", files[0])
	}

	// Test Delete
	err = storage.Delete("test.txt")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file is deleted
	exists = storage.Exists("test.txt")
	if exists {
		t.Error("File should not exist after delete")
	}
}

func TestStorageManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create storage config
	config := StorageConfig{
		Type:      "local",
		BasePath:  tempDir,
		PublicURL: "http://localhost:8080/storage",
	}

	// Create storage manager
	manager, err := NewStorageManagerWithDefault(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test Save
	testContent := "Storage Manager Test"
	err = manager.Save("manager/test.txt", []byte(testContent))
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Test Load
	data, err := manager.Load("manager/test.txt")
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}
	if string(data) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(data))
	}

	// Test URL generation
	url := manager.GetURL("manager/test.txt")
	expectedURL := "http://localhost:8080/storage/manager/test.txt"
	if url != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, url)
	}

	// Test path helpers
	coursePath := manager.GetCoursePath("course123")
	expectedCoursePath := "courses/course123"
	if coursePath != expectedCoursePath {
		t.Errorf("Expected course path '%s', got '%s'", expectedCoursePath, coursePath)
	}

	lessonPath := manager.GetLessonPath("course123", "lesson456")
	expectedLessonPath := "courses/course123/lessons/lesson456"
	if lessonPath != expectedLessonPath {
		t.Errorf("Expected lesson path '%s', got '%s'", expectedLessonPath, lessonPath)
	}

	// Test provider switching
	s3Config := StorageConfig{
		Type: "s3",
		Settings: map[string]interface{}{
			"bucket": "test-bucket",
			"region": "us-east-1",
		},
		PublicURL: "https://test-bucket.s3.amazonaws.com",
	}

	// Test provider switching - this may succeed if AWS credentials are configured
	err = manager.SwitchProvider(s3Config)
	// We don't check for error here since it might succeed if AWS is configured
	if err != nil {
		t.Logf("Expected error when switching to S3 (no credentials): %v", err)
	}

	// Test GetProvider
	defaultProvider := manager.GetProvider("")
	if defaultProvider == nil {
		t.Error("Default provider should not be nil")
	}

	// Test DefaultProvider
	defaultProvider2 := manager.DefaultProvider()
	if defaultProvider2 == nil {
		t.Error("Default provider should not be nil")
	}

	// Test path helper functions
	storagePath := GetStoragePath("course123", "videos", "lesson1.mp4")
	expectedStoragePath := "courses/course123/videos/lesson1.mp4"
	if storagePath != expectedStoragePath {
		t.Errorf("Expected storage path '%s', got '%s'", expectedStoragePath, storagePath)
	}

	videoPath := GetVideoStoragePath("course123", "lesson456", "video.mp4")
	expectedVideoPath := "courses/course123/lessons/lesson456/videos/video.mp4"
	if videoPath != expectedVideoPath {
		t.Errorf("Expected video path '%s', got '%s'", expectedVideoPath, videoPath)
	}

	audioPath := GetAudioStoragePath("course123", "lesson456", "audio.mp3")
	expectedAudioPath := "courses/course123/lessons/lesson456/audio/audio.mp3"
	if audioPath != expectedAudioPath {
		t.Errorf("Expected audio path '%s', got '%s'", expectedAudioPath, audioPath)
	}

	subtitlePath := GetSubtitleStoragePath("course123", "lesson456", "en", "subs.srt")
	expectedSubtitlePath := "courses/course123/lessons/lesson456/subtitles/en/subs.srt"
	if subtitlePath != expectedSubtitlePath {
		t.Errorf("Expected subtitle path '%s', got '%s'", expectedSubtitlePath, subtitlePath)
	}

	backgroundPath := GetBackgroundStoragePath("course123", "bg.jpg")
	expectedBackgroundPath := "courses/course123/assets/backgrounds/bg.jpg"
	if backgroundPath != expectedBackgroundPath {
		t.Errorf("Expected background path '%s', got '%s'", expectedBackgroundPath, backgroundPath)
	}

	// Test default filename behavior
	defaultVideoPath := GetVideoStoragePath("course123", "lesson456", "")
	expectedDefaultVideoPath := "courses/course123/lessons/lesson456/videos/video.mp4"
	if defaultVideoPath != expectedDefaultVideoPath {
		t.Errorf("Expected default video path '%s', got '%s'", expectedDefaultVideoPath, defaultVideoPath)
	}
}

func TestLocalStorage_ErrorConditions(t *testing.T) {
	tempDir := t.TempDir()

	config := StorageConfig{
		Type:      "local",
		BasePath:  tempDir,
		PublicURL: "http://localhost:8080/storage",
	}
	storage := NewLocalStorage(config)

	// Test Load non-existent file
	_, err := storage.Load("nonexistent.txt")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	// Test Delete non-existent file
	err = storage.Delete("nonexistent.txt")
	if err == nil {
		t.Error("Expected error when deleting non-existent file")
	}

	// Test GetSize non-existent file
	_, err = storage.GetSize("nonexistent.txt")
	if err == nil {
		t.Error("Expected error when getting size of non-existent file")
	}

	// Test List non-existent directory
	_, err = storage.List("nonexistent_dir")
	if err == nil {
		t.Error("Expected error when listing non-existent directory")
	}

	// Test GetFile non-existent file
	_, err = storage.GetFile("nonexistent.txt")
	if err == nil {
		t.Error("Expected error when getting info of non-existent file")
	}
}

func TestStorageManager_MultipleProviders(t *testing.T) {
	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	configs := map[string]StorageConfig{
		"primary": {
			Type:      "local",
			BasePath:  tempDir1,
			PublicURL: "http://localhost:8080/storage1",
		},
		"secondary": {
			Type:      "local",
			BasePath:  tempDir2,
			PublicURL: "http://localhost:8080/storage2",
		},
	}

	manager, err := NewStorageManager(configs)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test GetProvider with specific name
	primaryProvider := manager.GetProvider("primary")
	if primaryProvider == nil {
		t.Error("Primary provider should not be nil")
	}

	secondaryProvider := manager.GetProvider("secondary")
	if secondaryProvider == nil {
		t.Error("Secondary provider should not be nil")
	}

	// Test saving to different providers
	err = manager.Save("primary/test.txt", []byte("primary content"))
	if err != nil {
		t.Fatalf("Failed to save to primary: %v", err)
	}

	// Switch to secondary provider
	err = manager.SwitchProvider(configs["secondary"])
	if err != nil {
		t.Fatalf("Failed to switch provider: %v", err)
	}

	err = manager.Save("secondary/test.txt", []byte("secondary content"))
	if err != nil {
		t.Fatalf("Failed to save to secondary: %v", err)
	}

	// Verify files were saved in correct locations
	data, err := manager.Load("secondary/test.txt")
	if err != nil {
		t.Fatalf("Failed to load from secondary: %v", err)
	}
	if string(data) != "secondary content" {
		t.Errorf("Expected 'secondary content', got '%s'", string(data))
	}
}

func TestDefaultStorageConfig(t *testing.T) {
	config := DefaultStorageConfig()

	if config.Type != "local" {
		t.Errorf("Expected type 'local', got '%s'", config.Type)
	}

	if config.BasePath != "./storage" {
		t.Errorf("Expected base path './storage', got '%s'", config.BasePath)
	}

	if config.PublicURL != "http://localhost:8080/storage" {
		t.Errorf("Expected public URL 'http://localhost:8080/storage', got '%s'", config.PublicURL)
	}
}

func TestS3Storage_NewS3Storage(t *testing.T) {
	// Test S3 storage creation without bucket (should fail)
	config := StorageConfig{
		Type: "s3",
		Settings: map[string]interface{}{
			"region": "us-east-1",
		},
	}

	_, err := NewS3Storage(config)
	if err == nil {
		t.Error("Expected error when creating S3 storage without bucket")
	}

	// Test S3 storage creation with bucket (may fail due to AWS config)
	configWithBucket := StorageConfig{
		Type: "s3",
		Settings: map[string]interface{}{
			"bucket": "test-bucket",
			"region": "us-east-1",
		},
	}

	s3Storage, err := NewS3Storage(configWithBucket)
	if err != nil {
		// This is expected if AWS credentials are not configured
		t.Logf("S3 storage creation failed (expected without AWS config): %v", err)
		return
	}

	if s3Storage == nil {
		t.Error("S3 storage should not be nil when created successfully")
	}

	if s3Storage.bucket != "test-bucket" {
		t.Errorf("Expected bucket 'test-bucket', got '%s'", s3Storage.bucket)
	}
}

func TestCreateProvider(t *testing.T) {
	// Test creating local provider
	localConfig := StorageConfig{
		Type:      "local",
		BasePath:  "./test-storage",
		PublicURL: "http://localhost:8080/storage",
	}

	provider, err := createProvider(localConfig)
	if err != nil {
		t.Fatalf("Failed to create local provider: %v", err)
	}

	if provider == nil {
		t.Error("Local provider should not be nil")
	}

	// Test creating unsupported provider
	unsupportedConfig := StorageConfig{
		Type: "unsupported",
	}

	_, err = createProvider(unsupportedConfig)
	if err == nil {
		t.Error("Expected error for unsupported storage type")
	}
}

func TestLocalStorage_GetFile(t *testing.T) {
	tempDir := t.TempDir()

	config := StorageConfig{
		Type:      "local",
		BasePath:  tempDir,
		PublicURL: "http://localhost:8080/storage",
	}
	storage := NewLocalStorage(config)

	// Create a test file
	testContent := "Test file content"
	err := storage.Save("test_file.txt", []byte(testContent))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test GetFile
	file, err := storage.GetFile("test_file.txt")
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if file.Name != "test_file.txt" {
		t.Errorf("Expected file name 'test_file.txt', got '%s'", file.Name)
	}

	if file.Path != "test_file.txt" {
		t.Errorf("Expected file path 'test_file.txt', got '%s'", file.Path)
	}

	if file.Size != int64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), file.Size)
	}

	expectedURL := "http://localhost:8080/storage/test_file.txt"
	if file.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, file.URL)
	}

	if file.Modified == "" {
		t.Error("Modified time should not be empty")
	}
}

func TestLocalStorage_GetURL_EmptyPublicURL(t *testing.T) {
	tempDir := t.TempDir()

	config := StorageConfig{
		Type:     "local",
		BasePath: tempDir,
		// PublicURL is empty
	}
	storage := NewLocalStorage(config)

	url := storage.GetURL("test.txt")
	if url != "" {
		t.Errorf("Expected empty URL when publicURL is not set, got '%s'", url)
	}
}

func TestStorageManager_DelegateMethods(t *testing.T) {
	tempDir := t.TempDir()

	config := StorageConfig{
		Type:      "local",
		BasePath:  tempDir,
		PublicURL: "http://localhost:8080/storage",
	}

	manager, err := NewStorageManagerWithDefault(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test SaveReader delegate
	readerContent := "Reader content"
	reader := strings.NewReader(readerContent)
	err = manager.SaveReader("reader.txt", reader)
	if err != nil {
		t.Fatalf("Failed to save reader: %v", err)
	}

	// Test Exists delegate
	exists := manager.Exists("reader.txt")
	if !exists {
		t.Error("File should exist")
	}

	// Test List delegate
	files, err := manager.List("")
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Should have at least one file")
	}

	// Test CreateDir delegate
	err = manager.CreateDir("test_dir")
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Test GetSize delegate
	size, err := manager.GetSize("reader.txt")
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}
	if size != int64(len(readerContent)) {
		t.Errorf("Expected size %d, got %d", len(readerContent), size)
	}

	// Test Delete delegate
	err = manager.Delete("reader.txt")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify deletion
	exists = manager.Exists("reader.txt")
	if exists {
		t.Error("File should not exist after deletion")
	}
}

func TestUtilityFunctions_DefaultFilenames(t *testing.T) {
	// Test GetStoragePath with empty filename
	path := GetStoragePath("course123", "videos", "")
	expected := "courses/course123/videos/default"
	if path != expected {
		t.Errorf("Expected '%s', got '%s'", expected, path)
	}

	// Test GetAudioStoragePath with empty filename
	audioPath := GetAudioStoragePath("course123", "lesson456", "")
	expectedAudio := "courses/course123/lessons/lesson456/audio/audio.mp3"
	if audioPath != expectedAudio {
		t.Errorf("Expected '%s', got '%s'", expectedAudio, audioPath)
	}

	// Test GetSubtitleStoragePath with empty filename
	subtitlePath := GetSubtitleStoragePath("course123", "lesson456", "en", "")
	expectedSubtitle := "courses/course123/lessons/lesson456/subtitles/en/subtitles.srt"
	if subtitlePath != expectedSubtitle {
		t.Errorf("Expected '%s', got '%s'", expectedSubtitle, subtitlePath)
	}

	// Test GetBackgroundStoragePath with empty filename
	backgroundPath := GetBackgroundStoragePath("course123", "")
	expectedBackground := "courses/course123/assets/backgrounds/background.jpg"
	if backgroundPath != expectedBackground {
		t.Errorf("Expected '%s', got '%s'", expectedBackground, backgroundPath)
	}
}

func TestNewStorageManager_ErrorHandling(t *testing.T) {
	// Test with invalid provider config
	configs := map[string]StorageConfig{
		"invalid": {
			Type: "unsupported-type",
		},
	}

	_, err := NewStorageManager(configs)
	if err == nil {
		t.Error("Expected error when creating storage manager with invalid provider")
	}
}
