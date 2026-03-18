-- name: ListCourses :many
SELECT id, name, street, city, state, zip_code, phone, cost
FROM courses
ORDER BY id;
