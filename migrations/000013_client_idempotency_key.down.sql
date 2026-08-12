-- Reverse 000013_client_idempotency_key.

DROP INDEX IF EXISTS uq_reservations_client_idempotency;

ALTER TABLE reservations
    DROP COLUMN IF EXISTS client_idempotency_key;
