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
check_port 4204 || { echo "Please stop the service using port 4204"; exit 1; }
check_port 4203 || { echo "Please stop the service using port 4203"; exit 1; }
check_port 4200 || { echo "Please stop the service using port 4200"; exit 1; }

# Check MongoDB availability
echo "🗄️  Checking MongoDB..."
if command -v mongod > /dev/null 2>&1; then
    if ! pgrep -x "mongod" > /dev/null; then
        echo "   Starting MongoDB daemon..."
        mkdir -p ./data/db ./data/logs
        mongod --dbpath ./data/db --logpath ./data/logs/mongod.log --fork
        sleep 3
    else
        echo "   MongoDB is already running"
    fi
else
    echo "   ⚠️  MongoDB not found locally. Using external MongoDB."
    echo "   Please ensure MongoDB is running on localhost:27017"
fi

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

# Build Angular apps first
echo "🔨 Building Angular apps..."
cd ui
if [ ! -d "node_modules" ]; then
    echo "📦 Installing UI dependencies..."
    npm install
fi
ng build kiosk
ng build mobile
ng build tv
ng build backoffice
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

# Start Mobile app server
echo "📱 Starting Mobile app server..."
cd ui/projects/mobile
node server.js &
MOBILE_PID=$!
cd ../..

# Start TV app server
echo "📺 Starting TV app server..."
cd ui
npx serve -s dist/tv/browser -l 4203 &
TV_PID=$!
cd ..

# Start Backoffice app server
echo "🏢 Starting Backoffice app server..."
cd ui
npx serve -s dist/backoffice/browser -l 4200 &
BACKOFFICE_PID=$!
cd ..

echo ""
echo "🎉 System is ready!"
echo ""
echo "📱 Kiosk: http://localhost:4201"
echo "📱 Mobile: http://localhost:4204"
echo "📺 TV Display: http://localhost:4203"
echo "🏢 Backoffice: http://localhost:4200"
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
echo "   - Watch the kiosk for card data and ticket generation"
echo "   - Scan QR code on mobile to track queue position"
echo "   - Use backoffice to manage the queue"
echo "   - TV display shows current queue status"
echo ""
echo "Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    kill $API_PID $KIOSK_PID $MOBILE_PID $TV_PID $BACKOFFICE_PID $CARD_READER_PID 2>/dev/null
    echo "✅ All services stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Wait for user to stop
wait
