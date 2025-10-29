#!/bin/bash
set -e

echo "🚀 Starting Waiting Room System (Development Mode with Hot Reload)..."

# Kill any existing processes
echo "🧹 Cleaning up existing processes..."
pkill -f "go run cmd/api" 2>/dev/null || true
pkill -f "npx serve" 2>/dev/null || true
pkill -f "node websocket-server.js" 2>/dev/null || true
pkill -f "api-server" 2>/dev/null || true
pkill -f "nx serve" 2>/dev/null || true
# Kill processes on specific ports
lsof -ti :8080 2>/dev/null | xargs kill -9 2>/dev/null || true
lsof -ti :4200 2>/dev/null | xargs kill -9 2>/dev/null || true
lsof -ti :4201 2>/dev/null | xargs kill -9 2>/dev/null || true
lsof -ti :4203 2>/dev/null | xargs kill -9 2>/dev/null || true
lsof -ti :4204 2>/dev/null | xargs kill -9 2>/dev/null || true
lsof -ti :4205 2>/dev/null | xargs kill -9 2>/dev/null || true
sleep 3

# Start API server in background
echo "🔧 Starting API server..."
cd api
go run ./cmd/api &
API_PID=$!
cd ..

# Wait for API to be ready
echo "⏳ Waiting for API server..."
for i in {1..30}; do
  if curl -s http://localhost:8080/health >/dev/null 2>&1; then
    echo "✅ API server is ready!"
    break
  fi
  echo "   Attempt $i/30..."
  sleep 2
done

# Build libraries and kiosk first (required for WebSocket server)
echo "📦 Building shared libraries and kiosk..."
cd ui-nx
npx nx run-many --target=build --projects=api-client,primeng-components --prod
npx nx build kiosk --configuration=development

echo "✅ Libraries and kiosk built successfully!"

# Start WebSocket server (serves kiosk on port 4201)
echo "🌐 Starting WebSocket server for kiosk..."
node websocket-server.js &
KIOSK_PID=$!

# Start other apps in development mode (with hot reload)
echo "🌐 Starting other frontend development servers..."

npx nx serve admin --configuration=development &
ADMIN_PID=$!

npx nx serve backoffice --configuration=development &
BACKOFFICE_PID=$!

npx nx serve tv --configuration=development &
TV_PID=$!

npx nx serve mobile --configuration=development &
MOBILE_PID=$!

# Wait a moment for all servers to start
sleep 5

cd ..

echo ""
echo "🎉 System started successfully!"
echo ""
echo "📱 Applications (with hot reload):"
echo "   Kiosk:      http://localhost:4201"
echo "   Admin:      http://localhost:4205"
echo "   Backoffice: http://localhost:4200"
echo "   TV:         http://localhost:4203"
echo "   Mobile:     http://localhost:4204"
echo "   API:        http://localhost:8080"
echo ""
echo "🔥 Hot reload is enabled - changes will automatically reload!"
echo ""
echo "Press Ctrl+C to stop all services"

# Cleanup function
cleanup() {
  echo ""
  echo "🛑 Stopping services..."
  kill $API_PID 2>/dev/null || true
  kill $KIOSK_PID $ADMIN_PID $BACKOFFICE_PID $TV_PID $MOBILE_PID 2>/dev/null || true
  # Kill any remaining processes
  pkill -f "nx serve" 2>/dev/null || true
  pkill -f "websocket-server.js" 2>/dev/null || true
  echo "✅ All services stopped"
  exit 0
}

# Trap Ctrl+C
trap cleanup INT

# Wait for user to stop
wait
