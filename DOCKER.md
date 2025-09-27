# Docker Setup for Waiting Room System

This document describes the Docker setup for the complete waiting room management system.

## 🐳 Docker Architecture

The system is containerized with the following services:

- **mongodb**: MongoDB database with initialization scripts
- **api**: Go API server with queue management
- **kiosk**: Angular kiosk app with WebSocket server
- **mobile**: Angular mobile app for queue tracking
- **tv**: Angular TV display for queue status
- **backoffice**: Angular backoffice for queue management
- **card-reader**: Go card reader application (optional)

## 🚀 Quick Start

### Option 1: Full Docker Setup (with MongoDB)

```bash
# Start all services including MongoDB
./start-docker.sh
```

### Option 2: Development Setup (with existing MongoDB)

```bash
# Use your existing MongoDB Docker container
./start-docker-dev.sh
```

## 📁 Docker Files Structure

```
waiting-room/
├── docker-compose.yml          # Full setup with MongoDB
├── docker-compose.dev.yml      # Development setup (external MongoDB)
├── start-docker.sh            # Full Docker startup script
├── start-docker-dev.sh        # Development Docker startup script
├── scripts/
│   └── mongo-init.js          # MongoDB initialization script
├── api/
│   ├── Dockerfile             # API server container
│   └── .dockerignore          # API build optimization
├── card-reader/
│   └── Dockerfile             # Card reader container
└── ui/
    ├── Dockerfile.kiosk       # Kiosk application container
    ├── Dockerfile.mobile      # Mobile application container
    ├── Dockerfile.tv          # TV display container
    ├── Dockerfile.backoffice  # Backoffice container
    └── .dockerignore          # UI build optimization
```

## 🔧 Service Configuration

### MongoDB
- **Image**: mongo:7.0
- **Port**: 27017
- **Database**: waiting_room
- **Initialization**: Automatic with indexes and collections

### API Server
- **Port**: 8080
- **Environment**: MONGODB_URI, ADDR
- **Health Check**: /health endpoint
- **Dependencies**: MongoDB

### Angular Applications
- **Kiosk**: Port 4201 (with WebSocket server)
- **Mobile**: Port 4202
- **TV**: Port 4203
- **Backoffice**: Port 4200
- **Build**: Multi-stage build with Node.js 18

### Card Reader
- **Privileged**: true (for USB device access)
- **Devices**: /dev/bus/usb (for smart card readers)
- **Environment**: ROOM_ID, DEVICE_ID, WS_URL

## 🛠️ Development Commands

```bash
# Build and start all services
docker-compose up --build -d

# Start with external MongoDB
docker-compose -f docker-compose.dev.yml up --build -d

# View logs for specific service
docker-compose logs -f api
docker-compose logs -f kiosk
docker-compose logs -f mobile
docker-compose logs -f tv
docker-compose logs -f backoffice

# Stop all services
docker-compose down

# Restart specific service
docker-compose restart api

# Access container shell
docker-compose exec api sh
docker-compose exec kiosk sh

# Remove all containers and volumes
docker-compose down -v
```

## 🔍 Troubleshooting

### MongoDB Connection Issues
```bash
# Check if MongoDB is running
docker ps | grep mongo

# Check MongoDB logs
docker-compose logs mongodb

# Test MongoDB connection
docker-compose exec api sh -c "wget -qO- http://localhost:8080/health"
```

### Port Conflicts
```bash
# Check which ports are in use
lsof -i :4200
lsof -i :4201
lsof -i :4202
lsof -i :4203
lsof -i :8080
lsof -i :27017

# Stop conflicting services
docker-compose down
```

### Build Issues
```bash
# Clean build cache
docker-compose build --no-cache

# Remove unused images
docker image prune -f

# Remove all unused resources
docker system prune -f
```

### Card Reader Issues
```bash
# Check USB device access
docker-compose exec card-reader lsusb

# Check card reader logs
docker-compose logs card-reader

# Restart card reader with privileged access
docker-compose restart card-reader
```

## 📊 Monitoring

### Health Checks
All services include health checks:
- **API**: HTTP GET /health
- **Kiosk**: HTTP GET /health
- **Mobile/TV/Backoffice**: HTTP GET /

### Logs
```bash
# View all logs
docker-compose logs

# Follow logs in real-time
docker-compose logs -f

# View logs for specific service
docker-compose logs -f api
```

### Resource Usage
```bash
# View container resource usage
docker stats

# View container details
docker-compose ps
```

## 🔒 Security Considerations

- All containers run as non-root users
- MongoDB is not exposed externally in production
- Card reader requires privileged access for USB devices
- Health checks are configured for all services
- Proper network isolation with custom bridge network

## 🚀 Production Deployment

For production deployment:

1. Use environment-specific Docker Compose files
2. Configure proper secrets management
3. Set up reverse proxy (nginx/traefik)
4. Configure SSL/TLS certificates
5. Set up monitoring and logging
6. Configure backup strategies for MongoDB
7. Use Docker Swarm or Kubernetes for orchestration

## 📝 Environment Variables

### API Server
- `MONGODB_URI`: MongoDB connection string
- `ADDR`: Server address and port

### Card Reader
- `ROOM_ID`: Waiting room identifier
- `DEVICE_ID`: Card reader device ID
- `WS_URL`: WebSocket server URL

### MongoDB
- `MONGO_INITDB_DATABASE`: Database name
