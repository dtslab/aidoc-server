package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/core/ports"
	"go.uber.org/zap"
)

// MedicalHistoryService struct
type MedicalHistoryService struct {
	medicalHistoryRepo ports.MedicalHistoryRepository
	patientRepo        ports.PatientRepository
	log                *zap.Logger
	validate           *validator.Validate
	authorize          func(context.Context, int) bool
}

// NewMedicalHistoryService creates a new MedicalHistoryService. Injects dependencies, including authorize function.
func NewMedicalHistoryService(medicalHistoryRepo ports.MedicalHistoryRepository, patientRepo ports.PatientRepository, log *zap.Logger, validate *validator.Validate, authorize func(context.Context, int) bool) *MedicalHistoryService {
	return &MedicalHistoryService{
		medicalHistoryRepo: medicalHistoryRepo,
		patientRepo:        patientRepo,
		log:                log,
		validate:           validate,
		authorize:          authorize,
	}
}

func (s *MedicalHistoryService) CreateMedicalHistoryEntry(ctx context.Context, patientID int, req domain.CreateMedicalHistoryRequest) (*domain.MedicalHistoryEntry, error) {
	s.log.Info("CreateMedicalHistoryEntry service started", zap.Int("patientID", patientID), zap.Any("request", req))

	if err := s.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		s.log.Error("Input validation error", zap.Error(err), zap.Any("validationErrors", validationErrors))

		var errorDetails []string
		for _, err := range validationErrors {
			errorDetails = append(errorDetails, fmt.Sprintf("Field %s failed validation for tag %s", err.Field(), err.Tag()))
		}

		return nil, &domain.ValidationError{
			Code:    "INVALID_MEDICAL_HISTORY_DATA",
			Message: "Validation errors occurred",
			Details: errorDetails,
		}
	}

	_, err := s.patientRepo.GetPatient(ctx, patientID)
	if err != nil {
		if errors.Is(err, domain.ErrPatientNotFound) {
			return nil, domain.ErrPatientNotFound
		}
		s.log.Error("Failed to check patient existence", zap.Error(err))
		return nil, fmt.Errorf("failed to check patient existence: %w", err)
	}

	entry := &domain.MedicalHistoryEntry{
		PatientID:     patientID,
		Condition:     req.Condition,
		DiagnosisDate: req.DiagnosisDate,
		Status:        req.Status,
		Details:       req.Details,
	}

	createdEntry, err := s.medicalHistoryRepo.CreateMedicalHistoryEntry(ctx, entry)
	if err != nil {
		s.log.Error("Failed to create medical history entry in the repository", zap.Error(err))
		return nil, fmt.Errorf("create medical history entry error: %w", err)
	}

	s.log.Info("CreateMedicalHistoryEntry service completed successfully", zap.Int("createdEntryID", createdEntry.PatientMedicalHistoryID))
	return createdEntry, nil
}

func (s *MedicalHistoryService) GetMedicalHistoryEntries(ctx context.Context, patientID int) ([]*domain.MedicalHistoryEntry, error) {
	s.log.Info("GetMedicalHistoryEntries service started", zap.Int("patientID", patientID))

	_, err := s.patientRepo.GetPatient(ctx, patientID)
	if err != nil {
		if errors.Is(err, domain.ErrPatientNotFound) {
			return nil, domain.ErrPatientNotFound
		}
		s.log.Error("Failed to check patient existence", zap.Error(err))
		return nil, fmt.Errorf("failed to check patient existence: %w", err)

	}

	entries, err := s.medicalHistoryRepo.GetMedicalHistoryEntries(ctx, patientID)
	if err != nil {
		if errors.Is(err, domain.ErrMedicalHistoryEntryNotFound) {
			return nil, domain.ErrMedicalHistoryEntryNotFound
		}
		s.log.Error("Failed to get medical history entries from the repository", zap.Error(err))
		return nil, fmt.Errorf("get medical history entries error: %w", err)
	}

	s.log.Info("GetMedicalHistoryEntries service completed successfully")
	return entries, nil
}

