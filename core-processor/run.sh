#!/bin/bash

# Course Creator Setup Script
# This script helps set up the environment and run the Course Creator

set -e

echo "Course Creator Setup & Run Script"
echo "=============================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go (https://go.dev/) first."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Found Go version: $GO_VERSION"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ Created .env file. Please edit it with your configuration."
else
    echo "✅ .env file already exists"
fi

# Check if storage directory exists
if [ ! -d "./storage" ]; then
    echo "📁 Creating storage directory..."
    mkdir -p ./storage
    echo "✅ Created storage directory"
else
    echo "✅ Storage directory exists"
fi

# Check if data directory exists
if [ ! -d "./data" ]; then
    echo "📁 Creating data directory..."
    mkdir -p ./data
    echo "✅ Created data directory"
else
    echo "✅ Data directory exists"
fi

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy
echo "✅ Dependencies installed"

# Run tests
echo "🧪 Running tests..."
go test ./... -v
echo "✅ Tests completed"

# Ask what to do
echo ""
echo "What would you like to do?"
echo "1) Start API server"
echo "2) Generate course from markdown"
echo "3) Exit"
read -p "Choose option (1-3): " choice

case $choice in
    1)
        echo "🚀 Starting API server..."
        echo "Server will be available at: http://localhost:8080"
        echo "API endpoints: http://localhost:8080/api/v1/"
        echo "Storage files: http://localhost:8080/storage/"
        echo ""
        echo "Press Ctrl+C to stop the server"
        echo ""
        go run . server
        ;;
    2)
        echo "📚 Course Generation"
        read -p "Enter markdown file path: " markdown_file
        read -p "Enter output directory: " output_dir
        
        if [ ! -f "$markdown_file" ]; then
            echo "❌ Markdown file not found: $markdown_file"
            exit 1
        fi
        
        echo "🔄 Generating course..."
        go run . generate "$markdown_file" "$output_dir"
        echo "✅ Course generation completed"
        ;;
    3)
        echo "👋 Goodbye!"
        exit 0
        ;;
    *)
        echo "❌ Invalid option"
        exit 1
        ;;
esac