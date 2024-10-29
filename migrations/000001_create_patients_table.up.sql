-- 000001_create_patients_table.up.sql
CREATE TYPE sex_enum AS ENUM ('Male', 'Female', 'Other');
CREATE TYPE preferred_communication_enum AS ENUM ('Phone', 'Email', 'Text');
CREATE TYPE socioeconomic_status_enum AS ENUM ('Low', 'Middle', 'High', 'Decline to Answer');

CREATE TABLE patients (
    patient_id SERIAL PRIMARY KEY,
    user_id INT,
    full_name VARCHAR(255) NOT NULL,
    age INT,
    date_of_birth DATE NOT NULL,
    sex sex_enum NOT NULL,
    phone_number VARCHAR(20),
    email_address VARCHAR(255) UNIQUE,
    preferred_communication preferred_communication_enum DEFAULT 'Email',
    socioeconomic_status socioeconomic_status_enum DEFAULT 'Decline to Answer',
    geographic_location VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE patients
    ADD CONSTRAINT fk_patient_user
    FOREIGN KEY (user_id)
    REFERENCES Users(user_id)
    ON DELETE CASCADE;
