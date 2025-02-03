# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o storage-service ./cmd/server

# Final stage
FROM alpine:latest

# Install necessary certificates
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Create uploads directory
RUN mkdir -p /tmp/uploads

# Copy the built binary from the builder stage
COPY --from=builder /app/storage-service .

# Expose gRPC port
EXPOSE 50051

# Set environment variables
ENV UPLOAD_DIR=/tmp/uploads

# Run the binary
CMD ["./storage-service"]
