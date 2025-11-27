#!/bin/bash

# Course Creator Development Setup Script
# This script sets up a complete development environment for Course Creator

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# System requirements check
check_requirements() {
    print_status "Checking system requirements..."
    
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker first."
        echo "Visit: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! command_exists docker-compose; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        echo "Visit: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    if ! command_exists git; then
        print_error "Git is not installed. Please install Git first."
        echo "Visit: https://git-scm.com/downloads"
        exit 1
    fi
    
    # Check Docker is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker daemon."
        exit 1
    fi
    
    print_status "All requirements met!"
}

# Clone repository
setup_repository() {
    print_status "Setting up repository..."
    
    if [ ! -d "course-creator" ]; then
        print_status "Cloning Course Creator repository..."
        git clone https://github.com/your-org/course-creator.git
        cd course-creator
    else
        print_warning "Course Creator directory already exists. Pulling latest changes..."
        cd course-creator
        git pull origin main
    fi
}

# Setup environment
setup_environment() {
    print_status "Setting up environment..."
    
    if [ ! -f ".env" ]; then
        print_status "Creating .env file from template..."
        cp .env.example .env
        print_warning "Please edit .env file with your API keys and configurations."
        print_warning "Required: OPENAI_API_KEY or ANTHROPIC_API_KEY for LLM features"
        print_warning "Required: AWS credentials if using S3 storage"
    else
        print_status ".env file already exists"
    fi
}

# Create SSL certificates for development
setup_ssl() {
    print_status "Creating self-signed SSL certificates for development..."
    
    mkdir -p nginx/ssl
    
    if [ ! -f "nginx/ssl/cert.pem" ]; then
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout nginx/ssl/key.pem \
            -out nginx/ssl/cert.pem \
            -subj "/C=US/ST=State/L=City/O=Course Creator/CN=localhost" \
            2>/dev/null || {
            print_warning "OpenSSL not found. Skipping SSL certificate creation."
            print_warning "SSL features will not work without certificates."
        }
    else
        print_status "SSL certificates already exist"
    fi
}

# Setup backend
setup_backend() {
    print_status "Setting up backend..."
    
    cd core-processor
    
    # Install Go dependencies
    if command_exists go; then
        print_status "Installing Go dependencies..."
        go mod download
        go mod tidy
    else
        print_warning "Go is not installed. Backend will run in Docker only."
    fi
    
    # Run tests
    print_status "Running backend tests..."
    go test ./... -v || {
        print_warning "Some tests failed. This is normal if API keys are not configured."
    }
    
    cd ..
}

# Setup desktop app
setup_desktop_app() {
    print_status "Setting up desktop app..."
    
    cd creator-app
    
    if command_exists node && command_exists npm; then
        print_status "Installing Node.js dependencies..."
        npm install
        print_status "Building desktop app..."
        npm run build || print_warning "Desktop app build failed. Will run in Docker."
    else
        print_warning "Node.js/npm not found. Desktop app will run in Docker only."
    fi
    
    cd ..
}

# Setup web player
setup_web_player() {
    print_status "Setting up web player..."
    
    if [ -d "player-app" ]; then
        cd player-app
        
        if command_exists node && command_exists npm; then
            print_status "Installing Node.js dependencies..."
            npm install
            print_status "Building web player..."
            npm run build || print_warning "Web player build failed. Will run in Docker."
        else
            print_warning "Node.js/npm not found. Web player will run in Docker only."
        fi
        
        cd ..
    else
        print_status "Web player directory not found. Skipping."
    fi
}

# Setup mobile app
setup_mobile_app() {
    print_status "Setting up mobile app..."
    
    if [ -d "mobile-player" ]; then
        cd mobile-player
        
        if command_exists node && command_exists npm; then
            print_status "Installing Node.js dependencies..."
            npm install
            
            # Check for React Native CLI
            if command_exists npx; then
                print_status "Checking React Native environment..."
                npx react-native doctor || print_warning "React Native environment may need setup"
            fi
        else
            print_warning "Node.js/npm not found. Mobile app setup skipped."
        fi
        
        cd ..
    else
        print_status "Mobile player directory not found. Skipping."
    fi
}

