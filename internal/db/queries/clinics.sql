-- name: ListClinics :many
SELECT id, name, timezone, address, created_at, updated_at
FROM clinics
ORDER BY id;
