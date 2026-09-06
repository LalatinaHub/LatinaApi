# syntax=docker/dockerfile:1

# Multi-stage production build for LatinaApi
# Stage 1: Build statically linked binary
ARG GO_VERSION=1.27
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /build

# Install CA certificates and timezone database
RUN apk add --no-cache ca-certificates tzdata

# Leverage Docker cache for module dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy application source code
COPY . .

# Target OS and architecture provided automatically by Docker Buildx
ARG TARGETOS
ARG TARGETARCH

# Build statically linked binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -o /build/bin/api \
    ./cmd/api

# Stage 2: Minimal distroless runtime
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy timezone data & SSL certs from builder stage
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chown=nonroot:nonroot /build/bin/api /app/api

# Default container environment variables
ENV PORT=8080 \
    APP_ENV=production

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/api"]
