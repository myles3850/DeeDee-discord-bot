-- +goose Up
CREATE TABLE IF NOT EXISTS completed_channels (
	id SERIAL PRIMARY KEY,
	channel_id VARCHAR(255) NOT NULL UNIQUE,
	completed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS completed_channels;
