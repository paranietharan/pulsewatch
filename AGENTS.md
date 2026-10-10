# AGENTS.md — PulseWatch AI Agent Operating Guidelines

Welcome, Agent. This document establishes the development rules, operational standards, architectural conventions, and workflow procedures for working across the **PulseWatch** uptime monitoring platform codebase.

---

## 1. Project Mission & High-Level Architecture

**PulseWatch** is a distributed, backend-focused uptime monitoring platform built primarily in **Go**, utilizing **PostgreSQL** for persistent relational storage, **Redis** for distributed locking/caching, **RabbitMQ** for asynchronous job scheduling and result dispatch, and **Next.js** for the frontend analytical dashboard.

### Core Data & Event Flow
```
+-------------------------------------------------------------------------------+
|                                Next.js Dashboard                              |
+-------------------------------------------------------------------------------+
                                      ▲
                                      │ REST API (JSON)
                                      ▼
+-------------------------------------------------------------------------------+
|                            API Service (pulsewatch-api)                       |
+-------------------------------------------------------------------------------+
                                      │
                                      ▼
+---------------------+      +---------------------+      +---------------------+
| pulsewatch-scheduler|      |     PostgreSQL      |◄─────|pulsewatch-notif-svc |
+---------------------+      +---------------------+      +---------------------+
           │                                                         ▲
           │ publishes monitor.checks                                │ consumes monitor.results
           ▼                                                         │
   +---------------+                                         +---------------+
   |   RabbitMQ    |────────────────────────────────────────►|   RabbitMQ    |
   | monitor.checks|                                         |monitor.results|
   +---------------+                                         +---------------+
           │
           ▼ consumes check jobs
+-------------------------+
| pulsewatch-probe-worker |
+-------------------------+
```

---

## 2. Repository Structure

```
pulsewatch/
├── .env.example                       # Canonical template for all environment variables
├── .env                               # Local environment variables (git-ignored)
├── .gitignore                         # Project-wide git exclusion rules
├── docker-compose.yml                 # Local multi-container infrastructure (Postgres, RabbitMQ, etc.)
├── AGENTS.md                          # AI agent development rules & guidelines (this file)
├── context.md                         # Comprehensive platform context and architectural specifications
├── Readme.md                          # Project overview and entrypoint
├── docs/                              # Architectural diagrams and design specifications
│   └── overall diagram.png
├── pulse-watch-db/                    # PostgreSQL schema migrations and management tooling
│   ├── main.go                        # Go migration runner (up, down, status, force) using golang-migrate
│   ├── go.mod                         # Go module dependencies
│   ├── README.md                      # Migration CLI documentation
│   └── migrations/                    # Versioned up/down SQL migration files
├── pulsewatch-api/                    # Go REST API (Authentication, Monitor CRUD, History, Incidents)
│   ├── go.mod
│   └── Readme.md
├── pulsewatch-scheduler/              # Go service scheduling monitor probes into RabbitMQ
│   └── Readme.md
├── pulsewatch-probe-worker/           # Go concurrent worker probing HTTP/TCP endpoints
│   └── Readme.md
└── pulsewatch-notification-service/   # Go service processing results, incident state, and alerting
    └── Readme.md
```

---

## 3. Strict Development Rules for AI Agents

### 3.1 Environment & Secrets Management
1. **Never Hardcode Credentials**: Never write hardcoded database passwords, API tokens, JWT secrets, or connection strings into source code.
2. **Environment First**: All configurations must be sourced from environment variables (`.env`).
3. **Template Synchronization**: When adding any new configuration variable to any service, immediately update both `.env` and `.env.example`.

### 3.2 Database & Schema Migrations
1. **Single Source of Truth**: Schema definitions live exclusively in `pulse-watch-db/migrations/`.
2. **Deterministic Versioning**: All schema alterations must follow the pair convention:
   - `<version>_<name>.up.sql`
   - `<version>_<name>.down.sql`
3. **Always Include Rollbacks**: Every `.up.sql` must have a fully tested, functional `.down.sql` that cleanly reverses the operation.
4. **No In-Place Modifications**: Never edit an existing migration file that has already been executed in an environment; generate a new migration instead with the next sequential version number.
5. **Transactions & Idempotency**: Use explicit transactional patterns and `IF NOT EXISTS` / `IF EXISTS` where appropriate.

### 3.3 Go Backend Coding Standards
1. **Idiomatic Go**: Follow standard Go project structure and naming conventions (`camelCase` for unexported, `PascalCase` for exported).
2. **Context Propagation**: Always pass `context.Context` as the first argument in functions that perform I/O, database queries, HTTP calls, or RabbitMQ messaging.
3. **Graceful Shutdown**: All long-running Go processes (API server, scheduler ticker, probe workers, consumers) must listen for `SIGINT` / `SIGTERM` and shut down cleanly within a 5–10s timeout window.
4. **Error Handling**: Never discard errors with `_`. Wrap errors with contextual information using `fmt.Errorf("failed to execute probe: %w", err)`.
5. **Connection Pooling**: Use `database/sql` connection pooling settings (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`).

### 3.4 RabbitMQ Messaging Standards
1. **Explicit Routing Keys & Queues**:
   - `pulsewatch.direct` exchange
   - Queues: `monitor.checks`, `monitor.results`, `notifications.email`, `notifications.webhook`
2. **Structured JSON Payloads**: All messages must use strict JSON contracts with timestamps and correlation IDs.
3. **Consumer Idempotency**: Consumers must gracefully handle duplicate deliveries without creating duplicate incident alerts or corrupting metrics.
4. **Worker Concurrency**: Probe workers must use a configurable worker pool with bounded concurrency (`worker_concurrency` setting) to prevent resource exhaustion.

---

## 4. Key Agent Workflows

### 4.1 Running Database Migrations
```bash
# Check current status
go run pulse-watch-db/main.go status

# Apply pending migrations
go run pulse-watch-db/main.go up

# Rollback last migration
go run pulse-watch-db/main.go down
```

### 4.2 Validating Docker Compose
```bash
# Check compose configuration
docker compose config

# Start database service in background
docker compose up -d postgres
```

### 4.3 Creating a New Migration
```bash
# Create migration file pair
touch pulse-watch-db/migrations/007_add_probe_headers.up.sql
touch pulse-watch-db/migrations/007_add_probe_headers.down.sql

# Edit the generated .up.sql and .down.sql files in pulse-watch-db/migrations/
```

---

## 5. Security & Safety Checklist

- [ ] Parameterized SQL queries only (prevent SQL injection).
- [ ] Passwords stored using `bcrypt` or `argon2id`.
- [ ] SSRF protection in probe workers: prevent workers from querying private/internal IP ranges (`127.0.0.1`, `10.0.0.0/8`, `169.254.169.254`, etc.) unless explicitly permitted by an admin override flag.
- [ ] Safe JSON unmarshalling with size limits on request bodies.
- [ ] Rate limiting on public API endpoints.
