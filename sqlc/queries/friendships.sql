-- name: FindFriendship :one
SELECT id, follower_id, followee_id
FROM friendships
WHERE follower_id = $1 AND followee_id = $2;

-- name: CreateFriendship :one
INSERT INTO friendships (follower_id, followee_id, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
RETURNING id, follower_id, followee_id;

-- name: DeleteFriendship :exec
DELETE FROM friendships
WHERE follower_id = $1 AND followee_id = $2;

-- name: ListFolloweeIDsByFollowerID :many
SELECT followee_id
FROM friendships
WHERE follower_id = $1;
