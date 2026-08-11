-- Pre-launch integrity: FKs, membership uniqueness, friendship pair rules, status checks.
-- open_spots remains the column name but means capacity (max accepted players); remaining is derived.

-- ---------------------------------------------------------------------------
-- Clean orphans before adding FKs
-- ---------------------------------------------------------------------------

DELETE FROM player_events pe
WHERE pe.event_id IN (
  SELECT e.id FROM events e
  WHERE e.course_id IS NULL
     OR e.host_id IS NULL
     OR NOT EXISTS (SELECT 1 FROM courses c WHERE c.id = e.course_id)
     OR NOT EXISTS (SELECT 1 FROM players p WHERE p.id = e.host_id)
);

DELETE FROM events e
WHERE e.course_id IS NULL
   OR e.host_id IS NULL
   OR NOT EXISTS (SELECT 1 FROM courses c WHERE c.id = e.course_id)
   OR NOT EXISTS (SELECT 1 FROM players p WHERE p.id = e.host_id);

DELETE FROM friendships f
WHERE f.requester_id IS NULL
   OR f.addressee_id IS NULL
   OR f.requester_id = f.addressee_id
   OR NOT EXISTS (SELECT 1 FROM players p WHERE p.id = f.requester_id)
   OR NOT EXISTS (SELECT 1 FROM players p WHERE p.id = f.addressee_id);

-- Keep one row per unordered friendship pair (prefer lower id).
DELETE FROM friendships a
USING friendships b
WHERE a.id > b.id
  AND LEAST(a.requester_id, a.addressee_id) = LEAST(b.requester_id, b.addressee_id)
  AND GREATEST(a.requester_id, a.addressee_id) = GREATEST(b.requester_id, b.addressee_id);

-- Keep one player_event per (player, event).
DELETE FROM player_events a
USING player_events b
WHERE a.id > b.id
  AND a.player_id = b.player_id
  AND a.event_id = b.event_id;

-- ---------------------------------------------------------------------------
-- Align FK column types with referenced BIGSERIAL ids
-- ---------------------------------------------------------------------------

ALTER TABLE events
    ALTER COLUMN course_id TYPE BIGINT USING course_id::BIGINT,
    ALTER COLUMN host_id TYPE BIGINT USING host_id::BIGINT;

ALTER TABLE friendships
    ALTER COLUMN requester_id TYPE BIGINT USING requester_id::BIGINT,
    ALTER COLUMN addressee_id TYPE BIGINT USING addressee_id::BIGINT;

-- ---------------------------------------------------------------------------
-- NOT NULL + foreign keys
-- ---------------------------------------------------------------------------

ALTER TABLE events
    ALTER COLUMN course_id SET NOT NULL,
    ALTER COLUMN host_id SET NOT NULL;

ALTER TABLE events
    ADD CONSTRAINT fk_events_course FOREIGN KEY (course_id) REFERENCES courses(id),
    ADD CONSTRAINT fk_events_host FOREIGN KEY (host_id) REFERENCES players(id);

ALTER TABLE friendships
    ALTER COLUMN requester_id SET NOT NULL,
    ALTER COLUMN addressee_id SET NOT NULL,
    ALTER COLUMN status SET NOT NULL;

ALTER TABLE friendships
    ADD CONSTRAINT fk_friendships_requester FOREIGN KEY (requester_id) REFERENCES players(id),
    ADD CONSTRAINT fk_friendships_addressee FOREIGN KEY (addressee_id) REFERENCES players(id);

ALTER TABLE player_events
    ALTER COLUMN player_id SET NOT NULL,
    ALTER COLUMN event_id SET NOT NULL,
    ALTER COLUMN invite_status SET NOT NULL;

-- ---------------------------------------------------------------------------
-- Uniqueness & CHECKs
-- ---------------------------------------------------------------------------

CREATE UNIQUE INDEX IF NOT EXISTS uq_player_events_player_event
    ON player_events (player_id, event_id);

DROP INDEX IF EXISTS uq_friendships_requester_addressee;

CREATE UNIQUE INDEX IF NOT EXISTS uq_friendships_pair
    ON friendships (LEAST(requester_id, addressee_id), GREATEST(requester_id, addressee_id));

ALTER TABLE friendships
    ADD CONSTRAINT chk_friendships_not_self CHECK (requester_id <> addressee_id),
    ADD CONSTRAINT chk_friendships_status CHECK (status IN (0, 1));

ALTER TABLE player_events
    ADD CONSTRAINT chk_player_events_invite_status CHECK (invite_status IN (0, 1, 2, 3));
