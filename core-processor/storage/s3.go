package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements StorageInterface for AWS S3
type S3Storage struct {
	client   *s3.Client
	bucket   string
	basePath string
	publicURL string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(config StorageConfig) (*S3Storage, error) {
	// Load AWS configuration
	awsConfig, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	// Create S3 client
	client := s3.NewFromConfig(awsConfig)
	
	bucket, ok := config.Settings["bucket"].(string)
	if !ok {
		return nil, fmt.Errorf("S3 bucket name is required")
	}
	
	basePath, _ := config.Settings["basePath"].(string)
	if basePath == "" {
		basePath = ""
	}
	
	publicURL, _ := config.Settings["publicURL"].(string)
	
	return &S3Storage{
		client:   client,
		bucket:   bucket,
		basePath: basePath,
		publicURL: publicURL,
	}, nil
}

// Save writes data to S3
func (s3s *S3Storage) Save(path string, data []byte) error {
	fullPath := s3s.getFullPath(path)
	
	_, err := s3s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
		Body:   strings.NewReader(string(data)),
	})
	
	return err
}

// SaveReader writes data from a reader to S3
func (s3s *S3Storage) SaveReader(path string, reader io.Reader) error {
	fullPath := s3s.getFullPath(path)
	
	_, err := s3s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
		Body:   reader,
	})
	
	return err
}

// Load reads data from S3
func (s3s *S3Storage) Load(path string) ([]byte, error) {
	fullPath := s3s.getFullPath(path)
	
	result, err := s3s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
	})
	
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	
	return io.ReadAll(result.Body)
}

// Delete removes a file from S3
func (s3s *S3Storage) Delete(path string) error {
	fullPath := s3s.getFullPath(path)
	
	_, err := s3s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
	})
	
	return err
}

// Exists checks if a file exists in S3
func (s3s *S3Storage) Exists(path string) bool {
	fullPath := s3s.getFullPath(path)
	
	_, err := s3s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
	})
	
	return err == nil
}

// List returns a list of files in specified S3 "directory"
func (s3s *S3Storage) List(dir string) ([]string, error) {
	fullPath := s3s.getFullPath(dir)
	
	result, err := s3s.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:    aws.String(s3s.bucket),
		Prefix:    aws.String(fullPath),
		Delimiter: aws.String("/"),
	})
	
	if err != nil {
		return nil, err
	}
	
	var files []string
	for _, obj := range result.Contents {
		// Extract just the filename from the full key
		key := *obj.Key
		if fullPath != "" {
			key = strings.TrimPrefix(key, fullPath)
			key = strings.TrimPrefix(key, "/")
		}
		
		if key != "" {
			files = append(files, key)
		}
	}
	
	return files, nil
}

// CreateDir creates a "directory" in S3 by creating an empty object
func (s3s *S3Storage) CreateDir(path string) error {
	fullPath := s3s.getFullPath(path)
	
	// Ensure path ends with slash for S3 directories
	if !strings.HasSuffix(fullPath, "/") {
		fullPath += "/"
	}
	
	_, err := s3s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
		Body:   strings.NewReader(""),
	})
	
	return err
}

// GetURL returns a public URL for the S3 file
func (s3s *S3Storage) GetURL(path string) string {
	if s3s.publicURL == "" {
		return ""
	}
	
	fullPath := s3s.getFullPath(path)
	cleanPath := strings.TrimPrefix(fullPath, "/")
	
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s3s.publicURL, "/"), cleanPath)
}

// GetSize returns the size of the S3 file
func (s3s *S3Storage) GetSize(path string) (int64, error) {
	fullPath := s3s.getFullPath(path)
	
	result, err := s3s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
	})
	
	if err != nil {
		return 0, err
	}
	
	return *result.ContentLength, nil
}

// GetFile returns file metadata
func (s3s *S3Storage) GetFile(path string) (*File, error) {
	fullPath := s3s.getFullPath(path)
	
	result, err := s3s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(fullPath),
	})
	
	if err != nil {
		return nil, err
	}
	
	var modified string
	if result.LastModified != nil {
		modified = result.LastModified.Format(time.RFC3339)
	}
	
	return &File{
		Path:     path,
		Name:     getFilename(path),
		Size:     *result.ContentLength,
		Modified: modified,
		URL:      s3s.GetURL(path),
	}, nil
}

// getFullPath combines basePath with path for S3 keys
func (s3s *S3Storage) getFullPath(path string) string {
	if s3s.basePath == "" {
		return path
	}
	
	// Ensure no double slashes
	fullPath := strings.TrimPrefix(path, "/")
	if s3s.basePath != "" {
		fullPath = strings.Join([]string{s3s.basePath, fullPath}, "/")
	}
	
	return fullPath
}

// getFilename extracts filename from path
func getFilename(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}