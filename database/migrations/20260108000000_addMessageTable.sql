-- +goose Up
CREATE TABLE IF NOT EXISTS messages (
	id SERIAL PRIMARY KEY,
	discord_message_id VARCHAR(255) NOT NULL UNIQUE,
	channel_id VARCHAR(255) NOT NULL,
	author_id INT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (author_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS reactions (
	id SERIAL PRIMARY KEY,
	message_id INT NOT NULL,
	emoji VARCHAR(255) NOT NULL,
	reactor_id INT NOT NULL,
	FOREIGN KEY (message_id) REFERENCES messages(id),
	FOREIGN KEY (reactor_id) REFERENCES users(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_unique ON reactions(message_id, emoji, reactor_id);

-- +goose Down
DROP INDEX IF EXISTS idx_reactions_unique;
DROP TABLE IF EXISTS reactions;
DROP TABLE IF EXISTS messages;
