-- name: FindReaction :one
SELECT id, post_id, player_id, emoji
FROM reactions
WHERE post_id = $1 AND player_id = $2 AND emoji = $3;

-- name: CreateReaction :one
INSERT INTO reactions (post_id, player_id, emoji, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, post_id, player_id, emoji;

-- name: DeleteReaction :exec
DELETE FROM reactions
WHERE post_id = $1 AND player_id = $2 AND emoji = $3;

-- name: ListReactionsByPostID :many
SELECT r.id, r.post_id, r.player_id, r.emoji, pl.name AS player_name
FROM reactions r
JOIN players pl ON pl.id = r.player_id
WHERE r.post_id = $1
ORDER BY r.id;
