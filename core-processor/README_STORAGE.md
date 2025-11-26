# Course Creator Storage System Integration

## Overview

The Course Creator now features a comprehensive storage abstraction layer that supports multiple storage backends and provides a unified interface for file operations.

## Quick Start

### 1. Setup Environment
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 2. Run Application
```bash
# Using setup script
./run.sh

# Or directly
go run . server
```

### 3. Generate Course
```bash
# Using demo script
./demo.sh

# Or directly
go run . generate course.md output/
```

## Storage Features

### ✅ Implemented
- **Multiple Storage Backends**: Local filesystem, AWS S3
- **Unified Interface**: Single StorageInterface for all operations
- **Path Management**: Automatic organization of courses and lessons
- **URL Generation**: Public URLs for stored files
- **Configuration**: Environment-based configuration
- **Error Handling**: Graceful fallbacks and error recovery

### 📁 File Organization
```
storage/
└── courses/
    └── {course-id}/
        ├── metadata.json
        └── lessons/
            └── {lesson-id}/
                ├── content.md
                ├── audio.mp3
                └── video.mp4
```

## Configuration Options

### Local Storage (Default)
```bash
STORAGE_TYPE=local
STORAGE_BASE_PATH=./storage
STORAGE_PUBLIC_URL=http://localhost:8080/storage
```

### S3 Storage
```bash
STORAGE_TYPE=s3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
STORAGE_PUBLIC_URL=https://bucket.s3.amazonaws.com
```

## API Endpoints

- **Server**: http://localhost:8080
- **API**: http://localhost:8080/api/v1/
- **Storage**: http://localhost:8080/storage/
- **Health**: http://localhost:8080/api/v1/health

## Storage Operations

### Save File
```go
err := storageManager.Save("courses/test/metadata.json", data)
```

### Load File
```go
data, err := storageManager.Load("courses/test/metadata.json")
```

### Generate URL
```go
url := storageManager.GetURL("courses/test/video.mp4")
// Returns: "http://localhost:8080/storage/courses/test/video.mp4"
```

### Path Helpers
```go
coursePath := storageManager.GetCoursePath("course-123")
// Returns: "courses/course-123"

lessonPath := storageManager.GetLessonPath("course-123", "lesson-456")
// Returns: "courses/course-123/lessons/lesson-456"
```

## Testing

### Run All Tests
```bash
go test ./... -v
```

### Storage Tests
```bash
go test ./filestorage -v
```

### Demo Script
```bash
./demo.sh
```

## Migration Guide

### From Previous Version
1. Update imports to use `filestorage` package
2. Replace direct file operations with StorageInterface
3. Use path helpers for consistent structure
4. Update configuration to use environment variables

### Example Migration
```go
// Before
filePath := filepath.Join(outputDir, courseID, "video.mp4")
os.WriteFile(filePath, data, 0644)

// After
path := storageManager.GetLessonPath(courseID, lessonID)
storageManager.Save(filepath.Join(path, "video.mp4"), data)
```

## Troubleshooting

### Common Issues
1. **Storage directory not created**: Check write permissions
2. **S3 connection fails**: Verify AWS credentials
3. **Files not accessible**: Check STORAGE_PUBLIC_URL
4. **Build errors**: Ensure package imports are updated

### Debug Tips
- Check `generation.log` for course generation details
- Use `DB_DEBUG=true` for database queries
- Set `LOG_LEVEL=debug` for verbose logging

## Production Deployment

### Recommended Settings
1. **Use S3 storage** for scalability
2. **Set proper bucket permissions**
3. **Configure CORS origins**
4. **Set up monitoring**
5. **Use environment variables** for secrets

### Security Considerations
- Secure AWS credentials
- Validate file paths
- Set appropriate permissions
- Use HTTPS in production

## Documentation

- [Detailed Storage Guide](docs/STORAGE.md)
- [Configuration](.env.example)
- [API Documentation](api/handlers.go)

## Contributing

To contribute to storage system:
1. Implement StorageInterface for new provider
2. Add configuration options
3. Write comprehensive tests
4. Update documentation

## Support

For issues with storage system:
1. Check logs for error messages
2. Verify configuration
3. Run tests to isolate problem
4. Check documentation for best practices