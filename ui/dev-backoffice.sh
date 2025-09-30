#!/bin/bash

echo "🔨 Building UI library..."
ng build ui

echo "🚀 Starting Backoffice app in development mode..."
ng serve backoffice --port 4201
