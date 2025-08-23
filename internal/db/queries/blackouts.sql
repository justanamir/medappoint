-- name: IsDateBlackoutedForProvider :one
SELECT EXISTS (
  SELECT 1 FROM blackouts
  WHERE (provider_id = $1 OR (clinic_id = $2 AND provider_id IS NULL))
    AND date = $3
) AS blocked;
