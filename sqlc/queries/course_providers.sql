-- name: GetCourseByProviderExternalID :one
SELECT c.id, c.name, c.street, c.city, c.state, c.zip_code, c.phone, c.cost,
       c.country, c.latitude, c.longitude, c.timezone
FROM course_providers cp
JOIN courses c ON c.id = cp.course_id
WHERE cp.provider = $1 AND cp.external_id = $2;

-- name: UpsertCourseProvider :one
INSERT INTO course_providers (course_id, provider, external_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (provider, external_id) DO UPDATE
SET course_id = EXCLUDED.course_id,
    updated_at = NOW()
RETURNING id, course_id, provider, external_id;
