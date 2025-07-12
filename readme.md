# Small Payment System – End‑to‑End Roadmap

> **Goal**: Build a minimal yet realistic payment platform in Go 1.24.4 where **two microservices** communicate through **Kafka**, persist data in **PostgreSQL** + **DynamoDB**, and are delivered via **CI/CD → Docker images → AWS (simulated with LocalStack + Terraform)**—all developed on a macOS laptop with VS Code.

---

## 1. High‑Level Architecture

```text
┌──────────────┐        Kafka "payments.commands"          ┌────────────────┐
│  Payment API │ ───────────────────────────────────────▶ │ Payment Worker │
│  Service     │                                         │  Service       │
└─────┬────────┘                                         └──────┬─────────┘
      │ REST (JSON)                                           │
      ▼                                                       │
PostgreSQL  ←────────────────────── Kafka "payments.events" ──┘
      ▲                                                       ▼
      │                                              DynamoDB (idempotency / ledger snapshots)
 Local clients                                      
```

| Service            | Responsibility                                     | DB                        | External AWS‑like deps |
| ------------------ | -------------------------------------------------- | ------------------------- | ---------------------- |
| **payment‑api**    | AuthN/AuthZ, create/GET payments, publish commands | PostgreSQL (`payments`)   | S3 (invoice PDFs)      |
| **payment‑worker** | Consume commands, call mock gateway, emit events   | PostgreSQL (`tx_history`) | DynamoDB (ledger)      |

*Kafka* → Bitnami/Redpanda container in dev, MSK‑like config via LocalStack in infra.

---

## 2. Tech Stack

| Layer          | Choice (why)                                               |
| -------------- | ---------------------------------------------------------- |
| Language       | Go 1.24.4 – fits your background & strong concurrency      |
| HTTP Framework | [Gin](https://github.com/gin-gonic/gin) for API            |
| Kafka Client   | [Shopify/sarama](https://github.com/Shopify/sarama)        |
| DB Drivers     | `lib/pq` (Postgres), AWS SDK v2 for DynamoDB               |
| Infra as Code  | Terraform 1.9 with **localstack** provider overrides       |
| Containers     | Docker + docker‑compose (dev) & Kaniko/GitHub Actions (CI) |
| CI/CD          | GitHub Actions → build, test, Docker build → push to ECR   |
| Cloud Sim      | [LocalStack](https://localstack.cloud) (S3, Dynamo, ECR…)  |
| Orchestration  | docker‑compose for local; ECS‑Fargate (simulated) target   |
| Observability  | OpenTelemetry + Prometheus + Grafana (optional stretch)    |

---

## 3. Local Development Environment (macOS)

```bash
# 0. Prerequisites – Homebrew
brew install go@1.24 terraform awscli localstack/tap/localstack docker-compose

# 1. VS Code Extensions
code --install-extension golang.go ms-aws.aws-toolkit-vscode hashicorp.terraform
```

Create a **`dev` network** so all containers see each other:

```bash
docker network create dev
```

`docker-compose.yml` (root) will spin up:

* `localstack/localstack` (edge:4566)
* Postgres (bitnami/postgresql:16) → port 5432
* Kafka (bitnami/kafka\:latest) + ZK/Redpanda
* Optional: Grafana + Prometheus

---

## 4. Repository Layout

```text
payment-system/              # mono‑org folder (not mono-repo)
├── infra/                   # Terraform root (ECR, ECS, IAM, RDS…)
│   ├── modules/
│   └── main.tf
├── payment-api/             # service repo 1 (own GitHub repo)
│   ├── cmd/…
│   ├── internal/…
│   ├── Dockerfile
│   └── .github/workflows/
└── payment-worker/          # service repo 2 (own GitHub repo)
    ├── cmd/…
    ├── internal/…
    ├── Dockerfile
    └── .github/workflows/
```

> **Tip**: start each repo with `go mod init github.com/your‑org/payment‑api` etc.

---

## 5. Terraform & LocalStack Cheat‑Sheet

```hcl
provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  s3_force_path_style         = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  endpoints {
    dynamodb = "http://localhost:4566"
    s3       = "http://localhost:4566"
    ecr      = "http://localhost:4566"
  }
}

module "ecr" {
  source  = "terraform-aws-modules/ecr/aws"
  name    = "payment-api"
}
```

Run via:

```bash
export LOCALSTACK_HOST=localhost
localstack start -d   # background
cd infra && terraform init && terraform apply -auto-approve
```

> ⭐ Use `aws --endpoint-url=http://localhost:4566 ecr describe-repositories` to verify.

---

## 6. GitHub Actions Pipeline (sample for `payment-api`)

```yaml
name: CI
on:
  push:
    branches: [main]
jobs:
  build:
    runs-on: ubuntu-latest
    services:
      localstack:
        image: localstack/localstack:latest
        ports: ["4566:4566"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: {go-version: "1.24.4"}
      - run: go test ./...
      - name: Build image
        run: docker build -t payment-api:$GITHUB_SHA .
      - name: Login to ECR (LocalStack)
        run: |
          aws --endpoint-url=http://localhost:4566 ecr create-repository --repository-name payment-api || true
          aws --endpoint-url=http://localhost:4566 ecr get-login-password | \
            docker login --username AWS --password-stdin localhost:4566/payment-api
      - run: docker tag payment-api:$GITHUB_SHA localhost:4566/payment-api:$GITHUB_SHA
      - run: docker push localhost:4566/payment-api:$GITHUB_SHA
```

---

## 7. Step‑by‑Step Implementation Flow

| Week | Deliverable                                                         |
| ---- | ------------------------------------------------------------------- |
| 1    | Boilerplate repos, docker-compose, LocalStack up, Terraform apply   |
| 2    | **payment‑api**: REST CRUD, PostgreSQL migrations (golang‑migrate)  |
| 3    | Kafka producer integration (`sarama`), Outbox pattern demo          |
| 4    | **payment‑worker**: consumer & mock processor, events → DB + Dynamo |
| 5    | CI/CD pipelines (build/push), integration tests (TestContainers)    |
| 6    | Observability: OpenTelemetry traces, Prometheus metrics             |
| 7    | ECS deployment script (even if only hitting LocalStack)             |

> Adjust pace as per comfort; each step is independent & merge‑driven.

---

## 8. VS Code Tips

* **Dev Containers**: `.devcontainer.json` to open repo with Docker‑in‑Docker + LocalStack pre‑boot.
* **Go tools**: Enable `go.toolsManagement.autoUpdate`.
* **Terraform**: enable `terraform.languageServer` for linting.
* **AWS Toolkit**: point to `http://localhost:4566` to browse Dynamo/S3.

---

## 9. Next Actions for You

1. `git clone` the empty repos & set module names.
2. Build `docker-compose.yml` (copy from this doc’s snippet & tweak).
3. Initialise LocalStack & Terraform.
4. Scaffold `payment-api` with Gin router & health check.
5. Ping me back once step 4 runs (`curl :8080/healthz`), and we’ll layer Kafka publishing next! 🎯

---

### Happy building!
# payment-infra
# payment-infra
# payment-infra
