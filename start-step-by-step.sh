#!/bin/bash
set -e

echo "🚀 Starting Waiting Room System - Step by Step"
echo "=============================================="

# Kill any existing processes
echo "🧹 Step 1: Cleaning up existing processes..."
pkill -f "go run cmd/api" 2>/dev/null || true
pkill -f "npx serve" 2>/dev/null || true
pkill -f "node websocket-server.js" 2>/dev/null || true
sleep 2

# Start API server
echo "🔧 Step 2: Starting API server..."
cd api
go run ./cmd/api &
API_PID=$!
cd ..

# Wait for API
echo "⏳ Step 3: Waiting for API server..."
for i in {1..15}; do
  if curl -s http://localhost:8080/health >/dev/null 2>&1; then
    echo "✅ API server is ready!"
    break
  fi
  echo "   Attempt $i/15..."
  sleep 2
done

# Build applications
echo "🏗️  Step 4: Building applications..."
cd ui-nx

echo "   Building kiosk..."
npx nx build kiosk --prod
echo "   ✅ Kiosk built"

echo "   Building admin..."
npx nx build admin --prod
echo "   ✅ Admin built"

echo "   Building backoffice..."
npx nx build backoffice --prod
echo "   ✅ Backoffice built"

echo "   Building tv..."
npx nx build tv --prod
echo "   ✅ TV built"

echo "   Building mobile..."
npx nx build mobile --prod
echo "   ✅ Mobile built"

echo "✅ All applications built successfully!"

# Start serving apps
echo "🌐 Step 5: Starting frontend servers..."

echo "   Starting Kiosk server..."
npx serve -s dist/kiosk/browser -l 4201 &
KIOSK_PID=$!

echo "   Starting Admin server..."
npx serve -s dist/admin/browser -l 4205 &
ADMIN_PID=$!

echo "   Starting Backoffice server..."
npx serve -s dist/backoffice/browser -l 4200 &
BACKOFFICE_PID=$!

echo "   Starting TV server..."
npx serve -s dist/tv/browser -l 4203 &
TV_PID=$!

echo "   Starting Mobile server..."
npx serve -s dist/mobile/browser -l 4204 &
MOBILE_PID=$!

echo "   Starting WebSocket server..."
node websocket-server.js &
WS_PID=$!

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
  kill $API_PID $KIOSK_PID $ADMIN_PID $BACKOFFICE_PID $TV_PID $MOBILE_PID $WS_PID 2>/dev/null || true
  echo "✅ All services stopped"
  exit 0
}

# Trap Ctrl+C
trap cleanup INT

# Wait for user to stop
wait
