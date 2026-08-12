-- Tee time provider identities + cached inventory fields on tee_times.

ALTER TABLE tee_times
    ADD COLUMN IF NOT EXISTS capacity INTEGER,
    ADD COLUMN IF NOT EXISTS available_slots INTEGER,
    ADD COLUMN IF NOT EXISTS price_cents INTEGER,
    ADD COLUMN IF NOT EXISTS currency VARCHAR;

CREATE TABLE IF NOT EXISTS tee_time_providers (
    id BIGSERIAL PRIMARY KEY,
    tee_time_id BIGINT NOT NULL REFERENCES tee_times(id) ON DELETE CASCADE,
    provider VARCHAR NOT NULL,
    external_id VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tee_time_providers_provider_external UNIQUE (provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_tee_time_providers_tee_time_id
    ON tee_time_providers (tee_time_id);
