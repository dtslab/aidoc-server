-- name: CreatePatient :one
INSERT INTO patients (user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING patient_id, user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location, created_at, updated_at;


-- name: GetPatient :one
SELECT patient_id, user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location, created_at, updated_at
FROM patients
WHERE patient_id = $1;


-- name: UpdatePatient :one
UPDATE patients
SET full_name = $2,
    age = $3,
    date_of_birth = $4,
    sex = $5,
    phone_number = $6,
    email_address = $7,
    preferred_communication = $8,
    socioeconomic_status = $9,
    geographic_location = $10,
    updated_at = NOW()
WHERE patient_id = $1
RETURNING patient_id, user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location, created_at, updated_at;

