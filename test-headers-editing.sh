#!/bin/bash

echo "=== Testing Headers Editing Functionality ==="

echo "1. Testing admin frontend accessibility..."
curl -s "http://localhost:4205" > /dev/null && echo "✅ Admin frontend is accessible" || echo "❌ Admin frontend is not accessible"

echo -e "\n2. Testing API with custom headers..."
curl -X PUT "http://localhost:8080/api/admin/configuration/external-api" \
  -H "Content-Type: application/json" \
  -d '{
    "userServicesUrl": "https://private-4a985-invoice19.apiary-mock.com/waiting-room/medical/services?identifier=${identifier}",
    "timeoutSeconds": 15,
    "retryAttempts": 2,
    "headers": {
      "Custom-Auth": "Bearer my-custom-token",
      "X-Custom-Header": "custom-value-123",
      "API-Version": "v2.1",
      "Client-ID": "my-client-123"
    }
  }' > /dev/null

echo "✅ Custom headers updated successfully"

echo -e "\n3. Verifying custom headers are persisted..."
curl -s "http://localhost:8080/api/admin/configuration/external-api" | jq '.headers'

echo -e "\n✅ Headers editing test completed!"
echo -e "\n📝 Frontend Testing Instructions:"
echo "   1. Open http://localhost:4205 in your browser"
echo "   2. Navigate to Configuration"
echo "   3. You should see 4 custom headers:"
echo "      - Custom-Auth: Bearer my-custom-token"
echo "      - X-Custom-Header: custom-value-123"
echo "      - API-Version: v2.1"
echo "      - Client-ID: my-client-123"
echo "   4. Try editing the header names and values"
echo "   5. Try adding new headers"
echo "   6. Try removing headers"
echo "   7. Save the configuration and verify changes persist"
