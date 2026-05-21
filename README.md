# Panzucha – Production‑Ready Go API

A modular, cloud‑native REST API that processes 1,000+ requests per second with structured logging, JWT authentication, and clean separation of concerns. Built to solve: “How do we build a maintainable, observable, and scalable API that can be deployed to Kubernetes without vendor lock‑in?”

## The Problem (Business Context)

A fast‑growing e‑commerce platform struggles with an unreliable monolithic API. Their legacy system suffers from:

| Issue | Business impact |
|-------|----------------|
| ❌ **Spaghetti code** – Business logic, HTTP handlers, and SQL queries all mixed together | Slow feature delivery & high onboarding cost |
| ❌ **No structured logging** – Debugging requires manually grepping messy log files | 2‑hour mean time to recovery (MTTR) |
| ❌ **No request tracing** – Impossible to track a single request across service boundaries | No SLA visibility |
| ❌ **Dependency on Logstash** – API crashes when the log aggregator is down | Cascading failures |
| ❌ **Exposed internal models** – `User` object returned from API includes `password_hash` | Security risk & API versioning headache |
| ❌ **No graceful shutdown** – Rolling update kills in‑flight requests | Customer‑facing errors during deployment |
| **Business goal**: Build an API that handles 1,000+ RPS, survives its dependencies failing, and gives engineers full observability – without rewriting for every new database or log collector.

## Case Study: Rebuilding the UrbanCart E‑commerce API

*The following is a simulated case study based on real‑world challenges faced by a rapidly growing online marketplace.*

### 🏢 The Company: UrbanCart

A rapidly scaling online marketplace handling **500,000+ daily active users**. Their existing monolithic .NET API was becoming a bottleneck, slowing down feature development and causing outages during peak sales events.

### ⚠️ The "Black Friday Incident"

During a major sales event, the legacy API collapsed under **1,200 RPS** (requests per second). It took the engineering team over 4 hours to recover, resulting in an estimated **$500,000 in lost revenue**. The post‑mortem revealed three core failures:

1. **Cascading Failures** – The API crashed when the centralized Logstash cluster fell behind. The application’s logging calls would block and eventually cause a deadlock.
2. **Concurrency Disasters** – The inventory system had no locking mechanism, leading to overselling the same “Limited Edition” product to 50 customers when only 5 units were in stock.
3. **The "Black Box" Problem** – With no distributed tracing, the team could not determine why the database connection pool was being exhausted, leading to blind debugging.

### 💡 Our Solution: The Panzucha API

We partnered with UrbanCart to rebuild their core API from scratch in Go, focusing on **reliability**, **observability**, and **high performance**.

| Former Pain Point | Solution Implemented in Panzucha | Business Outcome |
| :--- | :--- | :--- |
| **Logstash Dependency** | Implemented **async, non‑blocking logging** (queue + worker). Logs write to `stdout` first; a background goroutine ships them. If the queue fills, logs are dropped (fail‑open). | API remained stable even when Logstash was offline. No more cascading failures. |
| **Inventory Overselling** | Added a `DecrementStock` method with **optimistic locking** using a `version` column in PostgreSQL. All stock updates run in a transaction. | Concurrent orders for the last product result in only one success; others receive a `409 Conflict`. **Zero overselling**. |
| **Blind Debugging** | Integrated **OpenTelemetry tracing** with Jaeger to visualise request flow across handlers, services, and DB. Added **structured JSON logs** with `request_id` to correlate logs with traces. | Mean Time To Resolution (MTTR) dropped from 2 hours to 15 minutes. Root causes identified instantly. |
| **Feature Delivery** | Adopted **Clean Architecture** – business logic separated from HTTP and DB concerns. Swapping external dependencies requires zero changes to core business rules. | New features ship in **days instead of weeks**. Accelerated time‑to‑market. |

### 🚀 The Results (Pre‑Launch Load Test)

| Metric | Legacy API | Panzucha API | Improvement |
| :--- | :--- | :--- | :--- |
| **Throughput (RPS)** | 1,150 (max) | 4,200+ (stable) | **265% Increase** |
| **P99 Latency** | 850ms | 110ms | **87% Improvement** |
| **Recovery Time (outage)** | 4 hours | N/A (graceful degradation) | **100% Availability** |

**By implementing these changes, UrbanCart’s API can now handle an unexpected viral campaign or flash sale without collapsing. The result? Higher revenue and a dependable customer experience.**

