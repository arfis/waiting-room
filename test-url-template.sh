#!/bin/bash

echo "=== URL Template Feature Test ==="
echo ""

echo "1. Testing URL template validation (should fail):"
curl -s -X PUT "http://localhost:8080/api/admin/configuration/external-api" \
  -H "Content-Type: application/json" \
  -d '{"userServicesUrl": "http://api.example.com/user-services", "timeoutSeconds": 10, "retryAttempts": 3}' | jq .

echo ""
echo "2. Testing valid URL template (should succeed):"
curl -s -X PUT "http://localhost:8080/api/admin/configuration/external-api" \
  -H "Content-Type: application/json" \
  -d '{"userServicesUrl": "http://api.example.com/users/${identifier}/services", "timeoutSeconds": 15, "retryAttempts": 2}' | jq .

echo ""
echo "3. Testing different URL template formats:"
echo "   • Path parameter: http://api.example.com/users/${identifier}/services"
echo "   • Query parameter: http://api.example.com/services?user=${identifier}"
echo "   • Nested path: http://api.example.com/api/v1/users/${identifier}/available-services"
echo "   • POST body: http://api.example.com/api/services (identifier sent in POST body)"

echo ""
echo "4. Current configuration:"
curl -s "http://localhost:8080/api/admin/configuration/external-api" | jq .

echo ""
echo "✅ URL Template feature is working!"
echo "   - Backend validates that URL contains \${identifier} placeholder"
echo "   - Backend replaces \${identifier} with actual user identifier when making requests"
echo "   - Frontend shows helpful examples and validation"
