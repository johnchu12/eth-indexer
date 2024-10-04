FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o indexer cmd/indexer/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/internal/indexer/config.json ./internal/indexer/config.json

COPY --from=builder /app/indexer .

CMD ["./indexer"]