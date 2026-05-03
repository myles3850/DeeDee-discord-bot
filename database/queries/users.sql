-- name: GetUser :one
SELECT id, discord_id, discord_user
FROM users
WHERE discord_id = $1;

-- name: SaveUser :one
INSERT INTO users (discord_id, discord_user)
VALUES ($1, $2)
ON CONFLICT (discord_id) DO UPDATE
    SET discord_user = EXCLUDED.discord_user
RETURNING id;
