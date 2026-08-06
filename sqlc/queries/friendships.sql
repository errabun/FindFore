-- name: FindFriendship :one
SELECT id, requester_id, addressee_id, status
FROM friendships
WHERE requester_id = $1 AND addressee_id = $2;

-- name: FindFriendshipBetween :one
SELECT id, requester_id, addressee_id, status
FROM friendships
WHERE (requester_id = $1 AND addressee_id = $2)
   OR (requester_id = $2 AND addressee_id = $1)
LIMIT 1;

-- name: GetFriendshipByID :one
SELECT id, requester_id, addressee_id, status
FROM friendships
WHERE id = $1;

-- name: CreateFriendship :one
INSERT INTO friendships (requester_id, addressee_id, status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, requester_id, addressee_id, status;

-- name: UpdateFriendshipStatus :one
UPDATE friendships
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, requester_id, addressee_id, status;

-- name: DeleteFriendshipByID :exec
DELETE FROM friendships
WHERE id = $1;

-- name: ListAcceptedFriendIDs :many
SELECT CASE
    WHEN requester_id = $1 THEN addressee_id
    ELSE requester_id
END::integer AS friend_id
FROM friendships
WHERE status = 1
  AND (requester_id = $1 OR addressee_id = $1);

-- name: ListIncomingPendingFriendships :many
SELECT id, requester_id, addressee_id, status
FROM friendships
WHERE addressee_id = $1 AND status = 0
ORDER BY id DESC;

-- name: ListOutgoingPendingFriendships :many
SELECT id, requester_id, addressee_id, status
FROM friendships
WHERE requester_id = $1 AND status = 0
ORDER BY id DESC;
