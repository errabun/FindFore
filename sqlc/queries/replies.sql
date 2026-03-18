-- name: CreateReply :one
INSERT INTO replies (post_id, player_id, body, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, post_id, player_id, body, created_at;

-- name: GetReplyByID :one
SELECT r.id, r.post_id, r.player_id, r.body, r.created_at, pl.name AS player_name
FROM replies r
JOIN players pl ON pl.id = r.player_id
WHERE r.id = $1;

-- name: ListRepliesByPostID :many
SELECT r.id, r.post_id, r.player_id, r.body, r.created_at, pl.name AS player_name
FROM replies r
JOIN players pl ON pl.id = r.player_id
WHERE r.post_id = $1
ORDER BY r.created_at ASC;

-- name: DeleteReply :exec
DELETE FROM replies WHERE id = $1 AND player_id = $2;
