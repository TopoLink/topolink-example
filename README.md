# Topolink Example — Order & Settlement Service

A small but **realistic** Go microservice used to demo and validate [Topolink](https://github.com/TopoLink/topolink).

It is deliberately a normal service — layered packages, JWT auth, gRPC, Kafka, PostgreSQL, Redis, migrations, structured logs, graceful shutdown and tests — with a few risky patterns that real order systems tend to accumulate.

> ⚠️ The service intentionally contains anti-patterns (full-table updates, remote calls inside transactions, weak legacy auth). Do **not** copy it into production.

[![Go](https://img.shields.io/badge/go-1.23-blue)](https://go.dev/)
[![Analyzed by Topolink](https://img.shields.io/badge/analyzed%20by-Topolink-22d3ee)](https://github.com/TopoLink/topolink)

## Stack

| Layer | Choice |
|---|---|
| HTTP | Gin + JWT middleware |
| gRPC | internal order lookup |
| Database | PostgreSQL + GORM |
| Cache / locks | Redis (`go-redis/v9`) |
| Events | Kafka (`segmentio/kafka-go`) |
| Ops | `log/slog`, graceful shutdown, SQL migrations, Docker Compose |

## Architecture

```
cmd/api            HTTP + gRPC servers, graceful shutdown
cmd/consumer       Kafka payment.succeeded consumer
internal/
  config           env configuration
  model            order / payment / ledger / settlement
  repository       GORM + Redis persistence
  service          order, refund, settlement use cases
  gateway          payment provider HTTP client
  transport/http   Gin router, JWT auth, handlers
  transport/grpc   internal order lookup
  event            Kafka consumer
  platform         DB / Redis / logger wiring
migrations         PostgreSQL schema
deploy             Dockerfile + docker-compose
```

## Run locally

```bash
cp .env.example .env
docker compose -f deploy/docker-compose.yml up -d postgres redis redpanda

go run ./cmd/api        # HTTP :8080, gRPC :9090
go run ./cmd/consumer   # payment.succeeded consumer
```

## API

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/healthz`, `/readyz` | — | liveness / readiness |
| GET | `/api/v1/catalog` | — | public catalog |
| GET | `/api/v1/orders` | JWT | list the current user's orders |
| POST | `/api/v1/orders` | JWT | place an order |
| GET | `/api/v1/orders/:id` | JWT | order detail (cached) |
| POST | `/api/v1/orders/:id/cancel` | JWT | cancel a pending order |
| POST | `/api/v1/orders/:id/refund` | JWT | refund a paid order |
| GET | `/api/v1/admin/orders` | admin | all orders |
| POST | `/api/v1/admin/orders/bulk-status` | admin | bulk status update |
| POST | `/api/v1/admin/settlement/run` | admin | run a settlement batch |
| GET | `/legacy/orders` | legacy | deprecated mobile endpoint |

## Run Topolink

No configuration is required — the ecosystems are detected from `go.mod`.

```bash
# full scan
go run github.com/topolink/topolink/cmd/topolink@v1 -repo . -format markdown

# with the pre-built semantic cache (adds business meaning + weak-auth detection)
go run github.com/topolink/topolink/cmd/topolink@v1 -repo . \
  -semantic testdata/semantic.json -format markdown

# diff-driven blast radius
go run github.com/topolink/topolink/cmd/topolink@v1 -repo . \
  -diff-file testdata/change.patch -format markdown
```

Or use the GitHub Action:

```yaml
- uses: TopoLink/topolink@v1
  with:
    fail-on: none
    diff: auto
```

## What Topolink finds here

The service is intentionally shaped so that every rule fires on a **realistic** code path:

| Rule | Where | Why it matters |
|---|---|---|
| `write-without-where` | `OrderRepository.BulkSetOrderStatus` | admin bulk update can rewrite every order |
| `write-in-loop` | `OrderRepository.InsertOrder` | one INSERT per line item |
| `query-in-loop` | `OrderService.ListOrders` | one item query per order (N+1) |
| `remote-in-transaction` | `OrderService.RefundOrder` | payment provider call keeps the DB transaction open |
| `remote-under-lock` | `SettlementService.RunSettlement` | bank call while holding the process mutex |
| `missing-timeout` | `gateway.PaymentClient` | `http.Post` uses the default client (no timeout) |
| `money-write` | `LedgerRepository` / `accounts` | balance changes need idempotency + audit |
| `cache-no-ttl` | `Cache.SetOrder` | order snapshots never expire |
| `unchecked-write` | `OrderRepository.TouchOrder` | best-effort update ignores the error |
| `missing-auth` | `GET /api/v1/catalog` | public by design; Topolink surfaces it for review |
| `weak-auth` | `LegacyAuth` middleware | logs a missing token and continues |

Example output (14 findings):

```
## Topolink — Code Blast Radius Report
**14 finding(s)** — S0:1 · S1:7 · S2:6 · S3:0

| Rule | Severity | Entry | Location |
|------|----------|-------|----------|
| write-without-where | critical | http:POST:/api/v1/admin/orders/bulk-status | internal/repository/order.go:93 |
| remote-in-transaction | high | http:POST:/api/v1/orders/:id/refund  | internal/gateway/payment.go:37 |
| remote-under-lock    | high | http:POST:/api/v1/admin/settlement/run | internal/gateway/payment.go:57 |
| money-write          | high | http:POST:/api/v1/orders/:id/refund  | internal/repository/ledger.go:23 |
| ...                  |      |                                      | |
| weak-auth            | medium | http:GET:/legacy/orders            | internal/transport/httpapi/router.go:52 |
```

## Semantic cache

`testdata/semantic.json` is a **pre-generated AI annotation cache**. It was produced once with a model, then committed so CI can demonstrate annotation-driven rules (`weak-auth`) and the `Condition → Action → Business` effects without an API key.

To regenerate it with a real model:

```bash
OPENAI_API_KEY=sk-... go run github.com/topolink/topolink/cmd/topolink@v1 -repo . \
  -semantic testdata/semantic.json -ai-model gpt-4o-mini
```

The cache is keyed by a content fingerprint, so unchanged code is never re-sent to the model.

## Blast-radius demo

`testdata/change.patch` modifies `OrderService.RefundOrder`. With `-diff-file`, Topolink narrows the report to the entries reachable from that change:

```bash
go run github.com/topolink/topolink/cmd/topolink@v1 -repo . \
  -diff-file testdata/change.patch -format markdown
```

Expected: the refund entry (`http:POST:/api/v1/orders/:id/refund`) and its transaction/remote-call findings, not the whole repository.

## CI

`.github/workflows/demo.yml` runs:

1. `go build ./...` and `go test ./...`
2. `scripts/verify.sh` — asserts that all 11 rules are present in the Topolink report
3. the Topolink GitHub Action in report-only mode (`fail-on: none`)

## A Java sibling

The same scenario will be mirrored as `topolink-example-java` (Spring Boot + MyBatis + Redis + Kafka) once the Java grammar lands.

## License

Apache-2.0
