set dotenv-load
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# start hot-reload dev server (requires db to be running)
dev:
    go tool air

# start just the database container
db:
    docker compose up db -d

# stop and remove all containers
down:
    docker compose down

# scaffold a new migration file: just migrate-create <name>
migrate-create name:
    go tool goose -dir ./database/migrations create {{name}} sql

# run migrations: just migrate [up|down|reset|status]
migrate status="up":
    go tool goose -dir ./database/migrations postgres "host={{env('DATABASE_HOST')}} port={{env('DATABASE_PORT')}} user={{env('DATABASE_USER')}} password={{env('DATABASE_PASS')}} dbname={{env('DATABASE_DB')}} sslmode=disable" {{status}}

# generate sqlc query code from database/queries/
generate:
    go tool sqlc generate
