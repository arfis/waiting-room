#!/bin/bash

# Waiting Room System Startup Script
# This script starts all components of the waiting room system

echo "🚀 Starting Waiting Room System..."

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        echo "⚠️  Port $port is already in use"
        return 1
    else
        return 0
    fi
}

# Function to wait for a service to be ready
wait_for_service() {
    local url=$1
    local name=$2
    local max_attempts=30
    local attempt=1
    
    echo "⏳ Waiting for $name to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            echo "✅ $name is ready!"
            return 0
        fi
        echo "   Attempt $attempt/$max_attempts..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ $name failed to start after $max_attempts attempts"
    return 1
}

# Check if required ports are available
echo "🔍 Checking port availability..."
check_port 8080 || { echo "Please stop the service using port 8080"; exit 1; }
check_port 4201 || { echo "Please stop the service using port 4201"; exit 1; }

# Start API server
echo "🌐 Starting API server..."
cd api
go build -o api-server cmd/api/main.go
./api-server &
API_PID=$!
cd ..

# Wait for API to be ready
wait_for_service "http://localhost:8080/health" "API Server" || {
    echo "❌ API server failed to start"
    kill $API_PID 2>/dev/null
    exit 1
}

# Build Angular kiosk app first
echo "🔨 Building Angular kiosk app..."
cd ui
if [ ! -d "node_modules" ]; then
    echo "📦 Installing UI dependencies..."
    npm install
fi
ng build kiosk
cd ..

# Start Kiosk WebSocket server
echo "🖥️  Starting Kiosk WebSocket server..."
cd ui/projects/kiosk

# Install WebSocket server dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing kiosk WebSocket server dependencies..."
    npm install
fi

# Start WebSocket server
node websocket-server.js &
KIOSK_PID=$!
cd ../../..

# Wait for Kiosk to be ready
wait_for_service "http://localhost:4201/health" "Kiosk WebSocket Server" || {
    echo "❌ Kiosk WebSocket server failed to start"
    kill $API_PID $KIOSK_PID 2>/dev/null
    exit 1
}

echo ""
echo "🎉 System is ready!"
echo ""
echo "📱 Kiosk: http://localhost:4201"
echo "🔌 API: http://localhost:8080"
echo "📡 WebSocket: ws://localhost:4201/ws/card-reader"
echo ""
echo "💳 Starting card reader..."
cd card-reader
go run main.go &
CARD_READER_PID=$!
cd ..

echo "✅ Card reader started (PID: $CARD_READER_PID)"
echo ""
echo "🎯 System is fully ready!"
echo "   - Insert a smart card to test"
echo "   - Watch the kiosk for card data"
echo ""
echo "Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    kill $API_PID $KIOSK_PID $CARD_READER_PID 2>/dev/null
    echo "✅ All services stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Wait for user to stop
wait
