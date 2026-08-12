-- name: ListAllEvents :many
SELECT e.id, e.course_id, e.open_spots, e.number_of_holes,
       e.private, e.host_id, e.planned_starts_at, e.tee_time_id,
       c.name AS course_name, c.timezone AS course_timezone, p.name AS host_name
FROM events e
JOIN courses c ON c.id = e.course_id
JOIN players p ON p.id = e.host_id
ORDER BY e.planned_starts_at, e.id;

-- name: ListPublicEvents :many
SELECT e.id, e.course_id, e.open_spots, e.number_of_holes,
       e.private, e.host_id, e.planned_starts_at, e.tee_time_id,
       c.name AS course_name, c.timezone AS course_timezone, p.name AS host_name
FROM events e
JOIN courses c ON c.id = e.course_id
JOIN players p ON p.id = e.host_id
WHERE e.private = false
ORDER BY e.planned_starts_at, e.id;

-- name: ListEventsByPlayerID :many
SELECT e.id, e.course_id, e.open_spots, e.number_of_holes,
       e.private, e.host_id, e.planned_starts_at, e.tee_time_id,
       c.name AS course_name, c.timezone AS course_timezone, p.name AS host_name
FROM events e
JOIN courses c ON c.id = e.course_id
JOIN players p ON p.id = e.host_id
JOIN player_events pe ON pe.event_id = e.id
WHERE pe.player_id = $1
ORDER BY e.planned_starts_at, e.id;

-- name: GetEventByID :one
SELECT e.id, e.course_id, e.open_spots, e.number_of_holes,
       e.private, e.host_id, e.planned_starts_at, e.tee_time_id,
       c.name AS course_name, c.timezone AS course_timezone, p.name AS host_name
FROM events e
JOIN courses c ON c.id = e.course_id
JOIN players p ON p.id = e.host_id
WHERE e.id = $1;

-- name: CreateEvent :one
INSERT INTO events (
    course_id, open_spots, number_of_holes, private, host_id, planned_starts_at, tee_time_id,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING id, course_id, open_spots, number_of_holes, private, host_id, planned_starts_at, tee_time_id;

-- name: UpdateEvent :exec
UPDATE events
SET course_id = $2, open_spots = $3, number_of_holes = $4, private = $5,
    planned_starts_at = $6, tee_time_id = $7, updated_at = NOW()
WHERE id = $1;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = $1;

-- name: DeletePastEvents :exec
DELETE FROM events WHERE planned_starts_at < NOW();

-- name: ListFriendsAvailableEventIDs :many
SELECT DISTINCT e.id
FROM events e
JOIN player_events pe ON pe.event_id = e.id AND pe.invite_status = 1
WHERE pe.player_id IN (
  SELECT CASE
    WHEN requester_id = $1 THEN addressee_id
    ELSE requester_id
  END
  FROM friendships
  WHERE status = 1
    AND (requester_id = $1 OR addressee_id = $1)
)
AND NOT EXISTS (
  SELECT 1 FROM player_events pe2
  WHERE pe2.event_id = e.id AND pe2.player_id = $2
)
AND e.open_spots > (
  SELECT COUNT(*) FROM player_events pe3
  WHERE pe3.event_id = e.id AND pe3.invite_status = 1
);
