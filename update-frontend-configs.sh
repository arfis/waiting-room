#!/bin/bash

echo "=== Updating Frontend Configurations ==="

# Read the shared config
CONFIG_FILE="ui/shared-config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: $CONFIG_FILE not found!"
    exit 1
fi

# Extract API URL from config
API_URL=$(jq -r '.development.apiUrl' "$CONFIG_FILE")
WS_URL=$(jq -r '.development.wsUrl' "$CONFIG_FILE")

echo "API URL: $API_URL"
echo "WebSocket URL: $WS_URL"

# Update admin environment
echo "Updating admin environment..."
cat > ui/projects/admin/src/environments/environment.ts << EOF
export const environment = {
  production: false,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'Admin Panel'
};
EOF

cat > ui/projects/admin/src/environments/environment.prod.ts << EOF
export const environment = {
  production: true,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'Admin Panel'
};
EOF

# Update backoffice environment
echo "Updating backoffice environment..."
cat > ui/projects/backoffice/src/environments/environment.ts << EOF
export const environment = {
  production: false,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'Backoffice'
};
EOF

cat > ui/projects/backoffice/src/environments/environment.prod.ts << EOF
export const environment = {
  production: true,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'Backoffice'
};
EOF

# Update kiosk environment
echo "Updating kiosk environment..."
cat > ui/projects/kiosk/src/environments/environment.ts << EOF
export const environment = {
  production: false,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'Kiosk'
};
EOF

cat > ui/projects/kiosk/src/environments/environment.prod.ts << EOF
export const environment = {
  production: true,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'Kiosk'
};
EOF

# Update tv environment
echo "Updating tv environment..."
cat > ui/projects/tv/src/environments/environment.ts << EOF
export const environment = {
  production: false,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'TV Display'
};
EOF

cat > ui/projects/tv/src/environments/environment.prod.ts << EOF
export const environment = {
  production: true,
  apiUrl: '$API_URL',
  wsUrl: '$WS_URL',
  appName: 'TV Display'
};
EOF

echo "✅ All frontend configurations updated!"
echo ""
echo "To change the backend URL, edit ui/shared-config.json and run this script again."
