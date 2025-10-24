#!/bin/bash

echo "🔨 Building UI library..."
ng build ui

echo "🚀 Starting Admin app in development mode..."
ng serve admin --port 4205