## How We Solve It – Clean Architecture in Go

We built a modular API using **clean architecture** and **dependency injection**. Key design decisions:

- **Domain‑driven packages** – Business logic (`services`), data access (`repositories`), and HTTP handling (`handlers`) live in separate packages. No `models` or `utils` dumping ground.
- **Interface‑first design** – `ProductRepository` interface in the `domain` package; actual PostgreSQL implementation lives in `repositories/postgres`. Swap databases without changing a single line of business logic.
- **Structured JSON logging** – Logs go to `stdout` only. No direct Logstash calls. Each log entry contains `category`, `sub_category`, `duration_ms`, `performance_bucket` (Fast/Normal/Slow/Critical), `request_id`, `user_id`, and operation‑specific fields.
- **Graceful shutdown** – Captures `SIGINT`/`SIGTERM`, stops accepting new requests, waits for in‑flight requests (up to 30s), closes database pool, then exits.
- **Async non‑blocking logging** – Queue + worker goroutine prevents slow log sinks from blocking API responses. When the queue is full, older logs are dropped (fail‑open) to maintain API availability.
- **Context propagation** – `context.Context` passes request scopes (timeout, user ID, request ID) from handler → service → repository. All DB operations honour context cancellation.
- **No voodoo** – All code is explicit, easy to trace, and logs exactly what happens.

## Architecture

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#1e293b', 'primaryTextColor': '#ffffff', 'primaryBorderColor': '#475569', 'lineColor': '#64748b', 'background': '#0f172a'}}}%%
flowchart TB

  subgraph Client_Layer["Client Layer"]
    REST[REST Clients<br>curl, Postman, Web]
  end

  subgraph API_Gateway["API Gateway / Ingress (EKS)"]
    LB[Load Balancer]
    Ingress[Nginx Ingress Controller]
  end

  subgraph Panzucha_API["Panzucha API"]
    direction TB
    Middleware[Middleware Stack<br>• Request ID<br>• Recovery<br>• Timeout<br>• JWT Auth<br>• Structured Logging]
    Handlers[HTTP Handlers<br>• UserHandler<br>• ProductHandler<br>• OrderHandler]
    Services[Service Layer<br>• Business logic<br>• Orchestration]
    Domain[Domain Models<br>• Entities<br>• Repository Interfaces]
    Repo[Repository Layer<br>• PostgresUserRepo<br>• PostgresProductRepo]
    LoggerPkg[Logger Package<br>• Async queue<br>• stdout + ELK]
  end

  subgraph Storage_Observability["Storage & Observability"]
    PG[(PostgreSQL)]
    Prometheus[Prometheus]
    Grafana[Grafana]
    Jaeger[Jaeger]
    Splunk[Splunk / ELK]
  end

  %% Routing Connections
  REST --> LB
  LB --> Ingress
  Ingress --> Middleware

  %% API Internal Flow
  Middleware --> Handlers
  Handlers --> Services
  Services --> Domain
  Domain --> Repo
  Repo --> PG

  %% Telemetry and Logging Connections
  Services --> LoggerPkg
  Handlers --> LoggerPkg
  Repo --> LoggerPkg
  LoggerPkg -->|JSON logs| Splunk

  Handlers -->|/metrics| Prometheus
  Prometheus -->|visualise| Grafana
  Services -->|OpenTelemetry traces| Jaeger

  %% Class Styling for High Contrast
  classDef default fill:#2d3748,stroke:#4a5568,stroke-width:2px,color:#ffffff;
  classDef client fill:#312e81,stroke:#4338ca,stroke-width:2px,color:#ffffff;
  classDef gateway fill:#1e3a8a,stroke:#3b82f6,stroke-width:2px,color:#ffffff;
  classDef storage fill:#064e3b,stroke:#10b981,stroke-width:2px,color:#ffffff;
  classDef logs fill:#78350f,stroke:#f59e0b,stroke-width:2px,color:#ffffff;
  classDef metrics fill:#7f1d1d,stroke:#ef4444,stroke-width:2px,color:#ffffff;

  %% Apply Classes
  class REST client;
  class LB,Ingress,Middleware,Handlers,Services,Domain gateway;
  class Repo,PG storage;
  class LoggerPkg,Splunk logs;
  class Prometheus,Grafana,Jaeger metrics;
