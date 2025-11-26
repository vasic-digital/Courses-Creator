# Course Creator Storage System

This document explains how to configure and use the Course Creator storage system.

## Overview

The Course Creator uses an abstraction layer for storage operations, supporting multiple backends:
- **Local Storage**: Files are stored on the local filesystem
- **S3 Storage**: Files are stored in AWS S3 buckets

## Configuration

Storage is configured via environment variables:

### Local Storage (Default)
```bash
STORAGE_TYPE=local
STORAGE_BASE_PATH=./storage
STORAGE_PUBLIC_URL=http://localhost:8080/storage
```

### S3 Storage
```bash
STORAGE_TYPE=s3
STORAGE_BASE_PATH=                   # Base prefix within bucket (optional)
STORAGE_PUBLIC_URL=https://your-bucket.s3.amazonaws.com
AWS_REGION=us-east-1                # AWS region
AWS_ACCESS_KEY_ID=your_access_key    # AWS access key
AWS_SECRET_ACCESS_KEY=your_secret_key  # AWS secret key
```

## Directory Structure

Files are organized in the following structure:

```
{base_path}/
└── courses/
    └── {course-id}/
        ├── metadata.json
        └── lessons/
            └── {lesson-id}/
                ├── content.md
                ├── audio.mp3
                └── video.mp4
```

## Using the Storage System

### Programmatic Usage

```go
import (
    storage "github.com/course-creator/core-processor/filestorage"
)

// Create storage configuration
config := storage.StorageConfig{
    Type:      "local",
    BasePath:  "./my_storage",
    PublicURL: "http://localhost:8080/my_storage",
}

// Create storage manager
sm, err := storage.NewStorageManagerWithDefault(config)
if err != nil {
    log.Fatal(err)
}

// Save a file
err = sm.Save("courses/test-course/metadata.json", []byte(metadata))
if err != nil {
    log.Fatal(err)
}

// Load a file
data, err := sm.Load("courses/test-course/metadata.json")
if err != nil {
    log.Fatal(err)
}

// Get public URL
url := sm.GetURL("courses/test-course/metadata.json")
// Returns: "http://localhost:8080/my_storage/courses/test-course/metadata.json"
```

## API Endpoints

When running the server, files are accessible via the `/storage` endpoint:

```
GET http://localhost:8080/storage/courses/{course-id}/metadata.json
GET http://localhost:8080/storage/courses/{course-id}/lessons/{lesson-id}/video.mp4
```

## Migration Guide

### From Direct Filesystem to Storage System

If you're migrating from the old direct filesystem approach:

1. **Update your code**: Use the StorageInterface instead of direct file operations
2. **Configure storage**: Set appropriate environment variables
3. **Move existing files**: Copy existing course files to the new storage structure
4. **Update URLs**: Use storage.GetURL() to generate public URLs

### Example Migration

Before (direct filesystem):
```go
// Old way
filePath := filepath.Join(outputDir, courseID, "metadata.json")
err = os.WriteFile(filePath, []byte(metadata), 0644)
publicURL := fmt.Sprintf("http://localhost:8080/static/%s", filePath)
```

After (storage system):
```go
// New way
config := storage.DefaultStorageConfig()
sm, _ := storage.NewStorageManagerWithDefault(config)
coursePath := sm.GetCoursePath(courseID)
err = sm.Save(filepath.Join(coursePath, "metadata.json"), []byte(metadata))
publicURL := sm.GetURL(filepath.Join(coursePath, "metadata.json"))
```

## Best Practices

1. **Use the StorageInterface**: Always use the abstraction layer, not direct filesystem access
2. **Generate paths with helpers**: Use `GetCoursePath()` and `GetLessonPath()` for consistent structure
3. **Handle errors properly**: Check all storage operations for errors
4. **Use appropriate storage type**: Use local storage for development, S3 for production
5. **Set proper permissions**: Ensure the storage directory has appropriate read/write permissions

## Troubleshooting

### Storage Directory Not Created

If the storage directory isn't being created:
1. Check if the base path has write permissions
2. Verify STORAGE_BASE_PATH environment variable is set
3. Check logs for storage initialization errors

### S3 Connection Issues

If S3 storage isn't working:
1. Verify AWS credentials are properly configured
2. Check if the bucket exists and has correct permissions
3. Ensure AWS region is correct
4. Check network connectivity to S3 endpoints

### File Access Issues

If files aren't accessible via URLs:
1. Verify STORAGE_PUBLIC_URL is correctly set
2. Check if the server is serving the storage directory
3. Ensure file paths match the expected structure

## Performance Considerations

- **Local Storage**: Fast for development, limited scalability
- **S3 Storage**: Slower for small files, excellent scalability and durability
- **Concurrent Access**: Storage interface is thread-safe for concurrent operations

## Security Notes

- Store AWS credentials securely (use IAM roles when possible)
- Set appropriate bucket policies for S3
- Consider encryption for sensitive course content
- Validate file paths to prevent directory traversal