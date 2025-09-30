#!/bin/bash

echo "🔨 Building UI library..."
ng build ui

echo "🚀 Starting TV app in development mode..."
ng serve tv --port 4203
