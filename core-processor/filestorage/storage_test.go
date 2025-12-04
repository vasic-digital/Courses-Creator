package storage

import (
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
}
