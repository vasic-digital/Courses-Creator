# Course Creator Deployment Guide

This guide provides comprehensive instructions for deploying the Course Creator system in various environments.

## Prerequisites

### System Requirements
- Docker 20.10+
- Docker Compose 2.0+
- Minimum 4GB RAM
- 20GB+ storage space
- FFmpeg installed (for local deployments)
- Git

### External Dependencies
- OpenAI API key (optional, for LLM features)
- Anthropic API key (optional, for LLM features)
- AWS credentials (if using S3 storage)

## Quick Start

### 1. Clone Repository
```bash
git clone https://github.com/your-org/course-creator.git
cd course-creator
```

### 2. Environment Configuration
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Start Services
```bash
# Production deployment
docker-compose up -d

# Development with desktop app
docker-compose --profile development up -d

# With monitoring
docker-compose --profile monitoring up -d

# Full development stack
docker-compose --profile development --profile monitoring up -d
```

### 4. Verify Deployment
```bash
# Check health
curl http://localhost:8080/api/v1/health

# View logs
docker-compose logs -f api

# Access services
# API: http://localhost:8080
# Desktop App: http://localhost:3000
# Web Player: http://localhost:3001
# Grafana: http://localhost:3002
# Prometheus: http://localhost:9090
```

## Environment Variables

### Database Configuration
```bash
DB_HOST=postgres
DB_PORT=5432
DB_NAME=course_creator
DB_USER=course_creator
DB_PASSWORD=your_secure_password
```

### Redis Configuration
```bash
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
```

### LLM Provider Keys
```bash
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
OLLAMA_BASE_URL=http://localhost:11434
```

### Storage Configuration
```bash
# Local Storage
STORAGE_TYPE=local
STORAGE_PATH=/app/storage

# S3 Storage (alternative)
STORAGE_TYPE=s3
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1
S3_BUCKET=course-creator-content
```

### Authentication
```bash
JWT_SECRET=your_super_secret_jwt_key_at_least_32_chars
JWT_EXPIRE_TIME=24h
```

## Deployment Configurations

### Production (Docker Compose)
```bash
# Production stack
docker-compose up -d

# Production with SSL and monitoring
docker-compose --profile production --profile monitoring up -d
```

Features:
- PostgreSQL database
- Redis job queue
- Nginx reverse proxy
- SSL termination
- Rate limiting
- Monitoring with Prometheus/Grafana

### Development Stack
```bash
# Development with hot reload
docker-compose --profile development up -d
```

Includes:
- All production services
- Desktop app with hot reload
- Debug logging enabled
- Volume mounts for live code editing

### Monitoring Stack
```bash
# Monitoring only
docker-compose --profile monitoring up -d
```

Provides:
- Prometheus metrics collection
- Grafana dashboards
- Alerting capabilities
- Performance visualization

## SSL/TLS Configuration

### Self-Signed Certificate (Development)
```bash
# Generate self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
```

### Let's Encrypt (Production)
```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Generate certificate
sudo certbot --nginx -d yourdomain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
```

## Scaling Considerations

### Horizontal Scaling
```yaml
# docker-compose.scale.yml
services:
  api:
    deploy:
      replicas: 3
  worker:
    deploy:
      replicas: 5
```

### Database Scaling
- Use managed PostgreSQL (AWS RDS, Google Cloud SQL)
- Configure connection pooling
- Enable read replicas for high read loads

### Storage Scaling
- Use S3-compatible storage for media files
- Configure CDN (CloudFront, Fastly)
- Implement image optimization pipeline

## Monitoring and Maintenance

### Health Checks
```bash
# Service health
curl http://localhost:8080/api/v1/health

# Docker health
docker-compose ps

# Resource usage
docker stats
```

### Log Management
```bash
# View logs
docker-compose logs -f api

# Log rotation (in docker-compose.yml)
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### Backup Procedures
```bash
# Database backup
docker-compose exec postgres pg_dump -U course_creator course_creator > backup.sql

# Volume backup
docker run --rm -v course-creator_storage_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/storage-backup.tar.gz -C /data .

# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec postgres pg_dump -U course_creator course_creator > "backup_${DATE}.sql"
```

## Troubleshooting

### Common Issues

#### API Not Starting
```bash
# Check logs
docker-compose logs api

# Verify configuration
docker-compose config

# Check port conflicts
netstat -tulpn | grep 8080
```

#### Database Connection Issues
```bash
# Verify database is running
docker-compose exec postgres pg_isready

# Check credentials
docker-compose exec postgres psql -U course_creator -d course_creator

# Reset database
docker-compose down -v
docker-compose up -d postgres
```

#### High Memory Usage
```bash
# Monitor resource usage
docker stats

# Adjust worker count
# Edit docker-compose.yml
environment:
  - WORKER_COUNT=2

# Optimize Go runtime
GOMAXPROCS=2
```

#### Slow Performance
```bash
# Check metrics
curl http://localhost:9090/targets

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile

# Database optimization
docker-compose exec postgres psql -U course_creator -d course_creator -c "
EXPLAIN ANALYZE SELECT * FROM courses LIMIT 10;"
```

### Performance Tuning

#### Database Optimization
```sql
-- Create indexes
CREATE INDEX idx_courses_user_id ON courses(user_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);

-- Vacuum and analyze
VACUUM ANALYZE;
```

#### Redis Optimization
```bash
# Set max memory
redis-cli CONFIG SET maxmemory 1gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

#### Application Tuning
```bash
# Go runtime optimization
GOGC=100
GOMAXPROCS=4

# Concurrency settings
WORKER_COUNT=4
MAX_CONCURRENT_JOBS=10
```

## Security Best Practices

### Network Security
- Use HTTPS in production
- Implement firewall rules
- Restrict database access
- Use VPC/subnet isolation

### Application Security
- Rotate secrets regularly
- Use environment variables for sensitive data
- Enable audit logging
- Implement rate limiting

### Container Security
```bash
# Use non-root user
USER 1001:1001

# Read-only filesystem
read_only: true

# Security scanning
docker scan course-creator:latest

# Minimal base image
FROM alpine:3.18
```

## Upgrades and Maintenance

### Application Updates
```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
docker-compose build --no-cache
docker-compose up -d

# Migrate database
docker-compose exec api ./course-creator migrate
```

### Zero-Downtime Deployment
```bash
# Blue-green deployment
docker-compose -f docker-compose.blue.yml up -d
# Test new version
# Switch traffic
docker-compose -f docker-compose.green.yml down
```

### Rollback Procedures
```bash
# Quick rollback
git checkout previous-tag
docker-compose up -d

# Database rollback
docker-compose exec postgres psql -U course_creator -d course_creator < rollback.sql
```

## Support and Resources

### Documentation
- [API Reference](./docs/api/)
- [Architecture Guide](./docs/architecture/)
- [Contributing Guide](./CONTRIBUTING.md)

### Monitoring Dashboards
- Grafana: http://localhost:3002
- Prometheus: http://localhost:9090
- Alert Manager: http://localhost:9093

### Community
- GitHub Issues: Report bugs and feature requests
- Discussions: General questions and support
- Wiki: Additional documentation and examples

For production deployments requiring high availability, scalability, or custom configurations, consider using Kubernetes with the provided Helm charts (see k8s/ directory).