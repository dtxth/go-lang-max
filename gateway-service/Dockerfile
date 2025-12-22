# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY gateway-service/go.mod gateway-service/go.sum ./

# Copy dependency modules
COPY auth-service/ ../auth-service/
COPY chat-service/ ../chat-service/
COPY employee-service/ ../employee-service/
COPY structure-service/ ../structure-service/

# Download dependencies
RUN go mod download

# Copy source code
COPY gateway-service/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/gateway

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S gateway && \
    adduser -u 1001 -S gateway -G gateway

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown -R gateway:gateway /app
USER gateway

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]