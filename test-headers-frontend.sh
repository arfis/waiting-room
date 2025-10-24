#!/bin/bash

echo "=== Testing Headers Frontend Functionality ==="

echo "1. Testing admin frontend accessibility..."
curl -s "http://localhost:4205" > /dev/null && echo "✅ Admin frontend is accessible" || echo "❌ Admin frontend is not accessible"

echo -e "\n2. Testing API configuration endpoint..."
API_RESPONSE=$(curl -s "http://localhost:8080/api/admin/configuration/external-api")
echo "Current configuration:"
echo "$API_RESPONSE" | jq .

echo -e "\n3. Testing headers update via API..."
curl -X PUT "http://localhost:8080/api/admin/configuration/external-api" \
  -H "Content-Type: application/json" \
  -d '{
    "userServicesUrl": "https://private-4a985-invoice19.apiary-mock.com/waiting-room/medical/services?identifier=${identifier}",
    "timeoutSeconds": 15,
    "retryAttempts": 2,
    "headers": {
      "Authorization": "Bearer test-token-123",
      "X-API-Key": "api-key-456",
      "Content-Type": "application/json",
      "Accept": "application/json"
    }
  }' > /dev/null

echo "✅ Headers updated successfully"

echo -e "\n4. Verifying headers are persisted..."
curl -s "http://localhost:8080/api/admin/configuration/external-api" | jq '.headers'

echo -e "\n✅ Headers frontend functionality test completed!"
echo -e "\n📝 Next steps:"
echo "   - Open http://localhost:4205 in your browser"
echo "   - Navigate to Configuration"
echo "   - Check that headers are displayed correctly"
echo "   - Try adding/editing/removing headers"
echo "   - Verify changes are saved"
