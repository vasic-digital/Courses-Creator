package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// StorageManager manages storage providers and provides a unified interface
type StorageManager struct {
	providers   map[string]StorageInterface
	defaultName string
}

// NewStorageManager creates a new storage manager
func NewStorageManager(configs map[string]StorageConfig) (*StorageManager, error) {
	sm := &StorageManager{
		providers: make(map[string]StorageInterface),
	}

	for name, config := range configs {
		provider, err := createProvider(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create storage provider %s: %w", name, err)
		}
		sm.providers[name] = provider

		// Set first provider as default
		if sm.defaultName == "" {
			sm.defaultName = name
		}
	}

	return sm, nil
}

// NewStorageManagerWithDefault creates a new storage manager with a single default provider
func NewStorageManagerWithDefault(config StorageConfig) (*StorageManager, error) {
	return NewStorageManager(map[string]StorageConfig{"default": config})
}

// GetProvider returns a storage provider by name, or default if name is empty
func (sm *StorageManager) GetProvider(name string) StorageInterface {
	if name == "" {
		return sm.providers[sm.defaultName]
	}
	return sm.providers[name]
}

// DefaultProvider returns the default storage provider
func (sm *StorageManager) DefaultProvider() StorageInterface {
	return sm.providers[sm.defaultName]
}

// Delegate methods to default provider
func (sm *StorageManager) Save(path string, data []byte) error {
	return sm.DefaultProvider().Save(path, data)
}

func (sm *StorageManager) SaveReader(path string, reader io.Reader) error {
	return sm.DefaultProvider().SaveReader(path, reader)
}

func (sm *StorageManager) Load(path string) ([]byte, error) {
	return sm.DefaultProvider().Load(path)
}

func (sm *StorageManager) Delete(path string) error {
	return sm.DefaultProvider().Delete(path)
}

func (sm *StorageManager) Exists(path string) bool {
	return sm.DefaultProvider().Exists(path)
}

func (sm *StorageManager) List(dir string) ([]string, error) {
	return sm.DefaultProvider().List(dir)
}

func (sm *StorageManager) CreateDir(path string) error {
	return sm.DefaultProvider().CreateDir(path)
}

func (sm *StorageManager) GetURL(path string) string {
	return sm.DefaultProvider().GetURL(path)
}

func (sm *StorageManager) GetSize(path string) (int64, error) {
	return sm.DefaultProvider().GetSize(path)
}

// SwitchProvider changes the default provider
func (sm *StorageManager) SwitchProvider(config StorageConfig) error {
	provider, err := createProvider(config)
	if err != nil {
		return err
	}

	providerName := "switched"
	sm.providers[providerName] = provider
	sm.defaultName = providerName

	return nil
}

// GetCoursePath generates a storage path for course files
func (sm *StorageManager) GetCoursePath(courseID string) string {
	return filepath.Join("courses", courseID)
}

// GetLessonPath generates a storage path for lesson files
func (sm *StorageManager) GetLessonPath(courseID, lessonID string) string {
	return filepath.Join("courses", courseID, "lessons", lessonID)
}

// createProvider creates a storage provider based on configuration
func createProvider(config StorageConfig) (StorageInterface, error) {
	switch strings.ToLower(config.Type) {
	case "local":
		if config.BasePath == "" {
			config.BasePath = "./storage"
		}

		// Ensure base path exists
		if err := os.MkdirAll(config.BasePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create base path: %w", err)
		}

		return NewLocalStorage(config), nil

	case "s3":
		return NewS3Storage(config)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// DefaultStorageConfig returns a default local storage configuration
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Type:      "local",
		BasePath:  "./storage",
		PublicURL: "http://localhost:8080/storage",
	}
}

// GetStoragePath generates a storage path for course content
func GetStoragePath(courseID, contentType, filename string) string {
	if filename == "" {
		filename = "default"
	}
	return filepath.Join("courses", courseID, contentType, filename)
}

// GetVideoStoragePath generates a storage path for video files
func GetVideoStoragePath(courseID, lessonID, filename string) string {
	if filename == "" {
		filename = "video.mp4"
	}
	return filepath.Join("courses", courseID, "lessons", lessonID, "videos", filename)
}

// GetAudioStoragePath generates a storage path for audio files
func GetAudioStoragePath(courseID, lessonID, filename string) string {
	if filename == "" {
		filename = "audio.mp3"
	}
	return filepath.Join("courses", courseID, "lessons", lessonID, "audio", filename)
}

// GetSubtitleStoragePath generates a storage path for subtitle files
func GetSubtitleStoragePath(courseID, lessonID, language, filename string) string {
	if filename == "" {
		filename = "subtitles.srt"
	}
	return filepath.Join("courses", courseID, "lessons", lessonID, "subtitles", language, filename)
}

// GetBackgroundStoragePath generates a storage path for background images
func GetBackgroundStoragePath(courseID, filename string) string {
	if filename == "" {
		filename = "background.jpg"
	}
	return filepath.Join("courses", courseID, "assets", "backgrounds", filename)
}
