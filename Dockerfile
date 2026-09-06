# Multi-stage production build for LatinaApi
# Stage 1: Build binary
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install security certificates
RUN apk add --no-cache ca-certificates tzdata

# Copy go modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /build/bin/api \
    ./cmd/api

# Stage 2: Minimal distroless runtime
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy timezone data & SSL certs
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/bin/api /app/api

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/api"]
