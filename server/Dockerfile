FROM golang:1.23.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/slow-server

FROM alpine:3.18

COPY --from=builder /app/slow-server /usr/local/bin/slow-server

ENV SERVER_PORT=8080 \
    MIN_DELAY=500 \
    MAX_DELAY=3000 \
    SIMULATE_ERRORS=true \
    ERROR_RATE=0.15

EXPOSE 8080

CMD ["slow-server"]
