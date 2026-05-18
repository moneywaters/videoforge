# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /svc-user ./cmd/main.go

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN adduser -D -u 1000 appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /svc-user /usr/local/bin/svc-user

# Copy migrations
COPY ./migrations /app/migrations

# Set environment
ENV PORT=8080
ENV DATABASE_URL=postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the service
CMD ["svc-user"]