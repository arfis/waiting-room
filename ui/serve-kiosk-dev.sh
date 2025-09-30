#!/bin/bash

# Development script to serve kiosk app with UI library in watch mode
echo "🔨 Building UI library in watch mode..."
ng build ui --watch &
UI_BUILD_PID=$!

# Wait a moment for the initial build
sleep 3

echo "🚀 Starting kiosk app in development mode..."
ng serve kiosk --port 4200

# Clean up when script exits
trap "kill $UI_BUILD_PID 2>/dev/null" EXIT
