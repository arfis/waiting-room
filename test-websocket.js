const WebSocket = require('ws');

// Test WebSocket connection and send mock card data
const ws = new WebSocket('ws://localhost:4201/ws/card-reader');

ws.on('open', () => {
  console.log('✅ Connected to WebSocket server');
  
  // Send mock card data
  const mockCardData = {
    deviceId: "reader-01",
    roomId: "triage-1",
    token: "TEST123456789",
    reader: "Test Reader",
    atr: "3B7F1800000031C0739E01000FE6509020080F",
    protocol: "T=1",
    occurredAt: new Date().toISOString(),
    cardData: {
      id_number: "123456789",
      first_name: "John",
      last_name: "Doe",
      date_of_birth: "1990-01-01",
      gender: "M",
      nationality: "US",
      address: "123 Main St, City, State",
      source: "test"
    }
  };
  
  console.log('📤 Sending mock card data...');
  ws.send(JSON.stringify(mockCardData));
  
  setTimeout(() => {
    console.log('🔌 Closing connection...');
    ws.close();
  }, 2000);
});

ws.on('message', (data) => {
  console.log('📥 Received:', JSON.parse(data));
});

ws.on('error', (error) => {
  console.error('❌ WebSocket error:', error);
});

ws.on('close', () => {
  console.log('🔌 Connection closed');
  process.exit(0);
});
