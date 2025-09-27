#!/bin/bash

echo "🔍 Testing Waiting Room System Components..."

# Test 1: API Health
echo "1. Testing API Health..."
if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo "   ✅ API is healthy"
else
    echo "   ❌ API is not responding"
    exit 1
fi

# Test 2: MongoDB Connection
echo "2. Testing MongoDB Connection..."
if nc -z localhost 27017 2>/dev/null; then
    echo "   ✅ MongoDB is accessible on port 27017"
else
    echo "   ❌ MongoDB is not accessible on port 27017"
    echo "   Please start MongoDB: docker run -d -p 27017:27017 --name mongodb mongo:7.0"
    exit 1
fi

# Test 3: API Ticket Generation
echo "3. Testing API Ticket Generation..."
response=$(curl -s -X POST http://localhost:8080/waiting-rooms/triage-1/swipe \
  -H "Content-Type: application/json" \
  -d '{"idCardRaw": "{\"id_number\":\"123456789\",\"first_name\":\"John\",\"last_name\":\"Doe\"}"}' \
  --max-time 10)

if [ $? -eq 0 ] && [ -n "$response" ]; then
    echo "   ✅ API ticket generation working"
    echo "   Response: $response"
else
    echo "   ❌ API ticket generation failed"
    echo "   This is likely a MongoDB connection issue"
fi

# Test 4: Frontend Applications
echo "4. Testing Frontend Applications..."
for port in 4200 4201 4204 4203; do
    if curl -s http://localhost:$port > /dev/null 2>&1; then
        echo "   ✅ Port $port is responding"
    else
        echo "   ❌ Port $port is not responding"
    fi
done

echo ""
echo "🎯 System Test Complete!"
echo ""
echo "If API ticket generation failed, the issue is likely:"
echo "1. MongoDB not running or not accessible"
echo "2. API server can't connect to MongoDB"
echo "3. Database permissions issue"
echo ""
echo "To fix MongoDB issues:"
echo "docker run -d -p 27017:27017 --name mongodb mongo:7.0"
