-- Reverse 000009_tee_time_providers_and_cache.

DROP INDEX IF EXISTS idx_tee_time_providers_tee_time_id;
DROP TABLE IF EXISTS tee_time_providers;

ALTER TABLE tee_times
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS price_cents,
    DROP COLUMN IF EXISTS available_slots,
    DROP COLUMN IF EXISTS capacity;
