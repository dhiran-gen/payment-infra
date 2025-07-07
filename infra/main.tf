terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
  required_version = ">= 1.3.0"
}

provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  endpoints {
    dynamodb = "http://localhost:4566"
    s3       = "http://localhost:4566"
    kinesis  = "http://localhost:4566"
    rds      = "http://localhost:4566"
    ecr      = "http://localhost:4566"
    kafka    = "http://localhost:4566"
  }
}

resource "aws_dynamodb_table" "payments" {
  name           = "payments"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"
  attribute {
    name = "id"
    type = "S"
  }
}

