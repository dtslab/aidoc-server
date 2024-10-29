-- name: CreateLifestyleEntry :one
INSERT INTO patient_lifestyle (patient_id, lifestyle_factor, value, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLifestyleEntries :many
SELECT *
FROM patient_lifestyle
WHERE patient_id = $1;

-- name: GetLifestyleEntry :one
SELECT * 
FROM patient_lifestyle
WHERE patient_lifestyle_id = $1;

-- name: UpdateLifestyleEntry :one
UPDATE patient_lifestyle
SET lifestyle_factor = $2,
    value = $3,
    start_date = $4,
    end_date = $5
WHERE patient_lifestyle_id = $1
RETURNING *;

-- name: DeleteLifestyleEntry :exec
DELETE FROM patient_lifestyle
WHERE patient_lifestyle_id = $1;