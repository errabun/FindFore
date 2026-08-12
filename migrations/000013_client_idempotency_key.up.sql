-- HTTP client idempotency key for POST /reservations (separate from provider_request_id).

ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS client_idempotency_key VARCHAR;

CREATE UNIQUE INDEX IF NOT EXISTS uq_reservations_client_idempotency
    ON reservations (booked_by_player_id, client_idempotency_key)
    WHERE client_idempotency_key IS NOT NULL;
