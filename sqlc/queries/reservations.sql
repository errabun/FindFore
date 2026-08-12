-- name: GetReservationByID :one
SELECT id, tee_time_id, booked_by_player_id, status, party_size, provider,
       external_reservation_id, hold_expires_at, failure_reason,
       provider_request_id, quoted_price_cents, quoted_currency,
       created_at, updated_at
FROM reservations
WHERE id = $1;

-- name: GetActiveReservationByTeeTimeID :one
SELECT id, tee_time_id, booked_by_player_id, status, party_size, provider,
       external_reservation_id, hold_expires_at, failure_reason,
       provider_request_id, quoted_price_cents, quoted_currency,
       created_at, updated_at
FROM reservations
WHERE tee_time_id = $1
  AND status IN ('pending', 'held', 'confirmed');

-- name: InsertReservation :one
INSERT INTO reservations (
    tee_time_id, booked_by_player_id, status, party_size, provider,
    external_reservation_id, hold_expires_at, failure_reason,
    provider_request_id, quoted_price_cents, quoted_currency,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
RETURNING id, tee_time_id, booked_by_player_id, status, party_size, provider,
          external_reservation_id, hold_expires_at, failure_reason,
          provider_request_id, quoted_price_cents, quoted_currency,
          created_at, updated_at;

-- name: UpdateReservation :one
UPDATE reservations
SET status = $2,
    external_reservation_id = $3,
    hold_expires_at = $4,
    failure_reason = $5,
    quoted_price_cents = $6,
    quoted_currency = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING id, tee_time_id, booked_by_player_id, status, party_size, provider,
          external_reservation_id, hold_expires_at, failure_reason,
          provider_request_id, quoted_price_cents, quoted_currency,
          created_at, updated_at;

-- name: ListReservationPlayers :many
SELECT id, reservation_id, player_id, guest_name, created_at, updated_at
FROM reservation_players
WHERE reservation_id = $1
ORDER BY id;

-- name: InsertReservationPlayer :one
INSERT INTO reservation_players (reservation_id, player_id, guest_name, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, reservation_id, player_id, guest_name, created_at, updated_at;
