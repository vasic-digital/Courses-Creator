#!/bin/bash

# Course Creator Demo Script
# Demonstrates complete course generation workflow

echo "Course Creator Demo"
echo "=================="

# Create demo markdown course
echo "📝 Creating demo course..."
cat > demo_course.md << 'EOF'
# Introduction to Web Development

This course covers the fundamentals of web development including HTML, CSS, and JavaScript.

## Module 1: HTML Basics
HTML (HyperText Markup Language) is the standard markup language for creating web pages. It describes the structure of a web page semantically.

### Key Concepts
- Elements and tags
- Attributes
- Document structure
- Semantic HTML

## Module 2: CSS Styling
CSS (Cascading Style Sheets) is used to style and layout web pages. It controls colors, fonts, spacing, and positioning.

### Key Concepts
- Selectors and properties
- Box model
- Flexbox
- Grid layout

## Module 3: JavaScript Programming
JavaScript adds interactivity to web pages. It's a programming language that runs in web browsers.

### Key Concepts
- Variables and data types
- Functions
- DOM manipulation
- Event handling
EOF

echo "✅ Demo course created: demo_course.md"

# Create output directory
echo "📁 Creating output directory..."
mkdir -p demo_output

# Generate course (this will show TTS errors which is expected)
echo "🔄 Generating course..."
echo "Note: You'll see TTS errors - this is normal if Bark isn't installed"
echo ""

# Run generation in background to capture output
go run . generate demo_course.md demo_output 2>&1 | tee generation.log

# Check results
echo ""
echo "📊 Generation Results:"
echo "===================="

if [ -d "storage" ]; then
    echo "✅ Storage directory created"
    echo "Storage structure:"
    find storage -type f 2>/dev/null | while read file; do
        echo "  - $file"
    done
else
    echo "❌ Storage directory not found"
fi

if [ -d "demo_output" ]; then
    echo "✅ Output directory created"
    echo "Output contents:"
    find demo_output -type f 2>/dev/null | while read file; do
        echo "  - $file"
    done
else
    echo "❌ Output directory not found"
fi

# Check database
echo ""
echo "💾 Database Check:"
if [ -f "data/course_creator.db" ]; then
    echo "✅ Database created"
    # Show database stats if sqlite3 is available
    if command -v sqlite3 &> /dev/null; then
        echo "Database tables:"
        sqlite3 data/course_creator.db ".tables"
    fi
else
    echo "❌ Database not found"
fi

# Show storage API URLs
echo ""
echo "🌐 Storage API URLs (when server is running):"
echo "=========================================="
echo "Storage base: http://localhost:8080/storage/"
echo "Course files would be accessible at:"
echo "  http://localhost:8080/storage/courses/{course-id}/"
echo "  http://localhost:8080/storage/courses/{course-id}/lessons/{lesson-id}/video.mp4"

# Cleanup options
echo ""
echo "🧹 Cleanup Options:"
echo "1) Keep all files"
echo "2) Remove generated files"
read -p "Choose option (1-2): " cleanup_choice

case $cleanup_choice in
    2)
        echo "🗑️  Cleaning up..."
        rm -rf demo_course.md demo_output storage generation.log
        echo "✅ Cleanup completed"
        ;;
    *)
        echo "✅ Keeping all files"
        echo "Files generated:"
        echo "  - demo_course.md (input)"
        echo "  - demo_output/ (output)"
        echo "  - storage/ (storage data)"
        echo "  - generation.log (generation log)"
        ;;
esac

echo ""
echo "Demo completed! 🎉"
echo ""
echo "Next steps:"
echo "1. Install Bark TTS if you want audio generation"
echo "2. Configure LLM providers for enhanced content"
echo "3. Run 'go run . server' to start API server"
echo "4. Visit http://localhost:8080/api/v1/ for API endpoints"