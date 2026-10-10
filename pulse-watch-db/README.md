# PulseWatch Database Migration Service 

## Commands

All commands can be executed from the repository root or from within `pulse-watch-db`:

### 1. Apply All Pending Migrations (`up`)
Applies all pending migrations in ascending order inside transactions:
```bash
# From pulse-watch-db/
go run main.go up

# Or from repository root
go run pulse-watch-db/main.go up
```

### 2. Check Migration Status (`status`)
Displays the current active version number and dirty state:
```bash
go run main.go status
```

### 3. Rollback Migrations (`down`)
Reverts the most recently applied migration:
```bash
go run main.go down
```
To revert multiple migrations (e.g., 2 steps):
```bash
go run main.go down 2
```


## Environment Configuration Variables

The migration runner reads configuration from `.env` or system environment variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `DATABASE_URL` | Complete Postgres connection URI (takes highest priority) | `postgres://pulsewatch:pulsewatch_secret@localhost:5432/pulsewatch?sslmode=disable` |
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Database username | `pulsewatch` |
| `POSTGRES_PASSWORD` | Database password | `pulsewatch_secret` |
| `POSTGRES_DB` | Database name | `pulsewatch` |
| `POSTGRES_SSLMODE` | SSL connection mode | `disable` |
