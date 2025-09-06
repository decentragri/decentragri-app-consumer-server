# Build stage
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install git for dependency resolution
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy environment file (optional, can be overridden)
COPY --from=builder /app/.env .env

# Expose port 9085
EXPOSE 9085

# Command to run the application
CMD ["./main"]
