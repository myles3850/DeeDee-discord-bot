# DeeDee

A Discord bot for community management and engagement. She helps to manage messages, giving fun facts, and boost engagement

## What she does

- Watches every message that gets sent, edited, or deleted — nothing slips past her
- Keeps a full edit trail so you can always see what a message used to say
- Calls out deleted messages in the mod channel so nothing quietly disappears
- Welcomes new members with a wave when they post their intro
- Spins the wheel, shakes the 8-ball, and throws around shake emotes when asked
- Lets admins manage emoji role restrictions without touching the Discord UI

## Prerequisites

Works the same on macOS, Linux, and Windows — every project command runs through `just`, which is configured to use the right shell on each platform (see `justfile`'s `windows-shell` setting).

| Tool | Purpose | Install |
|---|---|---|
| Go 1.26+ | Application runtime | [go.dev/dl](https://go.dev/dl/) |
| Docker | Runs the Postgres container | [docs.docker.com](https://docs.docker.com/get-docker/) — on Windows and macOS this means Docker Desktop (WSL2 backend on Windows) |
| just | Command runner (like `make`, but simpler — all project commands live in `justfile`) | macOS/Linux: `brew install just`. Windows: `winget install --id Casey.Just` or `scoop install just`. Anything else: [just.systems](https://just.systems/man/en/packages.html) |
| Air | Hot-reloads the app on file changes during development | No separate install — managed as a Go tool dependency (`go tool air`) |

On Windows, run these from PowerShell or Windows Terminal (not `cmd.exe`).

## Quick start

1. **Copy the example env file and fill in your credentials** (see [Environment variables](#environment-variables) below):
   ```sh
   cp .env.example .env
   ```

2. **Start the database container:**
   ```sh
   just db
   ```

3. **Start the app with hot reload:**
   ```sh
   just dev
   ```

The app will run migrations automatically on startup, then connect to Discord and begin listening.

## Environment variables

Copy `.env.example` to `.env` and fill in the values below.

| Variable | Required | Description |
|---|---|---|
| `DISCORD_BOT_TOKEN` | yes | Bot token from the [Discord Developer Portal](https://discord.com/developers/applications) — select your application → Bot → Reset Token |
| `DISCORD_GUILD_ID` | yes | The ID of your Discord server. Enable Developer Mode in Discord (Settings → Advanced), then right-click your server and select *Copy Server ID* |
| `DATABASE_HOST` | yes | Postgres host — use `localhost` when running via Docker locally |
| `DATABASE_PORT` | yes | Postgres port — `5434` when using the provided `compose.yaml` |
| `DATABASE_USER` | yes | Postgres user |
| `DATABASE_PASS` | yes | Postgres password |
| `DATABASE_DB` | yes | Postgres database name |
| `GOOGLE_CREDENTIALS_JSON` | yes | Service account credentials JSON — see [Google's guide to creating a service account](https://cloud.google.com/iam/docs/service-accounts-create). Paste the full JSON content as the value (not a file path) |
| `GOOGLE_SPREADSHEET_ID` | yes | The ID from the spreadsheet URL: `https://docs.google.com/spreadsheets/d/<ID>/edit`. The service account must have at least Viewer access to the sheet |
| `MIGRATE` | no | Migration action on startup: `up` (default when unset), `down`, `reset`, `status` |
| `MIGRATIONS_DIR` | no | Override the migrations path — set automatically inside the Docker container, leave unset locally |

## Common commands

All commands are run through `just`, which reads the `justfile` in the project root.

```sh
just dev                       # start hot-reload dev server (requires .env and db to be running)
just db                        # start the Postgres Docker container
just down                      # stop and remove all containers

just migrate-create <name>     # scaffold a new migration file
just migrate                   # apply pending migrations
just migrate down              # roll back one migration
just migrate status            # show migration state

just generate                  # regenerate sqlc query code after editing database/queries/
```

## Architecture

```
main.go
  ├── googleplatform   Google Sheets client (initialised first)
  ├── database         Postgres connection + migrations
  ├── discord          Discord bot (gateway events + slash commands)
  └── webapi           REST API (emoji / role management + health check)
```

`main.go` wires everything together: it passes the database and Google Sheets instances into the Discord bot, registers event handlers, then starts the HTTP server. The web API endpoints call back into the Discord instance for live guild data.

## Modules

| Module | README | Purpose |
|---|---|---|
| `database/` | [database/README.md](database/README.md) | Postgres layer — schema migrations (goose) and query generation (sqlc) |
| `discord/` | [discord/README.md](discord/README.md) | Gateway event handlers and slash commands |
| `webapi/` | [webapi/README.md](webapi/README.md) | REST API — emoji, role, and health endpoints |
| `googleplatform/` | [googleplatform/README.md](googleplatform/README.md) | Google Sheets client |
| `cmd/healthcheck/` | — | Standalone binary used by the Docker `HEALTHCHECK` (see [Docker](#docker) below) |

## Stack

- **Go** — application runtime
- **[discordgo](https://github.com/bwmarrin/discordgo)** — Discord gateway and REST client
- **PostgreSQL** — persistence
- **[goose](https://github.com/pressly/goose)** — schema migrations (run via `go tool goose`, no separate install)
- **[sqlc](https://sqlc.dev)** — generates type-safe Go query code from raw SQL (run via `go tool sqlc`, no separate install)
- **`net/http`** — the REST API is plain stdlib, no framework needed for four routes

## Docker

The `compose.yaml` defines two services: the bot application (mapped to port 3863) and Postgres 17 (mapped to port 5434). The database has a health check so the app waits until Postgres is ready before starting.

The bot image is built `FROM scratch`, so it has no shell, `curl`, or `wget` for a conventional health check command. Instead, the Dockerfile builds a second tiny binary from `cmd/healthcheck/` that hits the bot's own `GET /health` endpoint, and `HEALTHCHECK` runs that binary directly — this is what a Docker dashboard like [Arcane](https://getarcane.app/) uses to show the container's live status.

```sh
docker compose up      # start both services
docker compose down    # stop and remove containers
```
