-- +goose Up
ALTER TABLE messages ADD COLUMN IF NOT EXISTS content TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE messages DROP COLUMN IF EXISTS content;
