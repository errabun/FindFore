-- name: CreatePlayerEvent :one
INSERT INTO player_events (player_id, event_id, invite_status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, player_id, event_id, invite_status;

-- name: GetPlayerEvent :one
SELECT id, player_id, event_id, invite_status
FROM player_events
WHERE player_id = $1 AND event_id = $2;

-- name: UpdatePlayerEventStatus :one
UPDATE player_events
SET invite_status = $3, updated_at = NOW()
WHERE player_id = $1 AND event_id = $2
RETURNING id, player_id, event_id, invite_status;

-- name: ListPlayerIDsByEventAndStatus :many
SELECT player_id
FROM player_events
WHERE event_id = $1 AND invite_status = $2;

-- name: CountAcceptedForEvent :one
SELECT COUNT(*) FROM player_events
WHERE event_id = $1 AND invite_status = 1;

-- name: ClosePendingForEvent :exec
UPDATE player_events
SET invite_status = 3, updated_at = NOW()
WHERE event_id = $1 AND invite_status = 0;

-- name: ReopenClosedForEvent :exec
UPDATE player_events
SET invite_status = 0, updated_at = NOW()
WHERE event_id = $1 AND invite_status = 3;

-- name: ListAcceptedEventIDsByPlayerID :many
SELECT event_id
FROM player_events
WHERE player_id = $1 AND invite_status = 1;

-- name: ListPlayersExceptHost :many
SELECT id FROM players WHERE id != $1;
