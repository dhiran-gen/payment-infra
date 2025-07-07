# payment-worker

Go microservice for consuming payment events from Kafka.

## Build & Run

```sh
./config.sh && go build -o payment-worker ./cmd/main.go
./config.sh && ./payment-worker
```

## Docker

docker build -t payment-worker .

## Config
All config via environment variables (see `config.sh`).
