# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

# Install required system dependencies (if any)
# RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN go build -o main .

# Run Stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
# Copy config folder if needed (assuming config is loaded from file or env)
# COPY --from=builder /app/config ./config 

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
