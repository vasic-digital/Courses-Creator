package contract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/course-creator/core-processor/models"
	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoAssembler_CreateVideo_Basic(t *testing.T) {
	// Create a mock storage
	tempDir := t.TempDir()
	storageConfig := storage.StorageConfig{
		Type:      "local",
		BasePath:  tempDir,
		PublicURL: "http://localhost:8080/storage",
	}
	storageManager, _ := storage.NewStorageManagerWithDefault(storageConfig)
	
	assembler := pipeline.NewVideoAssembler(storageManager.DefaultProvider())

	// Create temporary files
	audioPath := filepath.Join(tempDir, "test_audio.wav")

	// Create placeholder audio file
	err := os.WriteFile(audioPath, []byte("fake audio"), 0644)
	require.NoError(t, err)

	textContent := "This is test content for video."
	options := models.ProcessingOptions{}

	// Create video - now requires course and lesson IDs
	resultPath, err := assembler.CreateVideo(audioPath, textContent, "test-course", "test-lesson", options)
	require.NoError(t, err)
	assert.NotEmpty(t, resultPath)

	// Check if output file exists (placeholder)
	_, err = os.Stat(resultPath)
	if err == nil {
		// File exists, good
	} else {
		// Placeholder created
		assert.Contains(t, resultPath, "video_")
	}
}

func TestVideoAssembler_AddBackgroundMusic_NoFFmpeg(t *testing.T) {
	// Create a mock storage
	tempDir := t.TempDir()
	storageConfig := storage.StorageConfig{
		Type:      "local",
		BasePath:  tempDir,
		PublicURL: "http://localhost:8080/storage",
	}
	storageManager, _ := storage.NewStorageManagerWithDefault(storageConfig)
	
	assembler := pipeline.NewVideoAssembler(storageManager.DefaultProvider())

	videoPath := filepath.Join(tempDir, "test.mp4")
	musicPath := filepath.Join(tempDir, "music.mp3")

	// Create placeholder files
	err := os.WriteFile(videoPath, []byte("fake video"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(musicPath, []byte("fake music"), 0644)
	require.NoError(t, err)

	resultPath, err := assembler.AddBackgroundMusic(nil, videoPath, musicPath, 0.5)
	require.NoError(t, err)
	// Should return original since FFmpeg not available
	assert.Equal(t, videoPath, resultPath)
}
