-- Reverse 000007_course_identity.

DROP INDEX IF EXISTS idx_course_providers_course_id;
DROP TABLE IF EXISTS course_providers;

ALTER TABLE courses
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS country;
