set dotenv-load

# start hot-reload dev server (requires db to be running)
dev:
    air

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
    MIGRATE={{status}} go run .

# generate sqlc query code from database/queries/
generate:
    go tool sqlc generate
