#!/bin/bash

echo "=== Testing Headers Configuration ==="

# Test 1: Get current configuration
echo "1. Getting current external API configuration..."
curl -s "http://localhost:8080/api/admin/configuration/external-api" | jq .

echo -e "\n2. Updating configuration with headers..."
# Test 2: Update configuration with headers
curl -X PUT "http://localhost:8080/api/admin/configuration/external-api" \
  -H "Content-Type: application/json" \
  -d '{
    "userServicesUrl": "https://private-4a985-invoice19.apiary-mock.com/waiting-room/medical/services?identifier=${identifier}",
    "timeoutSeconds": 15,
    "retryAttempts": 2,
    "headers": {
      "Authorization": "Bearer test-token-123",
      "X-API-Key": "api-key-456",
      "Content-Type": "application/json"
    }
  }' | jq .

echo -e "\n3. Verifying updated configuration..."
# Test 3: Verify the configuration was saved
curl -s "http://localhost:8080/api/admin/configuration/external-api" | jq .

echo -e "\n4. Testing external API call with headers..."
# Test 4: Test that the headers are actually used (this would be called by the kiosk service)
echo "Note: This would be tested when a card swipe occurs and the external API is called with the configured headers."

echo -e "\n✅ Headers configuration test completed!"
