-- Reverse 000011_events_planned_starts_at.

ALTER INDEX IF EXISTS idx_events_planned_starts_at RENAME TO idx_events_starts_at;

ALTER TABLE events
    RENAME COLUMN planned_starts_at TO starts_at;
