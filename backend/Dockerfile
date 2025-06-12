# Build stage
FROM golang:1.20-alpine AS build
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source
COPY . .

# Run static analysis and tests before build
RUN go vet ./... && go test ./...

# Build binary for Linux
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Minimal image
FROM alpine:3.18

WORKDIR /app

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Copy binary
COPY --from=build /app/server .

# Optionally set non-root user (create user:group 10001:10001)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Set permissions
RUN chmod +x /app/server

ENTRYPOINT ["./server"]
