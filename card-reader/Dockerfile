# Build stage
FROM golang:1.22-alpine AS builder
# (use whatever Go version you actually target, 1.25-alpine doesn't exist yet)

# Install build dependencies including PC/SC development libraries
RUN apk add --no-cache \
    git \
    pcsc-lite-dev \
    build-base        # <--- this brings in gcc, musl-dev, make, etc.

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
# GOOS=linux is already default in this image, but you can leave it if you like.
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install runtime dependencies (no compiler needed here)
RUN apk --no-cache add \
    ca-certificates \
    pcsc-lite \
    pcsc-lite-libs

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

EXPOSE 8081

CMD ["./main"]
