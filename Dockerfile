# Multi-service Dockerfile for VideoForge
# Build from workspace root: docker build --build-arg SERVICE=gateway -f Dockerfile .
ARG SERVICE

# Build stage
FROM golang:1.23-alpine AS builder
ARG SERVICE
WORKDIR /workspace

# Install git for module fetching
RUN apk add --no-cache git ca-certificates

# Copy workspace files (without go.work — service go.mod handles deps via replace)
COPY go.mod ./
COPY pkg/ ./pkg/
COPY svc-${SERVICE}/ ./svc-${SERVICE}/

# Build the service
WORKDIR /workspace/svc-${SERVICE}
ENV GOTOOLCHAIN=auto
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/main.go

# Final stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
