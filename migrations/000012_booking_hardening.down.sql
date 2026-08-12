-- Reverse 000012_booking_hardening.

DROP INDEX IF EXISTS uq_reservations_provider_request;

ALTER TABLE reservations
    DROP COLUMN IF EXISTS quoted_currency,
    DROP COLUMN IF EXISTS quoted_price_cents,
    DROP COLUMN IF EXISTS provider_request_id;

ALTER TABLE tee_times
    DROP COLUMN IF EXISTS last_synced_at;
