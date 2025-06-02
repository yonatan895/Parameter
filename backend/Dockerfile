FROM golang:1.20-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

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
