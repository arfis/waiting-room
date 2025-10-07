# Waiting Room System

A complete waiting room management system with real-time queue updates, card reader integration, and multiple client applications.

## System Architecture

The system consists of several interconnected components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Card Reader   │    │   Kiosk App     │    │   Mobile App    │
│   (Hardware)    │───▶│   (Angular)     │    │   (Angular)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────────────────────────────┐
                       │           API Server (Go)              │
                       │  - REST API endpoints                  │
                       │  - WebSocket for real-time updates     │
                       │  - MongoDB integration                 │
                       └─────────────────────────────────────────┘
                                │
                                ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Backoffice    │    │   TV Display    │
                       │   (Angular)     │    │   (Angular)     │
                       └─────────────────┘    └─────────────────┘
```

## Components Overview

### 1. API Server (Go)
- **Port**: 8080
- **Technology**: Go with Chi router, WebSocket support
- **Database**: MongoDB
- **Features**:
  - REST API for queue management
  - WebSocket for real-time updates
  - Card reader integration
  - Queue position management

### 2. Kiosk Application (Angular)
- **Port**: 4200
- **Purpose**: Card reading and ticket generation
- **Features**:
  - Card reader integration
  - QR code generation
  - Ticket display

### 3. Mobile Application (Angular)
- **Port**: 4204
- **Purpose**: Queue status for patients
- **Features**:
  - QR code scanning
  - Queue position display
  - Estimated wait time

### 4. Backoffice Application (Angular)
- **Port**: 4201
- **Purpose**: Queue management for staff
- **Features**:
  - Call next patient
  - Finish current patient
  - Queue overview
  - Real-time updates

### 5. TV Display (Angular)
- **Port**: 4203
- **Purpose**: Public queue display
- **Features**:
  - Current patient display
  - Waiting queue
  - Real-time updates

## Prerequisites

- **Go** 1.21+
- **Node.js** 18+
- **MongoDB** 6.0+
- **Angular CLI** 17+

## Quick Start

### Using Makefile (Recommended)

The system includes comprehensive Makefiles for easy management:

#### Root Level Commands
```bash
# Quick start
make setup          # Complete system setup
make start          # Start the entire system
make stop           # Stop the entire system
make restart        # Restart the entire system

# Development
make dev            # Start development environment
make dev-stop       # Stop development environment

# Docker
make docker-start   # Start with Docker
make docker-stop    # Stop Docker containers
make docker-build   # Build Docker images

# Testing
make test           # Test entire system
make status         # Show system status

# Utilities
make clean          # Clean all build artifacts
make logs           # Show system logs
make help           # Show all available commands
```

#### API Level Commands
```bash
cd api

# Development
make gen            # Generate API code from OpenAPI spec
make build          # Build API server
make run            # Run API server
make test           # Run API tests

# UI Management
make install-ui     # Install UI dependencies
make build-ui       # Build all UI applications
make serve-ui       # Start all UI applications

# Individual UI Apps
make serve-kiosk    # Start Kiosk application
make serve-mobile   # Start Mobile application
make serve-tv       # Start TV application
make serve-backoffice # Start Backoffice application

# System Management
make start-system   # Start complete system
make start-docker   # Start with Docker
make stop-docker    # Stop Docker containers

# Database
make mongo-start    # Start MongoDB
make mongo-stop     # Stop MongoDB
make mongo-shell    # Connect to MongoDB shell

# Testing
make test-system    # Test system components
make test-api       # Test API endpoints
make test-ui        # Test UI applications

# Development Workflow
make dev-setup      # Set up development environment
make dev-start      # Start development environment
make dev-stop       # Stop development environment
```

### Manual Setup (Alternative)

### 1. Clone and Setup

```bash
git clone <repository-url>
cd waiting-room
```

### 2. Start MongoDB

```bash
# Using Docker
docker run -d --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongodb:6.0

# Or using local MongoDB installation
mongod --dbpath /path/to/your/db
```

### 3. Start the System

```bash
# Start all services
./start-system.sh
```

This script will:
- Build the UI library
- Build all Angular applications
- Start the API server
- Start all frontend applications

### 4. Access the Applications

- **Kiosk**: http://localhost:4200
- **Backoffice**: http://localhost:4201
- **TV Display**: http://localhost:4203
- **Mobile**: http://localhost:4204
- **API**: http://localhost:8080

## Manual Setup

### 1. API Server

```bash
cd api
go mod download
go run cmd/api/main.go
```

### 2. Frontend Applications

```bash
cd ui

# Build shared UI library
ng build ui

# Start individual applications
ng serve kiosk --port 4200
ng serve backoffice --port 4201
ng serve tv --port 4203
ng serve mobile --port 4204
```

## Configuration

The system uses a YAML configuration file for all settings. The server looks for `config.yaml` in the API directory by default.

### Configuration Files

- `config.yaml` - Main configuration file
- `config.example.yaml` - Example configuration with all options
- `config.dev.yaml` - Development configuration
- `config.prod.yaml` - Production configuration

### Basic Configuration

```yaml
server:
  port: 8080
  host: "localhost"
  
database:
  mongodb:
    uri: "mongodb://admin:admin@localhost:27017/waiting_room?authSource=admin"
    database: "waiting_room"
    
cors:
  allowed_origins:
    - "http://localhost:4200"
    - "http://localhost:4201"
    - "http://localhost:4203"
    - "http://localhost:4204"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "*"

websocket:
  enabled: true
  path: "/ws/queue"

