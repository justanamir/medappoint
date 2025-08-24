-- name: ListProviderAppointmentsOnDate :many
SELECT id, clinic_id, provider_id, patient_id, service_id, start_time, end_time, status
FROM appointments
WHERE provider_id = $1
  AND start_time >= $2
  AND start_time <  $3
  AND status IN ('scheduled','completed'); -- cancelled doesn't block

-- name: CreateAppointment :one
INSERT INTO appointments (clinic_id, provider_id, patient_id, service_id, start_time, end_time, status, notes)
VALUES ($1,$2,$3,$4,$5,$6,'scheduled',$7)
RETURNING id, clinic_id, provider_id, patient_id, service_id, start_time, end_time, status, notes, created_at, updated_at;

-- name: ListUpcomingAppointmentsByPatient :many
SELECT
  a.id, a.clinic_id, a.provider_id, a.patient_id, a.service_id,
  a.start_time, a.end_time, a.status, a.notes, a.created_at, a.updated_at,
  p.full_name   AS provider_name,
  s.name        AS service_name,
  c.name        AS clinic_name
FROM appointments a
JOIN providers p ON p.id = a.provider_id
JOIN services  s ON s.id = a.service_id
JOIN clinics   c ON c.id = a.clinic_id
WHERE a.patient_id = $1
  AND a.status IN ('scheduled','completed')
  AND a.start_time >= $2
ORDER BY a.start_time ASC
LIMIT $3 OFFSET $4;

-- name: GetAppointment :one
SELECT
  id, clinic_id, provider_id, patient_id, service_id,
  start_time, end_time, status, notes, created_at, updated_at
FROM appointments
WHERE id = $1;

-- name: CancelAppointment :one
UPDATE appointments
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1 AND status = 'scheduled'
RETURNING
  id, clinic_id, provider_id, patient_id, service_id,
  start_time, end_time, status, notes, created_at, updated_at;

-- name: ListAppointmentsByProviderOnDate :many
SELECT
  a.id, a.clinic_id, a.provider_id, a.patient_id, a.service_id,
  a.start_time, a.end_time, a.status, a.notes, a.created_at, a.updated_at,
  p.full_name  AS patient_name,
  s.name       AS service_name
FROM appointments a
LEFT JOIN patients p ON p.id = a.patient_id
JOIN services  s ON s.id = a.service_id
WHERE a.provider_id = $1
  AND a.start_time >= $2
  AND a.start_time <  $3
ORDER BY a.start_time ASC;

-- name: ListAllAppointmentsOnDate :many
SELECT
  a.id, a.clinic_id, a.provider_id, a.patient_id, a.service_id,
  a.start_time, a.end_time, a.status, a.notes, a.created_at, a.updated_at,
  pr.full_name AS provider_name,
  pa.full_name AS patient_name,
  s.name       AS service_name,
  c.name       AS clinic_name
FROM appointments a
JOIN providers pr ON pr.id = a.provider_id
LEFT JOIN patients pa ON pa.id = a.patient_id
JOIN services  s  ON s.id = a.service_id
JOIN clinics   c  ON c.id = a.clinic_id
WHERE a.start_time >= $1
  AND a.start_time <  $2
ORDER BY pr.id, a.start_time;
