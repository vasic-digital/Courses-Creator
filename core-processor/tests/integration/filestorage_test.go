package integration

import (
	"bytes"
	"strings"
	"testing"

	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_Integration(t *testing.T) {
	t.Run("Local Storage Basic Operations", func(t *testing.T) {
		// Create local storage
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  t.TempDir(),
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err := storage.NewStorageManagerWithDefault(config)
		require.NoError(t, err)

		// Test Save and Load
		testContent := "Hello, World!"
		err = storageManager.Save("test.txt", []byte(testContent))
		require.NoError(t, err)

		data, err := storageManager.Load("test.txt")
		require.NoError(t, err)
		assert.Equal(t, testContent, string(data))

		// Test Exists
		exists := storageManager.Exists("test.txt")
		assert.True(t, exists)

		// Test GetSize
		size, err := storageManager.GetSize("test.txt")
		require.NoError(t, err)
		assert.Equal(t, int64(len(testContent)), size)

		// Test GetURL
		url := storageManager.GetURL("test.txt")
		assert.Equal(t, "http://localhost:8080/storage/test.txt", url)

		// Test SaveReader
		readerContent := "Reader test content"
		reader := strings.NewReader(readerContent)
		err = storageManager.SaveReader("reader_test.txt", reader)
		require.NoError(t, err)

		data, err = storageManager.Load("reader_test.txt")
		require.NoError(t, err)
		assert.Equal(t, readerContent, string(data))

		// Test List
		files, err := storageManager.List("")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(files), 2)

		// Test CreateDir
		err = storageManager.CreateDir("test_dir")
		require.NoError(t, err)

		// Test saving in directory
		err = storageManager.Save("test_dir/nested.txt", []byte("nested content"))
		require.NoError(t, err)

		// Test Delete
		err = storageManager.Delete("test.txt")
		require.NoError(t, err)

		exists = storageManager.Exists("test.txt")
		assert.False(t, exists)
	})

	t.Run("Storage Manager with Multiple Providers", func(t *testing.T) {
		// Create multiple storage providers
		configs := map[string]storage.StorageConfig{
			"primary": {
				Type:      "local",
				BasePath:  t.TempDir(),
				PublicURL: "http://localhost:8080/primary",
			},
			"secondary": {
				Type:      "local",
				BasePath:  t.TempDir(),
				PublicURL: "http://localhost:8080/secondary",
			},
		}

		storageManager, err := storage.NewStorageManager(configs)
		require.NoError(t, err)

		// Test saving to default provider
		err = storageManager.Save("default_test.txt", []byte("default content"))
		require.NoError(t, err)

		// Test switching providers
		err = storageManager.SwitchProvider(configs["secondary"])
		require.NoError(t, err)

		err = storageManager.Save("secondary_test.txt", []byte("secondary content"))
		require.NoError(t, err)

		// Verify files in correct locations
		data, err := storageManager.Load("secondary_test.txt")
		require.NoError(t, err)
		assert.Equal(t, "secondary content", string(data))
	})

	t.Run("Path Helper Functions", func(t *testing.T) {
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  t.TempDir(),
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err := storage.NewStorageManagerWithDefault(config)
		require.NoError(t, err)

		// Test course path
		coursePath := storageManager.GetCoursePath("course123")
		assert.Equal(t, "courses/course123", coursePath)

		// Test lesson path
		lessonPath := storageManager.GetLessonPath("course123", "lesson456")
		assert.Equal(t, "courses/course123/lessons/lesson456", lessonPath)

		// Test utility functions
		storagePath := storage.GetStoragePath("course123", "videos", "lesson1.mp4")
		assert.Equal(t, "courses/course123/videos/lesson1.mp4", storagePath)

		videoPath := storage.GetVideoStoragePath("course123", "lesson456", "video.mp4")
		assert.Equal(t, "courses/course123/lessons/lesson456/videos/video.mp4", videoPath)

		audioPath := storage.GetAudioStoragePath("course123", "lesson456", "audio.mp3")
		assert.Equal(t, "courses/course123/lessons/lesson456/audio/audio.mp3", audioPath)

		subtitlePath := storage.GetSubtitleStoragePath("course123", "lesson456", "en", "subs.srt")
		assert.Equal(t, "courses/course123/lessons/lesson456/subtitles/en/subs.srt", subtitlePath)

		backgroundPath := storage.GetBackgroundStoragePath("course123", "bg.jpg")
		assert.Equal(t, "courses/course123/assets/backgrounds/bg.jpg", backgroundPath)
	})

	t.Run("Error Conditions", func(t *testing.T) {
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  t.TempDir(),
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err := storage.NewStorageManagerWithDefault(config)
		require.NoError(t, err)

		// Test loading non-existent file
		_, err = storageManager.Load("nonexistent.txt")
		assert.Error(t, err)

		// Test deleting non-existent file
		err = storageManager.Delete("nonexistent.txt")
		assert.Error(t, err)

		// Test getting size of non-existent file
		_, err = storageManager.GetSize("nonexistent.txt")
		assert.Error(t, err)

		// Test listing non-existent directory
		_, err = storageManager.List("nonexistent_dir")
		assert.Error(t, err)
	})

	t.Run("File Operations with Course Context", func(t *testing.T) {
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  t.TempDir(),
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err := storage.NewStorageManagerWithDefault(config)
		require.NoError(t, err)

		courseID := "test-course-123"
		lessonID := "test-lesson-456"

		// Create course directory structure
		coursePath := storageManager.GetCoursePath(courseID)
		err = storageManager.CreateDir(coursePath)
		require.NoError(t, err)

		// Create lesson directory
		lessonPath := storageManager.GetLessonPath(courseID, lessonID)
		err = storageManager.CreateDir(lessonPath)
		require.NoError(t, err)

		// Save course metadata
		courseMetadata := `{"title": "Test Course", "description": "Test Description"}`
		err = storageManager.Save(coursePath+"/metadata.json", []byte(courseMetadata))
		require.NoError(t, err)

		// Save lesson content
		lessonContent := `{"title": "Test Lesson", "content": "Lesson content here"}`
		err = storageManager.Save(lessonPath+"/content.json", []byte(lessonContent))
		require.NoError(t, err)

		// Verify files exist
		assert.True(t, storageManager.Exists(coursePath+"/metadata.json"))
		assert.True(t, storageManager.Exists(lessonPath+"/content.json"))

		// Load and verify content
		data, err := storageManager.Load(coursePath + "/metadata.json")
		require.NoError(t, err)
		assert.Equal(t, courseMetadata, string(data))

		data, err = storageManager.Load(lessonPath + "/content.json")
		require.NoError(t, err)
		assert.Equal(t, lessonContent, string(data))

		// List course files
		files, err := storageManager.List(coursePath)
		require.NoError(t, err)
		assert.Contains(t, files, "metadata.json")

		// List lesson files
		files, err = storageManager.List(lessonPath)
		require.NoError(t, err)
		assert.Contains(t, files, "content.json")
	})

	t.Run("Large File Operations", func(t *testing.T) {
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  t.TempDir(),
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err := storage.NewStorageManagerWithDefault(config)
		require.NoError(t, err)

		// Create large content (1MB)
		largeContent := strings.Repeat("A", 1024*1024)

		// Save large file
		err = storageManager.Save("large_file.txt", []byte(largeContent))
		require.NoError(t, err)

		// Load and verify
		data, err := storageManager.Load("large_file.txt")
		require.NoError(t, err)
		assert.Equal(t, len(largeContent), len(data))
		assert.Equal(t, largeContent, string(data))

		// Verify size
		size, err := storageManager.GetSize("large_file.txt")
		require.NoError(t, err)
		assert.Equal(t, int64(len(largeContent)), size)
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  t.TempDir(),
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err := storage.NewStorageManagerWithDefault(config)
		require.NoError(t, err)

		// Create multiple files concurrently (simulated with sequential writes)
		files := []string{"file1.txt", "file2.txt", "file3.txt"}
		contents := []string{"content1", "content2", "content3"}

		for i, filename := range files {
			err := storageManager.Save(filename, []byte(contents[i]))
			require.NoError(t, err)
		}

		// Verify all files exist
		for _, filename := range files {
			assert.True(t, storageManager.Exists(filename))
		}

		// List all files
		listedFiles, err := storageManager.List("")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listedFiles), len(files))

		// Delete all files
		for _, filename := range files {
			err := storageManager.Delete(filename)
			require.NoError(t, err)
			assert.False(t, storageManager.Exists(filename))
		}
	})

	t.Run("Binary File Operations", func(t *testing.T) {
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  t.TempDir(),
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err := storage.NewStorageManagerWithDefault(config)
		require.NoError(t, err)

		// Create binary data
		binaryData := make([]byte, 256)
		for i := 0; i < 256; i++ {
			binaryData[i] = byte(i)
		}

		// Save binary file
		err = storageManager.Save("binary.bin", binaryData)
		require.NoError(t, err)

		// Load binary file
		loadedData, err := storageManager.Load("binary.bin")
		require.NoError(t, err)

		// Verify binary data
		assert.Equal(t, len(binaryData), len(loadedData))
		assert.Equal(t, binaryData, loadedData)

		// Test SaveReader with bytes.Buffer (binary)
		buffer := bytes.NewBuffer(binaryData)
		err = storageManager.SaveReader("binary_from_buffer.bin", buffer)
		require.NoError(t, err)

		// Verify buffer save worked
		loadedFromBuffer, err := storageManager.Load("binary_from_buffer.bin")
		require.NoError(t, err)
		assert.Equal(t, binaryData, loadedFromBuffer)
	})
}
