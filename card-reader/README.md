# Card Reader Application

This is a standalone Go application that reads smart cards and sends the data via WebSocket to the kiosk application.

## Features

- PC/SC smart card reader support
- Automatic card detection and reading
- Multiple card data extraction methods (PKCS#11 certificates, CPLC, UID, ATR hash)
- WebSocket communication with kiosk
- Configurable via environment variables

## Prerequisites

- Go 1.25.0 or later
- PC/SC smart card reader hardware
- OpenSC or compatible PKCS#11 middleware (optional, for certificate reading)

## Installation

1. Install dependencies:
```bash
go mod tidy
```

2. Build the application:
```bash
go build -o card-reader main.go
```

## Configuration

Set the following environment variables:

- `ROOM_ID`: Room identifier (default: "triage-1")
- `DEVICE_ID`: Device identifier (default: "reader-01")
- `READER_NAME`: Specific reader name (optional, uses first available if not set)
- `WS_URL`: WebSocket URL to send card data (default: "ws://localhost:4201/ws/card-reader")
- `PKCS11_MODULE`: Path to PKCS#11 module (optional, auto-detected if not set)

## Usage

1. Connect your smart card reader
2. Start the kiosk WebSocket server (see kiosk README)
3. Run the card reader:
```bash
./card-reader
```

The application will:
- Detect available card readers
- Wait for card insertion
- Read card data when inserted
- Send data to the kiosk via WebSocket
- Display card data in console

## Card Data Sources

The application tries to read card data in this order:

1. **PKCS#11 Certificates**: Public certificates from the card (requires PKCS#11 middleware)
2. **CPLC (Card Production Life Cycle)**: Chip serial number via GlobalPlatform
3. **UID**: Contactless card UID (vendor-specific command)
4. **ATR Hash**: Fallback using ATR (Answer To Reset) hash

## Troubleshooting

- **No readers found**: Ensure PC/SC service is running and reader is connected
- **PKCS#11 errors**: Install OpenSC or compatible middleware
- **WebSocket connection failed**: Ensure kiosk WebSocket server is running
- **Card reading fails**: Check card compatibility and reader support

## Development

To run in development mode with verbose logging:
```bash
go run main.go
```
