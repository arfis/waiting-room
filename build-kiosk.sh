#!/bin/bash

# Build Kiosk Application Script
echo "🔨 Building Kiosk Application..."

# Check if we're in the right directory
if [ ! -d "ui" ]; then
    echo "❌ Please run this script from the project root directory"
    exit 1
fi

# Install UI dependencies if needed
echo "📦 Installing UI dependencies..."
cd ui
if [ ! -d "node_modules" ]; then
    npm install
fi

# Build the kiosk app
echo "🏗️  Building kiosk app..."
ng build kiosk

if [ $? -eq 0 ]; then
    echo "✅ Kiosk app built successfully!"
    echo "📁 Built files are in: ui/dist/kiosk/browser/"
else
    echo "❌ Failed to build kiosk app"
    exit 1
fi

# Install WebSocket server dependencies
echo "📦 Installing WebSocket server dependencies..."
cd projects/kiosk
if [ ! -d "node_modules" ]; then
    npm install
fi

echo "🎉 Kiosk is ready to run!"
echo "💡 To start the WebSocket server: cd ui/projects/kiosk && node websocket-server.js"
