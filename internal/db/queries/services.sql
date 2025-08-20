-- name: ListServices :many
SELECT
  s.id,
  s.clinic_id,
  s.name,
  s.description,
  s.duration_min,
  c.name     AS clinic_name,
  c.timezone AS clinic_timezone
FROM services s
JOIN clinics  c ON c.id = s.clinic_id
ORDER BY s.id;