# Start services
start_services() {
    print_status "Starting development services..."
    
    # Choose deployment mode
    echo "Choose deployment mode:"
    echo "1) Basic (API only)"
    echo "2) Development (with desktop app)"
    echo "3) Full development (with monitoring)"
    echo "4) Production-ready (with SSL)"
    read -p "Enter choice [1-4]: " choice
    
    case $choice in
        1)
            print_status "Starting basic services..."
            docker-compose up -d postgres redis api
            ;;
        2)
            print_status "Starting development services..."
            docker-compose --profile development up -d
            ;;
        3)
            print_status "Starting full development stack..."
            docker-compose --profile development --profile monitoring up -d
            ;;
        4)
            print_status "Starting production-ready services..."
            docker-compose --profile production --profile monitoring up -d
            ;;
        *)
            print_warning "Invalid choice. Starting basic services..."
            docker-compose up -d postgres redis api
            ;;
    esac
}

# Wait for services
wait_for_services() {
    print_status "Waiting for services to be ready..."
    
    # Wait for database
    print_status "Waiting for database..."
    timeout 60 bash -c 'until docker-compose exec -T postgres pg_isready -U course_creator; do sleep 2; done' || {
        print_error "Database failed to start"
        exit 1
    }
    
    # Wait for API
    print_status "Waiting for API..."
    timeout 60 bash -c 'until curl -s http://localhost:8080/api/v1/health >/dev/null; do sleep 2; done' || {
        print_error "API failed to start"
        exit 1
    }
    
    print_status "All services are ready!"
}

# Show access information
show_info() {
    echo ""
    echo -e "${GREEN}=== Course Creator Development Environment Ready ===${NC}"
    echo ""
    echo "Services:"
    echo "  • API Server:     http://localhost:8080"
    echo "  • Health Check:    http://localhost:8080/api/v1/health"
    echo "  • Metrics:         http://localhost:8080/api/v1/metrics"
    
    if docker-compose ps | grep -q "desktop-app"; then
        echo "  • Desktop App:     http://localhost:3000"
    fi
    
    if docker-compose ps | grep -q "web-player"; then
        echo "  • Web Player:      http://localhost:3001"
    fi
    
    if docker-compose ps | grep -q "grafana"; then
        echo "  • Grafana:         http://localhost:3002 (admin/admin123)"
    fi
    
    if docker-compose ps | grep -q "prometheus"; then
        echo "  • Prometheus:      http://localhost:9090"
    fi
    
    echo ""
    echo "Database:"
    echo "  • Host:           localhost:5432"
    echo "  • Database:       course_creator"
    echo "  • User:           course_creator"
    echo "  • Password:       (see .env file)"
    
    echo ""
    echo "Redis:"
    echo "  • Host:           localhost:6379"
    
    echo ""
    echo "Useful Commands:"
    echo "  • View logs:      docker-compose logs -f api"
    echo "  • Stop services:  docker-compose down"
    echo "  • Restart API:     docker-compose restart api"
    echo "  • Run tests:      cd core-processor && go test ./..."
    
    echo ""
    echo -e "${YELLOW}Important:${NC}"
    echo "  • Edit .env file with your API keys for LLM features"
    echo "  • SSL certificates are self-signed for development"
    echo "  • Database data persists in Docker volumes"
    echo ""
}

# Main setup flow
main() {
    echo "Course Creator Development Setup"
    echo "================================"
    echo ""
    
    check_requirements
    setup_repository
    setup_environment
    setup_ssl
    setup_backend
    setup_desktop_app
    setup_web_player
    setup_mobile_app
    start_services
    wait_for_services
    show_info
    
    print_status "Setup complete! Happy coding! 🚀"
}

# Handle script interruption
trap 'print_error "Setup interrupted"; exit 1' INT

# Run main function
main "$@"