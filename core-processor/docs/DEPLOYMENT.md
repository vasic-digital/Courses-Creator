# Course Creator Production Deployment Guide

This guide covers deploying Course Creator in production with various storage backends.

## Overview

Course Creator is designed to be deployed as a containerized service with configurable storage backends and optional LLM/TTS providers.

## Prerequisites

- Docker and Docker Compose (recommended)
- Node.js 18+ (for development)
- Access to cloud storage (AWS S3, GCS, etc.)
- Optional: LLM API keys (OpenAI, Anthropic)
- Optional: TTS server (Bark, SpeechT5)

## Deployment Options

### 1. Docker Compose (Recommended)

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  course-creator:
    build: .
    ports:
      - "8080:8080"
    environment:
      # Server
      HOST: 0.0.0.0
      PORT: 8080
      
      # Database
      DB_TYPE: sqlite
      DB_PATH: /app/data/course_creator.db
      
      # Storage Configuration
      STORAGE_TYPE: s3  # or "local"
      STORAGE_BASE_PATH: courses/
      STORAGE_PUBLIC_URL: https://your-bucket.s3.amazonaws.com
      
      # AWS Credentials (if using S3)
      AWS_REGION: us-east-1
      AWS_ACCESS_KEY_ID: ${AWS_ACCESS_KEY_ID}
      AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY}
      
      # LLM Providers
      OPENAI_API_KEY: ${OPENAI_API_KEY}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY}
      
      # TTS Configuration
      TTS_SERVER_URL: http://bark-server:8765
      
      # Security
      JWT_SECRET: ${JWT_SECRET:-default-secret}
      
      # CORS
      CORS_ORIGINS: https://yourdomain.com,https://app.yourdomain.com
      
      # Environment
      ENVIRONMENT: production
      LOG_LEVEL: info
    
    volumes:
      - ./data:/app/data
      - ./storage:/app/storage
    
    depends_on:
      - bark-server

  bark-server:
    image: ghcr.io/suno-ai/bark:latest
    ports:
      - "8765:8765"
    environment:
      - CUDA_VISIBLE_DEVICES=0
    volumes:
      - ./bark-models:/app/models
```

### 2. Kubernetes Deployment

Create Kubernetes manifests:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: course-creator
spec:
  replicas: 3
  selector:
    matchLabels:
      app: course-creator
  template:
    metadata:
      labels:
        app: course-creator
    spec:
      containers:
      - name: course-creator
        image: course-creator:latest
        ports:
        - containerPort: 8080
        env:
        - name: STORAGE_TYPE
          value: "s3"
        - name: AWS_REGION
          value: "us-east-1"
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: access-key
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: aws-credentials
              key: secret-key
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-keys
              key: openai
```

### 3. Manual Deployment

```bash
# Build application
go build -o course-creator

# Set environment variables
export STORAGE_TYPE=s3
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret

# Run application
./course-creator server
```

## Storage Configuration

### AWS S3 Setup

1. Create S3 bucket:
```bash
aws s3api create-bucket \
    --bucket course-creator-storage \
    --region us-east-1
```

2. Configure bucket policy:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadGetObject",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::course-creator-storage/*"
        }
    ]
}
```

3. Enable CORS on bucket:
```bash
aws s3api put-bucket-cors \
    --bucket course-creator-storage \
    --cors-configuration file://s3-cors.json
```

### Google Cloud Storage Setup

1. Create GCS bucket:
```bash
gsutil mb gs://course-creator-storage
```

2. Make objects public:
```bash
gsutil iam ch allUsers:objectViewer gs://course-creator-storage
```

## Security Configuration

### HTTPS Setup (Nginx)

```nginx
server {
    listen 443 ssl;
    server_name api.yourdomain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location /api/v1/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /storage/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
    }
}
```

### API Gateway Setup

Using AWS API Gateway for API protection:

```yaml
Resources:
  CourseCreatorAPI:
    Type: AWS::Serverless::Api
    Properties:
      StageName: prod
      Cors:
        AllowMethods: "'GET,POST,PUT,DELETE,OPTIONS'"
        AllowHeaders: "'Content-Type,Authorization'"
        AllowOrigin: "'https://yourdomain.com'"
```

## Monitoring and Logging

### Prometheus Metrics

Add to your application:
```go
// Prometheus metrics collection
var (
    coursesGenerated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "courses_generated_total",
            Help: "Total number of courses generated",
        },
        []string{"status"},
    )
    
    storageOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "storage_operations_total",
            Help: "Total storage operations",
        },
        []string{"operation", "backend"},
    )
)
```

### Log Aggregation

Configure structured logging:
```yaml
# docker-compose.yml
services:
  course-creator:
    environment:
      LOG_FORMAT: json
      LOG_LEVEL: info
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Performance Optimization

### Caching

Add Redis for caching:
```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  
  course-creator:
    environment:
      REDIS_URL: redis://redis:6379
      CACHE_TTL: 3600
```

### CDN Setup

Use CloudFront with S3:
```bash
aws cloudfront create-distribution \
    --distribution-config file://cloudfront-config.json
```

## Backup and Recovery

### Database Backup

```bash
# SQLite backup
sqlite3 /app/data/course_creator.db ".backup backup-$(date +%Y%m%d).db"

# S3 backup
aws s3 sync /app/data/ s3://backup-bucket/course-creator/
```

### Automated Backups

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)

# Backup database
sqlite3 /app/data/course_creator.db ".backup /backup/db_$DATE.db"

# Backup storage
tar -czf /backup/storage_$DATE.tar.gz /app/storage/

# Upload to S3
aws s3 cp /backup/db_$DATE.db s3://backup-bucket/
aws s3 cp /backup/storage_$DATE.tar.gz s3://backup-bucket/

# Cleanup old backups (keep 30 days)
find /backup -name "*.db" -mtime +30 -delete
find /backup -name "*.tar.gz" -mtime +30 -delete
```

## Scaling Considerations

### Horizontal Scaling

- Use load balancer for multiple instances
- Share storage via S3 for consistency
- Use external database for state

### Resource Limits

```yaml
# kubernetes/deployment.yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## Troubleshooting

### Common Production Issues

1. **Storage Access Errors**
   ```bash
   # Check S3 permissions
   aws s3 ls s3://your-bucket/
   aws s3api get-bucket-policy --bucket your-bucket
   ```

2. **Database Lock Issues**
   ```bash
   # Check for SQLite locks
   lsof /app/data/course_creator.db
   ```

3. **Memory Leaks**
   ```bash
   # Monitor memory usage
   docker stats course-creator
   ```

4. **Performance Issues**
   ```bash
   # Profile application
   go tool pprof http://localhost:8080/debug/pprof/profile
   ```

## Health Checks

### Readiness Probe

```go
// In main.go
func healthCheck(w http.ResponseWriter, r *http.Request) {
    // Check database
    if db == nil {
        http.Error(w, "Database not connected", http.StatusServiceUnavailable)
        return
    }
    
    // Check storage
    if storageManager == nil {
        http.Error(w, "Storage not available", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### Liveness Probe

```yaml
# Kubernetes
livenessProbe:
  httpGet:
    path: /api/v1/health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

## Migration Guide

### From Version 1.0 to 2.0

1. Update Docker image
2. Run database migration
3. Migrate storage files
4. Update environment variables

```bash
# Migration script
./migrate.sh --from=1.0 --to=2.0
```