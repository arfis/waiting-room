#!/bin/bash

echo "🔍 Testing API endpoints..."

# Test 1: Health endpoint
echo "1. Testing health endpoint..."
health_response=$(curl -s http://localhost:8080/health)
if [ "$health_response" = '{"status":"ok"}' ]; then
    echo "   ✅ Health endpoint working"
else
    echo "   ❌ Health endpoint failed: $health_response"
    exit 1
fi

# Test 2: Simple POST without MongoDB
echo "2. Testing simple POST endpoint..."
simple_response=$(curl -s -X POST http://localhost:8080/waiting-rooms/triage-1/swipe \
  -H "Content-Type: application/json" \
  -d '{"idCardRaw": "test"}' \
  --max-time 5)

if [ $? -eq 0 ] && [ -n "$simple_response" ]; then
    echo "   ✅ POST endpoint working"
    echo "   Response: $simple_response"
else
    echo "   ❌ POST endpoint failed or timed out"
fi

# Test 3: Check if API server is responsive
echo "3. Testing API server responsiveness..."
start_time=$(date +%s)
curl -s http://localhost:8080/health > /dev/null
end_time=$(date +%s)
response_time=$((end_time - start_time))

if [ $response_time -lt 2 ]; then
    echo "   ✅ API server is responsive (${response_time}s)"
else
    echo "   ⚠️  API server is slow (${response_time}s)"
fi

echo ""
echo "🎯 API Test Complete!"
