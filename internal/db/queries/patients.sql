-- name: GetPatientByUserID :one
SELECT id, user_id, full_name, phone, created_at, updated_at
FROM patients
WHERE user_id = $1;
