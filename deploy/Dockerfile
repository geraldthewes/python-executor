# Multi-stage Dockerfile for python-executor server

# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o python-executor-server ./cmd/server

# Stage 2: Runtime image
FROM docker:27-dind

# Install Python for testing/debugging
RUN apk add --no-cache python3 py3-pip ca-certificates

# Copy binary from builder
COPY --from=builder /build/python-executor-server /usr/local/bin/python-executor-server

# Create non-root user (note: server needs to run as root for Docker-in-Docker)
# But executed code will run as UID 1000

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
ENTRYPOINT ["/usr/local/bin/python-executor-server"]
