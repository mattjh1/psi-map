# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies including swag for API docs
RUN apk add --no-cache git ca-certificates tzdata && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    mkdir -p /app

WORKDIR /app

# Copy dependency files and download deps in one layer
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source
COPY . .

# Generate swagger docs for API
RUN swag init -g cmd/api/main.go -o docs/

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Build both CLI and API binaries
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -trimpath \
    -o psi-map cmd/cli/main.go && \
    CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -trimpath \
    -o psi-map-api cmd/api/main.go

# Final stage
FROM gcr.io/distroless/static-debian12:nonroot

# Copy everything needed in one layer
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/psi-map /app/psi-map
COPY --from=builder /app/psi-map-api /app/psi-map-api
COPY --from=builder /app/internal/server/static /app/internal/server/static
COPY --from=builder /app/internal/server/templates /app/internal/server/templates
COPY --from=builder /app/docs /app/docs

WORKDIR /app

# Expose port
EXPOSE 8080

# Health check - defaults to API health but can be overridden
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/psi-map-api", "--help"] || exit 1

# Default to CLI mode (maintains backward compatibility)
ENTRYPOINT ["/app/psi-map"]
CMD ["--help"]
