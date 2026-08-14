-- Scope posts to a golfer group. NULL group_id = community feed (unchanged).
ALTER TABLE posts
    ADD COLUMN group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE;

CREATE INDEX idx_posts_group_id_created
    ON posts (group_id, created_at DESC)
    WHERE group_id IS NOT NULL;
