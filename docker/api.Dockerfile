FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/env ./env

COPY --from=builder /app/api .

EXPOSE 8080

CMD ["./api"]