rooms:
  default_room: "triage-1"  # Default room ID
  allow_wildcard: false     # Restrict access to configured rooms
  rooms:
    - id: "triage-1"
      name: "Triage Room"
      service_points:
        - id: "window-1"
          name: "Main Triage Desk"
          description: "Primary service desk"
        - id: "window-2"
          name: "Secondary Triage Desk"
          description: "Overflow service desk"

logging:
  level: "info"
  format: "text"
```

### Configuration Sections

#### Server Configuration
- `port`: Server port (default: 8080)
- `host`: Server host (default: localhost)

#### Database Configuration
- `mongodb.uri`: MongoDB connection string
- `mongodb.database`: Database name

#### CORS Configuration
- `allowed_origins`: List of allowed origins for CORS
- `allowed_methods`: List of allowed HTTP methods
- `allowed_headers`: List of allowed headers

#### WebSocket Configuration
- `enabled`: Enable/disable WebSocket support
- `path`: WebSocket endpoint path

#### Rooms Configuration
- `default_room`: Default room ID used by the system
- `allow_wildcard`: Allow any room ID (.* pattern) - set to false for strict mode

**Dynamic Room Management**: The system supports dynamic room creation. Set `allow_wildcard: true` to allow any room ID to be used dynamically (useful for prototyping).

**Strict Mode**: With `allow_wildcard: false` (default when rooms are defined), only the configured rooms are accepted by the API.

#### Logging Configuration
- `level`: Log level (debug, info, warn, error)
- `format`: Log format (text, json)

### Using Different Configuration Files

```bash
# Use development configuration
CONFIG_PATH=config.dev.yaml go run cmd/api/main.go

# Use production configuration
CONFIG_PATH=config.prod.yaml go run cmd/api/main.go
```

## API Endpoints

### Configuration
- `GET /api/config` - Retrieve default room, available rooms, and websocket path

### Queue Management
- `GET /api/waiting-rooms/{roomId}/queue` - Get queue entries for any room
- `POST /api/waiting-rooms/{roomId}/swipe` - Create new queue entry in any room
- `POST /api/waiting-rooms/{roomId}/next` - Call next patient in any room
- `POST /api/waiting-rooms/{roomId}/finish` - Finish current patient in any room

### WebSocket
- `WS /ws/queue/{roomId}` - Real-time queue updates for any room

### Dynamic Room Examples
```bash
# Use different rooms dynamically
curl -X POST http://localhost:8080/api/waiting-rooms/emergency/swipe
curl -X POST http://localhost:8080/api/waiting-rooms/pediatrics/swipe
curl -X POST http://localhost:8080/api/waiting-rooms/consultation-1/swipe

# Get queue for any room
curl http://localhost:8080/api/waiting-rooms/emergency/queue
curl http://localhost:8080/api/waiting-rooms/pediatrics/queue
```

### Health Check
- `GET /health` - Server health status

## Usage Flow

### 1. Patient Arrival
1. Patient approaches kiosk
2. Staff swipes patient's card
3. System generates ticket number and QR code
4. Patient receives ticket with QR code

### 2. Queue Management
1. Staff uses backoffice to call next patient
2. System updates queue positions
3. All displays update in real-time
4. Patient can scan QR code to check position

### 3. Service Completion
1. Staff finishes current patient
2. System marks patient as completed
3. Queue positions recalculate automatically

## Development

### Project Structure

```
waiting-room/
├── api/                    # Go API server
│   ├── cmd/api/           # Main application
│   ├── internal/          # Internal packages
│   │   ├── queue/         # Queue service
│   │   ├── rest/          # HTTP handlers
│   │   ├── repository/    # Data access layer
│   │   └── service/       # Business logic
│   └── config.yaml        # Configuration file
├── ui/                    # Angular applications
│   ├── projects/
│   │   ├── kiosk/         # Kiosk application
│   │   ├── mobile/        # Mobile application
│   │   ├── backoffice/    # Backoffice application
│   │   ├── tv/            # TV display
│   │   └── ui/            # Shared UI library
│   └── dist/              # Built applications
└── start-system.sh        # System startup script
```

### Building for Production

```bash
# Build API server
cd api
go build -o waiting-room-api cmd/api/main.go

# Build Angular applications
cd ui
ng build kiosk --configuration production
ng build mobile --configuration production
ng build backoffice --configuration production
ng build tv --configuration production
```

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Kill processes on specific ports
   lsof -ti:8080 | xargs kill -9
   lsof -ti:4200 | xargs kill -9
   ```

2. **MongoDB connection failed**
   - Check if MongoDB is running
   - Verify connection string in config.yaml
   - Check firewall settings

3. **WebSocket connection failed**
   - Check if API server is running
   - Verify CORS settings
   - Check browser console for errors

4. **Angular build errors**
   ```bash
   # Clean and rebuild
   cd ui
   rm -rf dist/
   ng build ui
   ```

### Logs

- **API Server**: Check terminal output
- **Angular Apps**: Check browser console
- **MongoDB**: Check MongoDB logs

## Environment Variables

You can override configuration using environment variables:

```bash
# Server configuration
export WAITING_ROOM_PORT=8080
export WAITING_ROOM_HOST=localhost

# Database configuration
export MONGODB_URI="mongodb://admin:admin@localhost:27017/waiting_room?authSource=admin"
export MONGODB_DATABASE="waiting_room"

# Logging configuration
export LOG_LEVEL="debug"
export LOG_FORMAT="json"

# WebSocket configuration
export WEBSOCKET_ENABLED="true"

# Room configuration
export DEFAULT_ROOM="triage-1"

# Configuration file path
export CONFIG_PATH="config.dev.yaml"
```

### Environment Variable Priority

1. Environment variables (highest priority)
2. Configuration file values
3. Default values (lowest priority)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

[Add your license information here]
