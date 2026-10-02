package database

import (
	"database/sql"
	"os"
	"strings"

	"github.com/pressly/goose/v3"
)

func migrationsDir() string {
	if d := os.Getenv("MIGRATIONS_DIR"); d != "" {
		return d
	}
	return "./database/migrations"
}

func HandleMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	dir := migrationsDir()
	switch strings.ToLower(os.Getenv("MIGRATE")) {
	case "up", "true":
		return goose.Up(db, dir)
	case "down":
		return goose.Down(db, dir)
	case "reset":
		return goose.DownTo(db, dir, 0)
	case "status":
		return goose.Status(db, dir)
	default:
		return nil
	}
}
