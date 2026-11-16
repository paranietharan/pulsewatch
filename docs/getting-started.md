# Getting started

## Run with Docker Compose

Install Docker and Docker Compose, then from the repository root:

```bash
cp .env.example .env
docker compose up --build
```

Compose starts PostgreSQL, RabbitMQ, Mailpit, the migration runner, and the Go services. The migrator applies migrations before the API and scheduler start.

- API: http://localhost:8080/health
- RabbitMQ management: http://localhost:15672 (default credentials: `pulsewatch` / `pulsewatch_secret`)
- Mailpit inbox: http://localhost:8025
- PostgreSQL: `localhost:5432`

To stop the containers, run `docker compose down`.

## Web dashboard

The dashboard lives in [pulse-watch-web](https://github.com/paranietharan/pulse-watch-web). Follow its README to configure and run it; by default it connects to the API on port 8080.

## Running services on the host

Export the required variables from `.env` in each service's environment; Go processes do not automatically load the Compose environment file. Use host-accessible addresses such as `localhost` for the database and broker.

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

See [configuration](configuration.md) for environment settings and [the migration guide](../pulse-watch-db/README.md) for database commands. The Mailpit integration test accepts `MAILPIT_API_URL` (default `http://localhost:8025`).
