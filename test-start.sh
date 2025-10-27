#!/bin/bash
set -e

echo "🚀 Testing start script..."

# Kill any existing processes
echo "🧹 Cleaning up existing processes..."
pkill -f "go run cmd/api" 2>/dev/null || true
pkill -f "npx serve" 2>/dev/null || true
pkill -f "node websocket-server.js" 2>/dev/null || true
sleep 2

# Start API server in background
echo "🔧 Starting API server..."
cd api
go run ./cmd/api &
API_PID=$!
cd ..

# Wait for API to be ready
echo "⏳ Waiting for API server..."
for i in {1..10}; do
  if curl -s http://localhost:8080/health >/dev/null 2>&1; then
    echo "✅ API server is ready!"
    break
  fi
  echo "   Attempt $i/10..."
  sleep 2
done

# Test the build command
echo "🏗️  Testing build command..."
cd ui-nx

echo "Running: npx nx run-many --target=build --projects=kiosk,admin,backoffice,tv,mobile,ui,api-client,primeng-components --prod"
if npx nx run-many --target=build --projects=kiosk,admin,backoffice,tv,mobile,ui,api-client,primeng-components --prod; then
  echo "✅ Build command completed successfully!"
else
  echo "❌ Build command failed!"
  exit 1
fi

echo "🎉 Test completed successfully!"
