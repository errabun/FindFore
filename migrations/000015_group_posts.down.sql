DROP INDEX IF EXISTS idx_posts_group_id_created;

ALTER TABLE posts
    DROP COLUMN IF EXISTS group_id;