```

Data flow:

1. **Request flow** – Ingress receives REST request, applies rate limiting & JWT auth, calls handler. Handler extracts request metadata (request ID, user ID, client IP), validates DTO, maps to domain model, passes to service. Service applies business logic, uses repository interface to read/write PostgreSQL.
2. **Logging flow** – Handler, service, and repository call structured logging methods with category, sub_category, duration_ms, performance_bucket, request_id, user_id. Logs are written to stdout (immediate) and asynchronously queued for optional shipping to Logstash/Splunk. No dependency on external collector.
3. **Observability** – Prometheus scrapes /metrics for HTTP latency histograms, DB operation counters, goroutine count, and custom business metrics. Grafana visualises dashboards. OpenTelemetry traces propagate from handler → service → repository → PostgreSQL, exporting to Jaeger.
4. **Graceful shutdown** – On SIGTERM, API stops accepting new requests, finishes existing ones (max 30s), closes DB pool, then exits. K8s liveness/readiness probes prevent traffic during shutdown.

## ✨ Best Practices & Patterns Shown

| Pattern | Implementation |
|---------|----------------|
| **Clean architecture** | `domain` defines interfaces; `services` & `handlers` depend on them; infrastructure (`repositories`) plugs in. Swap PostgreSQL for MySQL without touching business logic. |
| **Structured logging** | JSON logs with `category`, `sub_category`, `duration_ms`, `performance_bucket`, `request_id`, `user_id`. Each layer (API, DB, business) logs its own slice. |
| **Graceful shutdown** | Listen for signals, call `http.Server.Shutdown()` with 30s timeout, close DB pool, flush logs. |
| **Context propagation** | `context.Context` passes request scopes (timeout, user ID, request ID, roles) from middleware → handler → service → repository. |
| **Async non‑blocking logging** | Queue + worker goroutine; logs are sent to external aggregator without blocking the response. On queue full, logs are dropped (fail‑open). |
| **Request‑scoped metadata** | `ExtractRequestInfo(r)` helper gathers `request_id`, `client_ip`, `user_agent`, `start_time`, `user_id`. Used by all handlers. |
| **Interface‑based mocking** | `ProductRepository` interface allows `mockProductRepository` in unit tests. No real database needed. |
| **Repository logging** | Every `SELECT`, `INSERT`, `UPDATE`, `DELETE` logs duration, rows affected, and error. |
| **Partial updates with DTOs** | `UpdateProductRequest` uses pointers (`*string`, `*float64`) to detect which fields were sent, skipping zero‑value ambiguity. |
| **Idempotency** (ready) | DB unique constraint on `idempotency_key` prevents duplicate order creation on retry. |

## Tech Stack

- **Language**: Go 1.23+ (with toolchain)
- **Router**: [chi](https://github.com/go-chi/chi) – lightweight, 100% compatible with `net/http`, excellent middleware support
- **Database**: PostgreSQL + `pgxpool` (connection pool)
- **ORM / SQL**: Pure `database/sql` + `pgx` – no ORM, full control
- **Validation**: `go-playground/validator`
- **Authentication**: `golang-jwt/jwt` (JWT with roles)
- **Logging**: Structured JSON to `stdout` + async queue for external shipping (Splunk/Logstash optional)
- **Metrics**: `prometheus/client_golang`
- **Tracing**: OpenTelemetry + Jaeger (optional)
- **Deployment**: Docker + Kubernetes (Helm chart included)
- **Testing**: `testing` package + `testcontainers` (integration tests) + `go test -race`
- **CI/CD**: GitHub Actions – test, race detection, build, push to GHCR

## Getting Started (local development)

### Prerequisites

- Go 1.23+
- Docker (PostgreSQL & optional Logstash)
- `make` (optional, but convenient)

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/panzucha
cd panzucha
```

### 2. Start PostgreSQL (with Docker Compose)

```bash
docker-compose up -d postgres
```

### 3. Run migrations
```bash
# Using golang-migrate
migrate -path migrations -database "postgres://postgres:password@localhost:5432/panzucha?sslmode=disable" up

```

### 4. Copy configuration
```bash
cp .env.example .env
# Edit .env with your database credentials
```

### 5. Run the API
```bash
go run ./cmd/api
```
The API will start on http://localhost:8080

### 6. Test with curl
```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","name":"Alice","password":"secret123"}'

# Login and get JWT token
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'

# Create a product (authenticated)
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","price":999.99}'

```
