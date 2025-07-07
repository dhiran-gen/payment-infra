# payment-api

Go microservice for payment CRUD, Kafka producer, and DynamoDB cache.

## Build & Run

```sh
# Build
./config.sh && go build -o payment-api ./cmd/main.go
# Run
./config.sh && ./payment-api
```

## Docker

```sh
docker build -t payment-api .
```

## Endpoints
- `GET /healthz`
- `POST /payments`
- `GET /payments/:id`
- `PUT /payments/:id`
- `DELETE /payments/:id`
- `GET /payments`

## Config
All config via environment variables (see `config.sh`).
# payment-infra
