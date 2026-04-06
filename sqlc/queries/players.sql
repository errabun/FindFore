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

-- name: GetPlayerByUsername :one
SELECT id, name, phone, email, username, password_digest
FROM players
WHERE username = $1;

-- name: CreatePlayer :one
INSERT INTO players (name, phone, email, username, password_digest, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, name, phone, email, username;

-- name: UpdatePlayer :one
UPDATE players
SET name = $2, phone = $3, email = $4, username = $5, updated_at = NOW()
WHERE id = $1
RETURNING id, name, phone, email, username;

-- name: UpdatePlayerPassword :exec
UPDATE players
SET password_digest = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetPlayerPasswordByID :one
SELECT password_digest
FROM players
WHERE id = $1;
