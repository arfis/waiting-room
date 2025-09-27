#!/bin/bash

# Waiting Room System Docker Development Startup Script
# This script starts all components using Docker Compose with external MongoDB

echo "🐳 Starting Waiting Room System with Docker (Development Mode)..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if MongoDB is accessible
echo "🔍 Checking MongoDB connection..."
if ! nc -z localhost 27017 2>/dev/null; then
    echo "⚠️  MongoDB is not accessible on localhost:27017"
    echo "   Please ensure your MongoDB Docker container is running"
    echo "   You can start it with: docker run -d -p 27017:27017 --name mongodb mongo:7.0"
    exit 1
fi
echo "✅ MongoDB is accessible"

# Initialize MongoDB if needed
echo "🗄️  Initializing MongoDB..."
docker run --rm --network host mongo:7.0 mongosh --host localhost:27017 --eval "
db = db.getSiblingDB('waiting_room');
if (db.queue_entries.countDocuments() === 0) {
  db.createCollection('queue_entries');
  db.queue_entries.createIndex({ 'waitingRoomId': 1, 'status': 1 });
  db.queue_entries.createIndex({ 'qrToken': 1 }, { unique: true });
  db.queue_entries.createIndex({ 'ticketNumber': 1 }, { unique: true });
  db.queue_entries.createIndex({ 'createdAt': 1 });
  print('MongoDB initialized successfully!');
} else {
  print('MongoDB already initialized');
}
" > /dev/null 2>&1

# Build and start all services
echo "🔨 Building and starting all services..."
docker-compose -f docker-compose.dev.yml up --build -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."

# Wait for API
echo "   Waiting for API server..."
until curl -s http://localhost:8080/health > /dev/null 2>&1; do
    sleep 2
done
echo "   ✅ API server is ready"

# Wait for Kiosk
echo "   Waiting for Kiosk..."
until curl -s http://localhost:4201/health > /dev/null 2>&1; do
    sleep 2
done
echo "   ✅ Kiosk is ready"

# Wait for Mobile
echo "   Waiting for Mobile app..."
until curl -s http://localhost:4202 > /dev/null 2>&1; do
    sleep 2
done
echo "   ✅ Mobile app is ready"

# Wait for TV
echo "   Waiting for TV display..."
until curl -s http://localhost:4203 > /dev/null 2>&1; do
    sleep 2
done
echo "   ✅ TV display is ready"

# Wait for Backoffice
echo "   Waiting for Backoffice..."
until curl -s http://localhost:4200 > /dev/null 2>&1; do
    sleep 2
done
echo "   ✅ Backoffice is ready"

echo ""
echo "🎉 System is ready!"
echo ""
echo "📱 Kiosk: http://localhost:4201"
echo "📱 Mobile: http://localhost:4202"
echo "📺 TV Display: http://localhost:4203"
echo "🏢 Backoffice: http://localhost:4200"
echo "🔌 API: http://localhost:8080"
echo "🗄️  MongoDB: localhost:27017 (external)"
echo ""
echo "🎯 System is fully ready!"
echo "   - Insert a smart card to test"
echo "   - Watch the kiosk for card data and ticket generation"
echo "   - Scan QR code on mobile to track queue position"
echo "   - Use backoffice to manage the queue"
echo "   - TV display shows current queue status"
echo ""
echo "📊 To view logs: docker-compose -f docker-compose.dev.yml logs -f [service-name]"
echo "🛑 To stop: docker-compose -f docker-compose.dev.yml down"
echo "🔄 To restart: docker-compose -f docker-compose.dev.yml restart [service-name]"
echo ""
echo "Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping Docker services..."
    docker-compose -f docker-compose.dev.yml down
    echo "✅ All services stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Wait for user to stop
wait
