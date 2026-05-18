# -----------------------------
# Stage 1: Build the Go binary
# -----------------------------
FROM golang:1.25-alpine AS builder

# Install certificates for HTTPS support
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
# CGO disabled so it can run in scratch
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o rideshare .

# -----------------------------
# Stage 2: Minimal runtime
# -----------------------------
FROM scratch

# Copy SSL certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /app/rideshare /rideshare

# Expose app port if needed
EXPOSE 5000
EXPOSE 5001

# Start application
ENTRYPOINT ["/rideshare", "--config", "/configuration.json"]