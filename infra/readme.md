# Infra

This directory contains Terraform scripts to provision AWS resources in LocalStack for local development.

## Resources Provisioned
- DynamoDB table (payments)
- RDS Postgres instance
- MSK (Kafka) cluster
- ECR repositories for payment-api and payment-worker

## Usage

1. Ensure LocalStack is running (see root docker-compose.yml).
2. Run `terraform init` and `terraform apply` in this directory.
3. Use the outputs to configure your services.
