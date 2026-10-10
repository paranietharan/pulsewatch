# PulseWatch Database Migrations

## Commands

```bash
# Apply pending migrations
go run main.go up

# Rollback last migration
go run main.go down

# Rollback multiple migrations
go run main.go down <steps>

# Check migration status
go run main.go status

# Force version (fix dirty state)
go run main.go force <version>
```
