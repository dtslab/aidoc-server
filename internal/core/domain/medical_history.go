package domain

import (
	"time"
)

// MedicalHistoryEntry represents the medical history data model
type MedicalHistoryEntry struct {
	PatientMedicalHistoryID int       `db:"patient_medical_history_id" json:"patient_medical_history_id"`
	PatientID               int       `db:"patient_id" json:"patient_id" validate:"required"`
	Condition               string    `db:"condition" json:"condition" validate:"required"`
	DiagnosisDate           time.Time `db:"diagnosis_date" json:"diagnosis_date" validate:"omitempty,pastdate"` // optional, and must be in the past if provided
	Status                  string    `db:"status" json:"status" validate:"required,oneof=Active Inactive Resolved"`
	Details                 string    `db:"details" json:"details"`
	CreatedAt               time.Time `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time `db:"updated_at" json:"updated_at"`
}

type CreateMedicalHistoryRequest struct {
	Condition     string    `json:"condition" validate:"required"`
	DiagnosisDate time.Time `json:"diagnosis_date" validate:"omitempty,pastdate"`
	Status        string    `json:"status" validate:"required,oneof=Active Inactive Resolved"`
	Details       string    `json:"details"`
}

type UpdateMedicalHistoryRequest struct {
	Condition     string    `json:"condition"`
	DiagnosisDate time.Time `json:"diagnosis_date" validate:"omitempty,pastdate"`
	Status        string    `json:"status" validate:"omitempty,oneof=Active Inactive Resolved"`
	Details       string    `json:"details"`
}
