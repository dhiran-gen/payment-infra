# Dockerfile for payment-worker
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git && go mod download && go build -o payment-worker ./cmd/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/payment-worker .
COPY config.sh .
CMD ["/bin/sh", "-c", ". ./config.sh && ./payment-worker"]
