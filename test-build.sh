#!/bin/bash
set -e

echo "Testing Nx build process..."

cd ui-nx

echo "Building kiosk..."
npx nx build kiosk

echo "Building admin..."
npx nx build admin

echo "Building all apps..."
npx nx run-many --target=build --all

echo "All builds completed successfully!"
