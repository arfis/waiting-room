// MongoDB initialization script for waiting room system

// Switch to the waiting_room database
db = db.getSiblingDB('waiting_room');

// Create collections with proper indexes
db.createCollection('queue_entries');

// Create indexes for optimal performance
db.queue_entries.createIndex(
  { "waitingRoomId": 1, "status": 1 },
  { name: "waiting_room_status_idx" }
);

db.queue_entries.createIndex(
  { "qrToken": 1 },
  { unique: true, name: "qr_token_unique_idx" }
);

db.queue_entries.createIndex(
  { "ticketNumber": 1 },
  { unique: true, name: "ticket_number_unique_idx" }
);

db.queue_entries.createIndex(
  { "createdAt": 1 },
  { name: "created_at_idx" }
);

// Create a sample waiting room configuration
db.createCollection('waiting_rooms');
db.waiting_rooms.insertOne({
  _id: "triage-1",
  name: "Triage Room 1",
  description: "Emergency department triage",
  isActive: true,
  createdAt: new Date(),
  updatedAt: new Date()
});

print('Waiting room database initialized successfully!');
