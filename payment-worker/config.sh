# payment-worker config for local/dev
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export POSTGRES_DB=payments
export KAFKA_BROKERS=kafka:9092
export DYNAMO_ENDPOINT=http://localhost:8000
export DYNAMO_TABLE=payments_cache
export AWS_ACCESS_KEY_ID=fake
export AWS_SECRET_ACCESS_KEY=fake
export AWS_REGION=us-east-1
