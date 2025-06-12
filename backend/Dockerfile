FROM golang:1.20-alpine AS build
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the code
COPY . .

RUN go vet ./...
RUN --mount=type=cache,target=/root/.cache/go-build \
    go test ./...
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o server .

FROM alpine
WORKDIR /app
COPY --from=build /app/server ./server
CMD ["./server"]
