#!/bin/bash

# Aether Development Script
# This script helps with local development using docker-compose

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[STATUS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if required tools are installed
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check if docker-compose is installed
    if ! command -v docker-compose &> /dev/null; then
        print_error "docker-compose is not installed. Please install docker-compose first."
        exit 1
    fi
    
    print_status "All prerequisites are met!"
}

# Function to build and start services
start_services() {
    print_status "Starting Aether services..."
    
    # Check if .env file exists in synaptic-core
    if [ ! -f "./synaptic-core/.env" ]; then
        print_warning "No .env file found in synaptic-core. Creating one from example..."
        if [ -f "./synaptic-core/.env.example" ]; then
            cp ./synaptic-core/.env.example ./synaptic-core/.env
            print_status "Created .env file from .env.example"
        else
            touch ./synaptic-core/.env
            print_warning "Created empty .env file. Please add your GEMINI_API_KEY."
        fi
    fi
    
    # Build and start services
    docker-compose up --build -d
    
    print_status "Services started successfully!"
    print_status "Frontend: http://localhost:3000"
    print_status "Backend API: http://localhost:8080"
    print_status "Backend WebSocket: ws://localhost:8080/ws"
}

# Function to stop services
stop_services() {
    print_status "Stopping Aether services..."
    docker-compose down
    print_status "Services stopped successfully!"
}

# Function to view logs
view_logs() {
    print_status "Viewing logs for all services..."
    docker-compose logs -f
}

# Function to restart services
restart_services() {
    print_status "Restarting Aether services..."
    docker-compose restart
    print_status "Services restarted successfully!"
}

# Function to show service status
show_status() {
    print_status "Showing service status..."
    docker-compose ps
}

# Function to run tests
run_tests() {
    print_status "Running tests..."
    # Run backend tests
    if [ -d "./synaptic-core" ]; then
        print_status "Running backend tests..."
        docker-compose exec synaptic-core go test ./...
    fi
    
    # Run frontend tests
    if [ -d "./interactive-carapace" ]; then
        print_status "Running frontend tests..."
        docker-compose exec interactive-carapace pnpm test
    fi
}

# Main script logic
case "$1" in
    start)
        check_prerequisites
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        check_prerequisites
        start_services
        ;;
    logs)
        view_logs
        ;;
    status)
        show_status
        ;;
    test)
        run_tests
        ;;
    *)
        echo "Aether Development Script"
        echo "Usage: $0 {start|stop|restart|logs|status|test}"
        echo ""
        echo "Commands:"
        echo "  start   - Build and start all services"
        echo "  stop    - Stop all services"
        echo "  restart - Restart all services"
        echo "  logs    - View logs for all services"
        echo "  status  - Show status of all services"
        echo "  test    - Run tests for all services"
        exit 1
        ;;
esac