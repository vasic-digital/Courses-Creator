package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage implements StorageInterface for local filesystem
type LocalStorage struct {
	basePath string
	publicURL string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(config StorageConfig) *LocalStorage {
	return &LocalStorage{
		basePath: config.BasePath,
		publicURL: config.PublicURL,
	}
}

// Save writes data to local filesystem
func (ls *LocalStorage) Save(path string, data []byte) error {
	fullPath := filepath.Join(ls.basePath, path)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	return os.WriteFile(fullPath, data, 0644)
}

// SaveReader writes data from a reader to local filesystem
func (ls *LocalStorage) SaveReader(path string, reader io.Reader) error {
	fullPath := filepath.Join(ls.basePath, path)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	_, err = io.Copy(file, reader)
	return err
}

// Load reads data from local filesystem
func (ls *LocalStorage) Load(path string) ([]byte, error) {
	fullPath := filepath.Join(ls.basePath, path)
	return os.ReadFile(fullPath)
}

// Delete removes a file from local filesystem
func (ls *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(ls.basePath, path)
	return os.Remove(fullPath)
}

// Exists checks if a file exists in local filesystem
func (ls *LocalStorage) Exists(path string) bool {
	fullPath := filepath.Join(ls.basePath, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// List returns a list of files in specified directory
func (ls *LocalStorage) List(dir string) ([]string, error) {
	fullDir := filepath.Join(ls.basePath, dir)
	
	entries, err := os.ReadDir(fullDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// CreateDir creates a directory in local filesystem
func (ls *LocalStorage) CreateDir(path string) error {
	fullPath := filepath.Join(ls.basePath, path)
	return os.MkdirAll(fullPath, 0755)
}

// GetURL returns a public URL for file (for local storage, returns path-based URL)
func (ls *LocalStorage) GetURL(path string) string {
	if ls.publicURL == "" {
		return ""
	}
	
	// Ensure path doesn't start with slash to avoid double slashes
	cleanPath := strings.TrimPrefix(path, "/")
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(ls.publicURL, "/"), cleanPath)
}

// GetSize returns the size of the file
func (ls *LocalStorage) GetSize(path string) (int64, error) {
	fullPath := filepath.Join(ls.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFile returns file metadata
func (ls *LocalStorage) GetFile(path string) (*File, error) {
	fullPath := filepath.Join(ls.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	
	return &File{
		Path:     path,
		Name:     filepath.Base(path),
		Size:     info.Size(),
		Modified: info.ModTime().Format(time.RFC3339),
		URL:      ls.GetURL(path),
	}, nil
}