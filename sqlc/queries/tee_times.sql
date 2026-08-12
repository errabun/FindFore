-- name: GetTeeTimeByID :one
SELECT id, course_id, starts_at, holes, status, capacity, available_slots, price_cents, currency,
       last_synced_at, created_at, updated_at
FROM tee_times
WHERE id = $1;

-- name: ListTeeTimesByCourseAndWindow :many
SELECT id, course_id, starts_at, holes, status, capacity, available_slots, price_cents, currency,
       last_synced_at, created_at, updated_at
FROM tee_times
WHERE course_id = $1
  AND starts_at >= $2
  AND starts_at < $3
ORDER BY starts_at, id;

-- name: InsertTeeTime :one
INSERT INTO tee_times (
    course_id, starts_at, holes, status, capacity, available_slots, price_cents, currency,
    last_synced_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
RETURNING id, course_id, starts_at, holes, status, capacity, available_slots, price_cents, currency,
          last_synced_at, created_at, updated_at;

-- name: UpdateTeeTimeCache :one
UPDATE tee_times
SET holes = $2,
    status = $3,
    capacity = $4,
    available_slots = $5,
    price_cents = $6,
    currency = $7,
    starts_at = $8,
    last_synced_at = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING id, course_id, starts_at, holes, status, capacity, available_slots, price_cents, currency,
          last_synced_at, created_at, updated_at;

-- name: UpdateTeeTimeStatus :one
UPDATE tee_times
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, course_id, starts_at, holes, status, capacity, available_slots, price_cents, currency,
          last_synced_at, created_at, updated_at;

-- name: GetTeeTimeByProviderExternalID :one
SELECT t.id, t.course_id, t.starts_at, t.holes, t.status, t.capacity, t.available_slots,
       t.price_cents, t.currency, t.last_synced_at, t.created_at, t.updated_at
FROM tee_time_providers tp
JOIN tee_times t ON t.id = tp.tee_time_id
WHERE tp.provider = $1 AND tp.external_id = $2;

-- name: GetTeeTimeProvider :one
SELECT id, tee_time_id, provider, external_id
FROM tee_time_providers
WHERE provider = $1 AND external_id = $2;

-- name: GetTeeTimeProviderByTeeTimeAndProvider :one
SELECT id, tee_time_id, provider, external_id
FROM tee_time_providers
WHERE tee_time_id = $1 AND provider = $2;

-- name: InsertTeeTimeProvider :one
INSERT INTO tee_time_providers (tee_time_id, provider, external_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, tee_time_id, provider, external_id;
