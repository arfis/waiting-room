#!/bin/bash

echo "=== Current MongoDB Configuration ==="
mongosh "mongodb://admin:admin@localhost:27017/waiting_room?authSource=admin" --quiet --eval "db.system_configuration.findOne({}, {externalAPI: 1})"

echo ""
echo "=== Testing Configuration Update ==="
echo "1. The configuration is now cached in memory"
echo "2. When you update it via the admin panel, it will:"
echo "   - Update MongoDB immediately"
echo "   - Reload the cache immediately"
echo "   - Use the new configuration for all subsequent API calls"
echo ""
echo "3. The system will also auto-reload every 30 seconds to catch external changes"
echo ""
echo "=== Current API Logs ==="
tail -10 /tmp/api.log
