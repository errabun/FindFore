-- Booking hardening: persisted idempotency, quoted price, cache freshness.
-- Additive only — does not redesign tee_time_providers / events.

ALTER TABLE tee_times
    ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ;

ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS provider_request_id UUID,
    ADD COLUMN IF NOT EXISTS quoted_price_cents INTEGER,
    ADD COLUMN IF NOT EXISTS quoted_currency VARCHAR;

-- Backfill any pre-hardening rows (none expected in prod yet; safe for empty tables).
UPDATE reservations
SET provider_request_id = gen_random_uuid()
WHERE provider_request_id IS NULL;

ALTER TABLE reservations
    ALTER COLUMN provider_request_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_reservations_provider_request
    ON reservations (provider, provider_request_id);
