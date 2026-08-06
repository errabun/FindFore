DROP INDEX IF EXISTS index_friendships_requester_status;
DROP INDEX IF EXISTS index_friendships_addressee_status;
DROP INDEX IF EXISTS uq_friendships_requester_addressee;

ALTER TABLE friendships
    RENAME COLUMN addressee_id TO followee_id;

ALTER TABLE friendships
    RENAME COLUMN requester_id TO follower_id;

ALTER TABLE friendships
    DROP COLUMN IF EXISTS status;
