# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

# Set working directory
WORKDIR /app

# Copy Go module files first for layer caching
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code into builder image
COPY src/ ./

# Build the statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o route-validator .

# Stage 2: Runtime container
FROM alpine:3.18

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Create a non-root user for the app
RUN adduser -D -g '' appuser

# Create mount point for TLS certs
RUN mkdir /certs

# Copy the built binary
COPY --from=builder /app/route-validator /route-validator

# Expose TLS cert path
VOLUME ["/certs"]

# Use non-root user
USER appuser

# Run the admission controller
ENTRYPOINT ["/route-validator"]