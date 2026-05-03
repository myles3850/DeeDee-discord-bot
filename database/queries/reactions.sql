-- name: SaveReaction :exec
INSERT INTO reactions (message_id, emoji, reactor_id)
VALUES ($1, $2, $3)
ON CONFLICT (message_id, emoji, reactor_id) DO UPDATE
    SET emoji      = EXCLUDED.emoji,
        reactor_id = EXCLUDED.reactor_id;

-- name: GetReactionsByMessage :many
SELECT
    r.id,
    r.emoji,
    u.discord_user
FROM reactions r
JOIN users u ON r.reactor_id = u.id
WHERE r.message_id = $1;
