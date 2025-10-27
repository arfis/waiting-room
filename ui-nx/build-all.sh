#!/bin/bash

# Build all applications and libraries
echo "Building all applications and libraries..."

# Build libraries first
echo "Building shared libraries..."
npx nx build api-client
npx nx build ui

# Build applications
echo "Building applications..."
npx nx build admin
npx nx build backoffice
npx nx build kiosk
npx nx build mobile
npx nx build tv

echo "All builds completed successfully!"
