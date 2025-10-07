#!/bin/bash

# Build script for all projects that depend on api-client
# This ensures api-client is built first, then all dependent projects

echo "🔨 Building api-client library..."
ng build api-client

if [ $? -ne 0 ]; then
    echo "❌ Failed to build api-client"
    exit 1
fi

echo "✅ api-client built successfully"
echo ""

echo "🔨 Building all dependent projects..."

# Build all projects that depend on api-client
echo "Building backoffice..."
ng build backoffice

echo "Building kiosk..."
ng build kiosk

echo "Building mobile..."
ng build mobile

echo "Building tv..."
ng build tv

echo ""
echo "✅ All projects built successfully!"
echo ""
echo "📁 Build outputs:"
echo "  - api-client: dist/api-client/"
echo "  - backoffice: dist/backoffice/"
echo "  - kiosk: dist/kiosk/"
echo "  - mobile: dist/mobile/"
echo "  - tv: dist/tv/"
