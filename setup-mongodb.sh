#!/bin/bash

echo "🗄️  Setting up MongoDB for Waiting Room System..."

# Check if MongoDB Docker container is running
if docker ps | grep -q mongodb; then
    echo "✅ MongoDB Docker container is running"
else
    echo "🚀 Starting MongoDB Docker container..."
    docker run -d -p 27017:27017 --name mongodb mongo:7.0
    sleep 5
fi

# Wait for MongoDB to be ready
echo "⏳ Waiting for MongoDB to be ready..."
until docker exec mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; do
    sleep 2
done
echo "✅ MongoDB is ready"

# Initialize the database
echo "🔧 Initializing database..."
docker exec mongodb mongosh --eval "
db = db.getSiblingDB('waiting_room');
if (db.queue_entries.countDocuments() === 0) {
  db.createCollection('queue_entries');
  db.queue_entries.createIndex({ 'waitingRoomId': 1, 'status': 1 });
  db.queue_entries.createIndex({ 'qrToken': 1 }, { unique: true });
  db.queue_entries.createIndex({ 'ticketNumber': 1 }, { unique: true });
  db.queue_entries.createIndex({ 'createdAt': 1 });
  print('Database initialized successfully!');
} else {
  print('Database already initialized');
}
"

echo "🎉 MongoDB setup complete!"
echo "   Connection string: mongodb://localhost:27017/waiting_room"
echo "   Database: waiting_room"
echo "   Collections: queue_entries"
