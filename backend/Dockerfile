FROM golang:1.20-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go vet ./... && go test ./...
RUN go build -o server .

FROM alpine
WORKDIR /app
COPY --from=build /app/server ./server
CMD ["./server"]
