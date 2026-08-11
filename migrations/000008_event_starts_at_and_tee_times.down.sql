-- Reverse 000008_event_starts_at_and_tee_times.

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS date VARCHAR,
    ADD COLUMN IF NOT EXISTS tee_time VARCHAR;

-- Best-effort restore wall-clock strings in America/Denver for down migration.
UPDATE events
SET date = to_char(starts_at AT TIME ZONE 'America/Denver', 'YYYY-MM-DD'),
    tee_time = to_char(starts_at AT TIME ZONE 'America/Denver', 'HH24:MI')
WHERE starts_at IS NOT NULL;

ALTER TABLE events
    DROP COLUMN IF EXISTS tee_time_id;

DROP INDEX IF EXISTS idx_events_starts_at;

ALTER TABLE events
    DROP COLUMN IF EXISTS starts_at;

DROP INDEX IF EXISTS idx_tee_times_course_starts_at;
DROP TABLE IF EXISTS tee_times;
