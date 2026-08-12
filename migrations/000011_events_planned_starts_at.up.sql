-- Option B: events store planned social time; linked tee_times.starts_at is play time.

ALTER TABLE events
    RENAME COLUMN starts_at TO planned_starts_at;

ALTER INDEX IF EXISTS idx_events_starts_at RENAME TO idx_events_planned_starts_at;
