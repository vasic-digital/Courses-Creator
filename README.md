# Course Creator

A comprehensive video course creation system that converts markdown scripts into professional video courses with AI-powered audio and visual enhancements.

## Features

### Core Functionality
- **Markdown to Video**: Convert markdown scripts into video courses
- **AI-Powered TTS**: Text-to-speech using Bark and SpeechT5 models
- **Visual Enhancements**: Dynamic backgrounds, text overlays, and illustrations
- **Audio Features**: Background music mixing and high-quality narration
- **Cross-Platform**: Desktop (Electron) and mobile (React Native) players

### Technical Stack
- **Backend**: Go with Gin web framework
- **MCP Integration**: Model Context Protocol for AI services
- **Video Processing**: FFmpeg for video assembly
- **Desktop App**: Electron + React + TypeScript
- **Mobile App**: React Native + TypeScript
- **Testing**: Go testing with testify

## Project Structure

```
course-creator/
├── core-processor/          # Go backend and processing engine
│   ├── api/                 # REST API handlers
│   ├── cmd/                 # CLI commands
│   ├── mcp_servers/         # MCP server implementations
│   ├── models/              # Data models
│   ├── pipeline/            # Processing pipeline
│   ├── tests/               # Test suites
│   └── utils/               # Utilities
├── creator-app/             # Electron desktop application
│   └── src/
│       ├── main/            # Electron main process
│       └── renderer/        # React UI
├── mobile-player/           # React Native mobile app
│   └── src/
│       ├── screens/         # App screens
│       ├── services/        # API services
│       └── types/           # TypeScript types
├── shared/                  # Shared TypeScript types
├── specs/                   # Project specifications
└── README.md
```

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- FFmpeg (optional, for video processing)
- React Native development environment (for mobile)

### Backend Setup

```bash
cd core-processor
go mod tidy
go build .
./core-processor server
```

The API server will start on http://localhost:8080

### Desktop App Setup

```bash
cd creator-app
npm install
npm run build
npm start
```

### Mobile App Setup

```bash
cd mobile-player
npm install
# For iOS
npm run ios
# For Android
npm run android
```

## Usage

### Creating a Course

1. **Write Markdown**: Create a course script in markdown format
2. **Use Desktop App**: Open the creator app, select your markdown file
3. **Configure Options**: Choose voice, quality, background music
4. **Generate**: Click generate to create the video course
5. **Play**: Use the mobile or desktop player to view the course

### Markdown Format

```markdown
# Course Title

This is the course description.

## Introduction

Welcome to the course!

## Main Content

This is the main content section.
```

## API Reference

### Generate Course
```http
POST /api/v1/courses/generate
Content-Type: application/json

{
  "markdown_path": "/path/to/course.md",
  "output_dir": "/path/to/output",
  "options": {
    "voice": "bark",
    "backgroundMusic": true,
    "languages": ["en"],
    "quality": "standard"
  }
}
```

### Get Courses
```http
GET /api/v1/courses
```

### Get Course
```http
GET /api/v1/courses/{id}
```

## Architecture

### Processing Pipeline
1. **Markdown Parsing**: Extract structure, content, and metadata
2. **TTS Generation**: Convert text to speech using MCP servers
3. **Video Assembly**: Combine audio, visuals, and text overlays
4. **Post-Processing**: Add subtitles, background music, final packaging

### MCP Servers
- **Bark TTS**: High-quality text-to-speech
- **SpeechT5**: Alternative TTS engine
- **Suno**: Background music generation
- **LLaVA**: Image analysis and description
- **Pix2Struct**: UI parsing for diagrams

## Development

### Running Tests
```bash
cd core-processor
go test ./...
```

### Building
```bash
# Backend
cd core-processor && go build .

# Desktop app
cd creator-app && npm run build

# Mobile app
cd mobile-player && npm run android # or ios
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] Real-time video preview
- [ ] Advanced video editing tools
- [ ] Multi-language support
- [ ] Cloud storage integration
- [ ] Collaborative editing
- [ ] Analytics and engagement tracking