FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o server main.go

FROM alpine
WORKDIR /app
COPY --from=build /app/server ./server
CMD ["./server"]
