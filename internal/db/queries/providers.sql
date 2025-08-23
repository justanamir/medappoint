-- name: ListProviders :many
SELECT
  p.id,
  p.full_name,
  p.speciality,
  p.clinic_id,
  c.name       AS clinic_name,
  c.timezone   AS clinic_timezone
FROM providers p
JOIN clinics   c ON c.id = p.clinic_id
ORDER BY p.id;

-- name: GetProvider :one
SELECT id, user_id, full_name, speciality, clinic_id, created_at, updated_at
FROM providers
WHERE id = $1;

-- name: GetProviderWeekdayAvailability :many
SELECT id, provider_id, weekday, start_hhmm, end_hhmm
FROM availabilities
WHERE provider_id = $1 AND weekday = $2
ORDER BY start_hhmm;
