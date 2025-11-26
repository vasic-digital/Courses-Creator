package utils

import (
	"context"
	"os/exec"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HashString creates a simple hash for unique filenames
func HashString(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}

// GenerateID generates a unique ID
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ExecuteCommand creates a command for execution
func ExecuteCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd
}

// ExecuteCommandWithOutput executes a command and returns output
func ExecuteCommandWithOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := ExecuteCommand(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

// EnsureDir ensures a directory exists
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// CleanTempFiles removes temporary files older than specified duration
func CleanTempFiles(dir string, olderThan time.Duration) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(file)
		}
	}

	return nil
}

// SanitizeFilename sanitizes a string for use as a filename
func SanitizeFilename(name string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Limit length
	if len(result) > 100 {
		result = result[:100]
	}

	return result
}

// GetFileExtension returns the file extension
func GetFileExtension(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

// GetFileSize returns the size of a file
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// Retry executes a function with retries
func Retry(attempts int, sleep time.Duration, f func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}

// SafeClose safely closes a Closer interface
func SafeClose(c io.Closer) {
	if c != nil {
		c.Close()
	}
}

// WriteFile writes content to a file
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
