# Course Creator Examples

This directory contains example courses that demonstrate the Course Creator system's capabilities. Each example is a complete markdown file that can be converted into a professional video course.

## Available Examples

### Programming Courses
- **[go-programming.md](go-programming.md)** - Complete Go programming tutorial (26 lessons)
- **[python-basics.md](python-basics.md)** - Python fundamentals for beginners
- **[javascript-essentials.md](javascript-essentials.md)** - JavaScript core concepts

### Web Development
- **[html-css-fundamentals.md](html-css-fundamentals.md)** - HTML and CSS basics
- **[react-quickstart.md](react-quickstart.md)** - React.js introduction

### Data Science
- **[data-analysis-python.md](data-analysis-python.md)** - Data analysis with Python
- **[machine-learning-intro.md](machine-learning-intro.md)** - ML fundamentals

### Quick Examples
- **[hello-world.md](hello-world.md)** - Minimal 3-lesson example
- **[quick-demo.md](quick-demo.md)** - Single lesson demonstration

## How to Generate Courses

### Using the Desktop App
1. Open the Course Creator desktop application
2. Select any `.md` file from this directory
3. Choose output directory
4. Configure options (voice, quality, background music)
5. Click "Generate Course"

### Using the Command Line
```bash
cd core-processor

# Generate Go programming course
go run main.go generate ../examples/go-programming.md ../output/go-course

# Generate Python basics course
go run main.go generate ../examples/python-basics.md ../output/python-course

# Generate quick demo
go run main.go generate ../examples/quick-demo.md ../output/demo
```

### Using the API
```bash
# Start the server
go run main.go server

# In another terminal, generate course
curl -X POST http://localhost:8080/api/v1/courses/generate \
  -H "Content-Type: application/json" \
  -d '{
    "markdown_path": "examples/python-basics.md",
    "output_dir": "output/python-course",
    "options": {
      "voice": "bark",
      "backgroundMusic": true,
      "quality": "standard"
    }
  }'
```

## Example Output Structure

After generation, each course will have:
```
output/course-name/
├── video_*.mp4          # Individual lesson videos
├── bg_*.png            # Background images
└── course.json         # Course metadata
```

## Customization Options

### Voice Selection
- `bark` - High-quality AI voice (default)
- `speecht5` - Alternative AI voice

### Quality Settings
- `standard` - 1080p, good quality
- `high` - Higher quality settings

### Audio Features
- Background music mixing
- Multiple language support
- Custom voice presets

## Tips for Creating Your Own Courses

1. **Structure**: Use `#` for title, `##` for lesson headings
2. **Content**: Keep lessons focused and concise
3. **Code**: Use triple backticks for code blocks
4. **Images**: Reference images with `![alt](url)` syntax
5. **Metadata**: Add course description after title

Example structure:
```markdown
# Course Title

Course description here.

## Lesson 1: Introduction

Lesson content...

## Lesson 2: Advanced Topics

More content...
```

## Troubleshooting

### FFmpeg Not Available
If FFmpeg is not installed, the system will create placeholder video files. Install FFmpeg for full video generation:

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt install ffmpeg

# Windows (via Chocolatey)
choco install ffmpeg
```

### API Server Issues
Ensure the Go server is running on port 8080:
```bash
cd core-processor && go run main.go server
```

### File Permissions
Ensure read/write permissions for input files and output directories.