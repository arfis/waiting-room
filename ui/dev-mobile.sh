#!/bin/bash

echo "🔨 Building UI library..."
ng build ui

echo "🚀 Starting Mobile app in development mode..."
ng serve mobile --port 4204
