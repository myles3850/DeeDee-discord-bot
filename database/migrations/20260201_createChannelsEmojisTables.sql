-- figure out how to save handler function return
CREATE TABLE IF NOT EXISTS "channels_emojis" (
    discord_channel_id varchar(255) NOT NULL,
    discord_emoji_id varchar(255) NOT NULL,
    PRIMARY KEY (discord_channel_id, discord_emoji_id)
    FOREIGN KEY (discord_channel_id) REFERENCES channel_name(discord_id)
);

CREATE TABLE IF NOT EXISTS "commands" {
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
}

CREATE TABLE IF NOT EXISTS "channel_command_handlers" {
    id SERIAL PRIMARY KEY,
    command_id INT NOT NULL,
    channel_id INT NOT NULL,
    handler_type VARCHAR(255 NOT NULL),
    handler_func TEXT NOT NULL,
    FOREIGN KEY (command_id) REFERENCES commands(id)
    FOREIGN KEY (channel_id) REFERENCES channel_name(id)
}

CREATE UNIQUE INDEX IF NOT EXISTS idx_discord_channel_id_discord_emoji_id_unique ON channels_emojis(discord_channel_id, discord_emoji_id);