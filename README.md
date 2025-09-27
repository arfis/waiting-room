# Waiting Room Management System

A complete waiting room management system with smart card reading, queue management, and real-time updates.

## 🏗️ System Architecture

The system consists of several components:

- **Card Reader** (Go): Standalone smart card reader that communicates via WebSocket
- **API Server** (Go): REST API with MongoDB for queue management
- **Kiosk App** (Angular): Card reading interface with ticket generation
- **Mobile App** (Angular): Queue position tracking via QR codes
- **TV Display** (Angular): Real-time queue status display
- **Backoffice** (Angular): Queue management interface

## 🚀 Quick Start

### Option 1: Docker (Recommended)

1. **Prerequisites**:
   - Docker & Docker Compose
   - Smart card reader hardware (optional)

2. **Start with Docker (full setup)**:
   ```bash
   chmod +x start-docker.sh
   ./start-docker.sh
   ```

3. **Start with Docker (using existing MongoDB)**:
   ```bash
   chmod +x start-docker-dev.sh
   ./start-docker-dev.sh
   ```

### Option 2: Local Development

1. **Prerequisites**:
   - Go 1.25+
   - Node.js 18+
   - MongoDB (or Docker MongoDB)
   - Smart card reader hardware

2. **Start the system**:
   ```bash
   chmod +x start-system.sh
   ./start-system.sh
   ```

### Access the applications:
- **Kiosk**: http://localhost:4201
- **Mobile**: http://localhost:4204
- **TV Display**: http://localhost:4203
- **Backoffice**: http://localhost:4200
- **API**: http://localhost:8080

## 📱 User Flow

1. **Card Reading**: User inserts smart card at kiosk
2. **Ticket Generation**: System generates ticket number and QR code
3. **Queue Entry**: User is added to waiting queue
4. **Mobile Tracking**: User scans QR code to track position
5. **Queue Management**: Staff uses backoffice to call next person
6. **TV Display**: Shows current queue status to all users

## 🔧 Technical Details

### Card Reader
- Reads smart cards via PC/SC
- Supports PKCS#11 certificates and CPLC data
- Real-time WebSocket communication
- Automatic state management

### API Endpoints
- `POST /waiting-rooms/{roomId}/swipe` - Create queue entry
- `GET /queue-entries/token/{qrToken}` - Get queue position
- `POST /waiting-rooms/{roomId}/next` - Call next person
- `GET /api/waiting-rooms/{roomId}/queue` - Get current queue

### Database Schema
```javascript
{
  _id: ObjectId,
  waitingRoomId: String,
  ticketNumber: String,
  qrToken: String,
  status: String, // WAITING, CALLED, IN_SERVICE, COMPLETED, etc.
  position: Number,
  createdAt: Date,
  updatedAt: Date,
  cardData: {
    idNumber: String,
    firstName: String,
    lastName: String,
    // ... other card fields
  }
}
```

## 🛠️ Development

### Docker Management

```bash
# Start all services
docker-compose up -d

# Start with existing MongoDB
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose logs -f [service-name]

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up --build -d

# Access container shell
docker-compose exec [service-name] sh
```

### Building Individual Components

```bash
# API
cd api && go build ./cmd/api

# Card Reader
cd card-reader && go run main.go

# Angular Apps
cd ui
ng build kiosk
ng build mobile
ng build tv
ng build backoffice
```

### Environment Variables

- `MONGODB_URI`: MongoDB connection string (default: mongodb://localhost:27017)
- `ROOM_ID`: Waiting room identifier (default: triage-1)
- `DEVICE_ID`: Card reader device ID (default: reader-01)

## 📋 Features

- ✅ Smart card reading with multiple data sources
- ✅ Real-time WebSocket communication
- ✅ QR code generation and scanning
- ✅ Queue position tracking
- ✅ Staff queue management interface
- ✅ Public queue display
- ✅ MongoDB persistence
- ✅ Responsive design
- ✅ Auto-refresh and real-time updates

## 🔒 Security Considerations

- Card data is stored securely in MongoDB
- QR tokens are UUIDs for security
- API endpoints include proper validation
- CORS configured for development

## 📝 License

This project is for demonstration purposes.
