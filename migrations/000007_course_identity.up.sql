-- Canonical course identity: geo/timezone/country + multi-provider external IDs.
-- Additive only — no silent deletes of course or event data.

ALTER TABLE courses
    ADD COLUMN IF NOT EXISTS country VARCHAR,
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS timezone VARCHAR;

UPDATE courses
SET country = 'US'
WHERE country IS NULL;

CREATE TABLE IF NOT EXISTS course_providers (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    provider VARCHAR NOT NULL,
    external_id VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_course_providers_provider_external UNIQUE (provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_course_providers_course_id
    ON course_providers (course_id);
