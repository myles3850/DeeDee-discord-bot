# emojibot — backend

Discord bot backend. Tracks messages, reactions, and channel activity in Postgres.

## Stack

- **Go** — application runtime
- **discordgo** — Discord gateway and REST client
- **PostgreSQL** — persistence
- **goose** — schema migrations
- **sqlc** — type-safe query code generation

## Database strategy

Schema changes and query code are managed by two separate tools that work together:

- **goose** owns the schema. When you need to add a table or column, write a migration.
- **sqlc** owns the queries. When you need to read or write data, write SQL and run `just generate` to produce typed Go code.

See [database/README.md](database/README.md) for the full workflow.

## Quick start

```sh
just db        # start the Postgres container
just dev       # start the app with hot reload
```

## Common commands

```sh
just migrate-create <name>   # scaffold a new migration file
just migrate                 # run pending migrations (default: up)
just migrate down            # roll back one migration
just migrate status          # show migration state
just generate                # regenerate sqlc query code after editing database/queries/
```
