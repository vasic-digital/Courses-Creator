# Course Creator Development Setup

This guide will help you set up the Course Creator application for development and production.

## Prerequisites

- Go 1.21+
- Node.js 16+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose
- Make (optional but recommended)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd Course-Creator
make dev-setup
```

### 2. Configure Environment

Copy `.env.example` to `.env` and update with your configuration:

```bash
cp .env.example .env
```

Key settings to update:
- `DB_PASSWORD` - Set a secure database password
- `JWT_SECRET` - Generate a strong secret for JWT tokens
- `OPENAI_API_KEY` - Add your OpenAI API key for AI features
- `ANTHROPIC_API_KEY` - Add your Anthropic API key

### 3. Start Services

#### Option A: Docker (Recommended)

```bash
# Build and start all services
make docker-up

# Or use docker-compose directly
docker-compose up -d
```

#### Option B: Local Development

```bash
# Start database and Redis with Docker
docker-compose up -d postgres redis

# Start the API server
make run
```

### 4. Setup Database

Run database migrations:

```bash
make migrate-up
```

### 5. Start Frontend Applications

```bash
# Start creator app (Electron)
cd creator-app
npm run dev

# Start player app (Web)
cd ../player-app
npm start

# Start mobile app (React Native)
cd ../mobile-player
npm run start
```

## Application URLs

After starting services:

- API Server: http://localhost:8080
- Creator App: http://localhost:3000
- Player App: http://localhost:3001
- Grafana: http://localhost:3002
- Prometheus: http://localhost:9090

## API Documentation

The API documentation is available at:
- Swagger UI: http://localhost:8080/swagger/index.html
- OpenAPI Spec: http://localhost:8080/swagger/doc.json

## Development Workflows

### Running Tests

```bash
# Backend tests
make test

# With coverage
make test-cover

# Frontend tests
cd creator-app && npm test
cd player-app && npm test
```

### Code Quality

```bash
# Lint and format Go code
make lint
make format

# Frontend linting
cd creator-app && npm run lint
cd player-app && npm run lint
```

### Database Management

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
# Create SQL files in core-processor/migrations/ directory
```

## Architecture

### Backend (Go)
- **Location**: `core-processor/`
- **Framework**: Gin
- **ORM**: GORM
- **Auth**: JWT tokens with refresh support

### Frontend
- **Creator App**: Electron + React
- **Player App**: React SPA
- **Mobile App**: React Native

### Infrastructure
- **Database**: PostgreSQL
- **Cache**: Redis
- **Queue**: Redis based job queue
- **File Storage**: Local/S3 compatible
- **Monitoring**: Prometheus + Grafana

## Security Features

- JWT-based authentication with refresh tokens
- Role-based access control (RBAC)
- Rate limiting
- CORS protection
- SQL injection prevention
- XSS protection
- File upload security
- HTTPS with secure headers

## Deployment

### Docker Deployment

```bash
# Build for production
docker-compose -f docker-compose.prod.yml build

# Deploy
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes

Kubernetes manifests are in `k8s/` directory.

### Environment Variables

See `.env.example` for all available configuration options.

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Ensure PostgreSQL is running
   - Check DB_PASSWORD in .env
   - Run `make migrate-up` if tables are missing

2. **Permission Errors**
   - Ensure storage directories exist and are writable
   - Check file permissions in storage/

3. **API Key Errors**
   - Verify OPENAI_API_KEY and ANTHROPIC_API_KEY are set
   - Check that API keys are valid and have credits

4. **Docker Issues**
   - Ensure Docker is running
   - Check for port conflicts
   - Use `docker-compose logs [service]` to debug

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License.