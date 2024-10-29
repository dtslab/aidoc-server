package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	db "github.com/stackvity/aidoc-server/internal/platform/repository/sqlc"
	"go.uber.org/zap"
)

// PatientRepositoryImpl struct
type PatientRepositoryImpl struct {
	q   *db.Queries
	log *zap.Logger // Add logger
}

// NewPatientRepository creates a new PatientRepositoryImpl
func NewPatientRepository(q *db.Queries, log *zap.Logger) *PatientRepositoryImpl { // Add logger to constructor
	return &PatientRepositoryImpl{q: q, log: log}
}

// CreatePatient creates a new patient in the database
func (r *PatientRepositoryImpl) CreatePatient(ctx context.Context, patient *domain.Patient) (*domain.Patient, error) {
	r.log.Info("CreatePatient repository started") // Start logging
	arg := db.CreatePatientParams{
		UserID:                 sql.NullInt32{Int32: int32(patient.UserID), Valid: true},
		FullName:               patient.FullName,
		Age:                    sql.NullInt32{Int32: int32(patient.Age), Valid: true},
		DateOfBirth:            patient.DateOfBirth,
		Sex:                    db.SexEnum(patient.Sex),
		PhoneNumber:            sql.NullString{String: patient.PhoneNumber, Valid: patient.PhoneNumber != ""},
		EmailAddress:           sql.NullString{String: patient.EmailAddress, Valid: patient.EmailAddress != ""},
		PreferredCommunication: db.NullPreferredCommunicationEnum{PreferredCommunicationEnum: db.PreferredCommunicationEnum(patient.PreferredCommunication), Valid: patient.PreferredCommunication != ""},
		SocioeconomicStatus:    db.NullSocioeconomicStatusEnum{SocioeconomicStatusEnum: db.SocioeconomicStatusEnum(patient.SocioeconomicStatus), Valid: patient.SocioeconomicStatus != ""},
		GeographicLocation:     sql.NullString{String: patient.GeographicLocation, Valid: patient.GeographicLocation != ""},
	}

	createdPatient, err := r.q.CreatePatient(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				return nil, fmt.Errorf("patient with that email already exists: %w", err)
			} else if pgErr.Code == "23503" { // foreign_key_violation
				return nil, domain.ErrPatientNotFound // Or a custom Foreign Key error
			}

		}
		r.log.Error("Create patient error", zap.Error(err)) // Log the error
		return nil, fmt.Errorf("failed to create patient: %w", err)
	}
	r.log.Info("CreatePatient repository completed successfully") // Log success
	return convertDbPatientToDomain(createdPatient), nil
}

// GetPatient retrieves a patient from the database
func (r *PatientRepositoryImpl) GetPatient(ctx context.Context, patientID int) (*domain.Patient, error) {
	r.log.Info("GetPatient repository started", zap.Int("patientID", patientID)) // Logging

	dbPatient, err := r.q.GetPatient(ctx, int32(patientID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPatientNotFound
		}
		r.log.Error("failed to get patient", zap.Error(err), zap.Int("patient_id", patientID)) // Log with patientID
		return nil, fmt.Errorf("failed to get patient: %w", err)                               // Wrap
	}

	r.log.Info("GetPatient repository completed successfully", zap.Int("patient_id", patientID)) // Log with patientID
	return convertDbPatientToDomain(dbPatient), nil
}

// UpdatePatient updates a patient in the database
func (r *PatientRepositoryImpl) UpdatePatient(ctx context.Context, patientID int, patient *domain.Patient) (*domain.Patient, error) {
	r.log.Info("UpdatePatient repository started", zap.Int("patientID", patientID)) // Log with patientID
	arg := db.UpdatePatientParams{
		PatientID:              int32(patientID),
		FullName:               patient.FullName,
		Age:                    sql.NullInt32{Int32: int32(patient.Age), Valid: true},
		DateOfBirth:            patient.DateOfBirth,
		Sex:                    db.SexEnum(patient.Sex),
		PhoneNumber:            sql.NullString{String: patient.PhoneNumber, Valid: patient.PhoneNumber != ""},
		EmailAddress:           sql.NullString{String: patient.EmailAddress, Valid: patient.EmailAddress != ""},
		PreferredCommunication: db.NullPreferredCommunicationEnum{PreferredCommunicationEnum: db.PreferredCommunicationEnum(patient.PreferredCommunication), Valid: patient.PreferredCommunication != ""},
		SocioeconomicStatus:    db.NullSocioeconomicStatusEnum{SocioeconomicStatusEnum: db.SocioeconomicStatusEnum(patient.SocioeconomicStatus), Valid: patient.SocioeconomicStatus != ""},
		GeographicLocation:     sql.NullString{String: patient.GeographicLocation, Valid: patient.GeographicLocation != ""},
	}
	updatedPatient, err := r.q.UpdatePatient(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				return nil, fmt.Errorf("email address already in use: %w", err) // More specific error message. Updated
			}
		}
		r.log.Error("failed update patient", zap.Error(err), zap.Int("patient_id", patientID)) // Log the error details. Updated
		return nil, fmt.Errorf("failed to update patient: %w", err)                            // Wrap
	}

	r.log.Info("UpdatePatient repository completed successfully") // Log success
	return convertDbPatientToDomain(updatedPatient), nil
}

// convertDbPatientToDomain converts a database patient to a domain patient
func convertDbPatientToDomain(dbPatient db.Patient) *domain.Patient {
	// ... (No changes in the conversion logic)
	return &domain.Patient{
		PatientID:              int(dbPatient.PatientID),
		UserID:                 int(dbPatient.UserID.Int32),
		FullName:               dbPatient.FullName,
		Age:                    int(dbPatient.Age.Int32),
		DateOfBirth:            dbPatient.DateOfBirth,
		Sex:                    string(dbPatient.Sex),
		PhoneNumber:            dbPatient.PhoneNumber.String,
		EmailAddress:           dbPatient.EmailAddress.String,
		PreferredCommunication: string(dbPatient.PreferredCommunication.PreferredCommunicationEnum),
		SocioeconomicStatus:    string(dbPatient.SocioeconomicStatus.SocioeconomicStatusEnum),
		GeographicLocation:     dbPatient.GeographicLocation.String,
		CreatedAt:              dbPatient.CreatedAt.Time,
		UpdatedAt:              dbPatient.UpdatedAt.Time,
	}
}
