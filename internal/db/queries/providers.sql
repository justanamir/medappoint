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
