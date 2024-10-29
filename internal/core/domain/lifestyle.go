package domain

import (
	"time"
)

// LifestyleEntry represents an entry for a patient's lifestyle factor
type LifestyleEntry struct {
	PatientLifestyleID int       `db:"patient_lifestyle_id" json:"patient_lifestyle_id"`
	PatientID          int       `db:"patient_id" json:"patient_id"`
	LifestyleFactor    string    `db:"lifestyle_factor" json:"lifestyle_factor" validate:"required"`
	Value              string    `db:"value" json:"value"` // Can be a string, number, or other value depending on the factor
	StartDate          time.Time `db:"start_date" json:"start_date"`
	EndDate            time.Time `db:"end_date" json:"end_date"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

type CreateLifestyleRequest struct {
	LifestyleFactor string    `json:"lifestyle_factor" validate:"required"`
	Value           string    `json:"value"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}

type UpdateLifestyleRequest struct {
	LifestyleFactor string    `json:"lifestyle_factor"`
	Value           string    `json:"value"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}
