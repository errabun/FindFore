-- Event starts_at + tee_times skeleton.
-- Pre-launch clean cut: drop events.date / events.tee_time after backfill.
-- Reservations and tee_time_providers are intentionally deferred.

CREATE TABLE IF NOT EXISTS tee_times (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id),
    starts_at TIMESTAMPTZ NOT NULL,
    holes VARCHAR,
    status VARCHAR NOT NULL DEFAULT 'available',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_tee_times_status CHECK (status IN ('available', 'held', 'booked', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_tee_times_course_starts_at
    ON tee_times (course_id, starts_at);

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS starts_at TIMESTAMPTZ;

-- Backfill from ISO date + HH:MM in the course timezone (or America/Denver).
UPDATE events e
SET starts_at = (
    (e.date || ' ' || COALESCE(NULLIF(TRIM(e.tee_time), ''), '00:00'))::timestamp
    AT TIME ZONE COALESCE(NULLIF(TRIM(c.timezone), ''), 'America/Denver')
)
FROM courses c
WHERE c.id = e.course_id
  AND e.starts_at IS NULL
  AND e.date ~ '^\d{4}-\d{2}-\d{2}$'
  AND (
      e.tee_time IS NULL
      OR TRIM(e.tee_time) = ''
      OR e.tee_time ~ '^\d{1,2}:\d{2}'
  );

UPDATE events
SET starts_at = created_at AT TIME ZONE 'UTC'
WHERE starts_at IS NULL;

ALTER TABLE events
    ALTER COLUMN starts_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_events_starts_at ON events (starts_at);

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS tee_time_id BIGINT REFERENCES tee_times(id);

ALTER TABLE events
    DROP COLUMN IF EXISTS date,
    DROP COLUMN IF EXISTS tee_time;
