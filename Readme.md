# PulseWatch

PulseWatch is a distributed, high-performance website and API uptime monitoring platform built with **Go**, **PostgreSQL**, **RabbitMQ**, and **Mailpit**.

---

## Configuration

PulseWatch services are configured via environment variables. You can copy `.env.example` to `.env` in the project root to configure all services locally.

### 1. PostgreSQL Database Configuration

| Config name | Example value | description |
|---|---|---|
| `DATABASE_URL` | `postgres://pulsewatch:pulsewatch_secret@localhost:5432/pulsewatch?sslmode=disable` | Canonical PostgreSQL connection string used by Go services and migration runners |
| `POSTGRES_USER` | `pulsewatch` | PostgreSQL username for container and local authentication |
| `POSTGRES_PASSWORD` | `pulsewatch_secret` | PostgreSQL user password |
| `POSTGRES_DB` | `pulsewatch` | Primary database name |
| `POSTGRES_HOST` | `localhost` | Database host hostname or IP |
| `POSTGRES_PORT` | `5432` | PostgreSQL listener port |
| `POSTGRES_SSLMODE` | `disable` | SSL connection mode (`disable`, `require`, `verify-full`) |

---

### 2. RabbitMQ Message Broker Configuration

| Config name | Example value | description |
|---|---|---|
| `RABBITMQ_URL` | `amqp://pulsewatch:pulsewatch_secret@localhost:5672/` | Full AMQP connection URI used by all publishers and consumers |
| `RABBITMQ_HOST` | `localhost` | RabbitMQ server hostname or IP |
| `RABBITMQ_PORT` | `5672` | AMQP protocol client port |
| `RABBITMQ_DEFAULT_USER` | `pulsewatch` | RabbitMQ administrative user credentials |
| `RABBITMQ_DEFAULT_PASS` | `pulsewatch_secret` | RabbitMQ administrative user password |
| `RABBITMQ_MANAGEMENT_PORT`| `15672` | Web management dashboard and REST HTTP API port |

---

### 3. Redis Cache Configuration

| Config name | Example value | description |
|---|---|---|
| `REDIS_URL` | `redis://localhost:6379/0` | Full Redis connection URI with database index |
| `REDIS_HOST` | `localhost` | Redis server hostname or IP |
| `REDIS_PORT` | `6379` | Redis listener port |
| `REDIS_PASSWORD` | *(empty in dev)* | Authentication password for Redis (leave blank if none) |

---

### 4. API Service Configuration (`pulsewatch-api`)

| Config name | Example value | description |
|---|---|---|
| `API_PORT` | `8080` | Port for the HTTP REST API server |
| `API_HOST` | `0.0.0.0` | Bind address for the API server (`0.0.0.0` for all interfaces) |
| `JWT_SECRET` | `9h66d7LVUnjNeiMmtIATk3vaiwYIH5iHurEaa7SYYFe` | Cryptographic secret key used to sign and verify JWT authentication tokens (minimum 16 characters) |
| `ENVIRONMENT` | `development` | Runtime environment (`development`, `staging`, `production`) |
| `LOG_LEVEL` | `debug` | Minimum structured logging level (`debug`, `info`, `warn`, `error`) |
| `FRONTEND_URL` | `http://localhost:3000` | Frontend dashboard origin allowed for CORS requests and email action links |

---

### 5. Client IP Resolution & Security

| Config name | Example value | description |
|---|---|---|
| `CLIENT_IP_STRATEGY` | `remote_addr` | IP resolution method for rate limiting: `remote_addr` (direct), `cloudflare` (CDN), `header` (custom header), `xff` (trusted CIDRs), `proxy_hops` (hop count) |
| `CLIENT_IP_HEADER` | `CF-Connecting-IP` | Custom request header to inspect when strategy is `header` or `cloudflare` |
| `CLIENT_IP_TRUSTED_CIDRS` | `10.0.0.0/8,172.16.0.0/12` | Comma-separated list of trusted proxy CIDRs used when strategy is `xff` |
| `CLIENT_IP_TRUSTED_PROXIES`| `1` | Number of trusted reverse proxy hops between client and API when strategy is `proxy_hops` |

