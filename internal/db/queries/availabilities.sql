-- name: ListAvailabilitiesByProvider :many
SELECT
  a.id,
  a.provider_id,
  a.weekday,
  a.start_hhmm,
  a.end_hhmm,
  p.full_name AS provider_name
FROM availabilities a
JOIN providers p ON p.id = a.provider_id
WHERE a.provider_id = $1
ORDER BY a.weekday;
