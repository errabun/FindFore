-- name: GetCourseByProviderExternalID :one
SELECT c.id, c.name, c.street, c.city, c.state, c.zip_code, c.phone, c.cost,
       c.country, c.latitude, c.longitude, c.timezone
FROM course_providers cp
JOIN courses c ON c.id = cp.course_id
WHERE cp.provider = $1 AND cp.external_id = $2;

-- name: GetCourseProvider :one
SELECT id, course_id, provider, external_id
FROM course_providers
WHERE provider = $1 AND external_id = $2;

-- name: InsertCourseProvider :one
INSERT INTO course_providers (course_id, provider, external_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, course_id, provider, external_id;
