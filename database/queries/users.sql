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

-- name: SaveBirthdayForUser :exec
INSERT INTO users (discord_id, discord_user, birthday)
VALUES ($1, $2, $3)
ON CONFLICT (discord_id) DO UPDATE
    SET discord_user = EXCLUDED.discord_user,
        birthday = EXCLUDED.birthday;

-- name: GetBirthdayForUser :one
SELECT birthday
FROM users
WHERE discord_id = $1;

-- name: GetUsersWithBirthdayOn :many
SELECT discord_id
FROM users
WHERE EXTRACT(MONTH FROM birthday) = @month::int
  AND EXTRACT(DAY FROM birthday) = ANY(@days::int[]);

-- name: RemoveBirthdayForUser :execrows
UPDATE users
SET birthday = NULL
WHERE discord_id = $1
  AND birthday IS NOT NULL;