func (s *MedicalHistoryService) GetMedicalHistoryEntry(ctx context.Context, entryID int) (*domain.MedicalHistoryEntry, error) {
	s.log.Info("GetMedicalHistoryEntry service started", zap.Int("entryID", entryID))

	entry, err := s.medicalHistoryRepo.GetMedicalHistoryEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrMedicalHistoryEntryNotFound) {
			return nil, domain.ErrMedicalHistoryEntryNotFound
		}
		s.log.Error("Failed to get medical history entry", zap.Error(err), zap.Int("entry_id", entryID))
		return nil, fmt.Errorf("get medical history entry error: %w", err)

	}

	s.log.Info("GetMedicalHistoryEntry service completed successfully", zap.Int("entryID", entryID))
	return entry, nil
}

func (s *MedicalHistoryService) UpdateMedicalHistoryEntry(ctx context.Context, entryID int, req domain.UpdateMedicalHistoryRequest) (*domain.MedicalHistoryEntry, error) {
	s.log.Info("UpdateMedicalHistoryEntry service started", zap.Int("entryID", entryID))

	if err := s.validate.Struct(req); err != nil { // Perform input validation
		validationErrors := err.(validator.ValidationErrors)
		s.log.Error("Input validation error", zap.Error(err), zap.Any("validationErrors", validationErrors)) // Log validation errors

		var errorDetails []string
		for _, err := range validationErrors {
			errorDetails = append(errorDetails, fmt.Sprintf("Field %s failed validation for tag %s", err.Field(), err.Tag()))
		}

		return nil, &domain.ValidationError{ // Return a validation error
			Code:    "INVALID_MEDICAL_HISTORY_DATA",
			Message: "Validation errors occurred",
			Details: errorDetails,
		}
	}

	existingEntry, err := s.medicalHistoryRepo.GetMedicalHistoryEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrMedicalHistoryEntryNotFound) {
			return nil, domain.ErrMedicalHistoryEntryNotFound // Return not found error if entry not exist
		}

		s.log.Error("Failed to retrieve existing medical history entry", zap.Error(err), zap.Int("entry_id", entryID))
		return nil, fmt.Errorf("failed to retrieve existing medical history entry: %w", err)
	}

	if !s.authorize(ctx, existingEntry.PatientID) {
		return nil, domain.ErrForbidden // Return forbidden if unauthorized
	}

	// Update only the fields provided in the request
	if req.Condition != "" {
		existingEntry.Condition = req.Condition
	}
	if !req.DiagnosisDate.IsZero() {
		existingEntry.DiagnosisDate = req.DiagnosisDate
	}
	if req.Status != "" {
		existingEntry.Status = req.Status
	}
	if req.Details != "" {
		existingEntry.Details = req.Details
	}

	updatedEntry, err := s.medicalHistoryRepo.UpdateMedicalHistoryEntry(ctx, entryID, existingEntry)
	if err != nil {
		s.log.Error("Failed to update medical history entry in the repository", zap.Error(err))
		return nil, fmt.Errorf("update medical history entry error: %w", err)
	}

	s.log.Info("UpdateMedicalHistoryEntry service completed successfully", zap.Int("updatedEntryID", updatedEntry.PatientMedicalHistoryID)) // Log successful update. Updated.

	return updatedEntry, nil // Return the updated entry. Updated
}

func (s *MedicalHistoryService) DeleteMedicalHistoryEntry(ctx context.Context, entryID int) error {
	s.log.Info("DeleteMedicalHistoryEntry service started", zap.Int("entryID", entryID))

	existingEntry, err := s.medicalHistoryRepo.GetMedicalHistoryEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrMedicalHistoryEntryNotFound) {
			return domain.ErrMedicalHistoryEntryNotFound
		}
		s.log.Error("Failed to retrieve medical history entry before deletion", zap.Error(err))
		return fmt.Errorf("failed to retrieve medical history entry before deletion: %w", err)
	}

	if !s.authorize(ctx, existingEntry.PatientID) {
		return domain.ErrForbidden
	}

	err = s.medicalHistoryRepo.DeleteMedicalHistoryEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrMedicalHistoryEntryNotFound) { // Improved error handling
			return domain.ErrMedicalHistoryEntryNotFound // Return not found if it doesn't exist when attempting to delete. Updated
		}

		s.log.Error("Failed to delete medical history entry in the repository", zap.Error(err))
		return fmt.Errorf("delete medical history entry error: %w", err)
	}

	s.log.Info("DeleteMedicalHistoryEntry service completed successfully")
	return nil
}
