-- 000002_create_patient_medical_history_table.up.sql
CREATE TABLE patient_medical_history (
    patient_medical_history_id SERIAL PRIMARY KEY,
    patient_id INT,
    condition VARCHAR(255) NOT NULL,
    diagnosis_date DATE,
    status VARCHAR(255) DEFAULT 'Active',
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE patient_medical_history
    ADD CONSTRAINT fk_patient_medical_history_patient
    FOREIGN KEY (patient_id)
    REFERENCES patients(patient_id)
    ON DELETE CASCADE;
