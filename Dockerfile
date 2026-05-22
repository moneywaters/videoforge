# Multi-service Dockerfile for VideoForge
# Build from workspace root: docker build --build-arg SERVICE=ai-support -f Dockerfile .
ARG SERVICE

# Build stage
FROM golang:1.25-alpine AS builder
ARG SERVICE
WORKDIR /workspace

# Install git for module fetching
RUN apk add --no-cache git ca-certificates

# Copy Go workspace files to root
COPY go.work go.work

# Copy pkg module
COPY pkg/ ./pkg/

# Copy ALL svc-*/go.mod (for workspace resolution)
COPY svc-admin/go.mod svc-admin/go.mod
COPY svc-ai-support/go.mod svc-ai-support/go.mod
COPY svc-brief/go.mod svc-brief/go.mod
COPY svc-campaign/go.mod svc-campaign/go.mod
COPY svc-gateway/go.mod svc-gateway/go.mod
COPY svc-notification/go.mod svc-notification/go.mod
COPY svc-payout/go.mod svc-payout/go.mod
COPY svc-performance/go.mod svc-performance/go.mod
COPY svc-shopify/go.mod svc-shopify/go.mod
COPY svc-user/go.mod svc-user/go.mod
COPY svc-video/go.mod svc-video/go.mod

# Copy ALL service source code
COPY svc-admin/ ./svc-admin/
COPY svc-ai-support/ ./svc-ai-support/
COPY svc-brief/ ./svc-brief/
COPY svc-campaign/ ./svc-campaign/
COPY svc-gateway/ ./svc-gateway/
COPY svc-notification/ ./svc-notification/
COPY svc-payout/ ./svc-payout/
COPY svc-performance/ ./svc-performance/
COPY svc-shopify/ ./svc-shopify/
COPY svc-user/ ./svc-user/
COPY svc-video/ ./svc-video/

# Build the service (from workspace root so go.work is used)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./svc-${SERVICE}/cmd/main.go

# Final stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]