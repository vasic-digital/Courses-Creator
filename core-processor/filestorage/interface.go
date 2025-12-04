package storage

import (
	"io"
)

// StorageInterface defines the contract for file storage operations
type StorageInterface interface {
	// Save writes data to storage at the specified path
	Save(path string, data []byte) error

	// SaveReader writes data from a reader to storage at the specified path
	SaveReader(path string, reader io.Reader) error

	// Load reads data from storage at the specified path
	Load(path string) ([]byte, error)

	// Delete removes a file from storage
	Delete(path string) error

	// Exists checks if a file exists in storage
	Exists(path string) bool

	// List returns a list of files in the specified directory
	List(dir string) ([]string, error)

	// CreateDir creates a directory in storage
	CreateDir(path string) error

	// GetURL returns a public URL for the file (if applicable)
	GetURL(path string) string

	// GetSize returns the size of the file
	GetSize(path string) (int64, error)
}

// File represents a stored file with metadata
type File struct {
	Path     string
	Name     string
	Size     int64
	Modified string
	URL      string
}

// StorageConfig contains configuration for storage providers
type StorageConfig struct {
	Type      string                 // "local", "s3", "gcs"
	Settings  map[string]interface{} // Provider-specific settings
	BasePath  string                 // Base path for local storage
	PublicURL string                 // Base URL for public access
}
