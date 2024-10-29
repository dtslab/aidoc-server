package domain

import (
	"time"
)

type Patient struct {
	PatientID              int       `db:"patient_id" json:"patient_id"`
	UserID                 int       `db:"user_id" json:"user_id" validate:"required"`
	FullName               string    `db:"full_name" json:"full_name" validate:"required"`
	Age                    int       `db:"age" json:"age" validate:"omitempty,minage"`
	DateOfBirth            time.Time `db:"date_of_birth" json:"date_of_birth" validate:"required,pastdate,dateformat"`
	Sex                    string    `db:"sex" json:"sex" validate:"required,oneof=Male Female Other"`
	PhoneNumber            string    `db:"phone_number" json:"phone_number" validate:"omitempty,phoneNumber"`
	EmailAddress           string    `db:"email_address" json:"email_address" validate:"omitempty,email"`
	PreferredCommunication string    `db:"preferred_communication" json:"preferred_communication" validate:"omitempty,oneof=Phone Email Text"`
	SocioeconomicStatus    string    `db:"socioeconomic_status" json:"socioeconomic_status" validate:"omitempty,oneof=Low Middle High Decline to Answer"`
	GeographicLocation     string    `db:"geographic_location" json:"geographic_location"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time `db:"updated_at" json:"updated_at"`
}

type CreatePatientRequest struct {
	UserID                 int       `json:"user_id" validate:"required"`
	FullName               string    `json:"full_name" validate:"required"`
	Age                    int       `json:"age" validate:"omitempty,minage"`
	DateOfBirth            time.Time `json:"date_of_birth" validate:"required,pastdate,dateformat"`
	Sex                    string    `json:"sex" validate:"required,oneof=Male Female Other"`
	PhoneNumber            string    `json:"phone_number" validate:"omitempty,phoneNumber"`
	EmailAddress           string    `json:"email_address" validate:"required,email"`
	PreferredCommunication string    `json:"preferred_communication" validate:"omitempty,oneof=Phone Email Text"`
	SocioeconomicStatus    string    `json:"socioeconomic_status" validate:"omitempty,oneof=Low Middle High Decline to Answer"`
	GeographicLocation     string    `json:"geographic_location"`
}

type UpdatePatientRequest struct {
	FullName               string    `json:"full_name"`
	Age                    int       `json:"age" validate:"omitempty,minage"`
	DateOfBirth            time.Time `json:"date_of_birth" validate:"omitempty,pastdate,dateformat"`
	Sex                    string    `json:"sex" validate:"omitempty,oneof=Male Female Other"`
	PhoneNumber            string    `json:"phone_number" validate:"omitempty,phoneNumber"`
	EmailAddress           string    `json:"email_address" validate:"omitempty,email"`
	PreferredCommunication string    `json:"preferred_communication" validate:"omitempty,oneof=Phone Email Text"`
	SocioeconomicStatus    string    `json:"socioeconomic_status" validate:"omitempty,oneof=Low Middle High Decline to Answer"`
	GeographicLocation     string    `json:"geographic_location"`
}
