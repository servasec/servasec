# Migrations

Each migration is a versioned SQL file applied by [Goose](https://github.com/pressly/goose)
in sequential order when the backend starts.

## Mapping

| Migration | App version | Description |
|-----------|-------------|-------------|
| 001       | v1.0.0      | Initial schema, all tables and indexes |
| 002       | n/a           | *(placeholder)* |

## Adding a new migration

```bash
make migrate-create NAME=describe_change
```

Edit the generated file, then rebuild and restart:

```bash
make dev       # development
make community # production
```

The migration runs automatically on startup.
