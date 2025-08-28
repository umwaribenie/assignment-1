#!/bin/bash

# Microservices Startup Script
set -e

echo "🚀 Starting Microservices Architecture..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed."
    exit 1
fi

echo "📦 Building and starting services..."

# Start services
docker-compose up -d --build

echo "⏳ Waiting for services to be healthy..."

# Wait for services to be healthy
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker-compose ps | grep -q "healthy"; then
        echo "✅ Services are starting up..."
        break
    fi
    
    echo "⏳ Waiting for services... (${attempt}/${max_attempts})"
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "⚠️  Services are taking longer than expected to start."
    echo "📊 Current status:"
    docker-compose ps
    echo ""
    echo "📋 You can check logs with: docker-compose logs -f"
else
    echo ""
    echo "✅ Microservices are starting up!"
    echo ""
    echo "🔗 Service URLs:"
    echo "   • User Service:         http://localhost:8081"
    echo "   • Notification Service: http://localhost:8082"
    echo "   • Kafka UI:            http://localhost:8080"
    echo ""
    echo "🏥 Health Checks:"
    echo "   • User Service:         curl http://localhost:8081/health"
    echo "   • Notification Service: curl http://localhost:8082/health"
    echo ""
    echo "📊 View logs: docker-compose logs -f"
    echo "🛑 Stop services: docker-compose down"
fi