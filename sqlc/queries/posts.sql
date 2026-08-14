-- name: CreatePost :one
INSERT INTO posts (player_id, body, group_id, created_at, updated_at)
VALUES ($1, $2, sqlc.narg('group_id'), NOW(), NOW())
RETURNING id, player_id, body, group_id, created_at;

-- name: GetPostByID :one
SELECT p.id, p.player_id, p.body, p.group_id, p.created_at, pl.name AS player_name
FROM posts p
JOIN players pl ON pl.id = p.player_id
WHERE p.id = $1;

-- name: ListPosts :many
SELECT p.id, p.player_id, p.body, p.group_id, p.created_at, pl.name AS player_name
FROM posts p
JOIN players pl ON pl.id = p.player_id
WHERE p.group_id IS NULL
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListGroupPosts :many
SELECT p.id, p.player_id, p.body, p.group_id, p.created_at, pl.name AS player_name
FROM posts p
JOIN players pl ON pl.id = p.player_id
WHERE p.group_id = sqlc.arg('group_id')
ORDER BY p.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1 AND player_id = $2;

-- name: DeletePostByID :exec
DELETE FROM posts WHERE id = $1;
