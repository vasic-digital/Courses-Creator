#!/bin/bash

# Course Creator - Generate Example Courses Script
# This script generates video courses from all example markdown files

set -e  # Exit on any error

echo "🎬 Course Creator - Example Course Generator"
echo "=============================================="

# Check if core-processor is available
if [ ! -d "core-processor" ]; then
    echo "❌ Error: core-processor directory not found!"
    echo "Please run this script from the project root directory."
    exit 1
fi

# Create output directory for examples
OUTPUT_DIR="examples-output"
mkdir -p "$OUTPUT_DIR"

echo "📁 Created output directory: $OUTPUT_DIR"

# Function to generate course
generate_course() {
    local markdown_file="$1"
    local course_name="$2"

    echo ""
    echo "🎯 Generating course: $course_name"
    echo "📄 Source file: $markdown_file"

    local course_output="$OUTPUT_DIR/$course_name"

    # Generate the course
    if cd core-processor && go run main.go generate "../$markdown_file" "../$course_output"; then
        echo "✅ Successfully generated course: $course_name"
        echo "📂 Output location: $course_output"

        # Count generated files
        local video_count=$(find "../$course_output" -name "*.mp4" | wc -l)
        local bg_count=$(find "../$course_output" -name "*.png" | wc -l)

        echo "📊 Generated $video_count video files and $bg_count background images"
    else
        echo "❌ Failed to generate course: $course_name"
        return 1
    fi
}

# Generate all example courses
echo ""
echo "🚀 Starting course generation..."
echo ""

# Quick examples
generate_course "examples/hello-world.md" "hello-world"
generate_course "examples/quick-demo.md" "quick-demo"

# Programming courses
generate_course "examples/python-basics.md" "python-basics"
generate_course "examples/javascript-essentials.md" "javascript-essentials"
generate_course "examples/go-programming.md" "go-programming"

# Web development
generate_course "examples/html-css-fundamentals.md" "html-css-fundamentals"
generate_course "examples/react-quickstart.md" "react-quickstart"

# Data science
generate_course "examples/data-analysis-python.md" "data-analysis-python"
generate_course "examples/machine-learning-intro.md" "machine-learning-intro"

echo ""
echo "🎉 Course generation completed!"
echo ""
echo "📋 Summary:"
echo "=========="

# Generate summary
echo "Generated courses:"
for dir in "$OUTPUT_DIR"/*/; do
    if [ -d "$dir" ]; then
        course_name=$(basename "$dir")
        video_count=$(find "$dir" -name "*.mp4" | wc -l)
        bg_count=$(find "$dir" -name "*.png" | wc -l)
        echo "  • $course_name: $video_count videos, $bg_count backgrounds"
    fi
done

total_courses=$(find "$OUTPUT_DIR" -mindepth 1 -maxdepth 1 -type d | wc -l)
total_videos=$(find "$OUTPUT_DIR" -name "*.mp4" | wc -l)
total_backgrounds=$(find "$OUTPUT_DIR" -name "*.png" | wc -l)

echo ""
echo "📈 Totals:"
echo "  • Courses: $total_courses"
echo "  • Videos: $total_videos"
echo "  • Backgrounds: $total_backgrounds"

echo ""
echo "💡 Next steps:"
echo "  • Start the API server: cd core-processor && go run main.go server"
echo "  • Launch the desktop app: cd creator-app && npm start"
echo "  • Run the mobile app: cd mobile-player && npm run android"
echo ""
echo "🎬 Happy course creating!"