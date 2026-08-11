-- name: ListCourses :many
SELECT id, name, street, city, state, zip_code, phone, cost, country, latitude, longitude, timezone
FROM courses
ORDER BY id;

-- name: GetCourseByNameAndCity :one
SELECT id, name, street, city, state, zip_code, phone, cost, country, latitude, longitude, timezone
FROM courses
WHERE name = $1 AND city = $2;

-- name: GetCourseByID :one
SELECT id, name, street, city, state, zip_code, phone, cost, country, latitude, longitude, timezone
FROM courses
WHERE id = $1;

-- name: CreateCourse :one
INSERT INTO courses (
    name, street, city, state, zip_code, phone, cost,
    country, latitude, longitude, timezone,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
RETURNING id, name, street, city, state, zip_code, phone, cost, country, latitude, longitude, timezone;
