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
