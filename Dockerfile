# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o fibo-monitor ./cmd

# Final stage
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/fibo-monitor .
COPY config/config.yaml ./config/config.yaml

# Create logs directory
RUN mkdir logs

EXPOSE 8080 9090

ENTRYPOINT ["./fibo-monitor"]
