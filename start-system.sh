#!/usr/bin/env bash
set -Eeuo pipefail

# Waiting Room System Startup Script
# This script starts all components of the waiting room system

echo "Starting Waiting Room System..."

# Resolve repo root regardless of where script is invoked from
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if git -C "$SCRIPT_DIR" rev-parse --show-toplevel >/dev/null 2>&1; then
  REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
else
  REPO_ROOT="$SCRIPT_DIR"
fi
cd "$REPO_ROOT"

# Function to check if a port is in use
check_port() {
  local port=$1
  if lsof -Pi :"$port" -sTCP:LISTEN -t >/dev/null ; then
    echo "Port $port is already in use"
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
  echo "Waiting for $name to be ready..."
  while [ $attempt -le $max_attempts ]; do
    if curl -s "$url" > /dev/null 2>&1; then
      echo "$name is ready!"
      return 0
    fi
    echo "Attempt $attempt/$max_attempts..."
    sleep 2
    attempt=$((attempt + 1))
  done
  echo "$name failed to start after $max_attempts attempts"
  return 1
}

# Check if required ports are available
echo "Checking port availability..."
check_port 8080 || { echo "Please stop the service using port 8080"; exit 1; }
check_port 4201 || { echo "Please stop the service using port 4201"; exit 1; }
check_port 4204 || { echo "Please stop the service using port 4204"; exit 1; }
check_port 4203 || { echo "Please stop the service using port 4203"; exit 1; }
check_port 4200 || { echo "Please stop the service using port 4200"; exit 1; }

# Check MongoDB availability
echo "Checking MongoDB..."
if command -v mongod > /dev/null 2>&1; then
  if ! pgrep -x "mongod" > /dev/null; then
    echo "Starting MongoDB daemon..."
    mkdir -p ./data/db ./data/logs
    mongod --dbpath ./data/db --logpath ./data/logs/mongod.log --fork
    sleep 3
  else
    echo "MongoDB is already running"
  fi
else
  echo "MongoDB not found locally. Using external MongoDB."
  echo "Please ensure MongoDB is running on localhost:27017"
fi

# Start API server
echo "Starting API server..."
pushd api >/dev/null
if command -v go >/dev/null 2>&1; then
  go build -o api-server cmd/api/main.go
  ./api-server &
  API_PID=$!
else
  echo "Go toolchain not found. Please install Go to build the API."
  exit 1
fi
popd >/dev/null

# Wait for API to be ready
wait_for_service "http://localhost:8080/health" "API Server" || {
  echo "API server failed to start"
  kill ${API_PID:-} 2>/dev/null || true
  exit 1
}

# Build Angular apps first
echo "Building Angular apps..."
pushd ui >/dev/null
if [ ! -d "node_modules" ]; then
  echo "Installing UI dependencies..."
  if command -v npm >/dev/null 2>&1; then
    npm ci || npm install
  else
    echo "npm not found; cannot install UI dependencies."
    exit 1
  fi
fi
if command -v ng >/dev/null 2>&1; then
  ng build kiosk
  ng build mobile
  ng build tv
  ng build backoffice
  ng build admin
else
  echo "Angular CLI (ng) not found. Please install @angular/cli globally or provide a local script."
  exit 1
fi
popd >/dev/null

# Start Kiosk WebSocket server (now under scripts/)
echo "Starting Kiosk WebSocket server..."
if [ ! -d "node_modules" ]; then
  echo "Installing WebSocket server dependencies..."
  if command -v npm >/dev/null 2>&1; then
    npm ci || npm install
  else
    echo "npm not found; cannot install dependencies."
    exit 1
  fi
fi
node ui/projects/kiosk/websocket-server.js &
KIOSK_PID=$!

# Wait for Kiosk to be ready
wait_for_service "http://localhost:4201/health" "Kiosk WebSocket Server" || {
  echo "Kiosk WebSocket server failed to start"
  kill ${API_PID:-} ${KIOSK_PID:-} 2>/dev/null || true
  exit 1
}

# Start Mobile app server
echo "Starting Mobile app server..."
pushd ui/projects/mobile >/dev/null
node server.js &
MOBILE_PID=$!
popd >/dev/null

# Start TV app server
echo "Starting TV app server..."
pushd ui >/dev/null
npx serve -s dist/tv/browser -l 4203 &
TV_PID=$!
popd >/dev/null

# Start Backoffice app server
echo "Starting Backoffice app server..."
pushd ui >/dev/null
npx serve -s dist/backoffice/browser -l 4200 &
BACKOFFICE_PID=$!
popd >/dev/null

# Start Admin app server
echo "Starting admin app server..."
pushd ui >/dev/null
npx serve -s dist/admin/browser -l 4205 &
ADMIN_PID=$!
popd >/dev/null

printf "\nSystem is ready!\n\n"
echo "Kiosk:     http://localhost:4201"
echo "Mobile:    http://localhost:4204"
echo "TV Display: http://localhost:4203"
echo "Backoffice: http://localhost:4200"
echo "Admin:     http://localhost:4205"
echo "API:       http://localhost:8080"
echo "WebSocket: ws://localhost:4201/ws/card-reader"

echo
echo "Starting card reader..."
pushd card-reader >/dev/null
if command -v go >/dev/null 2>&1; then
  go run main.go &
  CARD_READER_PID=$!
else
  echo "Go toolchain not found; skipping card reader."
  CARD_READER_PID=""
fi
popd >/dev/null

if [[ -n "${CARD_READER_PID}" ]]; then
  echo "Card reader started (PID: ${CARD_READER_PID})"
fi

echo
printf "System is fully ready!\n"
echo " - Insert a smart card to test"
echo " - Watch the kiosk for card data and ticket generation"
echo " - Scan QR code on mobile to track queue position"
echo " - Use backoffice to manage the queue"
echo " - TV display shows current queue status"
echo

echo "Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
  echo
  echo "Stopping services..."
  kill ${API_PID:-} ${KIOSK_PID:-} ${MOBILE_PID:-} ${TV_PID:-} ${BACKOFFICE_PID:-} ${ADMIN_PID:-} ${CARD_READER_PID:-} 2>/dev/null || true
  echo "All services stopped"
}

# Set up signal handlers
trap cleanup EXIT INT TERM

# Wait for background jobs
wait
