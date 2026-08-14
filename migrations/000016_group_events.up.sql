-- Link social tee times to a golfer group. NULL = not a group round.
-- ON DELETE SET NULL keeps the host's event if the group is removed.
ALTER TABLE events
    ADD COLUMN group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL;

CREATE INDEX idx_events_group_id_starts
    ON events (group_id, planned_starts_at)
    WHERE group_id IS NOT NULL;