---

### 6. Notification Service Configuration (`pulsewatch-notification-service`)

| Config name | Example value | description |
|---|---|---|
| `SMTP_HOST` | `localhost` | Outgoing SMTP mail server hostname (Mailpit in development, Amazon SES/SendGrid in production) |
| `SMTP_PORT` | `1025` | SMTP listener port (`1025` for Mailpit, `587` for standard TLS SMTP) |
| `SMTP_USER` | *(empty in dev)* | SMTP authentication username |
| `SMTP_PASSWORD` | *(empty in dev)* | SMTP authentication password |
| `SMTP_FROM` | `PulseWatch <alerts@pulsewatch.dev>` | From address and display name attached to all outgoing emails |
| `MAILPIT_PORT` | `8025` | Mailpit web UI port for viewing captured emails locally |
| `MAILPIT_WEB_URL` | `http://localhost:8025` | URL to Mailpit browser interface |

---

### 7. Probe Worker Configuration (`pulsewatch-probe-worker`)

| Config name | Example value | description |
|---|---|---|
| `PROBE_WORKER_CONCURRENCY` | `20` | Size of the concurrent worker pool executing HTTP and TCP probe jobs |
| `PROBE_WORKER_PREFETCH` | `20` | RabbitMQ channel QoS prefetch count per worker instance |
| `PROBE_REGION` | `default` | Geographical region identifier tagged on probe results (e.g., `us-east-1`, `eu-west-1`) |
| `ALLOW_INTERNAL_PROBES` | `false` | SSRF protection toggle; set to `true` to allow probing RFC1918 / localhost IP addresses |

---

### 8. Scheduler Service Configuration (`pulsewatch-scheduler`)

| Config name | Example value | description |
|---|---|---|
| `SCHEDULER_POLL_INTERVAL_SEC` | `5` | Polling frequency in seconds for querying due monitors from the database |
| `SCHEDULER_BATCH_SIZE` | `50` | Maximum number of due monitors claimed and dispatched in a single batch sweep |

---

## Quick Start

You can run PulseWatch either entirely in Docker or in a hybrid local development setup.

### Run Entire Platform in Docker

Build and start all services, message broker, database, and automatic migration runner:

```bash
docker compose up --build
```

- **REST API**: http://localhost:8080 (Health check: `/health`, API docs: `/api/v1`)
- **RabbitMQ Management**: http://localhost:15672 (user: `pulsewatch`, pass: `pulsewatch_secret`)
- **Mailpit Web Inbox**: http://localhost:8025
- **PostgreSQL Database**: Exposed on host port `5432`

---

### Run Services on Host

#### Start Infrastructure Dependencies
```bash
docker compose up -d postgres rabbitmq mailpit
```

#### Run Database Migrations from Host
Because PostgreSQL is exposed on port `5432`, you can execute migrations directly from your terminal:
```bash
# Check current migration status
go run pulse-watch-db/main.go status

# Apply all pending migrations
go run pulse-watch-db/main.go up

# Rollback last migration
go run pulse-watch-db/main.go down
```

*Alternatively, run migrations inside Docker:*
```bash
docker compose run --rm migrator up
```

#### 3. Run Microservices
Run each service in separate terminal sessions:

```bash
# REST API (Port 8080)
go run pulsewatch-api/cmd/api/main.go

# Scheduler (Cron probe sweep)
go run pulsewatch-scheduler/cmd/scheduler/main.go

# Probe Worker (Job executor)
go run pulsewatch-probe-worker/cmd/worker/main.go

# Notification Service (Email & Webhook dispatcher)
go run pulsewatch-notification-service/main.go
```