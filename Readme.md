# PulseWatch

PulseWatch is a full-stack uptime monitoring tool built with **Go**, **Next.js (TypeScript)**, **PostgreSQL**, **RabbitMQ**, and **Mailpit**. It runs periodic HTTP/HTTPS checks against registered endpoints, tracks DNS/TCP/TLS latency breakdowns, manages incident state transitions (`UP`, `DOWN`, `DEGRADED`), and sends alerts over email and signed webhooks.

AWS EKS Terraform and Helm deployment configs live in the companion repository: [paranietharan/pulsewatch-infra](https://github.com/paranietharan/pulsewatch-infra).

---

## Architecture

![PulseWatch Architecture](docs/overall%20diagram.png)

### Services

| Service | Directory | Description |
| :--- | :--- | :--- |
| **REST API & Result Processor** | `pulsewatch-api` | Go (`chi/v5`) HTTP server handling auth, monitor CRUD, dashboard queries, a PostgreSQL transactional outbox publisher (with RabbitMQ publisher confirms), and the consumer that processes probe results and updates incident states. |
| **Scheduler** | `pulsewatch-scheduler` | Polls PostgreSQL for due monitors in batches using `FOR UPDATE SKIP LOCKED` so multiple scheduler instances can claim checks without scheduling duplicate jobs. |
| **Probe Worker** | `pulsewatch-probe-worker` | Consumes check jobs from RabbitMQ across a concurrent worker pool, measures DNS lookup, TCP connect, and TLS handshake latency using Go's `httptrace`, and blocks private/loopback IPs via a custom `SafeDialContext` to guard against SSRF requests. |
| **Notification Service** | `pulsewatch-notification-service` | Stateless RabbitMQ consumer that sends HTML emails over SMTP and delivers `HMAC-SHA256` signed webhook payloads with 3-attempt linear backoff (`attempt * 500ms`). |
| **Web Dashboard** | `pulse-watch-web` | Next.js 16 / React 19 / TypeScript frontend for viewing live uptime percentages, response-time charts, incident logs, and configuring monitors and alert channels. |
| **Migration CLI** | `pulse-watch-db` | SQL migration runner for applying or rolling back PostgreSQL schema changes. |

---

## Quick Start

### 1. Run Backend & Infrastructure in Docker

Copy `.env.example` to `.env` (or use the defaults in `docker-compose.yml`) and start the stack:

```bash
cp .env.example .env
docker compose up --build
```

Once running, the following endpoints are available locally:
- **REST API**: `http://localhost:8080` (`GET /health`, routes under `/api/v1`)
- **RabbitMQ Management UI**: `http://localhost:15672` (user: `pulsewatch`, password: `pulsewatch_secret`)
- **Mailpit Web Inbox**: `http://localhost:8025` (captures verification and incident emails locally)
- **PostgreSQL**: `localhost:5432`

### 2. Run the Next.js Web Dashboard

In a separate terminal:

```bash
cd pulse-watch-web
npm install
npm run dev
```

Open `http://localhost:3000` in your browser.

---

## Running Services Locally on the Host

If you prefer running the Go services directly on your machine while keeping PostgreSQL, RabbitMQ, and Mailpit in Docker:

```bash
# 1. Start infrastructure containers
docker compose up -d postgres rabbitmq mailpit

# 2. Apply database migrations
go run pulse-watch-db/main.go up

# 3. Start each Go service in its own terminal
go run pulsewatch-api/cmd/api/main.go
go run pulsewatch-scheduler/cmd/scheduler/main.go
go run pulsewatch-probe-worker/cmd/worker/main.go
go run pulsewatch-notification-service/main.go
```

Other migration commands:
```bash
go run pulse-watch-db/main.go status
go run pulse-watch-db/main.go down
```

---

## REST API Endpoints (`pulsewatch-api`)

### Health & Authentication
- `GET /health` — Service health check
- `POST /api/v1/auth/signup` — Create account and queue verification email via outbox
- `POST /api/v1/auth/verify-email` — Verify account email token
- `POST /api/v1/auth/resend-verification` — Resend verification email
- `POST /api/v1/auth/login` — Authenticate and receive access + refresh tokens
- `POST /api/v1/auth/refresh` — Rotate refresh token and issue new access token
- `POST /api/v1/auth/logout` — Revoke current session
- `POST /api/v1/auth/forgot-password` — Request password reset link
- `POST /api/v1/auth/reset-password` — Reset password using token
- `GET /api/v1/auth/me` — Get current authenticated user profile

### Monitors, Incidents & Alert Channels (Bearer Auth)
- `GET /api/v1/dashboard/summary` — Aggregate uptime stats and recent incidents
- `POST /api/v1/monitors` — Create an HTTP/HTTPS monitor
- `GET /api/v1/monitors` — List monitors for the current user
- `GET /api/v1/monitors/{monitorID}` — Get monitor configuration and current status
- `PATCH /api/v1/monitors/{monitorID}` — Update monitor interval, timeout, or thresholds
- `DELETE /api/v1/monitors/{monitorID}` — Delete a monitor
- `POST /api/v1/monitors/{monitorID}/enable` — Resume scheduled checks
- `POST /api/v1/monitors/{monitorID}/disable` — Pause scheduled checks
- `GET /api/v1/monitors/{monitorID}/results` — List recent probe check results and latency breakdowns
- `GET /api/v1/monitors/{monitorID}/incidents` — List incident history for a monitor
- `GET /api/v1/alert-channels` — List configured email and webhook alert channels
- `POST /api/v1/alert-channels` — Add an email or HMAC-signed webhook alert channel
- `DELETE /api/v1/alert-channels/{channelID}` — Remove an alert channel

---

## Environment Variables

### PostgreSQL
| Variable | Example Value | Description |
| :--- | :--- | :--- |
| `DATABASE_URL` | `postgres://pulsewatch:pulsewatch_secret@localhost:5432/pulsewatch?sslmode=disable` | PostgreSQL connection URI used by the API, scheduler, and migration runner |
| `POSTGRES_USER` | `pulsewatch` | PostgreSQL user |
| `POSTGRES_PASSWORD` | `pulsewatch_secret` | PostgreSQL password |
| `POSTGRES_DB` | `pulsewatch` | Database name |
| `POSTGRES_PORT` | `5432` | Exposed host port |

### RabbitMQ
| Variable | Example Value | Description |
| :--- | :--- | :--- |
| `RABBITMQ_URL` | `amqp://pulsewatch:pulsewatch_secret@localhost:5672/` | AMQP connection URI used by publishers and consumers |
| `RABBITMQ_DEFAULT_USER` | `pulsewatch` | RabbitMQ username |
| `RABBITMQ_DEFAULT_PASS` | `pulsewatch_secret` | RabbitMQ password |
| `RABBITMQ_PORT` | `5672` | AMQP port |
| `RABBITMQ_MANAGEMENT_PORT` | `15672` | Management UI port |

### API Service (`pulsewatch-api`)
| Variable | Example Value | Description |
| :--- | :--- | :--- |
| `API_HOST` | `0.0.0.0` | HTTP bind address |
| `API_PORT` | `8080` | HTTP server port |
| `JWT_SECRET` | `replace-with-32-char-secret` | Secret key for signing JWT access tokens |
| `FRONTEND_URL` | `http://localhost:3000` | Allowed CORS origin and base URL for email links |
| `CLIENT_IP_STRATEGY` | `remote_addr` | Client IP extraction strategy (`remote_addr`, `cloudflare`, `header`, `xff`, `proxy_hops`) |

### Scheduler, Probe Worker & Notifications
| Variable | Example Value | Description |
| :--- | :--- | :--- |
| `SCHEDULER_POLL_INTERVAL_SEC` | `5` | How often the scheduler polls PostgreSQL for due checks |
| `SCHEDULER_BATCH_SIZE` | `50` | Maximum monitors claimed per batch sweep |
| `PROBE_WORKER_CONCURRENCY` | `20` | Concurrent probe goroutines per worker instance |
| `PROBE_WORKER_PREFETCH` | `20` | RabbitMQ QoS prefetch count |
| `ALLOW_INTERNAL_PROBES` | `false` | Set to `false` in production to block private/loopback IPs (SSRF protection) |
| `SMTP_HOST` | `localhost` | SMTP server host (`mailpit` inside Docker) |
| `SMTP_PORT` | `1025` | SMTP server port (`1025` for Mailpit, `587` for TLS SMTP) |
| `SMTP_FROM` | `PulseWatch <alerts@pulsewatch.dev>` | Sender address for outgoing alert emails |
