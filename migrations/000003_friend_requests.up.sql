-- Promote one-way follows to mutual accepted friendships with request status.

ALTER TABLE friendships
    ADD COLUMN IF NOT EXISTS status INTEGER NOT NULL DEFAULT 1;

ALTER TABLE friendships
    RENAME COLUMN follower_id TO requester_id;

ALTER TABLE friendships
    RENAME COLUMN followee_id TO addressee_id;

-- If both A→B and B→A exist, keep the lower id and drop the duplicate.
DELETE FROM friendships a
USING friendships b
WHERE a.id > b.id
  AND a.status = b.status
  AND (
    (a.requester_id = b.addressee_id AND a.addressee_id = b.requester_id)
    OR (a.requester_id = b.requester_id AND a.addressee_id = b.addressee_id)
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_friendships_requester_addressee
    ON friendships (requester_id, addressee_id);

CREATE INDEX IF NOT EXISTS index_friendships_addressee_status
    ON friendships (addressee_id, status);

CREATE INDEX IF NOT EXISTS index_friendships_requester_status
    ON friendships (requester_id, status);
