-- Reverse 000006_integrity_constraints.

ALTER TABLE player_events
    DROP CONSTRAINT IF EXISTS chk_player_events_invite_status;

ALTER TABLE friendships
    DROP CONSTRAINT IF EXISTS chk_friendships_status,
    DROP CONSTRAINT IF EXISTS chk_friendships_not_self;

DROP INDEX IF EXISTS uq_friendships_pair;

CREATE UNIQUE INDEX IF NOT EXISTS uq_friendships_requester_addressee
    ON friendships (requester_id, addressee_id);

DROP INDEX IF EXISTS uq_player_events_player_event;

ALTER TABLE friendships
    DROP CONSTRAINT IF EXISTS fk_friendships_addressee,
    DROP CONSTRAINT IF EXISTS fk_friendships_requester;

ALTER TABLE events
    DROP CONSTRAINT IF EXISTS fk_events_host,
    DROP CONSTRAINT IF EXISTS fk_events_course;

ALTER TABLE player_events
    ALTER COLUMN invite_status DROP NOT NULL,
    ALTER COLUMN event_id DROP NOT NULL,
    ALTER COLUMN player_id DROP NOT NULL;

ALTER TABLE friendships
    ALTER COLUMN status DROP NOT NULL,
    ALTER COLUMN addressee_id DROP NOT NULL,
    ALTER COLUMN requester_id DROP NOT NULL;

ALTER TABLE events
    ALTER COLUMN host_id DROP NOT NULL,
    ALTER COLUMN course_id DROP NOT NULL;

-- Types stay BIGINT (safe widening); no downcast to INTEGER.
