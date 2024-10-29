-- migrations/000003_create_patient_lifestyle_table.up.sql
CREATE TABLE patient_lifestyle (
    patient_lifestyle_id SERIAL PRIMARY KEY,
    patient_id INT NOT NULL,
    lifestyle_factor VARCHAR(255) NOT NULL,
    value TEXT,
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (patient_id) REFERENCES patients(patient_id) ON DELETE CASCADE
);

