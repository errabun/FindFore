-- name: ListPlayers :many
SELECT id, name, phone, email, username
FROM players
ORDER BY id;

-- name: GetPlayerByID :one
SELECT id, name, phone, email, username
FROM players
WHERE id = $1;

-- name: GetPlayerByEmail :one
SELECT id, name, phone, email, username, password_digest
FROM players
WHERE email = $1;

-- name: CreatePlayer :one
INSERT INTO players (name, phone, email, username, password_digest, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, name, phone, email, username;
