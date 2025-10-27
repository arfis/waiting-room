#!/bin/bash
set -e

echo "🚀 Starting Waiting Room System (Simple Version)..."

# Kill any existing processes
echo "🧹 Cleaning up existing processes..."
pkill -f "go run cmd/api" 2>/dev/null || true
pkill -f "npx serve" 2>/dev/null || true
pkill -f "node websocket-server.js" 2>/dev/null || true
pkill -f "api-server" 2>/dev/null || true
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

# Build and serve frontend apps
echo "🏗️  Building and serving frontend apps..."

cd ui-nx

# Build all apps
echo "📦 Building all applications..."
npx nx run-many --target=build --projects=kiosk,admin,backoffice,tv,mobile,ui,api-client,primeng-components --prod

echo "✅ All applications built successfully!"

# Start serving apps
echo "🌐 Starting frontend servers..."

# Get absolute path to ui-nx directory
UI_NX_PATH="$(pwd)"

# Start WebSocket server (this also serves the kiosk app on port 4201)
node websocket-server.js &
KIOSK_PID=$!

# Admin
npx serve -s "$UI_NX_PATH/dist/apps/admin/browser/browser" -l 4205 &
ADMIN_PID=$!

# Backoffice
npx serve -s "$UI_NX_PATH/dist/backoffice/browser" -l 4200 &
BACKOFFICE_PID=$!

# TV
npx serve -s "$UI_NX_PATH/dist/tv/browser" -l 4203 &
TV_PID=$!

# Mobile
npx serve -s "$UI_NX_PATH/dist/mobile/browser" -l 4204 &
MOBILE_PID=$!

cd ..

echo ""
echo "🎉 System started successfully!"
echo ""
echo "📱 Applications:"
echo "   Kiosk:      http://localhost:4201"
echo "   Admin:      http://localhost:4205"
echo "   Backoffice: http://localhost:4200"
echo "   TV:         http://localhost:4203"
echo "   Mobile:     http://localhost:4204"
echo "   API:        http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop all services"

# Cleanup function
cleanup() {
  echo ""
  echo "🛑 Stopping services..."
  kill $API_PID $KIOSK_PID $ADMIN_PID $BACKOFFICE_PID $TV_PID $MOBILE_PID 2>/dev/null || true
  echo "✅ All services stopped"
  exit 0
}

# Trap Ctrl+C
trap cleanup INT

# Wait for user to stop
wait
