FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Tidy and download dependencies
RUN go mod tidy && go mod download

# Copy source code
COPY . .

# Run tidy again with all source files
RUN go mod tidy

# Build the application (without CGO)
RUN go build -o pdd-action ./cmd/pdd-action

# Use a small image for the final container
FROM alpine:latest

# Install git and certificates for HTTPS
RUN apk add --no-cache git ca-certificates

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /app/pdd-action /pdd-action

# Make the binary executable
RUN chmod +x /pdd-action

# Set the entrypoint
ENTRYPOINT ["/pdd-action"]