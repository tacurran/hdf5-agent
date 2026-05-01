#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        echo "Install from: https://golang.org/doc/install"
        exit 1
    fi
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go $GO_VERSION found"
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed"
        echo "Install from: https://nodejs.org/"
        exit 1
    fi
    NODE_VERSION=$(node --version)
    print_success "Node.js $NODE_VERSION found"
    
    # Check HDF5
    if ! pkg-config --exists hdf5; then
        print_error "HDF5 development libraries not found"
        echo ""
        echo "Install HDF5 libraries:"
        echo "  macOS:   brew install hdf5"
        echo "  Ubuntu:  sudo apt-get install libhdf5-dev"
        echo "  CentOS:  sudo yum install hdf5-devel"
        exit 1
    fi
    HDF5_VERSION=$(pkg-config --modversion hdf5)
    print_success "HDF5 $HDF5_VERSION found"
}

# Setup
setup() {
    print_header "Setting Up"
    
    # Create data directory
    if [ ! -d "data" ]; then
        mkdir -p data
        print_success "Created data directory"
    fi
    
    # Download Go dependencies
    print_info "Downloading Go dependencies..."
    go mod download
    print_success "Go dependencies ready"
    
    # Install npm dependencies
    print_info "Installing frontend dependencies..."
    cd frontend
    npm install > /dev/null 2>&1
    print_success "Frontend dependencies installed"
    cd ..
}

# Build
build() {
    print_header "Building"
    
    print_info "Building backend..."
    go build -o hdf5-agent
    print_success "Backend built: ./hdf5-agent"
    
    print_info "Building frontend..."
    cd ./frontend
    npm run build > /dev/null 2>&1
    print_success "Frontend built: ./frontend/dist/"
    cd ..
}

# Development server
dev() {
    print_header "Starting Development Environment"
    
    # Create data directory
    mkdir -p data
    cd data
    
    print_info "Backend running on: http://localhost:8080"
    print_info "Frontend running on: http://localhost:3000"
    print_info ""
    print_info "Press Ctrl+C in either terminal to stop"
    echo ""
    
    # Start backend and frontend in parallel
    (
        cd ..
        print_info "Starting backend..."
        go run main.go
    ) &
    BACKEND_PID=$!
    
    sleep 2
    
    (
        cd ./frontend
        print_info "Starting frontend..."
        npm run dev
    ) &
    FRONTEND_PID=$!
    
    # Handle Ctrl+C
    trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; echo ''; print_info 'Stopped both servers'" EXIT
    
    # Wait for any process to exit
    wait
}

# Docker
docker_dev() {
    print_header "Starting with Docker Compose"
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    
    docker-compose up
}

# Production build
production() {
    print_header "Building Production Container"
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    
    print_info "Building Docker image..."
    docker build -t hdf5-agent .
    print_success "Docker image built: hdf5-agent"
    
    echo ""
    print_info "Run with:"
    echo "  docker run -p 8080:8080 -v \$(pwd)/data:/root/data hdf5-agent"
}

# Main
main() {
    case "${1:-dev}" in
        check)
            check_prerequisites
            ;;
        setup)
            check_prerequisites
            setup
            ;;
        build)
            build
            ;;
        dev)
            check_prerequisites
            setup
            dev
            ;;
        docker)
            docker_dev
            ;;
        prod)
            production
            ;;
        *)
            print_header "HDF5 Agent - Startup Script"
            echo ""
            echo "Usage: ./run.sh [command]"
            echo ""
            echo "Commands:"
            echo "  check      - Check prerequisites"
            echo "  setup      - Setup dependencies"
            echo "  build      - Build both backend and frontend"
            echo "  dev        - Run development servers (default)"
            echo "  docker     - Run with Docker Compose"
            echo "  prod       - Build production Docker image"
            echo ""
            exit 1
            ;;
    esac
}

main "$@"
