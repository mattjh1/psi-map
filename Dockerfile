# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies and set up workspace in one layer
RUN apk add --no-cache git ca-certificates tzdata && \
    mkdir -p /app
WORKDIR /app

# Copy dependency files and download deps in one layer
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source and build in one layer
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -o psi-map .

# Final stage
FROM gcr.io/distroless/static-debian12:nonroot

# Copy everything needed in one layer
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/psi-map /app/psi-map
COPY --from=builder /app/internal/server/static /app/internal/server/static
COPY --from=builder /app/internal/server/templates /app/internal/server/templates

WORKDIR /app

# Expose port (adjust if your app uses a different port)
EXPOSE 8080

# Health check with wget alternative (distroless doesn't have wget)
# Using a simple network check instead
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/psi-map", "health"] || exit 1

# Set default command
ENTRYPOINT ["/app/psi-map"]

# Default arguments (can be overridden)
CMD ["server", "--port=8080"]
