-- name: GetMessageWithAuthor :one
SELECT
    m.id,
    m.discord_message_id,
    m.channel_id,
    m.content,
    m.created_at,
    m.edit_history,
    u.discord_id,
    u.discord_user
FROM messages m
JOIN users u ON m.author_id = u.id
WHERE m.discord_message_id = $1;

-- name: GetMessageForEdit :one
SELECT id, edit_history, content, discord_message_id
FROM messages
WHERE discord_message_id = $1;

-- name: SaveMessage :one
INSERT INTO messages (discord_message_id, channel_id, author_id, content, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (discord_message_id) DO UPDATE
    SET channel_id  = EXCLUDED.channel_id,
        author_id   = EXCLUDED.author_id,
        content     = EXCLUDED.content,
        created_at  = EXCLUDED.created_at
RETURNING id;

-- name: UpdateMessageContent :exec
UPDATE messages
SET content      = $1,
    edit_history = $2
WHERE id = $3;
