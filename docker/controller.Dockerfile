# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

# Set working directory
WORKDIR /app

# Copy Go module files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code into builder image
COPY . ./

# Build the statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o controller ./cmd/controller

# Stage 2: Runtime container
FROM alpine:3.18

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Create a non-root user for the app
RUN adduser --uid 1000 --disabled-password --gecos "" appuser

# Create mount point for TLS certs
RUN mkdir /certs

# Copy the built binary
COPY --from=builder /app/controller /usr/local/bin/controller

# Expose TLS cert path
VOLUME ["/certs"]

# Use non-root user
USER 1000

EXPOSE 8443 8443

# Run the admission controller
ENTRYPOINT ["controller"]