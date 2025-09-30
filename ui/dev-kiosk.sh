#!/bin/bash

echo "🔨 Building UI library..."
ng build ui

echo "🚀 Starting kiosk app in development mode..."
ng serve kiosk --port 4200
