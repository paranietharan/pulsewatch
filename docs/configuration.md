# Configuration

Copy `.env.example` to `.env` for Docker Compose and adjust the values for your environment. Compose sets container-specific connection addresses; host processes use the connection URLs configured for the host. See the service READMEs and `.env.example` for additional options.

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
