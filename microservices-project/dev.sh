#!/bin/bash

# Development helper script for microservices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to start all services
start_services() {
    print_status "Starting all microservices..."
    docker-compose up -d
    print_status "Waiting for services to be healthy..."
    sleep 10
    docker-compose ps
    print_status "All services started successfully!"
    print_status "Access points:"
    echo "  - API Gateway: http://localhost:8080"
    echo "  - User Service: http://localhost:8001"
    echo "  - Notification Service: http://localhost:8002"
    echo "  - Kafka UI: http://localhost:8090"
    echo "  - Redis Commander: http://localhost:8081"
    echo "  - pgAdmin: http://localhost:5050 (admin@admin.com/admin)"
}

# Function to stop all services
stop_services() {
    print_status "Stopping all microservices..."
    docker-compose down
    print_status "All services stopped."
}

# Function to restart a specific service
restart_service() {
    if [ -z "$1" ]; then
        print_error "Please specify a service name"
        exit 1
    fi
    print_status "Restarting $1..."
    docker-compose restart $1
    print_status "$1 restarted successfully!"
}

# Function to view logs
view_logs() {
    if [ -z "$1" ]; then
        docker-compose logs -f
    else
        docker-compose logs -f $1
    fi
}

# Function to rebuild services
rebuild_services() {
    print_status "Rebuilding all services..."
    docker-compose build
    print_status "Services rebuilt successfully!"
}

# Function to reset everything
reset_all() {
    print_warning "This will remove all containers, volumes, and data. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Resetting all services and data..."
        docker-compose down -v
        print_status "Reset complete."
    else
        print_status "Reset cancelled."
    fi
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    # Add migration commands here when implemented
    print_warning "Migrations not yet implemented"
}

# Function to create a test user
create_test_user() {
    print_status "Creating test user..."
    curl -X POST http://localhost:8080/api/v1/auth/register \
        -H "Content-Type: application/json" \
        -d '{
            "email": "test@example.com",
            "username": "testuser",
            "password": "password123",
            "first_name": "Test",
            "last_name": "User",
            "phone_number": "+1234567890"
        }'
    echo ""
    print_status "Test user created: test@example.com / password123"
}

# Main menu
show_menu() {
    echo ""
    echo "Microservices Development Helper"
    echo "================================"
    echo "1. Start all services"
    echo "2. Stop all services"
    echo "3. Restart a service"
    echo "4. View logs"
    echo "5. Rebuild services"
    echo "6. Reset everything (WARNING: Deletes all data)"
    echo "7. Create test user"
    echo "8. Exit"
    echo ""
}

# Check Docker before running
check_docker

# Main loop
if [ $# -eq 0 ]; then
    while true; do
        show_menu
        read -p "Select an option: " choice
        case $choice in
            1) start_services ;;
            2) stop_services ;;
            3) 
                read -p "Enter service name: " service
                restart_service $service 
                ;;
            4) 
                read -p "Enter service name (or press Enter for all): " service
                view_logs $service
                ;;
            5) rebuild_services ;;
            6) reset_all ;;
            7) create_test_user ;;
            8) 
                print_status "Goodbye!"
                exit 0 
                ;;
            *) print_error "Invalid option" ;;
        esac
    done
else
    # Command line arguments
    case $1 in
        start) start_services ;;
        stop) stop_services ;;
        restart) restart_service $2 ;;
        logs) view_logs $2 ;;
        rebuild) rebuild_services ;;
        reset) reset_all ;;
        test-user) create_test_user ;;
        *) 
            print_error "Unknown command: $1"
            echo "Usage: $0 [start|stop|restart <service>|logs [service]|rebuild|reset|test-user]"
            exit 1
            ;;
    esac
fi