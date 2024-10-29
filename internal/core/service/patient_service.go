package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/core/ports"
	"go.uber.org/zap"
)

// PatientService struct
type PatientService struct {
	patientRepo ports.PatientRepository
	log         *zap.Logger
	validate    *validator.Validate
}

// NewPatientService creates a new PatientService
func NewPatientService(patientRepo ports.PatientRepository, log *zap.Logger, validate *validator.Validate) *PatientService {
	return &PatientService{
		patientRepo: patientRepo,
		log:         log,
		validate:    validate,
	}
}

// CreatePatient creates a new patient
func (s *PatientService) CreatePatient(ctx context.Context, req domain.CreatePatientRequest) (*domain.Patient, error) {
	s.log.Info("CreatePatient service started")

	if err := s.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		s.log.Error("Input validation error", zap.Error(err), zap.Any("validationErrors", validationErrors))

		var errorDetails []string
		for _, err := range validationErrors {
			errorDetails = append(errorDetails, fmt.Sprintf("Field %s failed validation for tag %s", err.Field(), err.Tag()))
		}

		return nil, &domain.ValidationError{
			Code:    "INVALID_PATIENT_DATA",
			Message: "Validation errors occurred",
			Details: errorDetails,
		}
	}

	// Validate Age and DateOfBirth consistency
	if req.Age != 0 && !req.DateOfBirth.IsZero() {
		now := time.Now()
		expectedAge := now.Year() - req.DateOfBirth.Year()
		if now.YearDay() < req.DateOfBirth.YearDay() {
			expectedAge--
		}
		if expectedAge != req.Age {
			return nil, &domain.ValidationError{
				Code:    "INCONSISTENT_DATA",
				Message: "Age and DateOfBirth are inconsistent",
			}
		}
	}

	patient := &domain.Patient{
		UserID:                 req.UserID,
		FullName:               req.FullName,
		Age:                    req.Age,
		DateOfBirth:            req.DateOfBirth,
		Sex:                    req.Sex,
		PhoneNumber:            req.PhoneNumber,
		EmailAddress:           req.EmailAddress,
		PreferredCommunication: req.PreferredCommunication,
		SocioeconomicStatus:    req.SocioeconomicStatus,
		GeographicLocation:     req.GeographicLocation,
	}

	createdPatient, err := s.patientRepo.CreatePatient(ctx, patient)
	if err != nil {
		s.log.Error("Failed to create patient in the repository", zap.Error(err))
		return nil, fmt.Errorf("create patient error: %w", err)
	}

	s.log.Info("CreatePatient service completed successfully")
	return createdPatient, nil
}

// GetPatient retrieves a patient by ID
func (s *PatientService) GetPatient(ctx context.Context, patientID int) (*domain.Patient, error) {
	s.log.Info("GetPatient service started", zap.Int("patientID", patientID))

	patient, err := s.patientRepo.GetPatient(ctx, patientID)
	if err != nil {
		if errors.Is(err, domain.ErrPatientNotFound) {
			return nil, domain.ErrPatientNotFound
		}
		s.log.Error("failed to get patient", zap.Error(err), zap.Int("patient_id", patientID))
		return nil, fmt.Errorf("get patient error: %w", err)
	}

	s.log.Info("GetPatient service completed successfully")
	return patient, nil
}

// UpdatePatient updates an existing patient
func (s *PatientService) UpdatePatient(ctx context.Context, patientID int, req domain.UpdatePatientRequest) (*domain.Patient, error) {
	s.log.Info("UpdatePatient service started", zap.Int("patientID", patientID))

	if err := s.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		s.log.Error("Input validation error", zap.Error(err), zap.Any("validationErrors", validationErrors))

		var errorDetails []string
		for _, err := range validationErrors {
			errorDetails = append(errorDetails, fmt.Sprintf("Field %s failed validation for tag %s", err.Field(), err.Tag()))
		}
		return nil, &domain.ValidationError{
			Code:    "INVALID_PATIENT_DATA",
			Message: "Validation errors occurred",
			Details: errorDetails,
		}
	}

	// Validate Age and DateOfBirth consistency (similar to CreatePatient)
	if req.Age != 0 && !req.DateOfBirth.IsZero() {
		now := time.Now()
		expectedAge := now.Year() - req.DateOfBirth.Year()
		if now.YearDay() < req.DateOfBirth.YearDay() {
			expectedAge--
		}
		if expectedAge != req.Age {
			return nil, &domain.ValidationError{
				Code:    "INCONSISTENT_DATA",
				Message: "Age and DateOfBirth are inconsistent",
			}
		}
	}

	existingPatient, err := s.patientRepo.GetPatient(ctx, patientID) // Retrieve the existing patient. Updated
	if err != nil {
		return nil, fmt.Errorf("failed to get existing patient: %w", err)
	}

	// Update patient fields from the request. Updated
	if req.FullName != "" {
		existingPatient.FullName = req.FullName
	}
	if req.Age != 0 {
		existingPatient.Age = req.Age
	}
	if !req.DateOfBirth.IsZero() {
		existingPatient.DateOfBirth = req.DateOfBirth
	}
	if req.Sex != "" {
		existingPatient.Sex = req.Sex
	}
	if req.PhoneNumber != "" {
		existingPatient.PhoneNumber = req.PhoneNumber
	}
	if req.EmailAddress != "" {
		existingPatient.EmailAddress = req.EmailAddress
	}
	if req.PreferredCommunication != "" {
		existingPatient.PreferredCommunication = req.PreferredCommunication
	}
	if req.SocioeconomicStatus != "" {
		existingPatient.SocioeconomicStatus = req.SocioeconomicStatus
	}
	if req.GeographicLocation != "" {
		existingPatient.GeographicLocation = req.GeographicLocation
	}

	updatedPatient, err := s.patientRepo.UpdatePatient(ctx, patientID, existingPatient) // Pass the *domain.Patient
	if err != nil {
		s.log.Error("Failed to update patient in the repository", zap.Error(err))
		return nil, fmt.Errorf("update patient error: %w", err)
	}

	s.log.Info("UpdatePatient service completed successfully")
	return updatedPatient, nil
}
