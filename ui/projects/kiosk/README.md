# Kiosk Application with WebSocket Card Reader

This Angular application provides a kiosk interface that receives card data via WebSocket from the card reader application.

## Features

- Real-time card data reception via WebSocket
- Automatic card data display
- Connection status monitoring
- Fallback HTTP card reading
- Modern, responsive UI

## Prerequisites

- Node.js 18+ and npm/pnpm
- Angular CLI
- Card reader application running

## Setup

1. Install dependencies:
```bash
cd ui/projects/kiosk
npm install
```

2. Build the Angular application:
```bash
ng build kiosk
```

3. Install WebSocket server dependencies:
```bash
npm install express ws
```

## Running the Application

### Option 1: Development Mode

1. Start the WebSocket server:
```bash
npm start
```

2. In another terminal, serve the Angular app:
```bash
ng serve kiosk --port 4201
```

### Option 2: Production Mode

1. Build the Angular app:
```bash
ng build kiosk --configuration production
```

2. Start the WebSocket server (serves both WebSocket and static files):
```bash
npm start
```

The kiosk will be available at: http://localhost:4201

## WebSocket Communication

The kiosk connects to the WebSocket server at `ws://localhost:4201/ws/card-reader` and automatically receives card data when:

1. A card is inserted into the card reader
2. The card reader application reads the card data
3. The data is sent via WebSocket to the kiosk

## Card Data Display

When card data is received, the kiosk displays:

- Name (first and last)
- ID Number
- Date of Birth
- Gender
- Nationality
- Address
- Issue/Expiry dates (if available)

## Connection Status

The kiosk shows two connection statuses:

1. **Card Reader Status**: HTTP connection to the card reader service
2. **WebSocket Status**: Real-time connection for card data

## Configuration

The WebSocket URL can be configured in the `WebSocketService`:

```typescript
// In websocket.service.ts
connect(url: string = 'ws://localhost:4201/ws/card-reader')
```

## Troubleshooting

- **WebSocket not connecting**: Ensure the WebSocket server is running on port 4201
- **No card data received**: Check that the card reader application is running and connected
- **Card reader not detected**: Verify PC/SC service and hardware connection

## Development

To run in development mode with hot reload:

```bash
# Terminal 1: WebSocket server
npm run dev

# Terminal 2: Angular dev server
ng serve kiosk --port 4201
```
