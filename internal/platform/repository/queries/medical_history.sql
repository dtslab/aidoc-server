-- name: CreateMedicalHistoryEntry :one
INSERT INTO patient_medical_history (patient_id, condition, diagnosis_date, status, details)
VALUES ($1, $2, $3, $4, $5)
RETURNING patient_medical_history_id, patient_id, condition, diagnosis_date, status, details, created_at, updated_at;

-- name: GetMedicalHistoryEntries :many
SELECT patient_medical_history_id, patient_id, condition, diagnosis_date, status, details, created_at, updated_at
FROM patient_medical_history
WHERE patient_id = $1;

-- name: GetMedicalHistoryEntry :one
SELECT patient_medical_history_id, patient_id, condition, diagnosis_date, status, details, created_at, updated_at
FROM patient_medical_history
WHERE patient_medical_history_id = $1;

-- name: UpdateMedicalHistoryEntry :one
UPDATE patient_medical_history
SET condition = $2,
    diagnosis_date = $3,
    status = $4,
    details = $5,
    updated_at = NOW()
WHERE patient_medical_history_id = $1
RETURNING patient_medical_history_id, patient_id, condition, diagnosis_date, status, details, created_at, updated_at;

-- name: DeleteMedicalHistoryEntry :exec
DELETE FROM patient_medical_history
WHERE patient_medical_history_id = $1;