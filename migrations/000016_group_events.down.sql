DROP INDEX IF EXISTS idx_events_group_id_starts;

ALTER TABLE events
    DROP COLUMN IF EXISTS group_id;
