-- +goose Up
ALTER TABLE messages ADD COLUMN IF NOT EXISTS edit_history JSONB;

-- +goose Down
ALTER TABLE messages DROP COLUMN IF EXISTS edit_history;
