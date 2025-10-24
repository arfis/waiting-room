#!/bin/bash

echo "=== Current MongoDB Configuration ==="
mongosh "mongodb://admin:admin@localhost:27017/waiting_room?authSource=admin" --quiet --eval "db.system_configuration.find().pretty()"

echo ""
echo "=== Tailing API logs (press Ctrl+C to stop) ==="
tail -f /tmp/api.log

