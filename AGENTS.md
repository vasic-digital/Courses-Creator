# Course Creator - Agent Guide

This guide provides essential information for AI agents working on the Course Creator repository, a comprehensive system that converts markdown scripts into professional video courses.

## Project Overview

Course Creator is a multi-platform system with:
- **Backend**: Go with Gin web framework
- **Desktop App**: Electron + React + TypeScript
- **Mobile App**: React Native + TypeScript
- **Core Features**: AI-powered TTS, visual enhancements, video processing

## Essential Commands

### Backend (core-processor/)
```bash
# Build and run the API server
cd core-processor
go mod tidy
go build .
./core-processor server

# Generate course from command line
./core-processor generate <markdown-file> <output-dir>

# Run tests
go test ./...
go test ./tests/...
go test ./tests/integration/
go test ./tests/contract/
go test ./tests/unit/
```

### Desktop App (creator-app/)
```bash
# Install dependencies
npm install

# Development mode
npm run dev

# Build for production
npm run build

# Start the Electron app
npm start

# Run tests
npm test

# Linting and formatting
npm run lint
npm run format
```

### Mobile App (mobile-player/)
```bash
# Install dependencies
npm install

# Start Metro bundler
npm start

# Run on iOS
npm run ios

# Run on Android
npm run android

# Run tests
npm test

# Linting and formatting
npm run lint
npm run format
```

## Project Structure

```
course-creator/
├── core-processor/          # Go backend and processing engine
│   ├── api/                 # REST API handlers
│   ├── cmd/                 # CLI commands
│   ├── mcp_servers/         # MCP server implementations
│   ├── models/              # Data models
│   ├── pipeline/            # Processing pipeline
│   ├── tests/               # Test suites (unit, integration, contract)
│   └── utils/               # Utilities
├── creator-app/             # Electron desktop application
│   └── src/
│       ├── main/            # Electron main process
│       ├── renderer/        # React UI
│       ├── components/      # Reusable React components
│       ├── services/        # API services
│       └── pages/           # Page components
├── mobile-player/           # React Native mobile app
│   └── src/
│       ├── screens/         # App screens
│       ├── services/        # API services
│       ├── components/      # Reusable components
│       └── types/           # TypeScript types
├── shared/                  # Shared TypeScript types
├── specs/                   # Project specifications and plans
├── docs/                    # Documentation
└── examples/                # Example markdown courses
```

## Code Patterns and Conventions

### Go Backend
- Use standard Go package structure
- Error handling with explicit error returns
- Gin framework for REST API
- Testify for testing with BDD style
- Follow Go conventions: camelCase for variables, PascalCase for exports

### TypeScript (Frontend/Mobile)
- Functional React components with TypeScript
- Strict TypeScript configuration enabled
- React Navigation for mobile navigation
- Async/await for API calls
- Consistent naming: PascalCase for components, camelCase for functions/variables

### Testing
- **Go**: Use testify with require/assert functions
- **TypeScript**: Jest for unit and integration tests
- Test files end with `_test.go` (Go) or `.test.ts` (TypeScript)
- Test structure: Given/When/Then in test descriptions

## Key APIs and Endpoints

### Backend API
- `GET /api/v1/health` - Health check
- `POST /api/v1/courses/generate` - Generate course from markdown
- `GET /api/v1/courses` - List all courses
- `GET /api/v1/courses/{id}` - Get specific course

### Processing Pipeline
1. **Markdown Parsing**: Extract structure and content
2. **TTS Generation**: Convert text to speech via MCP
3. **Video Assembly**: Combine audio, visuals, text overlays
4. **Post-Processing**: Add subtitles, background music

## MCP (Model Context Protocol) Integration

The system uses MCP for AI services:
- **Bark TTS**: High-quality text-to-speech
- **SpeechT5**: Alternative TTS engine
- **Suno**: Background music generation
- **LLaVA**: Image analysis
- **Pix2Struct**: UI parsing for diagrams

MCP servers are implemented in `core-processor/mcp_servers/` following the base server pattern.

## Data Models

Key shared types (defined in `shared/types/index.ts`):
- **Course**: Main course entity with lessons and metadata
- **Lesson**: Individual video segments with content and media
- **ProcessingOptions**: Configuration for course generation
- **ProcessingResult**: Output of course generation process

## Configuration Files

- **Go**: `core-processor/go.mod` (Go 1.21+)
- **TypeScript**: `creator-app/tsconfig.json`, mobile uses React Native defaults
- **Electron**: Uses Webpack for bundling (creator-app/webpack.config.js)
- **React Native**: Standard Metro bundler configuration

## Gotchas and Important Notes

### Test-Driven Development
- The constitution mandates 100% test coverage
- Tests must be written BEFORE implementation
- Follow Red-Green-Refactor cycle strictly

### Cross-Platform Considerations
- Use absolute file paths for all file operations
- Handle platform-specific path separators (use path module)
- Consider differences in video/audio formats across platforms

### AI Service Integration
- Always check for AI service availability before use
- Implement fallback mechanisms for paid services
- Handle rate limits and API failures gracefully

### File Management
- All generated content goes to `output/` directory
- Temporary files should be cleaned up after processing
- Large media files require efficient handling

## Development Workflow

1. **Spec-Driven**: Features start with specifications in `specs/`
2. **TDD First**: Write failing tests before implementation
3. **Component-Based**: Build reusable components
4. **API Integration**: Backend provides REST APIs to frontend apps
5. **Cross-Platform**: Test on all target platforms

## Dependencies and Prerequisites

### System Requirements
- Go 1.21+
- Node.js 18+
- FFmpeg for video processing
- React Native development environment (for mobile)

### Key Libraries
- **Go**: Gin, testify, MCP SDK (when available)
- **Electron**: React, TypeScript, Axios
- **Mobile**: React Native, React Navigation

## Documentation and Resources

- Project README contains setup instructions and API examples
- `specs/` contains detailed feature specifications
- `examples/` contains sample markdown course files
- `docs/` contains additional documentation

## Constitutional Principles

The project is governed by a constitution emphasizing:
- Multimedia quality excellence (1080p+, professional audio)
- Cross-platform compatibility
- Ethical AI integration
- Test-driven development (100% coverage)
- Accessibility and inclusivity
- Performance and scalability

Always align development decisions with these constitutional principles.