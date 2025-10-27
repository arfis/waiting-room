const WebSocket = require('ws');
const http = require('http');
const express = require('express');
const path = require('path');

const app = express();
const server = http.createServer(app);

// Resolve kiosk dist directory relative to current directory (ui-nx)
const kioskDist = path.join(__dirname, 'dist', 'kiosk', 'browser');

// Serve static files from the kiosk dist directory
app.use(express.static(kioskDist));

// Create WebSocket server
const wss = new WebSocket.Server({ 
  server,
  path: '/ws/card-reader'
});

// Store connected clients
const clients = new Set();

wss.on('connection', (ws, req) => {
  console.log('New WebSocket connection from:', req.socket.remoteAddress);
  clients.add(ws);

  ws.on('message', (message) => {
    try {
      console.log('Received message type:', typeof message, 'Constructor:', message?.constructor?.name);

      // Convert to string if needed
      const messageStr = typeof message === 'string' ? message : message.toString();
      
      // Handle ping messages (check string content, not just exact match)
      if (messageStr === 'ping') {
        console.log('Received ping, sending pong');
        ws.send('pong');
        return;
      }

      console.log('Message content:', messageStr.substring(0, 100) + '...');

      // Try to parse as JSON
      let data;
      try {
        data = JSON.parse(messageStr);
      } catch (parseError) {
        console.log('Message is not JSON, treating as text message:', messageStr);
        // If it's not JSON, just forward it as-is
        clients.forEach(client => {
          if (client !== ws && client.readyState === WebSocket.OPEN) {
            client.send(messageStr);
          }
        });
        return;
      }

      console.log('Received card data:', data);

      // Broadcast to all connected clients (in this case, just the kiosk)
      clients.forEach(client => {
        if (client !== ws && client.readyState === WebSocket.OPEN) {
          client.send(messageStr);
        }
      });
    } catch (error) {
      console.error('Error handling message:', error);
    }
  });

  ws.on('close', () => {
    console.log('WebSocket connection closed');
    clients.delete(ws);
  });

  ws.on('error', (error) => {
    console.error('WebSocket error:', error);
    clients.delete(ws);
  });

  // Send welcome message
  ws.send(JSON.stringify({ 
    type: 'connection', 
    message: 'Connected to card reader WebSocket' 
  }));
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ 
    status: 'ok', 
    clients: clients.size,
    timestamp: new Date().toISOString()
  });
});

// Serve kiosk app for all other routes
app.get('*', (req, res) => {
  res.sendFile(path.join(kioskDist, 'index.html'));
});

const PORT = process.env.PORT || 4201;
server.listen(PORT, () => {
  console.log(`WebSocket server running on port ${PORT}`);
  console.log(`WebSocket endpoint: ws://localhost:${PORT}/ws/card-reader`);
  console.log(`Kiosk app: http://localhost:${PORT}`);
});
