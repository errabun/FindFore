-- name: CreatePost :one
INSERT INTO posts (player_id, body, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
RETURNING id, player_id, body, created_at;

-- name: GetPostByID :one
SELECT p.id, p.player_id, p.body, p.created_at, pl.name AS player_name
FROM posts p
JOIN players pl ON pl.id = p.player_id
WHERE p.id = $1;

-- name: ListPosts :many
SELECT p.id, p.player_id, p.body, p.created_at, pl.name AS player_name
FROM posts p
JOIN players pl ON pl.id = p.player_id
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1 AND player_id = $2;
