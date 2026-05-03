-- +goose Up
CREATE TABLE IF NOT EXISTS channel_name (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	discord_id VARCHAR(255) NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE IF EXISTS channel_name;
